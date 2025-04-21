package cfg_mod

import (
	"fmt"

	"github.com/rskv-p/mini/mod"
	"github.com/rskv-p/mini/mod/m_act/act_core"
	"github.com/rskv-p/mini/mod/m_act/act_type"
	"github.com/rskv-p/mini/mod/m_cfg/cfg_core"
	"github.com/rskv-p/mini/mod/m_cfg/cfg_type"
	"github.com/rskv-p/mini/mod/m_log/log_type"
	"github.com/rskv-p/mini/typ"
)

//---------------------
// ConfigModule
//---------------------

// ConfigModule представляет модуль конфигурации.
type ConfigModule struct {
	Module    typ.IModule            // Модуль, реализующий интерфейс IModule
	Client    cfg_type.IConfigClient // Клиент для работы с конфигурацией
	logClient log_type.ILogClient    // Client for logging
}

// Убедимся, что ConfigModule реализует интерфейс IModule.
var _ typ.IModule = (*ConfigModule)(nil)

//---------------------
// Жизненный цикл модуля
//---------------------

// Name возвращает имя модуля.
func (m *ConfigModule) GetName() string {
	return m.Module.GetName()
}

// Start запускает модуль (можно добавить специфическую логику остановки).
func (s *ConfigModule) Start() error {
	// Здесь можно реализовать логику старта, если нужно
	return nil
}

// Init инициализирует модуль (можно добавить логику инициализации).
func (s *ConfigModule) Init() error {
	// Здесь можно реализовать логику инициализации, если нужно
	return nil
}

// Stop останавливает модуль (полезно для очистки ресурсов).
func (m *ConfigModule) Stop() error {
	m.logClient.Info("Stopping Config module", map[string]interface{}{"module": m.GetName()})
	return nil
}

//---------------------
// Действия
//---------------------

// Actions возвращает список действий, которые модуль может выполнять.
func (m *ConfigModule) GetActions() []act_type.ActionDef {
	return []act_type.ActionDef{
		// Получение конфигурации
		{Name: "config.get", Func: HandleGet, Public: true},
		// Установка конфигурации
		{Name: "config.set", Func: HandleSet},
		// Удаление конфигурации
		{Name: "config.del", Func: HandleDel},
		// Получение всех конфигураций
		{Name: "config.all", Func: HandleAll, Public: true},
		// Перезагрузка конфигурации
		{Name: "config.reload", Func: HandleReload},
		// Публикация конфигурации через клиента
		{
			Name: "config.publish",
			Func: func(a act_type.IAction) any {
				action, err := castAction(a)
				if err != nil {
					m.logClient.Error("Invalid action type", map[string]interface{}{"error": err})
					return err
				}

				// Извлечение ключа и значения из действия
				key := action.InputString(0)
				val := action.Inputs[1]

				// Публикуем конфигурацию через клиента
				err = m.Client.PublishConfig(key, val)
				if err != nil {
					m.logClient.Error(fmt.Sprintf("Failed to publish config '%s'", key), map[string]interface{}{"key": key, "error": err})
					return err
				}

				m.logClient.Info("Config published successfully", map[string]interface{}{"key": key, "value": val})
				return true
			},
			Public: true,
		},
	}
}

//---------------------
// Создание модуля
//---------------------

// NewConfigModule создает новый экземпляр модуля ConfigModule и инициализирует его.
func NewConfigModule(service typ.IService, client cfg_type.IConfigClient, logClient log_type.ILogClient) *ConfigModule {
	// Создаем новый модуль с помощью NewModule
	module := mod.NewModule("config", service, nil, nil, nil)

	// Возвращаем ConfigModule с созданным модулем
	configModule := &ConfigModule{
		Module:    module,
		Client:    client,
		logClient: logClient,
	}

	// Регистрируем действия для модуля
	for _, action := range configModule.GetActions() {
		act_core.Register(action.Name, action.Func)
	}

	return configModule
}

//---------------------
// Вспомогательные функции
//---------------------

// castAction безопасно кастует IAction в *act_core.Action и возвращает ошибку, если кастинг не удался.
func castAction(a act_type.IAction) (*act_core.Action, error) {
	action, ok := a.(*act_core.Action)
	if !ok {
		return nil, fmt.Errorf("Invalid action type: expected *act_core.Action")
	}
	return action, nil
}

//---------------------
// Обработчики действий
//---------------------

// HandleGet получает значение конфигурации по заданному ключу.
func HandleGet(a act_type.IAction) any {
	action, err := castAction(a)
	if err != nil {
		return err
	}
	moduleName := action.InputString(0) // Получаем имя модуля
	key := action.InputString(1)        // Получаем ключ
	return cfg_core.Get(moduleName, key)
}

// HandleSet устанавливает пару ключ-значение в конфигурации.
func HandleSet(a act_type.IAction) any {
	action, err := castAction(a)
	if err != nil {
		return err
	}
	moduleName := action.InputString(0) // Получаем имя модуля
	key := action.InputString(1)        // Получаем ключ
	val := action.Inputs[2]             // Получаем значение
	cfg_core.Set(moduleName, key, val)
	return true
}

// HandleDel удаляет конфигурацию для заданного ключа.
func HandleDel(a act_type.IAction) any {
	action, err := castAction(a)
	if err != nil {
		return err
	}
	moduleName := action.InputString(0) // Получаем имя модуля
	key := action.InputString(1)        // Получаем ключ
	cfg_core.Delete(moduleName, key)
	return true
}

// HandleAll возвращает все конфигурации.
func HandleAll(a act_type.IAction) any {
	return cfg_core.All()
}

// HandleReload перезагружает конфигурацию.
func HandleReload(a act_type.IAction) any {
	return cfg_core.Reload("") == nil
}
