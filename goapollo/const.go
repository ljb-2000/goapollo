package goapollo

import (
	"errors"
	"time"
)

const (
	defaultNotificationId = -1
	longPoolInterval      = time.Second * 2
	longPoolTimeout       = time.Second * 90
	queryTimeout          = time.Second * 2
	fullInterval          = time.Second * 60
)

var (
	ErrConfigUnmodified = errors.New("Apollo configuration not changed. ")
)

const (
	F_FASTCGI ApolloLocalFileType = "fastcgi"
	F_INI                         = "ini"
	F_JSON                        = "json"
	F_YML                         = "yml"
	F_YAML                        = "yaml"
)

type ApolloLocalFileType string

const (
	C_TYPE_POROPERTIES ConfigType = "poroperties"
	C_TYPE_XML                    = "xml"
	C_TYPE_YAML                   = "yaml"
	C_TYPE_YML                    = "yml"
	C_TYPE_JSON                   = "json"
)

//配置类型.
type ConfigType string

const (
	C_UNCHANGED PropertyChangeType = iota
	C_ADDED
	C_MODIFIED
	C_DELETED
)

//配置变更类型
type PropertyChangeType int
