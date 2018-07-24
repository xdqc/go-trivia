package solver

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/xdqc/letterpress-solver/device"
)

func getQuizFromOCR() (quiz string, options []string) {
	// log.Println("Fetching quiz and options from screenshot OCR ...")
	// tx1 := time.Now()

	cfg := device.GetConfig()
	OCR := device.NewOcr(cfg)

	imgQuiz := make(chan string, 1)
	imgOptions1 := make(chan string, 1)
	imgOptions2 := make(chan string, 1)
	imgOptions3 := make(chan string, 1)
	imgOptions4 := make(chan string, 1)
	imgOptions := make(chan string, 1)

	var option1, option2, option3, option4 string

	var wig sync.WaitGroup
	if Hashquiz == 1 {
		wig.Add(5)
	} else {
		wig.Add(2)
	}

	go func() {
		defer wig.Done()
		quizText, err := OCR.GetText(<-imgQuiz)

		// for google api
		// buf := new(bytes.Buffer)
		// err := detectText(buf, <-imgQuiz)
		// quizText := buf.String()

		if err != nil {
			log.Println(err.Error())
			return
		}
		quiz = processQuiz(quizText)
	}()
	if Hashquiz == 1 {
		go func() {
			defer wig.Done()
			options1Text, err := OCR.GetText(<-imgOptions1)

			if err != nil {
				log.Println(err.Error())
				return
			}
			option1 = processQuiz(options1Text)
		}()
		go func() {
			defer wig.Done()
			options2Text, err := OCR.GetText(<-imgOptions2)

			if err != nil {
				log.Println(err.Error())
				return
			}
			option2 = processQuiz(options2Text)
		}()
		go func() {
			defer wig.Done()
			options3Text, err := OCR.GetText(<-imgOptions3)

			if err != nil {
				log.Println(err.Error())
				return
			}
			option3 = processQuiz(options3Text)
		}()
		go func() {
			defer wig.Done()
			options4Text, err := OCR.GetText(<-imgOptions4)

			if err != nil {
				log.Println(err.Error())
				return
			}
			option4 = processQuiz(options4Text)
		}()
	} else {
		go func() {
			defer wig.Done()
			optionsText, err := OCR.GetText(<-imgOptions)

			if err != nil {
				log.Println(err.Error())
				return
			}
			options = processOptions(optionsText)
		}()
	}
	go func() {
		if Hashquiz == 1 {
			imgQuiz <- device.QuestionImage
			imgOptions1 <- device.Answer1Image
			imgOptions2 <- device.Answer2Image
			imgOptions3 <- device.Answer3Image
			imgOptions4 <- device.Answer4Image
		} else {
			screenshot := device.NewScreenshot(cfg)
			png, err := screenshot.GetImage()
			// log.Printf("Image get time: %d ms\n", time.Now().Sub(tx1).Nanoseconds()/1e6)
			if err != nil {
				log.Println(err.Error())
				imgQuiz <- device.QuestionImage
				imgOptions <- device.AnswerImage
				return
			}
			err = device.SaveImage(png, cfg.APP, imgQuiz, imgOptions)
			if err != nil {
				log.Println(err.Error())
				imgQuiz <- device.QuestionImage
				imgOptions <- device.AnswerImage
				return
			}
			return
			// log.Printf("Image get+save time: %d ms\n", time.Now().Sub(tx1).Nanoseconds()/1e6)
		}
	}()
	wig.Wait()

	if Hashquiz == 1 {
		options = make([]string, 0)
		options = append(options, option1, option2, option3, option4)
	}

	// log.Printf("OCR time: %d ms\n", time.Now().Sub(tx1).Nanoseconds()/1e6)
	return
}

func processQuiz(text string) string {
	// text = strings.Replace(text, " ", "", -1)
	text = strings.Replace(text, "\"", "", -1)
	text = strings.Replace(text, "\\n", "", -1)
	text = strings.Replace(text, "\n", "", -1)
	return text
}

func processOptions(text string) []string {
	// println(text)
	text = strings.Replace(text, "\"", "", -1)
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
