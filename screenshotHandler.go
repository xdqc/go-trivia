package solver

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/xdqc/letterpress-solver/device"
)

var questionHash *Question

func handleScreenshotQuestionResp() {
	questionHash = &Question{}
	time.Sleep(time.Millisecond * time.Duration(4300))
	quiz, opt1, opt2, opt3, opt4, isImgQuiz := getHashFromScreenshot()
	questionHash.Data.Quiz = quiz
	questionHash.Data.Options = append(questionHash.Data.Options, opt1, opt2, opt3, opt4)
	questionHash.Data.CurTime = int(time.Now().Unix())
	for _, option := range questionHash.Data.Options {
		// skip blank option screenshot (shot too early)
		if option == "000000000000000000000000000000000000000000000000000000000000000000000000" {
			questionHash.Data.CurTime = -1
			// click answer
			if Autoclick == 1 {
				go clickProcess(0, questionHash)
			}
			return
		}
	}

	if isImgQuiz {
		questionHash.Data.ImageID = "_img"
		questionHash.Data.Quiz += questionHash.Data.ImageID
	} else {
		questionHash.Data.ImageID = ""
	}

	//Get the answer from the db if question fetched by MITM
	answer := FetchHashQuestion(questionHash.Data.Quiz)
	questionHash.CalData.Answer = "不知道"
	ansPos := 0

	hammings := make([]int, 0)
	if answer != "" {
		for _, option := range questionHash.Data.Options {
			ham, _ := Hamming(option, answer)
			hammings = append(hammings, ham)
		}
		min := math.MaxInt32
		for i, option := range questionHash.Data.Options {
			if hammings[i] < min && hammings[i] >= 0 {
				min = hammings[i]
				ansPos = i + 1
				questionHash.CalData.Answer = option
			}
		}
		if ansPos > 0 {
			go func() {
				f, err := os.OpenFile("./ranking.txt", os.O_APPEND|os.O_WRONLY, 0600)
				if err != nil {
					panic(err)
				}

				defer f.Close()

				if _, err = f.WriteString(quiz + "\n" + answer + "\n" + questionHash.CalData.Answer + "\n\n"); err != nil {
					panic(err)
				}
			}()
		}
	}

	if ansPos == 0 {
		questionDataQuiz, questionDataOptions := getQuizFromOCR()
		odds := make([]float32, len(questionDataOptions))
		answerItem := "不知道"
		answerItem, ansPos = getAnswerFromAPI(odds, questionDataQuiz, questionDataOptions, answer)
		fmt.Printf(" 【Q】 %v\n 【A】 %v\n", questionDataQuiz, answerItem)
	}

	// click answer
	if Autoclick == 1 {
		go clickProcess(ansPos, questionHash)
	}

}

//Hamming distance is simply the minimum number of substitutions required to change one string into the other.
func Hamming(a, b string) (int, error) {
	al := len(a)
	bl := len(b)

	if al != bl {
		return -1, fmt.Errorf("strings are not equal (len(a)=%d, len(b)=%d)", al, bl)
	}

	var difference = 0

	a = stringToBin(a)
	b = stringToBin(b)
	for i := range a {
		if a[i] != b[i] {
			difference = difference + 1
		}
	}

	return difference, nil
}

func stringToBin(hexStr string) (binString string) {
	for _, c := range hexStr {
		if s, err := strconv.ParseInt(string(c), 16, 4); err == nil {
			binString = fmt.Sprintf("%s%04s", binString, strconv.FormatInt(s, 2))
		}
	}
	return
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
	// log.Println("Hashing quiz and options from screenshot ...")
	// tx1 := time.Now()

	cfg := device.GetConfig()
	screenshot := device.NewScreenshot(cfg)
	png, err := screenshot.GetImage()
	if err != nil {
		log.Println(err.Error())
		return
	}

	//TODO: test the png image quiz or not
	sampleHash := ""
	quiz, opt1, opt2, opt3, opt4, sampleHash, err = device.GetImageHash(png, cfg.APP)
	if err != nil {
		log.Println(err.Error())
		return
	}
	if sampleHash == "0000000000000000" {
		isImgQuiz = false
	} else {
		isImgQuiz = true
		log.Println("deal with image quiz")
		quiz, opt1, opt2, opt3, opt4, sampleHash, err = device.GetImageHash(png, cfg.APP+"_img")
	}
	fmt.Printf("%v\n%v\n%v\n%v\n%v\n%v\n", quiz, opt1, opt2, opt3, opt4, sampleHash)
	// log.Printf("Image get+hash time: %d ms\n", time.Now().Sub(tx1).Nanoseconds()/1e6)
	return
}
