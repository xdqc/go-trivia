package solver

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/henson/Answer"
	"github.com/henson/Answer/util"
)

func getQuizFromOCR() (quiz string, options []string) {
	log.Println("Fetching quiz and options from screenshot OCR ...")
	tx1 := time.Now()

	cfg := util.GetConfig()
	OCR := Answer.NewOcr(cfg)

	imgQuiz := make(chan string, 1)
	imgOptions := make(chan string, 1)

	var wig sync.WaitGroup
	wig.Add(2)

	go func() {
		defer wig.Done()
		quizText, err := OCR.GetText(<-imgQuiz)
		if err != nil {
			log.Println(err.Error())
			return
		}
		quiz = processQuiz(quizText)
	}()
	go func() {
		defer wig.Done()
		optionsText, err := OCR.GetText(<-imgOptions)
		if err != nil {
			log.Println(err.Error())
			return
		}
		options = processOptions(optionsText)
	}()
	go func() {
		screenshot := Answer.NewScreenshot(cfg)
		png, err := screenshot.GetImage()
		if err != nil {
			log.Println(err.Error())
			imgQuiz <- util.QuestionImage
			imgOptions <- util.AnswerImage
			return
		}
		err = Answer.SaveImage(png, cfg, imgQuiz, imgOptions)
		if err != nil {
			log.Println(err.Error())
			imgQuiz <- util.QuestionImage
			imgOptions <- util.AnswerImage
			return
		}
	}()
	wig.Wait()

	log.Printf("OCR time: %d ms\n", time.Now().Sub(tx1).Nanoseconds()/1e6)
	return
}

func processQuiz(text string) string {
	text = strings.Replace(text, " ", "", -1)
	text = strings.Replace(text, "\n", "", -1)
	return text
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
