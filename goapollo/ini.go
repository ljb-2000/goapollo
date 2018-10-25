package goapollo

import (
	"path/filepath"
	log "github.com/sirupsen/logrus"
	"github.com/lifei6671/goini"
	"strings"
)

type IniHandler struct {
	body     *goini.IniContainer
	savePath string
}

func NewIniHandler(savePath string) (*IniHandler, error) {
	p, err := filepath.Abs(savePath)

	if err != nil {
		log.Error("解析配置文件路径失败 -> ", savePath, err)
		return nil, err
	}

	h := &IniHandler{
		savePath: p,
	}

	err = h.Open(h.savePath)

	return h, err
}

func (h *IniHandler) Handler(changeEvent *ConfigChangeEventArgs) (error) {
	if len(changeEvent.Values) > 0 {

		for k, v := range changeEvent.Values {
			log.Infof("Ini 变更信息 -> Key=%s NewValue=%s $ OldValue=%s $ ChangeType=%d", k, v.NewValue, v.OldValue, v.ChangeType)
			var keys []string
			key := strings.TrimSpace(k)
			if strings.Index(key, ".") > 0 {
				ks := strings.Split(key, ".")
				keys = []string{ks[0], strings.Join(keys[1:], ".")}
			} else {
				keys = []string{"", key}
			}

			if v.ChangeType == C_ADDED || v.ChangeType == C_MODIFIED {
				h.body.AddEntry(keys[0], keys[1], v.NewValue)

			} else if v.ChangeType == C_DELETED {
				h.body.DeleteKey(keys[0], keys[1])
			}
		}
		err := h.body.SaveFile(h.savePath)

		if err != nil {
			log.Errorf("保存配置失败 -> %s %s", h.savePath, err)
		}
		return err
	}
	return nil
}

//打开配置文件并解析.
func (h *IniHandler) Open(path string) error {

	ini, err := goini.LoadFromFile(h.savePath)

	if err != nil {
		log.Errorf("初始化 ini 配置失败 -> %s %s", h.savePath, err)
		ini = goini.NewConfig()
	}
	h.body = ini

	return nil
}
