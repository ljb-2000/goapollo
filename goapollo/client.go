package goapollo

import (
	"sync"
	log "github.com/sirupsen/logrus"
	"fmt"
	"os"
	"time"
)

type Client struct {
	mux *sync.RWMutex
	// Apollo 配置集合
	configurations map[string]*ApolloConfig
	//监听的 HTTP 请求端口
	Port int
	//监听的IP地址
	Addr string

	server *HttpServer
}

func NewClient(options ...func(client *Client)) *Client {
	client := &Client{mux: &sync.RWMutex{}, configurations: make(map[string]*ApolloConfig)}

	if len(options) > 0 {
		for _, option := range options {
			option(client)
		}
	}

	return client
}

func (c *Client) Run() {
	endRunning := make(chan bool, 1)

	if len(c.configurations) > 0 {
		for _, item := range c.configurations {

			if item.FullPullFromCacheInterval > 0 {
				//当启用了全量拉取，则启动一个goroutine来执行任务
				go func(config *ApolloConfig) {
					defer func() {
						if err := recover(); err != nil {
							log.Errorf("panic serving : %v %v", config, err)
						}
					}()
					channel := config.StartFullFromCacheFetchTaskWithIntervalAsync(item.FullPullFromCacheInterval)

					for {
						entity := <-channel

						log.Infof("全量拉取成功 -> %d - %v", config.client.config.NamespaceName, entity)
						//如果设置了配置文件则保存配置
						if config.LocalFiles != nil && len(config.LocalFiles) > 0 {
							for _, f := range config.LocalFiles {
								ExecuteHandler(entity, f)
							}
						}
					}
				}(item)
			}

			go func(config *ApolloConfig) {
				defer func() {
					if err := recover(); err != nil {
						log.Errorf("panic serving : %v %v", config, err)
					}
				}()

				channel := make(chan *ConfigEntity, 1)
				config.ListenRemoteConfigLongPollNotificationServiceAsync(channel)
				for {
					entity := <-channel
					log.Infof("变更事件触发拉取 -> %s - %v", config.client.config.NamespaceName, entity)
					//如果设置了配置文件则保存配置
					if config.LocalFiles != nil && len(config.LocalFiles) > 0 {
						for _, f := range config.LocalFiles {
							ExecuteHandler(entity, f)
						}
					}
				}
			}(item)
		}
	}

	if c.Port > 0 {
		server := NewHttpServer()
		go func() {
			if err := server.Start(fmt.Sprintf("%s:%d", c.Addr, c.Port)); err != nil {
				log.Errorf("ListenAndServe -> ", err, fmt.Sprintf("%d", os.Getpid()))
				time.Sleep(100 * time.Microsecond)
				endRunning <- true
			}

		}()
	}
	log.Infof("Apollo running.")
	<-endRunning
}

func (c *Client) AddApolloConfig(configs ...*ApolloConfig) {
	if len(configs) > 0 {
		c.mux.Lock()
		defer c.mux.Unlock()
		for _, config := range configs {
			key := fmt.Sprintf("%s.%s.%s", config.client.config.AppId, config.client.config.Cluster, config.client.config.NamespaceName)
			c.configurations[key] = config
		}
	}

}
