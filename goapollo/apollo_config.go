package goapollo

import (
	"time"
	"fmt"
	"net/http"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
	"encoding/json"
	"errors"
	"strings"
	"net/url"
	"sync"
)

var (
	ErrConfigUnmodified = errors.New("Apollo configuration not changed. ")
)

//Apollo 配置信息.
type ApolloConfig struct {
	//Apollo配置服务的地址.
	ConfigServerUrl string `json:"-"`
	//应用的appId
	AppId string `json:"appId"`
	// 集群名称，一般情况下传入 default 即可。
	// 如果希望配置按集群划分，可以参考集群独立配置说明做相关配置，然后在这里填入对应的集群名.
	ClusterName string `json:"cluster"`
	// 命名空间，如果没有新建过Namespace的话，传入application即可。
	// 如果创建了Namespace，并且需要使用该Namespace的配置，则传入对应的Namespace名字。
	// 需要注意的是对于properties类型的namespace，只需要传入namespace的名字即可，如application。
	// 对于其它类型的namespace，需要传入namespace的名字加上后缀名，如datasources.json
	NamespaceName string `json:"namespaceName"`
	// 应用部署的机器IP.
	// 这个参数是可选的，用来实现灰度发布。 如果不想传这个参数，请注意URL中从?号开始的query parameters整个都不要出现。
	IpAddress string `json:"ip"`

	Configurations map[string]string `json:"configurations"`
	//如果不是 properties 结构，则会储存在该字段内。
	ConfigFileContent string `json:"-"`

	ReleaseKey string `json:"releaseKey"`
	// 从缓存中全量拉取轮询的时间间隔.
	FullPullFromCacheInterval time.Duration `json:"-"`
	// 从数据库全量拉取轮询的时间间隔
	FullPullFromDbInterval time.Duration `json:"-"`
	//监听Apollo通知间隔.
	LongPollInterval time.Duration `json:"-"`
	//Apollo通知信息.
	notificationId int64 `json:"-"`

	mux *sync.RWMutex `json:"-"`
	//保存的目标文件.
	LocalFilePath string `json:"-"`
}

func NewApolloConfig(url, appId string) *ApolloConfig {
	config := &ApolloConfig{ConfigServerUrl: url,
		AppId: appId,
		ClusterName: "default",
		NamespaceName: "application",
		FullPullFromCacheInterval: time.Second * 60,
		LongPollInterval: time.Second * 30,
		notificationId: defaultNotificationId,
		mux: &sync.RWMutex{},
		Configurations: make(map[string]string),
	}

	return config
}


// 启动全量定时拉取任务.
func (c *ApolloConfig) StartFullFromCacheFetchTaskWithInterval(interval time.Duration, pairChannel chan <- KeyValuePair) {

	t := time.NewTimer(interval)

	for {
		select {
		case <-t.C:
			pair, err := c.GetApolloRemoteConfigFromCache()
			if err != nil {
				if err == ErrConfigUnmodified {
					log.Info(err.Error())
				} else {
					log.Errorf("get apollo remote config fail:%s", err)
				}
			} else {
				log.Info("From Apollo cache fetch config success. ")
				pairChannel <- pair
			}
			t.Reset(interval)
		}
	}
}

// 定时全量从Apollo数据库拉取配置信息.
func (c *ApolloConfig) StartFullFromDbFetchTaskWithInterval(interval time.Duration, pairChannel chan<- KeyValuePair) {
	t := time.NewTimer(interval)

	for {
		select {
		case <-t.C:
			pair, err := c.GetApolloRemoteConfigFromDb()
			if err != nil {
				if err == ErrConfigUnmodified {
					log.Info(err.Error())
				} else {
					log.Errorf("get apollo remote config fail:%s", err)
				}
			} else {
				log.Info("From Apollo cache fetch config success. ")
				pairChannel <- pair
			}
			t.Reset(interval)
		}
	}
}

