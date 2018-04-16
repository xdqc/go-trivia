package solver

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/henson/Answer"
	"github.com/henson/Answer/ocr"
	"github.com/henson/Answer/util"
)

func getQuizFromOCR() (quiz string, options []string) {
	log.Println("Fetching quiz and options from screenshot OCR ...")
	tx1 := time.Now()

	cfg := util.GetConfig()
	OCR := ocr.NewBaidu(cfg)

	imgChan1 := make(chan string, 1)
	imgChan2 := make(chan string, 1)

	var wig sync.WaitGroup
	wig.Add(3)
	go func() {
		defer wig.Done()
		//qText, err := tesseractOCR().GetText(util.QuestionImage)
		quizText, err := OCR.GetText(<-imgChan1)
		if err != nil {
			log.Panicf("识别题目失败，%v\n", err.Error())
			return
		}
		quiz = processQuiz(quizText)
	}()
	go func() {
		defer wig.Done()
		//answerText, err := baiduOCR().GetText(util.AnswerImage)
		optionsText, err := OCR.GetText(<-imgChan2)
		if err != nil {
			log.Panicf("识别答案失败，%v\n", err.Error())
			return
		}
		options = processOptions(optionsText)
	}()
	go func() {
		screenshot := Answer.NewScreenshot(cfg)
		png, err := screenshot.GetImage()
		if err != nil {
			log.Panicf("获取截图失败，%v\n", err.Error())
			return
		}
		err = Answer.SaveImage(png, cfg, imgChan1, imgChan2)
		if err != nil {
			log.Panicf("保存图片失败，%v\n", err.Error())
			return
		}
	}()
	wig.Wait()

	log.Printf("OCR time: %d ms\n", time.Now().Sub(tx1).Nanoseconds()/1e6)
	return
}

func processQuiz(text string) string {
	return strings.TrimSpace(text)
}

func processOptions(text string) []string {
	arr := strings.Split(text, "\n")
	textArr := []string{}
	for _, val := range arr {
		if strings.TrimSpace(val) == "" {
			continue
		}
		textArr = append(textArr, strings.TrimSpace(val))
	}
	return textArr
}
