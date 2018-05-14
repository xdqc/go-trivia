package solver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/xdqc/letterpress-solver/device"
)

func getQuizFromOCR() (quiz string, options []string) {
	log.Println("Fetching quiz and options from screenshot OCR ...")
	tx1 := time.Now()

	cfg := device.GetConfig()
	// OCR := Answer.NewOcr(cfg)

	imgQuiz := make(chan string, 1)
	imgOptions := make(chan string, 1)

	var wig sync.WaitGroup
	wig.Add(2)

	go func() {
		defer wig.Done()
		// quizText, err := OCR.GetText(<-imgQuiz)
		buf := new(bytes.Buffer)
		err := detectText(buf, <-imgQuiz)
		quizText := buf.String()

		if err != nil {
			log.Println(err.Error())
			return
		}
		quiz = processQuiz(quizText)
	}()
	go func() {
		defer wig.Done()
		// optionsText, err := OCR.GetText(<-imgOptions)
		buf := new(bytes.Buffer)
		err := detectText(buf, <-imgOptions)
		optionsText := buf.String()

		if err != nil {
			log.Println(err.Error())
			return
		}
		options = processOptions(optionsText)
	}()
	go func() {
		screenshot := device.NewScreenshot(cfg)
		png, err := screenshot.GetImage()
		if err != nil {
			log.Println(err.Error())
			imgQuiz <- device.QuestionImage
			imgOptions <- device.AnswerImage
			return
		}
		err = device.SaveImage(png, cfg, imgQuiz, imgOptions)
		if err != nil {
			log.Println(err.Error())
			imgQuiz <- device.QuestionImage
			imgOptions <- device.AnswerImage
			return
		}
	}()
	wig.Wait()

	log.Printf("OCR time: %d ms\n", time.Now().Sub(tx1).Nanoseconds()/1e6)
	return
}

func processQuiz(text string) string {
	text = strings.Replace(text, " ", "", -1)
	text = strings.Replace(text, "\"", "", -1)
	text = strings.Replace(text, "\\n", "", -1)
	text = strings.Replace(text, "\n", "", -1)
	return text
}

func processOptions(text string) []string {
	text = strings.Replace(text, "\"", "", -1)
	arr := strings.Split(text, "\\n")
	textArr := []string{}
	for _, val := range arr {
		if strings.TrimSpace(val) == "" {
			continue
		}
		textArr = append(textArr, strings.TrimSpace(val))
	}
	return textArr
}

func detectText(w io.Writer, file string) error {
	ctx := context.Background()

	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return err
	}

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	image, err := vision.NewImageFromReader(f)
	if err != nil {
		return err
	}
	annotations, err := client.DetectTexts(ctx, image, nil, 10)
	if err != nil {
		return err
	}

	if len(annotations) == 0 {
		fmt.Fprintln(w, "")
	} else {
		// fmt.Fprintln(w)
		// for _, annotation := range annotations {
		// 	fmt.Fprintf(w, "%q ", annotation.Description)
		// }
		fmt.Fprintf(w, "%q ", annotations[0].Description)
	}

	return nil
}