// 从远程服务器提供的接口获取配置信息.
func (c *ApolloConfig) GetApolloRemoteConfigFromCache() (KeyValuePair, error) {
	client := &http.Client{}
	serverUrl := c.getApolloRemoteConfigFromCacheUrl()
	req, err := http.NewRequest("GET", serverUrl, nil)

	if err != nil {
		log.Errorf("request error:%s %s", serverUrl, err)
		return nil, err
	}
	resp, err := client.Do(req)

	if err != nil {
		log.Errorf("response error:%s %s", serverUrl, err)
		return nil, err
	}
	defer resp.Body.Close()
	//如果服务器返回304，标识配置没有变更
	if resp.StatusCode == 304 {
		return nil, ErrConfigUnmodified
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("remote server response status code error: %s %d", serverUrl, resp.StatusCode))
	}
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Errorf("read response body error:%s %s", serverUrl, err)
		return nil, err
	}

	kv := make(map[string]string, 0)
	err = json.Unmarshal(body, &kv)
	if err != nil {
		log.Errorf("Unmarshal json result error:%s %s %s", serverUrl, string(body), err)
	}
	return kv, err
}

//通过不带缓存的Http接口从Apollo读取配置.
func (c *ApolloConfig) GetApolloRemoteConfigFromDb() (KeyValuePair, error) {
	client := &http.Client{}
	uri := c.getApolloRemoteConfigFromDbUrl()
	req, err := http.NewRequest("GET", uri, nil)

	if err != nil {
		log.Errorf("request error:%s %s", uri, err)
		return nil, err
	}
	resp, err := client.Do(req)

	if err != nil {
		log.Errorf("response error:%s %s", uri, err)
		return nil, err
	}
	defer resp.Body.Close()

	//如果服务器返回 304，标识配置没有变更
	if resp.StatusCode == 304 {
		return nil, ErrConfigUnmodified
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("remote server response status code error: %s %d", uri, resp.StatusCode))
	}
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Errorf("read response body error:%s %s", uri, err)
		return nil, err
	}

	config := ApolloConfig{}

	//解码后，如果解码成功，则将新值复制给原对象
	err = json.Unmarshal(body, &config)
	if err != nil {
		log.Errorf("Unmarshal json result error:%s %s %s", uri, string(body), err)
	} else {
		c.ReleaseKey = config.ReleaseKey
		return config.Configurations, nil
	}
	return nil, err
}



//监听 Apollo 的通知消息.
func (c *ApolloConfig) ListenRemoteConfigLongPollNotificationServiceAsync() (<-chan *ConfigChangeEventArgs) {

	channel := make(chan *ConfigChangeEventArgs, 1)

	if kv, err := c.GetApolloRemoteConfigFromDb(); err == nil {
		c.mux.Lock()
		for k, v := range kv {
			c.Configurations[k] = v
		}
		c.mux.Unlock()
	}

	log.Info("Start async notification listen.")
	go func() {
		t := time.NewTimer(c.LongPollInterval)

		for {
			select {
			case <-t.C:
				{
					configs, err := c.getApolloRemoteNotification()

					if err == nil && len(configs) > 0 {
						for _, config := range configs {
							if config.NamespaceName == c.NamespaceName {
								c.notificationId = config.NotificationId
								log.Infof("Apollo notification event: %s - %d", config.NamespaceName, config.NotificationId)
								kv, err := c.GetApolloRemoteConfigFromDb()
								if err == nil && len(kv) > 0 {
									channel <- c.updateConfigurationCache(kv)
								}
								break
							}
						}
					}
					t.Reset(c.LongPollInterval)
				}
			}
		}
	}()

	return channel
}

//从远程通知接口获取通知信息.
func (c *ApolloConfig) getApolloRemoteNotification() ([]*apolloNotify, error) {
	uri := c.getApolloRemoteNotificationUrl()

	log.Infof("Fetch Apollo remote server notification:%s", uri)

	client := &http.Client{}

	req, err := http.NewRequest("GET", uri, nil)

	if err != nil {
		log.Errorf("request error:%s %s", uri, err)
		return nil, err
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Errorf("response error:%s %s", uri, err)
		return nil, err
	}
	defer resp.Body.Close()

	//如果服务器返回 304，标识配置没有变更
	if resp.StatusCode == 304 {
		log.Info("Apollo configuration not changed.")
		return nil, ErrConfigUnmodified
	}

	if resp.StatusCode != 200 {
		log.Errorf("remote server response status code error: %s %d", uri, resp.StatusCode)
		return nil, errors.New(fmt.Sprintf("remote server response status code error: %s %d", uri, resp.StatusCode))
	}
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Errorf("read response body error:%s %s", uri, err)
		return nil, err
	}
	remoteConfig := make([]*apolloNotify, 0)

	err = json.Unmarshal(body, &remoteConfig)

	if err != nil {
		log.Error("Unmarshal Msg Fail,Error:", err)
		return nil, err
	}
	return remoteConfig, nil
}

