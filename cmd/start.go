package cmd

import (
	"gopkg.in/urfave/cli.v2"
	"time"
	"github.com/lifei6671/goapollo/goapollo"
	"os"
	log "github.com/sirupsen/logrus"
	logxin "github.com/lifei6671/goapollo/log"
	"github.com/lifei6671/goini"
	"strings"
)

var Start = &cli.Command{
	Name:        "run",
	Usage:       "Run apollo client",
	Description: `Run the apollo client with the specified parameters.`,
	Action:      runStart,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Value:   "",
			Usage:   "Custom configuration file path",
			EnvVars: []string{"APOLLO_CONFIG"},
		},
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"f"},
			Value:   "",
			Usage:   "Save config to file path.",
			EnvVars: []string{"APOLLO_SAVE_PATH"},
		},
		&cli.IntFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Value:   8088,
			Usage:   "Http listen port.",
			EnvVars: []string{"APOLLO_HTTP_PORT"},
		},
		&cli.StringFlag{
			Name:    "addr",
			Value:   "",
			Usage:   "Http listen addr",
			EnvVars: []string{"APOLLO_HTTP_ADDR"},
		},
		&cli.StringFlag{
			Name:    "log_path",
			Value:   "./runtime/logs/",
			Usage:   "Log output directory.",
			EnvVars: []string{"APOLLO_LOG_PATH"},
		},
		&cli.StringFlag{
			Name:    "app_id",
			Aliases: []string{"id"},
			Value:   "",
			Usage:   "Apollo application id value.",
			EnvVars: []string{"APOLLO_APP_ID"},
		},
		&cli.StringFlag{
			Name:    "cluster",
			Value:   "default",
			Usage:   "Apollo cluster name.",
			EnvVars: []string{"APOLLO_CLUSTER"},
		},
		&cli.StringFlag{
			Name:    "namespace",
			Aliases: []string{"ns"},
			Usage:   "Apollo namespace name.",
			EnvVars: []string{"APOLLO_NAMESPACE_NAME"},
		},
		&cli.StringFlag{
			Name:    "server_url",
			Aliases: []string{"u"},
			Usage:   "Apollo server url.",
			EnvVars: []string{"APOLLO_SERVER_URL"},
		},
		&cli.IntFlag{
			Name:    "long_interval",
			Value:   1,
			Usage:   "Timed notification time interval. ",
			EnvVars: []string{"APOLLO_LONG_INTERVAL"},
		},
		&cli.IntFlag{
			Name:    "full_interval",
			Usage:   "Timed full pull time interval.",
			EnvVars: []string{"APOLLO_FULL_INTERVAL"},
		},
		&cli.StringFlag{
			Name:    "related",
			Usage:   "Associated namespace.",
			EnvVars: []string{"APOLLO_RELATED"},
		},
	},
}


func runStart(c *cli.Context) error {

	saveFile := c.String("file")

	configFile := c.String("config")

	configs := make([]*goapollo.ApolloConfig, 0)
	port := c.Int("port")
	addr := c.String("addr")
	logPath := ""

	//如果配置文件存在则初始化配置文件,如果没有指定配置文件或配置文件不存在则从其他参数中获取
	if _, err := os.Stat(configFile); err == nil {
		ini, err := goini.LoadFromFile(configFile);
		if err != nil {
			log.Fatalf("配置文件解析失败 -> %s %s", configFile, err)
		}
		port = ini.DefaultInt("port", port)
		addr = ini.DefaultString("addr", "")
		logPath = ini.DefaultString("log_path", "")

		ini.ForEach(func(section string) bool {
			if section != goini.DefaultSection && strings.HasPrefix(section, "app:") {
				appId := ini.GetString(section + "::appId")
				if appId == "" {
					log.Fatalf("Apollo aplication id not does empty -> [Section] %s", section)
					return true
				}
				serverUrl := ini.GetString(section + "::serverUrl")
				if serverUrl == "" {
					log.Fatalf("Apollo server url not does empty -> [Section] %s", section)
					return true
				}

				cluster := ini.DefaultString(section+"::cluster", "default")
				namespace := ini.DefaultString(section+"::namespace", "application")
				saveFile := ini.DefaultString(section+"::saveFile", "")

				config := goapollo.NewApolloConfig(appId, cluster, namespace, serverUrl)
				config.LocalFiles = goapollo.ApolloLocalFileFromString(saveFile)

				//轮询拉取时间间隔.
				if longPollInterval, err := ini.Int(section + "::longPollInterval"); err == nil && longPollInterval > 0 {
					config.LongPollInterval = time.Second * time.Duration(longPollInterval)
				}

				//全量拉取时间间隔
				if fullPullInterval, err := ini.Int(section + "::fullPullFromCacheInterval"); err == nil && fullPullInterval > 0 {
					config.FullPullFromCacheInterval = time.Second * time.Duration(fullPullInterval)
				}

				if related := ini.DefaultString(section+"::related", ""); related != "" {

					if relates := strings.Split(related, ";"); len(relates) > 0 {
						config.AppendNamespace(relates...)
					}
				}

				log.Infof("Add configuration application -> %v", config)
				configs = append(configs, config)
			}
			return true
		})
	} else if appId := c.String("app_id"); appId != "" {
		//从命令行中获取参数
		serverUrl := c.String("server_url")
		if serverUrl == "" {
			log.Error("Apollo server url not does empty.")
			os.Exit(1)
		}
		cluster := c.String("cluster");
		if cluster == "" {
			cluster = "default"
		}
		namespace := c.String("namespace");
		if namespace == "" {
			namespace = "application"
		}

		config := goapollo.NewApolloConfig(appId, cluster, namespace, serverUrl)

		config.LocalFiles = goapollo.ApolloLocalFileFromString(saveFile)
		//轮询拉取时间间隔.
		if longPollInterval := c.Int("long_interval"); longPollInterval > 0 {
			config.LongPollInterval = time.Duration(longPollInterval) * time.Second
		}
		//全量拉取时间间隔
		if fullPullInterval := c.Duration("full_interval"); fullPullInterval > 0 {
			config.FullPullFromCacheInterval = time.Duration(fullPullInterval) * time.Second
		}

		if related := c.String("related"); related != "" {
			if relates := strings.Split(related, ";"); len(relates) > 0 {
				config.AppendNamespace(relates...)
			}
		}
		configs = append(configs, config)

		logPath = c.String("log_path")
	} else {
		log.Fatal("Not found configuration file.")
	}

	if logPath != "" {
		os.MkdirAll(logPath, 0755)
		log.AddHook(&logxin.ContextHook{LogPath: logPath})
	}
	log.Info("开始监听配置变更.")

	config := func(client *goapollo.Client) {
		client.Port = port
		client.Addr = addr
	}

	apolloConfig := func(client *goapollo.Client) {
		client.AddApolloConfig(configs...)
	}

	client := goapollo.NewClient(config, apolloConfig)
	client.Run()

	return nil
}
