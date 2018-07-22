package solver

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/xdqc/letterpress-solver/device"
)

var questionHash *Question

func handleScreenshotQuestionResp() {
	questionHash = &Question{}
	time.Sleep(time.Millisecond * time.Duration(4200))
	quiz, opt1, opt2, opt3, opt4, isImgQuiz := getHashFromScreenshot()
	questionHash.Data.Quiz = quiz
	questionHash.Data.Options = append(questionHash.Data.Options, opt1, opt2, opt3, opt4)
	questionHash.Data.CurTime = int(time.Now().Unix())
	if isImgQuiz {
		questionHash.Data.ImageID = "_img"
		questionHash.Data.Quiz += "_img"
	} else {
		questionHash.Data.ImageID = ""
	}

	//Get the answer from the db if question fetched by MITM
	answer := FetchHashQuestion(questionHash.Data.Quiz)
	questionHash.CalData.Answer = "不知道"
	ansPos := 0

	if answer != "" {
		for i, option := range questionHash.Data.Options {
			// skip blank option screenshot (shot too early)
			if len(strings.Replace(option, "0", "", -1)) == 0 {
				questionHash.Data.CurTime = -1
				break
			}
			if option == answer {
				ansPos = i + 1
				questionHash.CalData.Answer = option
				break
			}
		}
	}
	fmt.Printf(" 【Q】 %v\n 【A】 %v\n", questionHash.Data.Quiz, questionHash.CalData.Answer)
	// click answer
	if Autoclick == 1 {
		go clickProcess(ansPos, questionHash)
	}

}

func handleScreenshotChooseResponse(bs []byte) {
	if questionHash == nil {
		log.Println("error getting question: nil questionHash")
		return
	}
	if questionHash.Data.CurTime == -1 {
		log.Println("error getting question: questionHash with blank options")
		return
	}
	chooseTime := int(time.Now().Unix())
	if chooseTime < questionHash.Data.CurTime || chooseTime-questionHash.Data.CurTime > 10 {
		log.Println("error getting question: questionHash expired")
		return
	}

	chooseResp := &ChooseResp{}
	json.Unmarshal(bs, chooseResp)

	//If the question fetched by MITM, save it; elif fetched by OCR(no roomID or Num), don't save
	questionHash.CalData.TrueAnswer = questionHash.Data.Options[chooseResp.Data.Answer-1]
	if chooseResp.Data.Yes {
		questionHash.CalData.TrueAnswer = questionHash.Data.Options[chooseResp.Data.Option-1]
	}
	log.Printf("[SaveHash]  %s -> %s\n\n", questionHash.Data.Quiz, questionHash.CalData.TrueAnswer)
	StoreHashQuestion(questionHash)
}

func getHashFromScreenshot() (quiz string, opt1 string, opt2 string, opt3 string, opt4 string, isImgQuiz bool) {
	log.Println("Hashing quiz and options from screenshot ...")
	tx1 := time.Now()

	cfg := device.GetConfig()
	screenshot := device.NewScreenshot(cfg)
	png, err := screenshot.GetImage()
	if err != nil {
		log.Println(err.Error())
		return
	}

	//TODO: test the png image quiz or not
	isImgQuiz = false

	log.Printf("Image get time: %d ms\n", time.Now().Sub(tx1).Nanoseconds()/1e6)

	quiz, opt1, opt2, opt3, opt4, err = device.GetImageHash(png, cfg)
	if err != nil {
		log.Println(err.Error())
		return
	}
	println(quiz, opt1, opt2, opt3, opt4)
	log.Printf("Image get+hash time: %d ms\n", time.Now().Sub(tx1).Nanoseconds()/1e6)
	return
}
