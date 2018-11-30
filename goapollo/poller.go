package goapollo

import (
	"time"
	"net/http"
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

type longPoll struct {
	serverUrl        string
	appId            string
	cluster          string
	namespace        []string
	requester        requester
	pollInterval     time.Duration
	notifications    *ApolloNotificationMessages
	notificationChan chan<- *ApolloNotificationMessages
}

type NotificationHandler func(message *ApolloNotificationMessages) error

func NewPoll(serverUrl, appId, cluster string, namespace []string, interval time.Duration, notificationChan chan<- *ApolloNotificationMessages) *longPoll {
	p := &longPoll{
		serverUrl:        serverUrl,
		appId:            appId,
		cluster:          cluster,
		namespace:        namespace,
		pollInterval:     interval,
		requester:        newHTTPRequester(&http.Client{Timeout: longPoolTimeout}),
		notifications:    NewApolloNotificationMessages(),
		notificationChan: notificationChan,
	}
	for _, name := range namespace {
		p.notifications.Put(name, defaultNotificationId)
	}
	return p
}

//异步监控通知.
func (p *longPoll) WatchNotification() *longPoll {

	go func() {
		timer := time.NewTimer(time.Second * 1)
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				message, err := p.request()
				if err != nil {
					if err == ErrConfigUnmodified {
						log.Info("没有最新通知 ->", err)
						break
					}
					log.Errorf("请求 Apollo 服务器出错 -> %s", err)
				} else if message != nil {

					select {
					case <-time.After(time.Second * 1):
						break
					case p.notificationChan <- message:
						//下次再监听通知需要用上一次的通知ID.
						message.Foreach(func(namespace string, id int64) {
							p.notifications.Put(namespace, id)
						})
						break
					}
				}
			}
			timer.Reset(p.pollInterval)
		}
	}()
	return p
}

func (p *longPoll) request() (*ApolloNotificationMessages, error) {

	serverUrl := getApolloRemoteNotificationUrl(p.serverUrl, p.appId, p.cluster, p.notifications)

	log.Info("正在发起通知获取请求 ->", serverUrl)

	bts, err := p.requester.request(serverUrl)

	if err != nil || len(bts) == 0 {
		return nil, err
	}
	ret := NewApolloNotificationMessages()

	var m []*ApolloConfigNotification

	if err := json.Unmarshal(bts, &m); err != nil {
		return nil, err
	}
	if m != nil {
		for _, item := range m {
			ret.Put(item.NamespaceName, item.NotificationId)
		}
	}
	return ret, nil
}
