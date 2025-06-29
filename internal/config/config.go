// internal/config/config.go
package config

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Platform PlatformConfig `yaml:"platform"`
	Log      LogConfig      `yaml:"log"`
}

type ServerConfig struct {
	Port             int `yaml:"port"`
	HTTPPort         int `yaml:"http_port"`
	MaxConnections   int `yaml:"maxConnections"`
	HeartbeatTimeout int `yaml:"heartbeatTimeout"`
}

type PlatformConfig struct {
	URL               string `yaml:"url"`           // 平台API地址
	MQTTBroker        string `yaml:"mqtt_broker"`   // MQTT服务器地址
	MQTTUsername      string `yaml:"mqtt_username"` // MQTT用户名
	MQTTPassword      string `yaml:"mqtt_password"` // MQTT密码
	ServiceIdentifier string `yaml:"service_identifier"`
}

type LogConfig struct {
	Level      string `yaml:"level"`
	FilePath   string `yaml:"filePath"`
	MaxSize    int    `yaml:"maxSize"`    // 每个日志文件的最大大小（MB）
	MaxBackups int    `yaml:"maxBackups"` // 保留的旧日志文件的最大数量
	MaxAge     int    `yaml:"maxAge"`     // 保留日志文件的最大天数
	Compress   bool   `yaml:"compress"`   // 是否压缩旧日志文件
}
