package goapollo

import "fmt"

const (
	C_UNCHANGED PropertyChangeType = iota
	C_ADDED
	C_MODIFIED
	C_DELETED
)

type PropertyChangeType int

type KeyValuePair map[string]string

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

//
type ConfigChangeEventArgs struct {
	NamespaceName string
	Values        map[string]*ConfigChangeEntry
}

func (args *ConfigChangeEventArgs) String() string {
	return fmt.Sprintf("ConfigChangeEventArgs(NamespaceName:%s,Values:%v)",
		args.NamespaceName,
		args.Values, )
}