// 获取URL
func (c *ApolloConfig) getApolloRemoteConfigFromCacheUrl() string {
	getApolloRemoteConfigFromCacheUrl := fmt.Sprintf("%sconfigfiles/json/%s/%s/%s",
		c.ConfigServerUrl,
		c.AppId,
		c.ClusterName,
		c.NamespaceName)

	if c.IpAddress != "" {
		getApolloRemoteConfigFromCacheUrl += "?ip=" + c.IpAddress
	}
	return getApolloRemoteConfigFromCacheUrl
}

// 生成通过不带缓存的Http接口从Apollo读取配置的链接
func (c *ApolloConfig) getApolloRemoteConfigFromDbUrl() string {
	getApolloRemoteConfigFromDbUrl := fmt.Sprintf("%sconfigs/%s/%s/%s?",
		c.ConfigServerUrl,
		c.AppId,
		c.ClusterName,
		c.NamespaceName)
	if c.ReleaseKey == "" {
		getApolloRemoteConfigFromDbUrl += "releaseKey=" + c.ReleaseKey + "&"
	}
	if c.IpAddress == "" {
		getApolloRemoteConfigFromDbUrl += "ip=" + c.IpAddress
	}
	getApolloRemoteConfigFromDbUrl = strings.TrimSuffix(getApolloRemoteConfigFromDbUrl, "?")

	return getApolloRemoteConfigFromDbUrl
}

//通知地址.
func (c *ApolloConfig) getApolloRemoteNotificationUrl() string {
	notificationUrl := fmt.Sprintf("%snotifications/v2?appId=%s&cluster=%s&notifications=%s",
		c.ConfigServerUrl,
		url.QueryEscape(c.AppId),
		url.QueryEscape(c.ClusterName),
		url.QueryEscape(fmt.Sprintf("[{\"namespaceName\":\"%s\",\"notificationId\":%d}]", c.NamespaceName, c.notificationId)), )

	return notificationUrl
}

//更新配置.
func (c *ApolloConfig) updateConfigurationCache(kv KeyValuePair) (*ConfigChangeEventArgs) {

	c.mux.Lock()
	defer c.mux.Unlock()

	args := NewConfigChangeEventArgs(c.NamespaceName, C_TYPE_POROPERTIES)

	if len(kv) <= 0 {
		return args
	}
	tempMap := make(map[string]string)

	for k, v := range c.Configurations {
		tempMap[k] = v
	}

	for k, v := range kv {
		//与已存在的配置信息作比较
		if vv, ok := c.Configurations[k]; ok {
			delete(tempMap, k)

			entry := &ConfigChangeEntry{
				NamespaceName: c.NamespaceName,
				PropertyName:  k,
				OldValue:      vv,
				NewValue:      v,
			}
			//如果值不相等，则是修改了值
			if vv != v {
				entry.ChangeType = C_MODIFIED
				c.Configurations[k] = v
			} else {
				entry.ChangeType = C_UNCHANGED
			}
			args.Values[k] = entry
		} else {
			entry := &ConfigChangeEntry{
				NamespaceName: c.NamespaceName,
				PropertyName:  k,
				OldValue:      vv,
				NewValue:      v,
				ChangeType:    C_ADDED,
			}
			args.Values[k] = entry
			c.Configurations[k] = v
		}
	}
	//删除的项
	for k, v := range tempMap {
		entry := &ConfigChangeEntry{
			NamespaceName: c.NamespaceName,
			PropertyName:  k,
			OldValue:      v,
			NewValue:      "",
			ChangeType:    C_DELETED,
		}
		args.Values[k] = entry
	}
	return args
}

// 获取 String.
func (c *ApolloConfig) String() string {
	return fmt.Sprintf("ApolloConfig{ appId='%s' , cluster='%s', namespaceName='%s', configurations=%v, releaseKey='%s' }",
		c.AppId,
		c.ClusterName,
		c.NamespaceName,
		c.Configurations,
		c.ReleaseKey)
}
