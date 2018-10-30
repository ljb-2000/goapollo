package goapollo

import "fmt"

const (
	C_UNCHANGED PropertyChangeType = iota
	C_ADDED
	C_MODIFIED
	C_DELETED
)

const (
	C_TYPE_POROPERTIES ConfigType = iota
	C_TYPE_XML
	C_TYPE_YAML
	C_TYPE_YML
	C_TYPE_JSON
)

//配置类型.
type ConfigType int

//配置变更类型
type PropertyChangeType int

type KeyValuePair map[string]string

type ConfigArgs struct {
	NamespaceName string
	ConfigType    ConfigType
}

//文件类型变动参数.
type ConfigFileChangeEventArgs struct {
	ConfigArgs
	FileContent string
}

func NewConfigFileChangeEventArgs (namespace string, configType ConfigType) *ConfigFileChangeEventArgs {

	return &ConfigFileChangeEventArgs{ ConfigArgs{ NamespaceName:namespace, ConfigType:configType}, ""}
}

// 配置变化的实体.
type ConfigChangeEntry struct {
	NamespaceName string
	PropertyName  string
	OldValue      string
	NewValue      string
	ChangeType    PropertyChangeType
}

func (e *ConfigChangeEntry) String() string {
	return fmt.Sprintf("ConfigChangeEntry(NamespaceName:%s,PropertyName:%s,OldValue:%s,NewValue:%s,ChangeType:%d)",
		e.NamespaceName,
		e.PropertyName,
		e.OldValue,
		e.NewValue,
		e.ChangeType, )
}

//键值类型结构.
type ConfigChangeEventArgs struct {
	ConfigArgs
	Values map[string]*ConfigChangeEntry
}

func NewConfigChangeEventArgs(namespace string, configType ConfigType) *ConfigChangeEventArgs {

	return &ConfigChangeEventArgs{ ConfigArgs{ NamespaceName:namespace, ConfigType:configType}, make(map[string]*ConfigChangeEntry) }
}

func (c *ConfigChangeEventArgs) String() string {
	return fmt.Sprintf("ConfigChangeEventArgs(NamespaceName:%s,Values:%v)",
		c.NamespaceName,
		c.Values, )
}
