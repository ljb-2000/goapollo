package goapollo

import (
	"sync"
	"fmt"
)
const (
	defaultNotificationId = -1
)


type apolloNotify struct {
	NotificationId int64  `json:"notificationId"`
	NamespaceName  string `json:"namespaceName"`
}

// Apollo 通知信息
type ApolloConfigNotification struct {
	NamespaceName  string
	NotificationId int64
}

func NewApolloConfigNotification() *ApolloConfigNotification {
	return &ApolloConfigNotification{NotificationId: defaultNotificationId}
}

// 添加通知.
//func (notify *ApolloConfigNotification) AddMessage(key string, notificationId int64) {
//	if notify.Messages == nil {
//		notify.Messages = NewApolloNotificationMessages()
//	}
//	notify.Messages.Put(key, notificationId)
//}

func (notify *ApolloConfigNotification) String() string {
	return fmt.Sprintf("ApolloConfigNotification{ namespaceName='%s', notificationId=%d }", notify.NamespaceName, notify.NotificationId);
}

// Apollo 通知消息集合.
type ApolloNotificationMessages struct {
	details map[string]int64
	mux     *sync.RWMutex
}

func NewApolloNotificationMessages() *ApolloNotificationMessages {
	return &ApolloNotificationMessages{details: make(map[string]int64), mux: &sync.RWMutex{}}
}

// 添加一个通知到集合中.
func (message *ApolloNotificationMessages) Put(key string, notificationId int64) {
	message.init()
	message.mux.Lock()
	defer message.mux.Unlock()

	message.details[key] = notificationId
}

func (message *ApolloNotificationMessages) Get(key string) (int64, bool) {
	message.init()
	if v, ok := message.details[key]; ok {
		return v, ok
	}
	return 0, false
}

func (message *ApolloNotificationMessages) Has(key string) bool {
	message.init()
	if _, ok := message.details[key]; ok {
		return ok
	}
	return false
}

func (message *ApolloNotificationMessages) MergeFrom(source *ApolloNotificationMessages) {
	message.init()
	source.mux.RLock()
	defer source.mux.RUnlock()
	message.mux.Lock()
	defer message.mux.Unlock()

	for k, v := range source.details {
		message.details[k] = v
	}
}

func (message *ApolloNotificationMessages) init() {
	if message.details == nil {
		message.details = make(map[string]int64)
	}
	if message.mux == nil {
		message.mux = &sync.RWMutex{}
	}
}
