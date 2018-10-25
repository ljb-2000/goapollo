package goapollo

import (
	"sync"
	log "github.com/sirupsen/logrus"
	"fmt"
)

type ApolloClient struct {
	mux *sync.RWMutex
	// Apollo 配置集合
	configurations map[string]*ApolloConfig
	//监听的 HTTP 请求端口
	Port int

	server *HttpServer
}

func NewApolloClient(options ...func(client *ApolloClient)) *ApolloClient {
	client := &ApolloClient{mux: &sync.RWMutex{}, configurations: make(map[string]*ApolloConfig)}

	if len(options) > 0 {
		for _, option := range options {
			option(client)
		}
	}

	return client
}

func (c *ApolloClient) Run() {
	endRunning := make(chan bool, 1)

	if len(c.configurations) > 0 {
		for _, item := range c.configurations {
			go func(config *ApolloConfig) {
				defer func() {
					if err := recover(); err != nil {
						log.Errorf("panic serving : %s %v", config.String(), err)
					}
				}()

				//if err := config.Start(); err != nil {
				//	log.Errorf("panic serving : %s %v", config.String(), err)
				//	endRunning <- true
				//}

				channel := config.ListenRemoteConfigLongPollNotificationServiceAsync()
				for {
					entry := <-channel
					log.Info(entry)
				}
			}(item)
		}
	}
	log.Infof("Apollo running.")
	<-endRunning
}

func (c *ApolloClient) AddApolloConfig(configs ...*ApolloConfig) {
	if len(configs) > 0{
		c.mux.Lock()
		defer c.mux.Unlock()
		for _,config := range configs {
			key := fmt.Sprintf("%s.%s.%s", config.AppId, config.ClusterName, config.NamespaceName)
			c.configurations[key] = config
		}
	}

}
