package goapollo

import (
	"os"
	log "github.com/sirupsen/logrus"
	"bufio"
	"io"
	"strings"
	"fmt"
	"sync"
	"path/filepath"
)

type NginxHandler struct {
	body     map[string]string
	savePath string
	mux      *sync.RWMutex
}

func NewNginxHandler(savePath string) (*NginxHandler, error) {
	p, err := filepath.Abs(savePath)

	if err != nil {
		log.Error("解析配置文件路径失败 -> ", savePath, err)
		return nil, err
	}
	h:= &NginxHandler{
		savePath: p,
		body:     make(map[string]string),
		mux:      &sync.RWMutex{},
	}

	err = h.Open(h.savePath)

	return h, err
}

//Nginx fastcgi 处理器.
// changeEvent Apollo 变动通知.
// savePath 将变动保存到的文件路径.
func (h *NginxHandler) Handler(changeEvent *ConfigChangeEventArgs) (error) {
	if len(changeEvent.Values) > 0 {

		h.mux.Lock()
		for k, v := range changeEvent.Values {
			log.Infof("Nginx 变更信息 -> Key=%s NewValue=%s $ OldValue=%s $ ChangeType=%d", k, v.NewValue, v.OldValue, v.ChangeType)
			if v.ChangeType == C_ADDED || v.ChangeType == C_MODIFIED {
				h.body[strings.TrimSpace(k)] = strings.TrimSpace(v.NewValue)
			} else if v.ChangeType == C_DELETED {
				delete(h.body, strings.TrimSpace(k))
			}
		}
		h.mux.Unlock()
		return h.save()

	}
	return nil
}

//解析 Nginx 配置文件
func (h *NginxHandler) Open(path string) error {
	p, err := filepath.Abs(path)

	if err != nil {
		log.Errorf("获取配置文件路径时出错 -> %s %s", path, err)
		return err
	}
	h.savePath = p

	f, err := os.OpenFile(h.savePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Errorf("打开配置文件时出错 -> %s %s", path, err)
		return err
	}
	defer f.Close()

	if h.body == nil {
		h.body = make(map[string]string)
	}
	rd := bufio.NewReader(f)

	for {
		line, err := rd.ReadString('\n') //以'\n'为结束符读入一行

		if err != nil || io.EOF == err {
			if io.EOF != err {
				log.Errorf("解析配置文件时出错 -> %s %s", path, err)
			}
			break
		}
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "fastcgi_param"))

		kv := strings.Split(line, " ")

		if len(kv) == 2 {
			h.body[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
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
