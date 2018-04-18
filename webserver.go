package solver

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
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
	r.HandleFunc("/match", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(matchInfo)
	}).Methods("GET")

	r.HandleFunc("/words", findWords).Methods("GET")
	r.HandleFunc("/word", deleteWord).Methods("DELETE")

	r.HandleFunc("/answer", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(questionInfo)
	}).Methods("GET")

	// r.HandleFunc("/idiom", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.Write(idiomInfo)
	// }).Methods("GET")

	r.HandleFunc("/brain-ocr", func(w http.ResponseWriter, r *http.Request) {
		handleQuestionResp([]byte{})
	}).Methods("PUT")

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

func fetchAnswerImage(ans string, quiz []string, quoted string) {
	tx1 := time.Now()
	values := url.Values{}
	searchStr := ans + " " + quoted
	re := regexp.MustCompile("[^\\p{Han}]+")
	hanRunes := re.ReplaceAllString(searchStr, "")
	if len([]rune(hanRunes)) < 2 && len(quiz) > 2 {
		searchStr += " " + strings.Join(quiz[:3], " ")
	}
	values.Add("q", searchStr)
	req, _ := http.NewRequest("GET", "http://image.so.com/i?"+values.Encode(), nil) //www.bing.com/images/search?
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if resp != nil {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		imgJSON := doc.Find("#initData").Text()
		images := &AnwserImage{}
		err := json.Unmarshal([]byte(imgJSON), images)
		tx2 := time.Now()
		log.Printf("Searching img time: %d ms\n", tx2.Sub(tx1).Nanoseconds()/1e6)
		if err == nil {
			if len(images.List) > 10 {
				// set timeout for http GET
				timeout := time.Duration(2 * time.Second)
				client := http.Client{
					Timeout: timeout,
				}
				rawImgReader := make(chan io.ReadCloser)
				for _, img := range images.List[0:10] {
					url := img.Thumb
					go func(c chan io.ReadCloser) {
						response, e := client.Get(url)
						if e != nil {
							log.Println(e.Error())
							return
						}
						if response != nil && response.StatusCode >= 200 && response.StatusCode < 299 {
							c <- response.Body
						}
						return
					}(rawImgReader)
				}

				//open a file for writing
				file, err := os.Create("./lpsolver/dist/assets/quiz.jpg")
				if err != nil {
					log.Println(err.Error())
				}
				// Use io.Copy to just dump the response body to the file. This supports huge files
				_, err = io.Copy(file, <-rawImgReader)
				if err != nil {
					log.Println(err.Error())
					return
				}
				// close(rawImgReader)
				file.Close()
				log.Printf("Total img save time: %d ms\n", time.Now().Sub(tx1).Nanoseconds()/1e6)
			}
		}
	}
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

// Image result from image.so.com
type AnwserImage struct {
	Total     int    `json:"total"`
	End       bool   `json:"end"`
	Sid       string `json:"sid"`
	Ran       int    `json:"ran"`
	Ras       int    `json:"ras"`
	Kn        int    `json:"kn"`
	Cn        int    `json:"cn"`
	Gn        int    `json:"gn"`
	Lastindex int    `json:"lastindex"`
	Ceg       string `json:"ceg"`
	List      []struct {
		ID            string `json:"id"`
		QqfaceDownURL bool   `json:"qqface_down_url"`
		Downurl       bool   `json:"downurl"`
		Grpmd5        bool   `json:"grpmd5"`
		Type          int    `json:"type"`
		Src           string `json:"src"`
		Color         int    `json:"color"`
		Index         int    `json:"index"`
		Title         string `json:"title"`
		Litetitle     string `json:"litetitle"`
		Width         string `json:"width"`
		Height        string `json:"height"`
		Imgsize       string `json:"imgsize"`
		Imgtype       string `json:"imgtype"`
		Key           string `json:"key"`
		Dspurl        string `json:"dspurl"`
		Link          string `json:"link"`
		Source        int    `json:"source"`
		Img           string `json:"img"`
		ThumbBak      string `json:"thumb_bak"`
		Thumb         string `json:"thumb"`
		ThumbBak_     string `json:"_thumb_bak"`
		Thumb_        string `json:"_thumb"`
		Imgkey        string `json:"imgkey"`
		ThumbWidth    int    `json:"thumbWidth"`
		Dsptime       string `json:"dsptime"`
		ThumbHeight   int    `json:"thumbHeight"`
		Grpcnt        string `json:"grpcnt"`
		FixedSize     bool   `json:"fixedSize"`
	} `json:"list"`
	Boxresult bool   `json:"boxresult"`
	Wordguess string `json:"wordguess"`
}
