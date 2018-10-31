package goapollo

import (
	log "github.com/sirupsen/logrus"
	"os"
)

type FileHandler struct {
	savePath string
}


func NewFileHandler(savePath string) (*FileHandler, error) {

	h := &FileHandler{
		savePath: savePath,
	}


	return h, nil
}

func (h *FileHandler) Handler(changeEvent *ConfigEntity) (error) {
	if changeEvent.ConfigType != C_TYPE_POROPERTIES {

		err := h.save(changeEvent.GetConfigFile())

		if err != nil {
			log.Errorf("保存配置失败 -> %s %s", h.savePath, err)
		}
		return err
	}
	return nil
}

//保存为配置文件.
func (h *FileHandler) save(body string) error {

	f, err := os.OpenFile(h.savePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Errorf("打开配置文件时出错 -> %s %s", h.savePath, err)
		return err
	}
	defer f.Close()
	_,err = f.WriteString(body)

	return err
}
