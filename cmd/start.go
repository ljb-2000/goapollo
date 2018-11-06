package cmd

import (
	"gopkg.in/urfave/cli.v2"
	"time"
	"github.com/lifei6671/goapollo/goapollo"
	"os"
	log "github.com/sirupsen/logrus"
	"github.com/lifei6671/goini"
	"strings"
	"path/filepath"
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
			Name:    "logger",
			Aliases: []string{"log"},
			Value:   "conf/seelog.xml",
			Usage:   "logrus configuration file path.",
			EnvVars: []string{"APOLLO_LOGGER"},
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
			Value:    60,
			Usage:   "Timed full pull time interval. ",
			EnvVars: []string{"APOLLO_LONG_INTERVAL"},
		},
		&cli.IntFlag{
			Name: "full_interval",
			Value: 60,
			Usage:"",
			EnvVars:[]string{"APOLLO_FULL_INTERVAL"},
		},
	},
}

/**
{
  "appId": "6e77bd897fe903ac",
  "cluster": "default",
  "namespaceName": "TEST1.ini",
  "ip": "http://dev.config.xin.com/",
  "configFilePath": "fastcgi_param.conf"
}
 */
func runStart(c *cli.Context) error {

	saveFile := c.String("file")

	configFile := c.String("config")

	configs := make([]*goapollo.ApolloConfig, 0)
	port := c.Int("port")
	addr := c.String("addr")

	//如果配置文件存在则初始化配置文件,如果没有指定配置文件或配置文件不存在则从其他参数中获取
	if _, err := os.Stat(configFile); err == nil {
		ini, err := goini.LoadFromFile(configFile);
		if err != nil {
			log.Fatalf("配置文件解析失败 -> %s %s", configFile, err)
		}
		port = ini.DefaultInt("port", port)
		addr = ini.DefaultString("addr", "")

		ini.ForEach(func(section string) bool {
			log.Info(section)
			if section != goini.DefaultSection && strings.HasPrefix(section, "app:") {
				appId := ini.GetString(section + "::appId")
				if appId == "" {
					log.Fatalf("Apollo aplication id not does empty -> [Section] %s", section)
					return true
				}
				url := ini.GetString(section + "::serverUrl")
				if url == "" {
					log.Fatalf("Apollo server url not does empty -> [Section] %s", section)
					return true
				}
				if !strings.HasSuffix(url, "/") {
					url += "/"
				}
				cluster := ini.DefaultString(section+"::cluster", "default")
				namespace := ini.DefaultString(section+"::namespace", "application")
				saveFile := ini.DefaultString(section+"::saveFile", "")

				if sf,err := filepath.Abs(saveFile); err == nil {
					dir := filepath.Dir(sf)

					if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
						os.MkdirAll(dir, 0755)
					}
					saveFile = sf
				}
				config := goapollo.NewApolloConfig(url, appId)
				config.LocalFilePath = saveFile
				config.NamespaceName = namespace
				config.ClusterName = cluster
				config.AppId = appId
				config.ConfigServerUrl = url
				//轮询拉取时间间隔.
				if longPollInterval,err := ini.Int("longPollInterval");err == nil && longPollInterval > 0 {
					config.LongPollInterval = time.Second * time.Duration(longPollInterval)
				}
				//全量拉取时间间隔
				if fullPullInterval,err := ini.Int("fullPullFromCacheInterval");err == nil && fullPullInterval > 0 {
					config.FullPullFromCacheInterval = time.Second * time.Duration(fullPullInterval)
				}

				log.Infof("Add configuration application -> %s", config.String())
				configs = append(configs, config)
			}
			return true
		})
	} else if appId := c.String("app_id"); appId != "" {
		//从命令行中获取参数
		url := c.String("server_url")
		if url == "" {
			log.Error("Apollo server url not does empty.")
			os.Exit(1)
		}

		config := goapollo.NewApolloConfig(url, appId)

		if cluster := c.String("cluster"); cluster != "" {
			config.ClusterName = cluster
		}
		if namespace := c.String("namespace"); namespace != "" {
			config.NamespaceName = namespace
		}
		//解析保存文件路径.
		if sf,err := filepath.Abs(saveFile); err == nil {
			dir := filepath.Dir(sf)

			if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
				os.MkdirAll(dir, 0755)
			}
			saveFile = sf
		}
		config.LocalFilePath = saveFile
		//轮询拉取时间间隔.
		if longPollInterval := c.Duration("long_interval"); longPollInterval > 0 {
			config.LongPollInterval = longPollInterval
		}
		//全量拉取时间间隔
		if fullPullInterval := c.Duration("full_interval"); fullPullInterval > 0 {
			config.FullPullFromCacheInterval = fullPullInterval
		}

		configs = append(configs, config)

	} else {
		log.Fatal("Not found configuration file.")
	}

	log.Info("开始监听配置变更.")

	config := func(client *goapollo.ApolloClient) {
		client.Port = port
		client.Addr = addr
	}

	apolloConfig := func(client *goapollo.ApolloClient) {
		client.AddApolloConfig(configs...)
	}

	client := goapollo.NewApolloClient(config, apolloConfig)
	client.Run()

	return nil
}
