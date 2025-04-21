package m_cfg

import (
	"fmt"

	"github.com/rskv-p/mini/act"
	"github.com/rskv-p/mini/mod"
	"github.com/rskv-p/mini/pkg/x_cfg"
	"github.com/rskv-p/mini/typ"
)

//---------------------
// ConfigModule
//---------------------

// ConfigModule представляет модуль конфигурации.
type ConfigModule struct {
	Module    typ.IModule       // Модуль, реализующий интерфейс IModule
	Client    typ.IConfigClient // Клиент для работы с конфигурацией
	logClient typ.ILogClient    // Client for logging
}

// Убедимся, что ConfigModule реализует интерфейс IModule.
var _ typ.IModule = (*ConfigModule)(nil)

//---------------------
// Жизненный цикл модуля
//---------------------

// Name возвращает имя модуля.
func (m *ConfigModule) Name() string {
	return m.Module.Name()
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
	m.logClient.Info("Stopping Config module", map[string]interface{}{"module": m.Name()})
	return nil
}

//---------------------
// Действия
//---------------------

// Actions возвращает список действий, которые модуль может выполнять.
func (m *ConfigModule) Actions() []typ.ActionDef {
	return []typ.ActionDef{
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
			Func: func(a typ.IAction) any {
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

				// Публикуем событие об обновлении конфигурации
				event := typ.Event{
					Name:    "config.updated",
					Payload: map[string]any{"key": key, "value": val},
				}
				err = m.Client.PublishConfig("events", event)
				if err != nil {
					m.logClient.Error("Failed to publish config update event", map[string]interface{}{"error": err})
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
func NewConfigModule(service typ.IService, client typ.IConfigClient, logClient typ.ILogClient) *ConfigModule {
	// Создаем новый модуль с помощью NewModule
	module := mod.NewModule("config", service, nil, nil, nil)

	// Возвращаем ConfigModule с созданным модулем
	configModule := &ConfigModule{
		Module:    module,
		Client:    client,
		logClient: logClient,
	}

	// Регистрируем действия для модуля
	for _, action := range configModule.Actions() {
		act.Register(action.Name, action.Func)
	}

	return configModule
}

//---------------------
// Вспомогательные функции
//---------------------

// castAction безопасно кастует IAction в *act.Action и возвращает ошибку, если кастинг не удался.
func castAction(a typ.IAction) (*act.Action, error) {
	action, ok := a.(*act.Action)
	if !ok {
		return nil, fmt.Errorf("Invalid action type: expected *act.Action")
	}
	return action, nil
}

//---------------------
// Обработчики действий
//---------------------

// HandleGet получает значение конфигурации по заданному ключу.
func HandleGet(a typ.IAction) any {
	action, err := castAction(a)
	if err != nil {
		return err
	}
	moduleName := action.InputString(0) // Получаем имя модуля
	key := action.InputString(1)        // Получаем ключ
	return x_cfg.Get(moduleName, key)
}

// HandleSet устанавливает пару ключ-значение в конфигурации.
func HandleSet(a typ.IAction) any {
	action, err := castAction(a)
	if err != nil {
		return err
	}
	moduleName := action.InputString(0) // Получаем имя модуля
	key := action.InputString(1)        // Получаем ключ
	val := action.Inputs[2]             // Получаем значение
	x_cfg.Set(moduleName, key, val)
	return true
}

// HandleDel удаляет конфигурацию для заданного ключа.
func HandleDel(a typ.IAction) any {
	action, err := castAction(a)
	if err != nil {
		return err
	}
	moduleName := action.InputString(0) // Получаем имя модуля
	key := action.InputString(1)        // Получаем ключ
	x_cfg.Delete(moduleName, key)
	return true
}

// HandleAll возвращает все конфигурации.
func HandleAll(a typ.IAction) any {
	return x_cfg.All()
}

// HandleReload перезагружает конфигурацию.
func HandleReload(a typ.IAction) any {
	return x_cfg.Reload("") == nil
}
