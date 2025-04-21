package cfg_type

// IConfigClient defines an interface for accessing and manipulating configurations.
type IConfigClient interface {
	GetConfig(key string) any
	SetConfig(key string, value any) error
	DeleteConfig(key string) error
	PublishConfig(key string, value any) error

	LoadConfig(file string) error // Загружать конфигурацию для модуля
	ReloadConfig() error          // Перезагружать конфигурацию

}
