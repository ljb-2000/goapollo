package goapollo

import (
	"time"
	log "github.com/sirupsen/logrus"
)

//Apollo 配置信息.
type ApolloConfig struct {
	// 从缓存中全量拉取轮询的时间间隔.
	FullPullFromCacheInterval time.Duration
	// 从数据库全量拉取轮询的时间间隔
	FullPullFromDbInterval time.Duration
	//监听Apollo通知间隔.
	LongPollInterval time.Duration
	//保存的目标文件.
	LocalFiles []*ApolloLocalFile
	client     *apolloClient
}

func NewApolloConfig(appId, cluster, namespace string, serverUrl string) *ApolloConfig {

	config := &ApolloConfig{
		client:                    NewApolloClient(NewApolloAppConfig(appId, cluster, namespace), serverUrl),
		FullPullFromCacheInterval: fullInterval,
		LongPollInterval:          longPoolInterval,
	}

	return config
}

// 启动全量定时拉取任务.
func (c *ApolloConfig) StartFullFromCacheFetchTaskWithIntervalAsync(interval time.Duration) (<-chan *ConfigEntity) {

	pairChannel := make(chan *ConfigEntity, 1)
	go func() {
		t := time.NewTimer(time.Second * 1)

		for {
			select {
			case <-t.C:
				entity, err := c.client.GetApolloRemoteConfigFromCache()
				if err != nil {
					if err == ErrConfigUnmodified {
						log.Info(err.Error())
					} else {
						log.Errorf("get apollo remote config fail:%s", err)
					}
				} else {
					pairChannel <- entity
				}
				t.Reset(interval)
			}
		}
	}()
	return pairChannel
}

//监听 Apollo 的通知消息.
func (c *ApolloConfig) ListenRemoteConfigLongPollNotificationServiceAsync(channel chan<- *ConfigEntity) {

	notificationChan := make(chan *ApolloNotificationMessages, 1)

	c.client.getLongPoll(c.LongPollInterval, notificationChan).WatchNotification()

	go func() {
		for {
			<-notificationChan
			if entity, err := c.client.GetApolloRemoteConfigFromDatabase(); err == nil {
				channel <- entity
			}
		}
	}()
}

//追加辅助的名命名空间.
func (c *ApolloConfig) AppendNamespace(namespace ...string) *ApolloConfig {
	for _, ns := range namespace {
		if ns != "" {
			c.client.config.NamespaceNames = append(c.client.config.NamespaceNames,ns)
		}
	}
	return c
}