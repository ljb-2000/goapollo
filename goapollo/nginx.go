package goapollo

import (
	"os"
	log "github.com/sirupsen/logrus"
	"strings"
	"fmt"
	"sync"
)

type NginxHandler struct {
	body     map[string]string
	savePath string
	mux      *sync.RWMutex
}

func NewNginxHandler(savePath string) (*NginxHandler, error) {

	h:= &NginxHandler{
		savePath: savePath,
		body:     make(map[string]string),
		mux:      &sync.RWMutex{},
	}

	return h, nil
}

//Nginx fastcgi 处理器.
// changeEvent Apollo 变动通知.
// savePath 将变动保存到的文件路径.
func (h *NginxHandler) Handler(changeEvent *ConfigEntity) (error) {
	if changeEvent.ConfigType == C_TYPE_POROPERTIES  {

		h.mux.Lock()
		for k, v := range changeEvent.GetValues() {
			h.body[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
		h.mux.Unlock()
		return h.save()

	}
	return nil
}


//保存为 Nginx fastcgi 配置文件.
func (h *NginxHandler) save() error {

	f, err := os.OpenFile(h.savePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Errorf("打开配置文件时出错 -> %s %s", h.savePath, err)
		return err
	}
	defer f.Close()
	h.mux.RLock()
	defer h.mux.RUnlock()

	for k, v := range h.body {
		if !strings.HasSuffix(v, ";") {
			v += ";"
		}
		if !strings.HasPrefix(k, "fastcgi_param") {
			k = "fastcgi_param\t" + k
		}
		_, err := f.WriteString(fmt.Sprintf("%s\t%s\n", k, v))
		if err != nil {
			log.Errorf("保存文件时出错 -> %s %s", h.savePath, err)
			return err
		}
	}
	log.Infof("变更已保存 -> %s", h.savePath)
	return nil
}
