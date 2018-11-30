package goapollo

import (
	"sync"
	"fmt"
	"encoding/json"
)


// Apollo 通知信息
type ApolloConfigNotification struct {
	NamespaceName  string `json:"namespaceName"`
	NotificationId int64 `json:"notificationId"`
}

func NewApolloConfigNotification() *ApolloConfigNotification {
	return &ApolloConfigNotification{NotificationId: defaultNotificationId}
}


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

func (message *ApolloNotificationMessages) Delete(key string) int64 {
	message.init()
	if v, ok := message.details[key]; ok {
		message.mux.Lock()
		defer message.mux.Unlock()
		delete(message.details,key)
		return v
	}
	return 0
}

func (message *ApolloNotificationMessages) Clear(key string) {
	message.mux.Lock()
	defer message.mux.Unlock()
	message.details = make(map[string]int64)
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

func (message *ApolloNotificationMessages)  Clone() *ApolloNotificationMessages {
	mm := NewApolloNotificationMessages()
	for k,v := range message.details  {
		mm.Put(k,v)
	}
	return mm
}

func (message *ApolloNotificationMessages) Foreach(fn func(namespace string,id int64)) {
	message.mux.RLock()
	defer message.mux.RUnlock()
	for k,v := range message.details {
		fn(k,v)
	}
}

func (message *ApolloNotificationMessages) String() string {
	message.mux.RLock()
	defer message.mux.RUnlock()
	details := make([]*ApolloConfigNotification,0)
	for k,v := range message.details {
		notify := &ApolloConfigNotification{
			NamespaceName:k,
			NotificationId:v,
		}
		details = append(details,notify)
	}

	if b,err := json.Marshal(details); err == nil {
		return string(b)
	}
	return ""
}
