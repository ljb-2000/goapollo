package goapollo

import (
	"errors"
	"sync"
	"strings"
)

var (
	handlerCache = map[string] FuncHandler{}
	mux          = &sync.RWMutex{}
	ErrHandlerExist = errors.New("handler already exists. ")
)

type FuncHandler func(p string)(Handler,error)

type Handler interface {
	Handler(changeEvent *ConfigEntity) (error)
}

//增加自定义的处理器.
func AddCustomHandler(name string, handler FuncHandler) error {
	if _, ok := handlerCache[name]; ok {
		return ErrHandlerExist
	}
	mux.Lock()
	handlerCache[name] = handler
	mux.Unlock()

	return nil
}

func ExecuteHandler(changeEvent *ConfigEntity, savePath string) (error) {
	mux.RLock()
	defer mux.RUnlock()
	var handler Handler
	var err error

	//如果是 Nginx 后缀的则当做Nginx配置文件
	if strings.HasSuffix(changeEvent.NamespaceName, ".nginx") {
		if nginx, ok := handlerCache["nginx"]; ok {
			handler,err = nginx(savePath)
			if err != nil {
				return err
			}
		}
	} else if strings.HasSuffix(changeEvent.NamespaceName, ".ini") || changeEvent.ConfigType == C_TYPE_POROPERTIES {
		if ini,ok := handlerCache["ini"]; ok {
			handler,err = ini(savePath)
			if err != nil {
				return err
			}
		}
	} else {
		if f,ok := handlerCache["default"]; ok {
			handler,err = f(savePath)
			if err != nil {
				return err
			}
		}
	}

	if handler != nil {
		return handler.Handler(changeEvent)
	}
	return errors.New("Not found handler -> " + changeEvent.NamespaceName)
}

func init() {
	handlerCache["nginx"] = func(p string) (Handler, error) {

		return NewNginxHandler(p)
	}
	handlerCache["ini"] = func(p string) (Handler, error) {

		return NewIniHandler(p)
	}
	handlerCache["default"] = func(p string) (Handler, error) {

		return NewFileHandler(p)
	}
}
