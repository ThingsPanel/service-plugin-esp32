// cmd/main.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"tp-plugin/internal/config"
	"tp-plugin/internal/handler"
	"tp-plugin/internal/pkg/logger"
	"tp-plugin/internal/platform"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

func main() {
	// 首先设置基本的日志格式
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	logrus.Info("=================== OPC-UA 插件服务启动 ===================")

	app := &cli.App{
		Name:    "tp-plugin",
		Usage:   "tp-plugin OPC-UA protocol plugin",
		Version: "1.0.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "../configs/config.yaml",
				Usage:   "config file path",
			},
		},
		Action: run,
	}

	if err := app.Run(os.Args); err != nil {
		logrus.WithError(err).Fatal("程序运行失败")
	}
}

func run(c *cli.Context) error {
	// 1. 配置文件检查
	configPath := c.String("config")
	logrus.Infof("正在检查配置文件路径: %s", configPath)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logrus.WithError(err).Errorf("配置文件不存在: %s", configPath)
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 2. 加载配置
	logrus.Info("开始加载配置文件...")
	cfg, err := loadConfig(configPath)
	if err != nil {
		logrus.WithError(err).Error("加载配置文件失败")
		return fmt.Errorf("加载配置文件失败: %v", err)
	}
	logrus.WithFields(logrus.Fields{
		"port":            cfg.Server.Port,
		"max_connections": cfg.Server.MaxConnections,
		"heartbeat":       cfg.Server.HeartbeatTimeout,
		"log_level":       cfg.Log.Level,
		"log_path":        cfg.Log.FilePath,
	}).Info("配置加载成功")

	// 3. 日志目录检查和初始化
	logrus.Info("正在初始化日志系统...")
	if err := ensureLogDir(cfg.Log.FilePath); err != nil {
		logrus.WithError(err).Error("创建日志目录失败")
		return fmt.Errorf("创建日志目录失败: %v", err)
	}
	logger.InitLogger(&cfg.Log)
	logrus.Info("日志系统初始化完成")

	// 4. 创建平台客户端
	logrus.Info("正在初始化平台客户端...")
	platformClient, err := platform.NewPlatformClient(platform.Config{
		BaseURL:      cfg.Platform.URL,
		MQTTBroker:   cfg.Platform.MQTTBroker,
		MQTTUsername: cfg.Platform.MQTTUsername,
		MQTTPassword: cfg.Platform.MQTTPassword,
	}, logrus.StandardLogger())
	if err != nil {
		return fmt.Errorf("创建平台客户端失败: %v", err)
	}
	defer platformClient.Close()
	logrus.Info("平台客户端初始化成功")

	// // 5. 创建并初始化服务管理器
	// logrus.Info("正在初始化服务管理器...")
	// serviceMgr := manager.NewServiceManager(
	// 	platformClient,
	// 	manager.Config{
	// 		UpdateInterval:  time.Minute * 1,        // 每分钟更新一次服务列表
	// 		ConnectTimeout:  time.Second * 30,       // 连接超时30秒
	// 		RequestTimeout:  time.Second * 10,       // 请求超时10秒
	// 		PublishInterval: time.Millisecond * 500, // 发布间隔500ms
	// 	},
	// 	logrus.StandardLogger(),
	// )

	// // 启动服务管理器
	// if err := serviceMgr.Start(); err != nil {
	// 	logrus.WithError(err).Error("启动服务管理器失败")
	// 	return fmt.Errorf("启动服务管理器失败: %v", err)
	// }
	// defer serviceMgr.Stop()
	// logrus.Info("服务管理器启动成功")

	// 6. 创建并启动HTTP服务
	httpHandler := handler.NewHTTPHandler(platformClient, logrus.StandardLogger())
	handlers := httpHandler.RegisterHandlers()
	httpPort := cfg.Server.HTTPPort
	go func() {
		logrus.Infof("正在启动HTTP服务，端口: %d", httpPort)
		if err := handlers.Start(fmt.Sprintf(":%d", httpPort)); err != nil {
			logrus.Errorf("HTTP服务启动失败: %v", err)
		}
	}()

	logrus.Info("插件HTTP服务启动成功")

	// 7. 阻塞主goroutine,等待信号
	select {}
}

func loadConfig(configPath string) (*config.Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func ensureLogDir(logPath string) error {
	dir := filepath.Dir(logPath)
	return os.MkdirAll(dir, 0755)
}
