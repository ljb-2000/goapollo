package goapollo

import (
	"strings"
	"path/filepath"
)

//判断是否是文件格式
func IsConfigFile(namespace string) bool {
	return strings.HasSuffix(namespace, ".json") ||
		strings.HasSuffix(namespace, ".yml") ||
		strings.HasSuffix(namespace, ".yaml") ||
		strings.HasSuffix(namespace, ".xml")
}

//获取命名空间的类型.
func GetConfigType(namespace string) ConfigType {
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
