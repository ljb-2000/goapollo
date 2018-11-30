package goapollo

import (
	"fmt"
	"strings"
	"path/filepath"
	"net/http"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"time"
)

type ApolloLocalFile struct {
	FileType ApolloLocalFileType
	FilePath string
}

func ApolloLocalFileFromString(str string) []*ApolloLocalFile {
	if str != "" {
		apolloLocalFiles := make([]*ApolloLocalFile, 0)
		//如果不存在分隔符，则默认为是ini类型
		if !strings.Contains(str, ";") && !strings.Contains(str, ":") {
			apolloLocalFile := &ApolloLocalFile{}
			if p, err := filepath.Abs(str); err == nil {
				apolloLocalFile.FilePath = p
				apolloLocalFile.FileType = F_INI
				apolloLocalFiles = append(apolloLocalFiles, apolloLocalFile)
			}
			return apolloLocalFiles
		}
		if saveFiles := strings.Split(str, ";"); len(saveFiles) >= 0 {
			for _, item := range saveFiles {
				if item != "" && strings.Contains(item, ":") {
					if files := strings.Split(item, ":"); len(files) == 2 {
						if files[0] == "" || files[1] == "" {
							continue
						}
						apolloLocalFile := &ApolloLocalFile{}

						switch strings.ToLower(files[0]) {
						case "ini":
							apolloLocalFile.FileType = F_INI
							break
						case "fastcgi", "nginx":
							apolloLocalFile.FileType = F_FASTCGI
							break
						case "yml":
							apolloLocalFile.FileType = F_YML
							break
						case "yaml":
							apolloLocalFile.FileType = F_YAML
						case "json":
							apolloLocalFile.FileType = F_JSON
							break
						default:
							apolloLocalFile.FileType = F_INI
						}
						if p, err := filepath.Abs(files[1]); err == nil {
							apolloLocalFile.FilePath = p
							apolloLocalFiles = append(apolloLocalFiles, apolloLocalFile)
						}
					}
				}
			}
			return apolloLocalFiles
		}
	}
	return nil
}

func (f *ApolloLocalFile) String() string {
	return fmt.Sprintf("{FileType='%s', FilePath='%s'}", string(f.FileType), f.FilePath)
}

//关联的Namespace.
type ApolloResult struct {
	ReleaseKey string `json:"release_key"`
	//应用的appId
	AppId string `json:"appId"`
	// 集群名称，一般情况下传入 default 即可。
	// 如果希望配置按集群划分，可以参考集群独立配置说明做相关配置，然后在这里填入对应的集群名.
	Cluster string `json:"cluster"`
	// 命名空间，如果没有新建过Namespace的话，传入application即可。
	// 如果创建了Namespace，并且需要使用该Namespace的配置，则传入对应的Namespace名字。
	// 需要注意的是对于properties类型的namespace，只需要传入namespace的名字即可，如application。
	// 对于其它类型的namespace，需要传入namespace的名字加上后缀名，如datasources.json
	NamespaceName string `json:"namespaceName"`
	//配置信息.
	Configurations map[string]string `json:"configurations"`
}

func NewApolloResult(appId, cluster, namespace string) *ApolloResult {
	config := &ApolloResult{
		AppId:          appId,
		Cluster:        cluster,
		NamespaceName:  namespace,
		Configurations: make(map[string]string, 0),
		ReleaseKey:     "",
	}

	return config
}

type ApolloAppConfig struct {
	AppId          string   `json:"appId,omitempty"`
	Cluster        string   `json:"cluster,omitempty"`
	NamespaceName  string   `json:"namespaceName,omitempty"`  //主命名空间，
	NamespaceNames []string `json:"namespaceNames,omitempty"` //辅命名空间
}

func NewApolloAppConfig(appId, cluster, namespace string) *ApolloAppConfig {
	return &ApolloAppConfig{
		AppId:          appId,
		Cluster:        cluster,
		NamespaceName:  namespace,
		NamespaceNames: make([]string, 0),
	}
}

type apolloClient struct {
	config *ApolloAppConfig
	//Apollo配置服务的地址.
	serverUrl      string
	releaseMap     map[string]string
	cacheRequester requester
	dataRequester  requester
}

func NewApolloClient(config *ApolloAppConfig, serverUrl string) *apolloClient {
	if !strings.HasSuffix(serverUrl, "/") {
		serverUrl += "/"
	}
	return &apolloClient{
		config:         config,
		serverUrl:      serverUrl,
		releaseMap:     make(map[string]string),
		cacheRequester: newHTTPRequester(&http.Client{Timeout: queryTimeout}),
		dataRequester:  newHTTPRequester(&http.Client{Timeout: queryTimeout}),
	}
}

