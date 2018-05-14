package device

import (
	"image"
)

//Screenshot 获取屏幕截图
type Screenshot interface {
	GetImage() (image.Image, error)
}

//NewScreenshot 根据手机系统区分
func NewScreenshot(cfg *Config) Screenshot {
	if cfg.Device == DeviceiOS {
		return NewIOS(cfg)
	}
	return NewAndroid(cfg)
}
