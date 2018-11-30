package goapollo

import (
	log "github.com/sirupsen/logrus"
	"github.com/lifei6671/goini"
	"strings"
	"path/filepath"
	"os"
)

type IniHandler struct {
	body     *goini.IniContainer
	savePath string
}

func NewIniHandler(savePath string) (*IniHandler, error) {

	base := filepath.Dir(savePath)
	if _,err := os.Stat(base); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(base,0755); err != nil {
			log.Errorf("创建目录失败 -> %s %s",base, err)
		}
	}

	h := &IniHandler{
		savePath: savePath,
	}

	err := h.Open(h.savePath)

	return h, err
}

func (h *IniHandler) Handler(changeEvent *ConfigEntity) (error) {
	if changeEvent.ConfigType == C_TYPE_POROPERTIES {

		for k, v := range changeEvent.GetValues() {
			var keys []string
			key := strings.TrimSpace(k)
			if strings.Index(key, ".") > 0 {
				ks := strings.Split(key, ".")
				keys = []string{ks[0], strings.Join(keys[1:], ".")}
			} else {
				keys = []string{"", key}
			}
			h.body.AddEntry(keys[0], keys[1], v)
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

	h.body = goini.NewConfig()

	return nil
}
