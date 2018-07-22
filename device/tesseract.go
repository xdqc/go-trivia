package device

import (
	"log"
	"os/exec"
	"strings"
	"time"
)

//Tesseract tesseract 识别
type Tesseract struct{}

//NewTesseract new
func NewTesseract() *Tesseract {
	return new(Tesseract)
}

//GetText 根据图片路径获取识别文字
func (tesseract *Tesseract) GetText(imgPath string) (string, error) {
	log.Println("tesseract OCR start ...")
	tx1 := time.Now()
	body, err := exec.Command("tesseract", imgPath, "stdout", "-l", "chi_sim").Output()
	if err != nil {
		return "", err
	}
	text := strings.Replace(string(body), " ", "", -1)
	text = strings.Replace(string(body), "_", "\n", -1)

	log.Printf("tesseract time: %d ms\n", time.Now().Sub(tx1).Nanoseconds()/1e6)

	return text, nil
}
