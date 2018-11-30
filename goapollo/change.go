package goapollo

import (
	"fmt"
	"encoding/json"
)

type KeyValuePair map[string]string

//配置实体.
type ConfigEntity struct {
	NamespaceName string      `json:"namespaceName"`
	ConfigType    ConfigType  `json:"configType"`
	Values        interface{} `json:"values"`
}

func NewConfigEntity() *ConfigEntity {
	return &ConfigEntity{}
}
func (c *ConfigEntity) GetValues() (map[string]string) {
	if c.ConfigType != C_TYPE_POROPERTIES {
		return nil
	}
	if kv, ok := c.Values.(map[string]string); ok {
		return kv
	}
	return nil
}

func (c *ConfigEntity) GetConfigFile() string {
	if c.ConfigType == C_TYPE_POROPERTIES {
		return ""
	}
	if kv, ok := c.Values.(string); ok {
		return kv
	}
	return ""
}

func (c *ConfigEntity) String() string {
	if b, err := json.Marshal(c); err == nil {
		return string(b)
	}
	return fmt.Sprintf("%v", *c)
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
	Values        map[string]*ConfigChangeEntry
	FileContent   string
}

func NewConfigChangeEventArgs(namespace string, configType ConfigType) *ConfigChangeEventArgs {
	return &ConfigChangeEventArgs{NamespaceName: namespace, ConfigType: configType, Values: make(map[string]*ConfigChangeEntry), FileContent: ""}
}

func (c *ConfigChangeEventArgs) String() string {
	return fmt.Sprintf("ConfigChangeEventArgs(NamespaceName:%s,Values:%v)",
		c.NamespaceName,
		c.Values, )
}
