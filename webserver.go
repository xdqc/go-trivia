package solver

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

var (
	matchInfo    []byte
	questionInfo []byte
	idiomInfo    []byte
)

type Words []struct {
	Word string `json:"word"`
}

//RunWeb run a webserver
func RunWeb(port string) {

	r := mux.NewRouter()
	// 1. LP solver
	r.HandleFunc("/match", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(matchInfo)
	}).Methods("GET")
	r.HandleFunc("/words", findWords).Methods("GET")
	r.HandleFunc("/word", deleteWord).Methods("DELETE")

	// 2. Brain solver
	r.HandleFunc("/answer", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(questionInfo)
	}).Methods("GET")

	r.HandleFunc("/brain-ocr", func(w http.ResponseWriter, r *http.Request) {
		handleQuestionResp([]byte{})
	}).Methods("POST")

	r.HandleFunc("/quizContextStream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		quizContextInfo, _ = json.Marshal(getQuizContext())
		w.Write(quizContextInfo)
	}).Methods("GET")

	// 3. Live Stream Questions
	r.HandleFunc("/nextQuiz", func(w http.ResponseWriter, r *http.Request) {
		go handleNextQuestion()
		w.Header().Set("Content-Type", "application/json")
		w.Write(nil)
	}).Methods("GET")
	r.HandleFunc("/currentQuizAnswer", func(w http.ResponseWriter, r *http.Request) {
		qNums, _ := r.URL.Query()["qNum"]
		qNum, _ := strconv.Atoi(qNums[0])
		go handleCurrentAnswer(qNum)
		w.Header().Set("Content-Type", "application/json")
		w.Write(nil)
	}).Methods("GET")

	r.PathPrefix("/solver/").Handler(http.StripPrefix("/solver/", http.FileServer(http.Dir("./lpsolver/dist"))))

	// Use default options
	handler := cors.AllowAll().Handler(r)

	log.Println("web server at port", port)
	http.ListenAndServe(":"+port, handler)

}

func findWords(w http.ResponseWriter, r *http.Request) {

	minLetters, _ := r.URL.Query()["selected"]
	maxLetters, _ := r.URL.Query()["letters"]

	loFreq := make(map[rune]int)
	hiFreq := make(map[rune]int)
	for _, c := range minLetters[0] {
		_, ok := loFreq[c]
		if ok {
			loFreq[c]++
		} else {
			loFreq[c] = 1
		}
	}
	for _, c := range maxLetters[0] {
		_, ok := hiFreq[c]
		if ok {
			hiFreq[c]++
		} else {
			hiFreq[c] = 1
		}
	}

	res := selectWordsDb(loFreq, hiFreq)
	ws, _ := json.Marshal(res)
	log.Println("Fourd words: ", len(res))
	w.Header().Set("Content-Type", "application/json")
	w.Write(ws)
}

func deleteWord(w http.ResponseWriter, r *http.Request) {
	word, _ := r.URL.Query()["delete"]
	log.Println(word[0])
	deleteWordDb(word[0])
}

func setMatch(jsonBytes []byte) {
	matchInfo = jsonBytes
}

func setIdiom(jsonBytes []byte) {
	idiomInfo = jsonBytes
}

func fetchAnswerImage(ans string, quiz []string, quoted string, imgTimeChan chan int64) {
	// tx1 := time.Now()
	values := url.Values{}
	// create search string
	searchStr := ans + " " + quoted
	re := regexp.MustCompile("[^\\p{Han}]+")
	hanRunes := re.ReplaceAllString(searchStr, "")
	if len([]rune(hanRunes)) < 5 {
		if len(quiz) > 8 {
			searchStr += " " + strings.Join(quiz[:9], " ")
		} else {
			searchStr += " " + strings.Join(quiz, " ")
		}
	}
	values.Add("q", searchStr)

	// HTTP request for img url
	// set timeout for http GET
	timeout := time.Duration(6 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	resp, e := client.Get("http://image.so.com/i?" + values.Encode()) //www.bing.com/images/search?
	if e != nil {
		log.Println("Get img URL err: " + e.Error())
		imgTimeChan <- 0
		return
	}
	if resp == nil || resp.Body == nil {
		imgTimeChan <- 0
		return
	}
	doc, e := goquery.NewDocumentFromReader(resp.Body)
	if e != nil {
		log.Println("Parse img url response body error: " + e.Error())
		imgTimeChan <- 0
		return
	}
	imgJSONode := doc.Find("#initData")
	if imgJSONode == nil {
		imgTimeChan <- 0
		return
	}
	imgJSON := imgJSONode.Text()
	resultImages := &AnwserImage{}
	err := json.Unmarshal([]byte(imgJSON), resultImages)
	if err != nil {
		log.Println("json error: " + err.Error())
		imgTimeChan <- 0
		return
	}
	// tx2 := time.Now()
	// log.Printf("Searching img time: %d ms\n", tx2.Sub(tx1).Nanoseconds()/1e6)

	//filter portrait/no-small images
	images := make([]Image, 0)
	for _, img := range resultImages.List {
		width, _ := strconv.Atoi(img.Width)
		height, _ := strconv.Atoi(img.Height)
		if width > height && height > 200 {
			images = append(images, img)
			if len(images) >= 5 {
				break
			}
		}
	}
	if len(images) < 0 {
		log.Println("..... not enough image result.")
		imgTimeChan <- 0
		return
	}

	rawImgReader := make(chan io.ReadCloser)
	done := false
	for _, img := range images {
		url := img.Img
		go func(c chan io.ReadCloser) {
			response, e := client.Get(url)
			if e != nil {
				return
			} else if response != nil && response.StatusCode >= 200 && response.StatusCode < 299 {
				if done {
					return
				}
				c <- response.Body
				done = true
			} else {
				log.Println("Get quiz img http request err: " + strconv.Itoa(response.StatusCode))
			}
		}(rawImgReader)
	}

	//open a file for writing
	file, err := os.Create("./lpsolver/dist/assets/quiz.jpg")
	if err != nil {
		log.Println("Create file error: " + err.Error())
		imgTimeChan <- 0
		return
	}
	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, <-rawImgReader)
	if err != nil {
		log.Println("Copy img error: " + err.Error())
		imgTimeChan <- 0
		return
	}
	// close(rawImgReader)
	file.Close()
	// log.Printf("Total img save time: %d ms\n", time.Now().Sub(tx1).Nanoseconds()/1e6)
	imgTimeChan <- time.Now().UTC().Unix()
}

