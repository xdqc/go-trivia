package solver

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

var matchInfo []byte

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

	r.PathPrefix("/solver/").Handler(http.StripPrefix("/solver/", http.FileServer(http.Dir("./lpsolver/dist"))))

	// Use default options
	handler := cors.Default().Handler(r)

	http.ListenAndServe(":"+port, handler)

}

func findWords(w http.ResponseWriter, r *http.Request) {

	minLetters, _ := r.URL.Query()["selected"]
	maxLetters, _ := r.URL.Query()["letters"]
	log.Println("params: ", minLetters[0], maxLetters[0])

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

	var words Words
	res := selectWords(loFreq, hiFreq)

	for i, word := range res {
		words[i].Word = word
	}
	w.Header().Set("Content-Type", "application/json")
	ws, _ := json.Marshal(words)
	w.Write(ws)
}

func getMatch() MatchInfo {
	matches := MatchInfo{}
	if matchInfo != nil {
		err := json.Unmarshal(matchInfo, &matches)
		if err != nil {
			log.Fatal("Error while parse matches info", err)
		}
	}
	return matches
}

func setMatch(jsonBytes []byte) {
	matchInfo = jsonBytes
}
