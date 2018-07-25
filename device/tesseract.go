package device

import (
	"os/exec"
	"strings"
)

//Tesseract tesseract 识别
type Tesseract struct{}

//NewTesseract new
func NewTesseract() *Tesseract {
	return new(Tesseract)
}

//GetText 根据图片路径获取识别文字
func (tesseract *Tesseract) GetText(imgPath string) (string, error) {
	// log.Println("tesseract OCR start ...")
	// tx1 := time.Now()
	body, err := exec.Command("tesseract", imgPath, "stdout", "-l", "chi_sim").Output()
	if err != nil {
		return "", err
	}
	text := strings.Replace(string(body), " ", "", -1)
	text = strings.Replace(text, "_", "\n", -1)
	text = strings.Replace(text, "J", "」", -1)
	text = strings.Replace(text, "′」\\", "小", -1)
	text = strings.Replace(text, "′」、", "小", -1)
	text = strings.Replace(text, "咽〖", "哪", -1)
	text = strings.Replace(text, "夕卜", "外", -1)
	text = strings.Replace(text, "1十", "什", -1)
	text = strings.Replace(text, "1立", "位", -1)
	text = strings.Replace(text, "1尔", "你", -1)
	text = strings.Replace(text, "1门", "们", -1)
	text = strings.Replace(text, "{氏", "低", -1)
	text = strings.Replace(text, "带‖", "制", -1)
	text = strings.Replace(text, "i寺", "诗", -1)
	text = strings.Replace(text, "才匕", "北", -1)
	text = strings.Replace(text, "届于", "属于", -1)

	// log.Printf("tesseract time: %d ms\n", time.Now().Sub(tx1).Nanoseconds()/1e6)

	return text, nil
}
