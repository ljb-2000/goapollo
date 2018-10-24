package goapollo

import (
	"time"
	"fmt"
	"net/http"
	"io/ioutil"
	log "github.com/cihub/seelog"
	"encoding/json"
	"errors"
)

var (
	GetApolloRemoteConfigFromCacheUrl    = "{config_server_url}/configfiles/json/{appId}/{clusterName}/{namespaceName}?ip={clientIp}"
	GetApolloRemoteConfigFromDatabaseUrl = "{config_server_url}/configs/{appId}/{clusterName}/{namespaceName}?releaseKey={releaseKey}&ip={clientIp}"
	ApolloHttpLongPollingUrl             = "{config_server_url}/notifications/v2?appId={appId}&cluster={clusterName}&notifications={notifications}"
)

//Apollo 配置信息.
type ApolloConfig struct {
	//Apollo配置服务的地址.
	ConfigServerUrl string
	//应用的appId
	AppId string
	// 集群名称，一般情况下传入 default 即可。
	// 如果希望配置按集群划分，可以参考集群独立配置说明做相关配置，然后在这里填入对应的集群名.
	ClusterName string
	// 命名空间，如果没有新建过Namespace的话，传入application即可。
	// 如果创建了Namespace，并且需要使用该Namespace的配置，则传入对应的Namespace名字。
	// 需要注意的是对于properties类型的namespace，只需要传入namespace的名字即可，如application。
	// 对于其它类型的namespace，需要传入namespace的名字加上后缀名，如datasources.json
	NamespaceName string
	// 应用部署的机器IP.
	// 这个参数是可选的，用来实现灰度发布。 如果不想传这个参数，请注意URL中从?号开始的query parameters整个都不要出现。
	IpAddress string

	Configurations map[string]string

	ReleaseKey string
	// 从缓存中全量拉取轮询的时间间隔.
	FullPullFromCacheInterval time.Duration
	// 从数据库全量拉取轮询的时间间隔
	FullPullFromDbInterval time.Duration
}

func NewApolloConfig(url, appId string) *ApolloConfig {
	return &ApolloConfig{ConfigServerUrl: url, AppId: appId, ClusterName: "default", NamespaceName: "application"}
}
// 启动监听.
func (c *ApolloConfig) Start() error {
	item,err := c.GetApolloRemoteConfigFromCache()

	if err != nil {
		return err
	}
	log.Infof("%v", item)
	return nil
}

// 从远程服务器提供的接口获取配置信息.
func (c *ApolloConfig) GetApolloRemoteConfigFromCache() (KeyValuePair, error) {
	client := &http.Client{}
	url := c.getApolloRemoteConfigFromCacheUrl()
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.Errorf("request error:%s %s", url, err)
		return nil, err
	}
	response, err := client.Do(req)

	if err != nil {
		log.Errorf("response error:%s %s", url, err)
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("remote server response status code error: %s %d", url, response.StatusCode))
	}
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Errorf("read response body error:%s %s", url, err)
		return nil, err
	}
	kv := make(map[string]string,0)

	err = json.Unmarshal(body, &kv)
	if err != nil {
		log.Errorf("Unmarshal json result error:%s %s %s", url, string(body), err)
	}
	return kv, err
}

// 获取URL
func (c *ApolloConfig) getApolloRemoteConfigFromCacheUrl() string {
	getApolloRemoteConfigFromCacheUrl := fmt.Sprintf("%s/configfiles/json/%s/%s/%s",
		c.ConfigServerUrl,
		c.AppId,
		c.ClusterName,
		c.NamespaceName)

	if c.IpAddress != "" {
		getApolloRemoteConfigFromCacheUrl += "?ip=" + c.IpAddress
	}
	return getApolloRemoteConfigFromCacheUrl
}

// 获取 String.
func (c *ApolloConfig) String() string {
	return fmt.Sprintf("ApolloConfig{ appId='%s' , cluster='%s', namespaceName='%s', configurations=%v, releaseKey='%s' }",
		c.AppId,
		c.ClusterName,
		c.NamespaceName,
		c.Configurations,
		c.ReleaseKey);
}