//MatchInfo ...
type MatchInfo struct {
	Success bool `json:"success"`
	Matches []struct {
		MatchID            string `json:"matchId"`
		MatchIDNumber      int    `json:"matchIdNumber"`
		MatchURL           string `json:"matchURL"`
		CreateDate         string `json:"createDate"`
		UpdateDate         string `json:"updateDate"`
		MatchStatus        int    `json:"matchStatus"`
		CurrentPlayerIndex int    `json:"currentPlayerIndex"`
		Letters            string `json:"letters"`
		RowCount           int    `json:"rowCount"`
		ColumnCount        int    `json:"columnCount"`
		TurnCount          int    `json:"turnCount"`
		MatchData          string `json:"matchData"`
		ServerData         struct {
			Language  int   `json:"language"`
			UsedTiles []int `json:"usedTiles"`
			Tiles     []struct {
				T string `json:"t"`
				O int    `json:"o"`
			} `json:"tiles"`
			UsedWords  []string `json:"usedWords"`
			MinVersion int      `json:"minVersion"`
		} `json:"serverData"`
		Participants []struct {
			UserID                string      `json:"userId"`
			UserName              string      `json:"userName"`
			PlayerIndex           int         `json:"playerIndex"`
			PlayerStatus          string      `json:"playerStatus"`
			LastTurnStatus        string      `json:"lastTurnStatus"`
			MatchOutcome          string      `json:"matchOutcome"`
			TurnDate              string      `json:"turnDate"`
			TimeoutDate           interface{} `json:"timeoutDate"`
			AvatarURL             string      `json:"avatarURL"`
			IsFavorite            bool        `json:"isFavorite"`
			UseBadWords           bool        `json:"useBadWords"`
			BlockChat             bool        `json:"blockChat"`
			DeletedFromPlayerList bool        `json:"deletedFromPlayerList"`
			Online                bool        `json:"online"`
			ChatsUnread           int         `json:"chatsUnread"`
			MuteChat              bool        `json:"muteChat"`
			AbandonedMatch        bool        `json:"abandonedMatch"`
			IsBot                 bool        `json:"isBot"`
			BannedChat            bool        `json:"bannedChat"`
		} `json:"participants"`
	} `json:"matches"`
}

//AnwserImage result from image.so.com
type AnwserImage struct {
	Total int `json:"total"`
	// End       bool   `json:"end"`
	// Sid       string `json:"sid"`
	// Ran       int    `json:"ran"`
	// Ras       int    `json:"ras"`
	// Kn        int    `json:"kn"`
	// Cn        int    `json:"cn"`
	// Gn        int    `json:"gn"`
	// Lastindex int    `json:"lastindex"`
	// Ceg       string `json:"ceg"`
	List []Image `json:"list"`
	// Boxresult bool `json:"boxresult"`
	// Wordguess string `json:"wordguess"`
}

//Image instance from image.so.com result
type Image struct {
	// ID            string `json:"id"`
	// QqfaceDownURL bool   `json:"qqface_down_url"`
	// Downurl       bool   `json:"downurl"`
	// Grpmd5        bool   `json:"grpmd5"`
	// Type          int    `json:"type"`
	// Src           string `json:"src"`
	// Color         int    `json:"color"`
	// Index         int    `json:"index"`
	// Title         string `json:"title"`
	// Litetitle     string `json:"litetitle"`
	Width   string `json:"width"`
	Height  string `json:"height"`
	Imgsize string `json:"imgsize"`
	// Imgtype       string `json:"imgtype"`
	// Key           string `json:"key"`
	// Dspurl        string `json:"dspurl"`
	// Link          string `json:"link"`
	// Source        int    `json:"source"`
	Img       string `json:"img"`
	ThumbBak  string `json:"thumb_bak"`
	Thumb     string `json:"thumb"`
	ThumbBak_ string `json:"_thumb_bak"`
	Thumb_    string `json:"_thumb"`
	// Imgkey        string `json:"imgkey"`
	ThumbWidth int `json:"thumbWidth"`
	// Dsptime       string `json:"dsptime"`
	ThumbHeight int `json:"thumbHeight"`
	// Grpcnt        string `json:"grpcnt"`
	// FixedSize     bool   `json:"fixedSize"`
}
