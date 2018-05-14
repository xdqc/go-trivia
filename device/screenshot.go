package device

import (
	"image"

	"github.com/henson/Answer/util"
	solver "github.com/xdqc/letterpress-solver"
	"github.com/xdqc/letterpress-solver/device"
)

//Screenshot 获取屏幕截图
type Screenshot interface {
	GetImage() (image.Image, error)
}

//NewScreenshot 根据手机系统区分
func NewScreenshot(cfg *solver.Config) Screenshot {
	if cfg.Device == util.DeviceiOS {
		return device.NewIOS(cfg)
	}
	return device.NewAndroid(cfg)
}
