package device

//Screenshot 获取屏幕截图
type Ocr interface {
	GetText(imgPath string) (string, error)
}

//NewScreenshot 根据手机系统区分
func NewOcr(cfg *Config) Ocr {
	if cfg.OCR == "baidu" {
		return NewBaidu(cfg)
	}
	return NewTesseract()
}
