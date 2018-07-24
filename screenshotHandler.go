package solver

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xdqc/letterpress-solver/device"
)

var questionHash *Question

func handleScreenshotQuestionResp() {
	questionHash = &Question{}
	time.Sleep(time.Millisecond * time.Duration(4100))
	quiz, opt1, opt2, opt3, opt4, isImgQuiz := getHashFromScreenshot()
	questionHash.HashData.Quiz = quiz
	questionHash.HashData.Options = append(questionHash.HashData.Options, opt1, opt2, opt3, opt4)
	questionHash.Data.CurTime = int(time.Now().Unix())
	for _, option := range questionHash.HashData.Options {
		// skip blank option screenshot (shot too early)
		if option == "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" {
			quiz, opt1, opt2, opt3, opt4, isImgQuiz = getHashFromScreenshot()
			questionHash.HashData.Quiz = quiz
			questionHash.HashData.Options = make([]string, 0)
			questionHash.HashData.Options = append(questionHash.HashData.Options, opt1, opt2, opt3, opt4)
			questionHash.Data.CurTime = int(time.Now().Unix())
			break
		}
	}
	// tx1 := time.Now()
	if isImgQuiz {
		questionHash.Data.ImageID = "_img"
		questionHash.HashData.Quiz += questionHash.Data.ImageID
	} else {
		questionHash.Data.ImageID = ""
	}

	answer := FetchHashQuestion(questionHash.HashData.Quiz)
	questionHash.CalData.Answer = "不知道"
	ansPos := 0

	if answer != "" {
		println("----------------", questionHash.HashData.Quiz, "----------------")
		hammings := make([]int, 0)
		for _, option := range questionHash.HashData.Options {
			ham := Hamming(option, answer)
			hammings = append(hammings, ham)
		}
		min := math.MaxInt32
		for i, option := range questionHash.HashData.Options {
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

				if _, err = f.WriteString(time.Now().Format(time.RFC3339) + "\t" + quiz + "\n\t" + answer + "\n\t" + questionHash.CalData.Answer + "\n\n"); err != nil {
					panic(err)
				}
				for _, opt := range questionHash.HashData.Options {
					if _, err = f.WriteString(strconv.Itoa(Hamming(answer, opt)) + "\t" + opt + "\n"); err != nil {
						panic(err)
					}
				}
				if _, err = f.WriteString("\n\n"); err != nil {
					panic(err)
				}
			}()
		}
	}

	// log.Printf("hash process time: %d ms\n", time.Now().Sub(tx1).Nanoseconds()/1e6)
	if ansPos == 0 {
		questionHash.Data.Quiz, questionHash.Data.Options = getQuizFromOCR()
		questionHash.Data.Quiz = strings.Replace(questionHash.Data.Quiz, "?", "？", -1)
		questionHash.Data.Quiz = strings.Replace(questionHash.Data.Quiz, ",", "，", -1)
		questionHash.Data.Quiz = strings.Replace(questionHash.Data.Quiz, "(", "（", -1)
		questionHash.Data.Quiz = strings.Replace(questionHash.Data.Quiz, ")", "）", -1)
		questionHash.Data.Quiz = strings.Replace(questionHash.Data.Quiz, "\"", "“", -1)
		questionHash.Data.Quiz = strings.Replace(questionHash.Data.Quiz, "'", "‘", -1)
		questionHash.Data.Quiz = strings.Replace(questionHash.Data.Quiz, "!", "！", -1)
		questionHash.Data.Quiz = strings.Replace(questionHash.Data.Quiz, "」」", "」", -1)
		odds := make([]float32, len(questionHash.Data.Options))

		answer := FetchQuestion(questionHash.Data.Quiz)

		answerItem := "不知道"
		if answer != "" {
			for i, option := range questionHash.Data.Options {
				if option == answer {
					ansPos = i + 1
					answerItem = option
					odds[i] = 888
					break
				}
			}
		}

		if ansPos == 0 {
			answerItem, ansPos = getAnswerFromAPI(odds, questionHash.Data.Quiz, questionHash.Data.Options, answer)
		}
		questionHash.CalData.TrueAnswer = answerItem
		fmt.Printf(" 【Q】 %v\n 【A】 %v\n", questionHash.Data.Quiz, answerItem)
	}
	// click answer
	if Autoclick == 1 {
		go clickProcess(ansPos, questionHash)
	}
}

//Hamming distance is simply the minimum number of substitutions required to change one string into the other.
func Hamming(a, b string) int {
	if a == b {
		return 0
	}

	al := len(a)
	bl := len(b)

	if al != bl {
		fmt.Errorf("strings are not equal (len(a)=%d, len(b)=%d)", al, bl)
		return -1
	}

	var difference = 0

	a = stringToBin(a)
	b = stringToBin(b)
	for i := range a {
		if a[i] != b[i] {
			difference = difference + 1
		}
	}

	return difference
}

func stringToBin(hexStr string) (binString string) {
	for _, c := range hexStr {
		if s, err := strconv.ParseInt(string(c), 16, 8); err == nil {
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
	if chooseTime < questionHash.Data.CurTime || chooseTime-questionHash.Data.CurTime > 15 {
		log.Println("error getting question: questionHash expired")
		return
	}

	chooseResp := &ChooseResp{}
	json.Unmarshal(bs, chooseResp)

	questionHash.HashData.TrueAnswer = questionHash.HashData.Options[chooseResp.Data.Answer-1]
	if chooseResp.Data.Yes {
		questionHash.HashData.TrueAnswer = questionHash.HashData.Options[chooseResp.Data.Option-1]
	}
	log.Printf("[SaveHash] %s ", questionHash.HashData.Quiz)
	go StoreHashQuestion(questionHash)

	if questionHash.Data.Quiz != "" && len(questionHash.Data.Options) == 4 {
		if chooseResp.Data.Yes {
			questionHash.CalData.TrueAnswer = questionHash.Data.Options[chooseResp.Data.Option-1]
		} else {
			questionHash.CalData.TrueAnswer = questionHash.Data.Options[chooseResp.Data.Answer-1]
		}
		log.Printf("[SaveQuiz] %s-> %s\n\n", questionHash.Data.Quiz, questionHash.CalData.TrueAnswer)
		go StoreQuestion(questionHash)
	} else {
		log.Println("[NoQuizSaved] - incomplete data\n")
	}
}

func getHashFromScreenshot() (quiz string, opt1 string, opt2 string, opt3 string, opt4 string, isImgQuiz bool) {
	// log.Println("Hashing quiz and options from screenshot ...")

	cfg := device.GetConfig()
	png, err := device.NewScreenshot(cfg).GetImage()
	if err != nil {
		log.Println(err.Error())
		return
	}
	// tx1 := time.Now()

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
	// fmt.Printf("%v\n%v\n%v\n%v\n%v\n%v\n", quiz, opt1, opt2, opt3, opt4, sampleHash)
	// log.Printf("Image get+hash time: %d ms\n", time.Now().Sub(tx1).Nanoseconds()/1e6)
	return
}
