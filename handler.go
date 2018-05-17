package solver

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	roomID       string
	storedAnsPos int
	randClicked  bool
	luckyPedias  []string
	answers      []string
)

func handleQuestionResp(bs []byte) {
	question := &Question{}
	storedAnsPos = 0
	if len(bs) > 0 {
		// Get quiz from MITM
		json.Unmarshal(bs, question)
	} else {
		// Get quiz from OCR
		question.Data.Quiz, question.Data.Options = getQuizFromOCR()
		if len(question.Data.Options) == 0 || question.Data.Quiz == "" {
			log.Println("No quiz or options found in screenshot...")
			return
		}
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

	question.CalData.RoomID = roomID
	question.CalData.quizNum = strconv.Itoa(question.Data.Num)

	//Get the answer from the db if question fetched by MITM
	answer := FetchQuestion(question)

	// fetch image of the quiz
	// keywords, quoted := preProcessQuiz(question.Data.Quiz, false)
	// imgTimeChan := make(chan int64)
	//go fetchAnswerImage(answer, keywords, quoted, imgTimeChan)

	// question.CalData.TrueAnswer = answer
	// question.CalData.Answer = answer
	go SetQuestion(question)

	answerItem := "不知道"
	ansPos := 0
	odds := make([]float32, len(question.Data.Options))
	if question.Data.Num == 1 {
		luckyPedias = make([]string, 0)
		answers = make([]string, 0)
	}

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
	storedAnsPos = ansPos

	// Put true here to force searching, even if found answer in db
	if true || storedAnsPos == 0 {
		var ret map[string]int
		ret, luckyStr := GetFromAPI(question.Data.Quiz, question.Data.Options)

		luckyPedias = append(luckyPedias, luckyStr)

		log.Printf("Google predict => %v\n", ret)
		total := 1

		for _, option := range question.Data.Options {
			total += ret[option]
		}
		if total != 1 {
			// total == 1 -> 0,0,0,0
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
		}
		// verify the stored answer
		if answer == answerItem {
			//good
			odds[ansPos-1] = 888
		} else {
			if answer != "" {
				// searched result could be wrong
				if storedAnsPos != 0 {
					re := regexp.MustCompile("\\p{Han}+")
					if odds[ansPos-1] < 5 || len(answer) > 6 || !re.MatchString(answer) {
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
			if len(odds) == 4 {
				storedAnsPos = ansPos
			}
		}
	}

	answers = append(answers, answerItem)
	if Mode == 1 {
		go clickProcess(ansPos, question)
	} // click answer

	log.Printf("Question answer predict =>\n 【Q】 %v\n 【A】 %v\n", question.Data.Quiz, answerItem)
	question.CalData.Answer = answerItem
	question.CalData.AnswerPos = ansPos
	question.CalData.Odds = odds
	questionInfo, _ = json.Marshal(question)

	// Image time and question core information may not be sent in one http GET response to client
	// question.CalData.ImageTime = <-imgTimeChan
	// questionInfo, _ = json.Marshal(question)
	question = nil
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

func clickProcess(ansPos int, question *Question) {
	var centerX = 540    // center of screen
	var firstItemY = 840 // center of first item (y)
	var optionHeight = 200
	var nextMatchY = 1650
	if ansPos >= 0 {
		if ansPos == 0 || (!randClicked && question.Data.Num != 5 && (question.Data.School == "娱乐" || question.Data.School == "流行")) {
			// click randomly, only do it once on first 4 quiz
			ansPos = rand.Intn(4) + 1
			// randClicked = true
		}
		time.Sleep(time.Millisecond * 1500)
		go clickAction(centerX, firstItemY+optionHeight*(ansPos-1))
		time.Sleep(time.Millisecond * 1000)
		go clickAction(centerX, firstItemY+optionHeight*(ansPos-1))
		time.Sleep(time.Millisecond * 1000)
		go clickAction(centerX, firstItemY+optionHeight*(4-1))
		if rand.Intn(100) < 20 {
			time.Sleep(time.Millisecond * 500)
			go clickEmoji()
		}
	} else {
		// go to next match
		randClicked = false

		// inputADBText()

		time.Sleep(time.Millisecond * 500)
		go swipeAction() // go back to game selection menu
		time.Sleep(time.Millisecond * 500)
		go clickAction(centerX, nextMatchY) // start new game
		time.Sleep(time.Millisecond * 1000)
		go clickAction(centerX, nextMatchY)
	}
}

func clickAction(posX int, posY int) {
	touchX, touchY := strconv.Itoa(posX+rand.Intn(400)-200), strconv.Itoa(posY+rand.Intn(50)-25)
	_, err := exec.Command("adb", "shell", "input", "tap", touchX, touchY).Output()
	if err != nil {
		log.Println("error: check adb connection.", err)
	}
}

func swipeAction() {
	_, err := exec.Command("adb", "shell", "input", "swipe", "75", "150", "75", "150", "0").Output() // swipe right, back
	if err != nil {
		log.Println("error: check adb connection.", err)
	}
}

func clickEmoji() {
	_, err := exec.Command("adb", "shell", "input", "tap", "100", "300").Output() // tap my avatar to summon emoji panel
	if err != nil {
		log.Println("error: check adb connection.", err)
	}
	time.Sleep(time.Millisecond * 200)
	fX, fY := 170, 560
	dX, dY := 150, 150
	touchX, touchY := strconv.Itoa(fX+dX*1), strconv.Itoa(fY+dY*2*rand.Intn(2))
	_, err = exec.Command("adb", "shell", "input", "tap", touchX, touchY).Output() // tap the emoji
	if err != nil {
		log.Println("error: check adb connection.", err)
	}
}

func inputADBText() {
	search := make(chan string, 5)
	done := make(chan bool, 1)
	count := cap(search)
	for i := 0; i < 5; i++ {
		reNum := regexp.MustCompile("[0-9]+")
		if !reNum.MatchString(answers[i]) {
			go searchBaiduBaike(answers, i+1, search)
		}
	}
	go func() {
		for {
			s, more := <-search
			if more {
				// The first 8 chars in text is the identifier of the search source, 4th is the index
				id := s[:8]
				idx, _ := strconv.Atoi(s[4:5])
				log.Println("search received...", id, idx)
				if idx <= len(luckyPedias) && len(luckyPedias[idx-1]) < 60 {
					luckyPedias[idx-1] = s[8:]
				}
				count--
				if count == 0 {
					done <- true
					return
				}
			}
		}
	}()

	time.Sleep(time.Millisecond * 500)
	exec.Command("adb", "shell", "input", "tap", "1000", "1050").Output() // tap `review current game`
	time.Sleep(time.Millisecond * 4000)

	select {
	case <-done:
		fmt.Println("search done")
	case <-time.After(2 * time.Second):
		fmt.Println("search timeout")
	}

	for index := 0; index < 5; index++ {
		exec.Command("adb", "shell", "input", "tap", "500", "1700").Output() // tap `input bar`
		time.Sleep(time.Millisecond * 200)
		re := regexp.MustCompile("[\\n\"]+")
		quoted := regexp.MustCompile("\\[[^\\]]+\\]")
		msg := re.ReplaceAllString(luckyPedias[index], "")
		msg = quoted.ReplaceAllString(msg, "")
		if len([]rune(msg)) > 500 {
			msg = string([]rune(msg)[:500])
		}
		println(msg)
		exec.Command("adb", "shell", "am", "broadcast", "-a ADB_INPUT_TEXT", "--es msg", "\""+msg+"\"").Output() // sending text input
		time.Sleep(time.Millisecond * 400)
		exec.Command("adb", "shell", "am", "broadcast", "-a ADB_EDITOR_CODE", "--ei code", "4").Output() // editor action `send`
		time.Sleep(time.Millisecond * 200)
		exec.Command("adb", "shell", "input", "swipe", "800", "470", "200", "470", "200").Output() // swipe left, forward
		time.Sleep(time.Millisecond * 400)
	}
	exec.Command("adb", "shell", "input", "tap", "500", "500").Output() // tap center, esc dialog box, to go back
	exec.Command("adb", "shell", "input", "tap", "75", "150").Output()  // tap esc arrow, go back
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
