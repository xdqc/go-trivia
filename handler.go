package solver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	brainID      string
	roomID       string
	storedAnsPos int
	selfScore    int
	oppoScore    int
	randClicked  bool // For click random answer

	hasReviewCommented    bool     // For input question comment
	isReviewCommentPassed bool     // For input question comment
	answers               []string // For input question comment

	prevQuizNum int // For getting random question from db for live streaming
)

func handleQuestionResp(bs []byte) {
	question := &Question{}
	storedAnsPos = 0
	if len(bs) > 0 && !strings.Contains(string(bs), "encryptedData") {
		// Get quiz from MITM
		json.Unmarshal(bs, question)
		// save question to buff, with the imageID
		if question.Data.ImageID != "" {
			question.Data.Quiz = question.Data.Quiz + question.Data.ImageID
		}
		// Get self and oppo score
		re := regexp.MustCompile(`"score":{"(\d+)":(\d+),"(\d+)":(\d+)}`)
		scores := re.FindStringSubmatch(string(bs))

		if len(scores) == 5 {
			if scores[1] == brainID {
				selfScore, _ = strconv.Atoi(scores[2])
				oppoScore, _ = strconv.Atoi(scores[4])
			} else if scores[3] == brainID {
				selfScore, _ = strconv.Atoi(scores[4])
				oppoScore, _ = strconv.Atoi(scores[2])
			}
		} else {
			selfScore, oppoScore = 0, 0
		}
	} else {
		if strings.Contains(string(bs), "encryptedData") {
			// Get quiz from OCR
			time.Sleep(time.Millisecond * time.Duration(4200))
			question.Data.Quiz, question.Data.Options = getQuizFromOCR()
		}
		if len(question.Data.Options) == 0 || question.Data.Quiz == "" {
			log.Println("No quiz or options found in screenshot...")
			// click answer
			if Autoclick == 1 {
				go clickProcess(0, question)
			}
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
		quiz = strings.Replace(quiz, "」」", "」", -1)
		question.Data.Quiz = quiz
	}

	question.CalData.RoomID = roomID
	question.CalData.quizNum = strconv.Itoa(question.Data.Num)

	//Get the answer from the db if question fetched by MITM
	answer := FetchQuestion(question)

	// fetch image of the quiz
	// keywords, quoted := preProcessQuiz(question.Data.Quiz, false)
	// imgTimeChan := make(chan int64)
	// go fetchAnswerImage(answer, keywords, quoted, imgTimeChan)

	answerItem := "不知道"
	ansPos := 0
	odds := make([]float32, len(question.Data.Options))
	if question.Data.Num == 1 {
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
	if storedAnsPos == 0 {
		answerItem, ansPos = getAnswerFromAPI(odds, question.Data.Quiz, question.Data.Options, answer)
	}

	// click answer
	if Autoclick == 1 {
		go clickProcess(ansPos, question)
	}

	go SetQuestion(question)

	// Determine the pedia term for quiz review comment
	// pediaTerm := quoted
	// if len(pediaTerm) == 0 || len([]rune(pediaTerm)) > 8 {
	// 	// Find keyword with minimum count in Corpus
	// 	minCount := math.MaxInt32
	// 	for _, kw := range keywords {
	// 		if word, ok := CorpusWord[kw]; ok {
	// 			if word.Count < minCount {
	// 				minCount = CorpusWord[kw].Count
	// 				pediaTerm = kw
	// 			}
	// 		}
	// 	}
	// 	// Use answer as the pedia item
	// 	reNum := regexp.MustCompile("[0-9]+")
	// 	if !reNum.MatchString(answerItem) {
	// 		if len([]rune(quoted)) > 8 {
	// 			pediaTerm = answerItem
	// 		} else if word, ok := CorpusWord[answerItem]; ok {
	// 			if word.Count < minCount {
	// 				pediaTerm = answerItem
	// 			}
	// 		} else if len([]rune(answerItem)) < 6 {
	// 			pediaTerm = answerItem
	// 		}
	// 	}
	// }
	// answers = append(answers, pediaTerm)

	fmt.Printf(" 【Q】 %v\n 【A】 %v\n", question.Data.Quiz, answerItem)
	question.CalData.Answer = answerItem
	question.CalData.AnswerPos = ansPos
	question.CalData.Odds = odds
	questionInfo, _ = json.Marshal(question)

	// Image time and question core information may not be sent in ONE http GET response to client
	// question.CalData.ImageTime = <-imgTimeChan
	// questionInfo, _ = json.Marshal(question)
	question = nil
}

func getAnswerFromAPI(odds []float32, questionDataQuiz string, questionDataOptions []string, answer string) (answerItem string, ansPos int) {
	ret := GetFromAPI(questionDataQuiz, questionDataOptions)

	log.Printf("Google predict => %v\n", ret)
	total := 1
	ansPos = 1

	for _, option := range questionDataOptions {
		total += ret[option]
	}
	if total != 1 {
		// total == 1 -> 0,0,0,0
		max := math.MinInt32
		for i, option := range questionDataOptions {
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
				if odds[ansPos-1] < 50 || len(answer) > 6 || !re.MatchString(answer) {
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
			// log.Println("new question got!")
		}
		if len(odds) == 4 {
			storedAnsPos = ansPos
		}
	}
	return
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

func handleNextQuestion(topic string) {
	question := &Question{}
	answers = make([]string, 0)

	// Get random question from db for live streaming
	question = FetchRandomQuestion(topic)
	ansPos := 0
	odds := make([]float32, len(question.Data.Options))
	for {
		question.Data.Num = rand.Intn(9) + 1
		if question.Data.Num != prevQuizNum {
			prevQuizNum = question.Data.Num
			break
		}
	}
	question.CalData.Odds = odds
	question.CalData.AnswerPos = ansPos
	questionInfo, _ = json.Marshal(question)

	answer := question.CalData.Answer

	// fetch image of the quiz
	keywords, quoted := preProcessQuiz(question.Data.Quiz, false)
	imgTimeChan := make(chan int64)
	go fetchAnswerImage(answer, keywords, quoted, imgTimeChan)

	// Image time and question core information may not be sent in ONE http GET response to client
	question.CalData.ImageTime = <-imgTimeChan
	questionInfo, _ = json.Marshal(question)
	question = nil
}

func handleCurrentAnswer(qNum int, user string, choice string) {
	question := &Question{}
	err := json.Unmarshal(questionInfo, question)
	if err != nil {
		log.Println(err.Error())
		return
	} else if question.Data.Num != qNum {
		log.Println("Question #id does not match current.")
		return
	} else if len(answers) > 0 {
		log.Println("Question has been answerd")
		return
	}

	if choice == "A" {
		question.CalData.Choice = 1
	} else if choice == "B" {
		question.CalData.Choice = 2
	} else if choice == "C" {
		question.CalData.Choice = 3
	} else if choice == "D" {
		question.CalData.Choice = 4
	} else {
		question.CalData.Choice = 0
	}
	answer := question.CalData.Answer
	ansPos := 0
	odds := make([]float32, len(question.Data.Options))
	for i, option := range question.Data.Options {
		if option == answer {
			ansPos = i + 1
			if ansPos == question.CalData.Choice {
				odds[i] = 666
				go recordCorrectUser(user)
			} else {
				odds[i] = 333
			}
			break
		}
	}
	question.CalData.Odds = odds
	question.CalData.AnswerPos = ansPos
	question.CalData.User = user
	questionInfo, _ = json.Marshal(question)
	question = nil

	answers = append(answers, answer)

	time.Sleep(10 * time.Second)
	handleNextQuestion("")
}

func clickProcess(ansPos int, question *Question) {
	var centerX = 540    // center of screen
	var firstItemY = 840 // center of first item (y)
	var optionHeight = 200
	var nextMatchY = 1150 // 1650 1400 1150 900
	// if rand.Intn(100) < 80 {
	// 	nextMatchY = 1400
	// }
	if ansPos >= 0 {
		if ansPos == 0 || ansPos > 4 {
			// click randomly, only do it once on first 4 quiz
			ansPos = rand.Intn(4) + 1
			randClicked = true
		}
		// if ansPos == 0 || selfScore-oppoScore > 500 || (question.Data.Num < 5 && selfScore-oppoScore > 220) {
		// 	// click randomly, only do it when have big advantage
		// 	correctAnsPos := ansPos
		// 	for {
		// 		ansPos = rand.Intn(4) + 1
		// 		if ansPos != correctAnsPos {
		// 			break
		// 		}
		// 	}
		// 	randClicked = true
		// }
		if question.Data.ImageID != "" {
			offsetX := -200
			offsetY := -200
			if ansPos%2 == 0 {
				offsetY = 0
			}
			if ansPos > 2 {
				offsetX = -offsetX
			}
			go clickAction(centerX+offsetX, firstItemY+optionHeight*(4-1)+offsetY) // click image option
		} else {
			go clickAction(centerX, firstItemY+optionHeight*(ansPos-1)) // click answer option
		}
		time.Sleep(time.Millisecond * 500)
		go clickAction(centerX, firstItemY+optionHeight*(4-1)) // click fourth option
		if rand.Intn(100) < 20 && question.Data.Num == 5 {
			time.Sleep(time.Millisecond * 400)
			go clickEmoji()
		}
	} else {
		// go to next match
		randClicked = false
		selfScore = 0
		oppoScore = 0

		// inputADBText()

		time.Sleep(time.Millisecond * 500)
		go swipeAction() // go back to game selection menu
		time.Sleep(time.Millisecond * 800)
		go clickAction(centerX, nextMatchY) // start new game
		time.Sleep(time.Millisecond * 1000)
		go clickAction(centerX, nextMatchY)
	}
}

func clickAction(posX int, posY int) {
	touchX, touchY := strconv.Itoa(posX+rand.Intn(40)-20), strconv.Itoa(posY+rand.Intn(20)-10)
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
	time.Sleep(time.Millisecond * 100)
	fX, fY := 170, 560
	dX, dY := 150, 150
	touchX, touchY := strconv.Itoa(fX+dX*(rand.Intn(2)*3+rand.Intn(2))), strconv.Itoa(fY+dY*rand.Intn(3))
	_, err = exec.Command("adb", "shell", "input", "tap", touchX, touchY).Output() // tap the emoji
	if err != nil {
		log.Println("error: check adb connection.", err)
	}
}

func inputADBText() {
	search := make(chan string, 5)
	done := make(chan bool, 1)
	count := cap(search)
	pediaContents := make([]string, len(answers)) // For input question comment

	for i := 0; i < len(answers); i++ {
		// donot search the pure number answer, meaningless
		reNum := regexp.MustCompile("[0-9]+")
		if !reNum.MatchString(answers[i]) {
			log.Println("Pedia item:", answers[i])
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
				pediaContents[idx-1] = s[8:]

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

	isReviewCommentPassed = true
	prevMsg := ""
	for index := 0; index < len(answers); index++ {
		// record the passed comments
		if isReviewCommentPassed && len(prevMsg) > 0 {
			f, _ := os.OpenFile("./dict/passed.txt", os.O_APPEND|os.O_WRONLY, 0644)
			defer f.Close()
			f.WriteString(prevMsg + "\n\n")
		} else if len(prevMsg) > 0 {
			break
		}

		re := regexp.MustCompile("[\\n\"]+")
		quoted := regexp.MustCompile("\\[[^\\]]+\\]")
		msg := re.ReplaceAllString(pediaContents[index], "")
		msg = quoted.ReplaceAllString(msg, "")
		msg, _ = filterManage.Filter().Replace(msg, '*') // filter sensitive words
		msg = strings.Replace(msg, "*", "", -1)
		if len([]rune(msg)) > 300 {
			msg = string([]rune(msg)[:300])
		}
		if hasReviewCommented {
			exec.Command("adb", "shell", "input", "tap", "500", "500").Output()                        // tap center, esc error msg dialog box
			exec.Command("adb", "shell", "input", "swipe", "800", "470", "200", "470", "200").Output() // swipe left, forward
			continue
		}
		println(msg)
		prevMsg = msg
		exec.Command("adb", "shell", "input", "tap", "500", "1700").Output() // tap `input bar`
		time.Sleep(time.Millisecond * 10)
		exec.Command("adb", "shell", "am", "broadcast", "-a ADB_INPUT_TEXT", "--es msg", "\""+msg+"\"").Output() // sending text input
		time.Sleep(time.Millisecond * 10)
		exec.Command("adb", "shell", "am", "broadcast", "-a ADB_EDITOR_CODE", "--ei code", "4").Output() // editor action `send`
		time.Sleep(time.Millisecond * 100)

		exec.Command("adb", "shell", "input", "tap", "500", "500").Output()                        // tap center, esc error msg dialog box
		exec.Command("adb", "shell", "input", "swipe", "800", "470", "200", "470", "200").Output() // swipe left, forward
		time.Sleep(time.Millisecond * 100)
	}
	// record the last passed and failed comments
	if isReviewCommentPassed && len(prevMsg) > 0 {
		f, _ := os.OpenFile("./dict/passed.txt", os.O_APPEND|os.O_WRONLY, 0644)
		defer f.Close()
		f.WriteString(prevMsg + "\n\n")
	} else if len(prevMsg) > 0 {
		f, _ := os.OpenFile("./dict/failed.txt", os.O_APPEND|os.O_WRONLY, 0644)
		defer f.Close()
		f.WriteString(prevMsg + "\n\n")
	}
	exec.Command("adb", "shell", "input", "tap", "500", "500").Output() // tap center, esc dialog box, to go back
	exec.Command("adb", "shell", "input", "tap", "75", "150").Output()  // tap esc arrow, go back
}

func recordCorrectUser(user string) {
	ranking := make(map[string]int)
	bs, _ := ioutil.ReadFile("ranking.txt")
	txt := string(bs)
	lines := strings.Split(txt, "\n")
	for _, line := range lines {
		if len(strings.Split(line, "\t")) == 2 {
			name := strings.Split(line, "\t")[1]
			count, _ := strconv.Atoi(strings.Split(line, "\t")[0])
			ranking[name] = count
			log.Println(name, count)
		}
	}

	if val, ok := ranking[user]; ok {
		ranking[user] = val + 1
	} else {
		ranking[user] = 1
	}

	type kv struct {
		Key   string
		Value int
	}

	var ss []kv
	for k, v := range ranking {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	contents := ""
	for _, kv := range ss {
		contents += fmt.Sprintf("%d\t%s\n", kv.Value, kv.Key)
	}
	contents += "|\n|\n|\n"
	ioutil.WriteFile("ranking.txt", []byte(contents), 0644)
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
		ImageID     string   `json:"imageId"`
	} `json:"data"`
	CalData struct {
		RoomID     string
		quizNum    string
		Answer     string
		AnswerPos  int
		TrueAnswer string
		Odds       []float32
		ImageTime  int64
		User       string
		Choice     int
		Voice      int
	} `json:"caldata"`
	Errcode int `json:"errcode"`
}

type EncodedQuestion struct {
	Data struct {
		EncryptedData string `json:"encryptedData"`
	} `json:"data"`
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
