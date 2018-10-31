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

//配置实体.
type ConfigEntity struct {
	NamespaceName string
	ConfigType    ConfigType
	Values interface{}
}

func NewConfigEntity() *ConfigEntity  {
	return &ConfigEntity{}
}
func (c *ConfigEntity) GetValues() (map[string]string) {
	if c.ConfigType != C_TYPE_POROPERTIES {
		return nil
	}
	if kv,ok := c.Values.(map[string]string); ok {
		return kv
	}
	return nil
}

func (c *ConfigEntity) GetConfigFile() string  {
	if c.ConfigType == C_TYPE_POROPERTIES {
		return ""
	}
	if kv,ok := c.Values.(string); ok {
		return kv
	}
	return ""
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
	NamespaceName string
	ConfigType    ConfigType
	Values map[string]*ConfigChangeEntry
	FileContent string
}

func NewConfigChangeEventArgs(namespace string, configType ConfigType) *ConfigChangeEventArgs {
	return &ConfigChangeEventArgs{  NamespaceName:namespace, ConfigType:configType,Values: make(map[string]*ConfigChangeEntry) ,FileContent:""}
}

func (c *ConfigChangeEventArgs) String() string {
	return fmt.Sprintf("ConfigChangeEventArgs(NamespaceName:%s,Values:%v)",
		c.NamespaceName,
		c.Values, )
}
