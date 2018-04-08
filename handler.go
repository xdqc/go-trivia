package solver

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
)

var (
	roomID string
)

//根据题目返回,进行答案搜索
func handleQuestionResp(bs []byte) (bsNew []byte, ansPos int) {
	bsNew = bs
	question := &Question{}
	json.Unmarshal(bs, question)
	question.CalData.RoomID = roomID
	question.CalData.quizNum = strconv.Itoa(question.Data.Num)

	//Get the answer from the db
	answer := FetchQuestion(question)

	// question.CalData.TrueAnswer = answer
	// question.CalData.Answer = answer
	go SetQuestion(question)

	ansPos = 0
	answerItem := "不知道"
	var odds [4]float32

	if answer != "" {
		for i, option := range question.Data.Options {
			if option == answer {
				// question.Data.Options[i] = option + "[.]"
				ansPos = i + 1
				answerItem = option
				odds[i] = 999
				break
			}
		}
	}

	if answerItem == "不知道" {
		var ret map[string]int
		ret = GetFromAPI(question.Data.Quiz, question.Data.Options)
		log.Printf("Google predict => %v\n", ret)
		total := 1

		for _, option := range question.Data.Options {
			total += ret[option]
		}
		max := math.MinInt32
		for i, option := range question.Data.Options {
			odd := float32(ret[option]) / float32(total-ret[option])
			fmt.Printf("|%-10s|%6.2f\n", option, odd)
			odds[i] = odd

			// question.Data.Options[i] = option + "[" + strconv.Itoa(ret[option]) + "]"
			if ret[option] > max && ret[option] != 0 {
				max = ret[option]
				ansPos = i + 1
				answerItem = option
			}
		}
	}

	log.Printf("Question answer predict =>\n 【Q】 %v\n 【A】 %v\n", question.Data.Quiz, answerItem)
	question.CalData.Answer = answerItem
	question.CalData.AnswerPos = ansPos
	question.CalData.Odds = odds
	questionInfo, _ = json.Marshal(question)
	// println(string(questionInfo))

	//返回答案
	return bs, ansPos
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
		Odds       [4]float32
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
