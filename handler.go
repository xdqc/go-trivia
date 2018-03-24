package solver

import (
	"encoding/json"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
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
	var ret map[string]int
	if answer == "" {
		tx := time.Now()
		ret = GetFromBaidu(question.Data.Quiz, question.Data.Options)
		tx2 := time.Now()
		log.Printf("Cost time %d ms\n", tx2.Sub(tx).Nanoseconds()/1e6)
		log.Printf("Google predict => %v\n", ret)
	}
	question.CalData.TrueAnswer = answer
	question.CalData.Answer = answer
	SetQuestion(question)

	ansPos = 0
	answerItem := "不知道"
	if question.CalData.TrueAnswer != "" {
		for i, option := range question.Data.Options {
			if option == question.CalData.TrueAnswer {
				// question.Data.Options[i] = option + "[.]"
				ansPos = i + 1
				answerItem = option
				break
			}
		}
	} else if strings.Contains(question.Data.Quiz, "不") && !strings.Contains(question.Data.Quiz, "「") {
		//当题目中有“不”时，选取百度结果中最罕见的选项
		var min = math.MaxInt32
		for i, option := range question.Data.Options {
			// question.Data.Options[i] = option + "[" + strconv.Itoa(ret[option]) + "]"
			if ret[option] < min {
				min = ret[option]
				ansPos = i + 1
				answerItem = option
			}
		}
	} else {
		var max int = 0
		for i, option := range question.Data.Options {
			if ret[option] > 0 {
				// question.Data.Options[i] = option + "[" + strconv.Itoa(ret[option]) + "]"
				if ret[option] > max {
					max = ret[option]
					ansPos = i + 1
					answerItem = option
				}
			}
		}
	}

	log.Printf("Question answer predict =>\n 【Q】 %v\n 【A】 %v\n", question.Data.Quiz, answerItem)
	question.CalData.Answer = answerItem
	question.CalData.AnswerPos = ansPos
	questionInfo, _ = json.Marshal(question)
	println(string(questionInfo))

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
