package goapollo

import (
	"errors"
	"sync"
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

func ExecuteHandler(changeEvent *ConfigEntity, file *ApolloLocalFile) (error) {
	mux.RLock()
	defer mux.RUnlock()
	var handler Handler
	var err error

	if fn, ok := handlerCache[string(file.FileType)]; ok {
		handler,err = fn(file.FilePath)
		if err != nil {
			return err
		}
	}  else {
		if fn,ok := handlerCache["default"]; ok {
			handler,err = fn(file.FilePath)
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
	handlerCache["fastcgi"] = func(p string) (Handler, error) {

		return NewNginxHandler(p)
	}
	handlerCache["ini"] = func(p string) (Handler, error) {

		return NewIniHandler(p)
	}
	handlerCache["default"] = func(p string) (Handler, error) {

		return NewFileHandler(p)
	}
}
