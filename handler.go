package solver

import (
	"encoding/json"
	"log"
	"math"
	"strconv"
	"strings"
)

var (
	roomID string
)

func handleQuestionResp(bs []byte) {
	question := &Question{}
	if len(bs) > 0 {
		// Get quiz from MITM
		json.Unmarshal(bs, question)
	} else {
		// Get quiz from OCR
		question.Data.Quiz, question.Data.Options = getQuizFromOCR()
		if len(question.Data.Options) == 0 || question.Data.Quiz == "" {
			log.Println("No quiz or options found in screenshot...")
			return
		} else {
			quiz := question.Data.Quiz
			quiz = strings.Replace(quiz, "?", "？", -1)
			quiz = strings.Replace(quiz, ",", "，", -1)
			quiz = strings.Replace(quiz, "(", "（", -1)
			quiz = strings.Replace(quiz, ")", "）", -1)
			quiz = strings.Replace(quiz, "\"", "“", -1)
			quiz = strings.Replace(quiz, "'", "‘", -1)
			quiz = strings.Replace(quiz, "!", "！", -1)
			question.Data.Quiz = quiz
		}
	}

	// fetch image of the quiz
	keywords, quoted := preProcessQuiz(question.Data.Quiz, false)
	imgTimeChan := make(chan int64)
	go fetchAnswerImage("", keywords, quoted, imgTimeChan)

	question.CalData.RoomID = roomID
	question.CalData.quizNum = strconv.Itoa(question.Data.Num)

	//Get the answer from the db if question fetched by MITM
	answer := FetchQuestion(question)
	ansPos := 0

	// question.CalData.TrueAnswer = answer
	// question.CalData.Answer = answer
	go SetQuestion(question)

	answerItem := "不知道"
	odds := make([]float32, len(question.Data.Options))

	if answer != "" {
		for i, option := range question.Data.Options {
			if option == answer {
				// question.Data.Options[i] = option + "[.]"
				ansPos = i + 1
				answerItem = option
				odds[i] = 888
				break
			}
		}
	}
	storedAnsPos := ansPos

	// Put true here to force searching, even if found answer in db
	if storedAnsPos == 0 {
		var ret map[string]int
		ret = GetFromAPI(question.Data.Quiz, question.Data.Options)
		log.Printf("Google predict => %v\n", ret)
		total := 1

		for _, option := range question.Data.Options {
			total += ret[option]
		}
		max := math.MinInt32
		for i, option := range question.Data.Options {
			odds[i] = float32(ret[option]) / float32(total-ret[option])
			// question.Data.Options[i] = option + "[" + strconv.Itoa(ret[option]) + "]"
			if ret[option] > max && ret[option] != 0 {
				max = ret[option]
				ansPos = i + 1
				answerItem = option
			}
		}
		// verify the stored answer
		if answer == answerItem {
			//good
			odds[ansPos-1] += 600
		} else {
			if answer != "" {
				// searched result could be wrong
				if storedAnsPos != 0 {
					if odds[ansPos-1] < 1 {
						log.Println("searched answer could be wrong...")
						answerItem = answer
						ansPos = storedAnsPos
						odds[ansPos-1] = 333
					} else {
						// stored answer may be corrupted
						log.Println("stored answer may be corrupted...")
						odds[ansPos-1] = 444
					}
				} else {
					// if storedAnsPos==0, the stored anser exists, but match nothing => the option words changed by the game
					log.Println("the previous option words changed by the game...")
				}
			} else {
				log.Println("new question got!")
			}
		}
	}

	log.Printf("Question answer predict =>\n 【Q】 %v\n 【A】 %v\n", question.Data.Quiz, answerItem)
	question.CalData.Answer = answerItem
	question.CalData.AnswerPos = ansPos
	question.CalData.Odds = odds
	questionInfo, _ = json.Marshal(question)

	// Image time and question core information may not be sent in one http GET response to client
	question.CalData.ImageTime = <-imgTimeChan
	questionInfo, _ = json.Marshal(question)

}

func handleChooseResponse(bs []byte) {
	chooseResp := &ChooseResp{}
	json.Unmarshal(bs, chooseResp)

	//log.Println("response choose", roomID, chooseResp.Data.Num, string(bs))
	question := GetQuestion(roomID, strconv.Itoa(chooseResp.Data.Num))
	if question == nil {
		log.Println("error getting question", chooseResp.Data.RoomID, chooseResp.Data.Num)
		return
	}
	//If the question fetched by MITM, save it; elif fetched by OCR(no roomID or Num), don't save
	question.CalData.TrueAnswer = question.Data.Options[chooseResp.Data.Answer-1]
	if chooseResp.Data.Yes {
		question.CalData.TrueAnswer = question.Data.Options[chooseResp.Data.Option-1]
	}
	log.Printf("[SaveData]  %s -> %s\n\n", question.Data.Quiz, question.CalData.TrueAnswer)
	StoreQuestion(question)
	StoreWholeQuestion(question)
}

type Question struct {
	Data struct {
		Quiz        string   `json:"quiz"`
		Options     []string `json:"options"`
		Num         int      `json:"num"`
		School      string   `json:"school"`
		Type        string   `json:"type"`
		Contributor string   `json:"contributor"`
		EndTime     int      `json:"endTime"`
		CurTime     int      `json:"curTime"`
	} `json:"data"`
	CalData struct {
		RoomID     string
		quizNum    string
		Answer     string
		AnswerPos  int
		TrueAnswer string
		Odds       []float32
		ImageTime  int64
	} `json:"caldata"`
	Errcode int `json:"errcode"`
}

type ChooseResp struct {
	Data struct {
		UID         int  `json:"uid"`
		Num         int  `json:"num"`
		Answer      int  `json:"answer"`
		Option      int  `json:"option"`
		Yes         bool `json:"yes"`
		Score       int  `json:"score"`
		TotalScore  int  `json:"totalScore"`
		RowNum      int  `json:"rowNum"`
		RowMult     int  `json:"rowMult"`
		CostTime    int  `json:"costTime"`
		RoomID      int  `json:"roomId"`
		EnemyScore  int  `json:"enemyScore"`
		EnemyAnswer int  `json:"enemyAnswer"`
	} `json:"data"`
	Errcode int `json:"errcode"`
}

//roomID=476376430&quizNum=4&option=4&uid=26394007&t=1515326786076&sign=3592b9d28d045f3465206b4147ea872b
