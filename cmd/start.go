package cmd

import (
	"gopkg.in/urfave/cli.v2"
	"time"
	"github.com/lifei6671/goapollo/goapollo"
	"os"
	log "github.com/sirupsen/logrus"
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
		&cli.StringFlag{
			Name:    "logger",
			Aliases: []string{"log"},
			Value:   "conf/seelog.xml",
			Usage:   "Seelog configuration file path.",
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
			Aliases: []string{"cus"},
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
		&cli.DurationFlag{
			Name:    "interval",
			Aliases: []string{"i"},
			Value:   time.Minute * 1,
			Usage:   "Timed full pull time interval. ",
			EnvVars: []string{"APOLLO_INTERVAL"},
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

	//如果配置文件存在则初始化配置文件,如果没有指定配置文件或配置文件不存在则从其他参数中获取
	if _, err := os.Stat(configFile); err == nil {

	} else if appId := c.String("app_id"); appId != "" {
		//从命令行中获取参数
		url := c.String("server_url")
		if url == "" {
			log.Error("Apollo ip address not does empty.")
			os.Exit(1)
		}

		config := goapollo.NewApolloConfig(url, appId)
		config.LocalFilePath = saveFile
		if cluster := c.String("cluster"); cluster != "" {
			config.ClusterName = cluster
		}
		if namespace := c.String("namespace"); namespace != "" {
			config.NamespaceName = namespace
		}
		configs = append(configs, config)

	} else {
		log.Error("Not found configuration file.")
		os.Exit(1)
	}

	//初始化日志
	//logConfigPath := c.String("logger")

	log.Info("开始监听配置变更.")

	config := func(client *goapollo.ApolloClient) {
		client.Port = 8080
	}

	apolloConfig := func(client *goapollo.ApolloClient) {
		client.AddApolloConfig(configs...)
	}

	client := goapollo.NewApolloClient(config, apolloConfig)
	client.Run()

	return nil
}