// 从远程服务器提供的接口获取配置信息.
func (c *apolloClient) GetApolloRemoteConfigFromCache() (*ConfigEntity, error) {

	configs := make(map[string]string)

	//只有键值结构才支持多命名空间.
	if getConfigType(c.config.NamespaceName) == C_TYPE_POROPERTIES && c.config.NamespaceNames != nil && len(c.config.NamespaceNames) > 0 {
		for _, ns := range c.config.NamespaceNames {
			kv, err := c.fetchFromCache(ns);
			if err != nil {
				log.Errorf("拉取配置信息时出错 ->", err)
				return nil, err
			}
			for k, v := range kv {
				configs[k] = v
			}
		}
	}
	//最后拉取主命名空间，用来覆盖关联的配置信息.
	kv, err := c.fetchFromCache(c.config.NamespaceName)
	if err != nil {
		log.Errorf("拉取配置信息时出错 ->", err)
		return nil, err
	}
	for k, v := range kv {
		configs[k] = v
	}

	entity := NewConfigEntity()
	entity.ConfigType = getConfigType(c.config.NamespaceName)
	entity.NamespaceName = c.config.NamespaceName
	if entity.ConfigType == C_TYPE_POROPERTIES {
		entity.Values = configs
	} else {
		entity.Values = configs["content"]
	}

	return entity, err
}

func (c *apolloClient) fetchFromCache(namespace string) (KeyValuePair, error) {
	serverUrl := getApolloRemoteConfigFromCacheUrl(c.serverUrl, c.config.AppId, c.config.Cluster, namespace)

	body, err := c.cacheRequester.request(serverUrl)

	if err != nil {
		if err == ErrConfigUnmodified {
			return nil, err
		}
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
func (c *apolloClient) GetApolloRemoteConfigFromDatabase() (*ConfigEntity, error) {
	configs := make(map[string]string)

	//只有键值结构才支持多命名空间.
	if getConfigType(c.config.NamespaceName) == C_TYPE_POROPERTIES && c.config.NamespaceNames != nil && len(c.config.NamespaceNames) > 0 {
		for _, ns := range c.config.NamespaceNames {
			ret, err := c.fetchFromDatabase(ns);
			if err != nil {
				log.Errorf("拉取配置信息时出错 ->", err)
				return nil, err
			}
			for k, v := range ret.Configurations {
				configs[k] = v
			}
		}
	}
	//最后拉取主命名空间，用来覆盖关联的配置信息.
	ret, err := c.fetchFromDatabase(c.config.NamespaceName)
	if err != nil {
		log.Errorf("拉取配置信息时出错 ->", err)
		return nil, err
	}
	for k, v := range ret.Configurations {
		configs[k] = v
	}
	entity := NewConfigEntity()
	entity.ConfigType = getConfigType(c.config.NamespaceName)
	entity.NamespaceName = c.config.NamespaceName

	if entity.ConfigType == C_TYPE_POROPERTIES {
		entity.Values = configs
	} else {
		entity.Values = configs["content"]
	}

	return entity, err
}

func (c *apolloClient) fetchFromDatabase(namespace string) (*ApolloResult, error) {

	serverUrl := getApolloRemoteConfigFromDbUrl(c.serverUrl, c.config.AppId, c.config.Cluster, namespace, c.getReleaseKey(namespace))

	body, err := c.dataRequester.request(serverUrl)

	if err != nil {
		if err == ErrConfigUnmodified {
			return nil, ErrConfigUnmodified
		}
		log.Errorf("Read response body error:%s %s", serverUrl, err)
		return nil, err
	}
	log.Infof("Remote response body -> %s", string(body))

	config := ApolloResult{}

	//解码后，如果解码成功，则将新值复制给原对象
	err = json.Unmarshal(body, &config)
	if err != nil {
		log.Errorf("Unmarshal json result error:%s %s %s", serverUrl, string(body), err)
	} else {
		c.releaseMap[namespace] = config.ReleaseKey
	}

	return &config, err
}

func (c *apolloClient) getReleaseKey(namespace string) string {
	if k, ok := c.releaseMap[namespace]; ok {
		return k
	}
	return ""
}

func (c *apolloClient) getNamespaceNames() []string {
	ns := []string{c.config.NamespaceName}
	if c.config.NamespaceNames != nil && len(c.config.NamespaceNames) > 0 {
		for _, s := range c.config.NamespaceNames {
			ns = append(ns, s)
		}
	}
	return ns
}

func (c *apolloClient) getLongPoll(longPollInterval time.Duration, notificationChan chan<- *ApolloNotificationMessages) *longPoll {

	return NewPoll(c.serverUrl, c.config.AppId, c.config.Cluster, c.getNamespaceNames(), longPollInterval, notificationChan)
}
