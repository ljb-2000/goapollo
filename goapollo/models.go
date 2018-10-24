package goapollo

const (
	C_ADDED PropertyChangeType=iota
	C_MODIFIED
	C_DELETED
)

type PropertyChangeType int

type KeyValuePair map[string]string

// 配置变化的实体.
type ConfigChangeEntry struct {
	NamespaceName string
	PropertyName string
	OldValue string
	NewValue string
	ChangeType PropertyChangeType
}

//
type ConfigChangeEventArgs struct {
	NamespaceName string
	Values map[string]*ConfigChangeEntry
}