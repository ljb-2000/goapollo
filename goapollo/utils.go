package goapollo

import (
	"strings"
	"path/filepath"
	"net"
	"fmt"
	"net/url"
)

//判断是否是文件格式
func IsConfigFile(namespace string) bool {
	return strings.HasSuffix(namespace, ".json") ||
		strings.HasSuffix(namespace, ".yml") ||
		strings.HasSuffix(namespace, ".yaml") ||
		strings.HasSuffix(namespace, ".xml")
}

//获取命名空间的类型.
func getConfigType(namespace string) ConfigType {
	ext := filepath.Ext(namespace)
	switch ext {
	case ".json":
		return C_TYPE_JSON
	case ".yml":
		return C_TYPE_YML
	case ".yaml":
		return C_TYPE_YAML
	case ".xml":
		return C_TYPE_XML
	default:
		return C_TYPE_POROPERTIES
	}
}

//获取本地的ip地址.
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		if ip4 := toIP4(a); ip4 != nil {
			return ip4.String()
		}
	}
	return ""
}

//将网络地址转换为ip地址.
func toIP4(addr net.Addr) net.IP {
	if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
		return ipnet.IP.To4()
	}
	return nil
}

//通知地址.
func getApolloRemoteNotificationUrl(serverUrl, appId, cluster string, messages *ApolloNotificationMessages) string {

	notificationUrl := fmt.Sprintf("%snotifications/v2?appId=%s&cluster=%s&notifications=%s",
		serverUrl,
		url.QueryEscape(appId),
		url.QueryEscape(cluster),
		url.QueryEscape(messages.String(), ))

	return notificationUrl
}

// 获取URL
func getApolloRemoteConfigFromCacheUrl(serverUrl, appId, cluster, namespace string) string {

	getApolloRemoteConfigFromCacheUrl := fmt.Sprintf("%sconfigfiles/json/%s/%s/%s?ip=%s",
		serverUrl,
		url.QueryEscape(appId),
		url.QueryEscape(cluster),
		url.QueryEscape(namespace),
		getLocalIP(), )

	return getApolloRemoteConfigFromCacheUrl
}

// 生成通过不带缓存的Http接口从Apollo读取配置的链接
func getApolloRemoteConfigFromDbUrl(serverUrl, appId, cluster, namespace, releaseKey string) string {
	getApolloRemoteConfigFromDbUrl := fmt.Sprintf("%sconfigs/%s/%s/%s?releaseKey=%s&ip=%s",
		serverUrl,
		url.QueryEscape(appId),
		url.QueryEscape(cluster),
		url.QueryEscape(namespace),
		releaseKey,
		getLocalIP(),
	)

	return getApolloRemoteConfigFromDbUrl
}
