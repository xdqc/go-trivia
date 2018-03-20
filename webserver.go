package solver

import (
	"encoding/json"
	"fmt"
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

//RunWeb run a webserver
func RunWeb(port string) {
	// The "HandleFunc" method accepts a path and a function as arguments
	// (Yes, we can pass functions as arguments, and even trat them like variables in Go)
	// However, the handler function has to have the appropriate signature (as described by the "handler" function below)
	// http.HandleFunc("/match", handler)

	// After defining our server, we finally "listen and serve" on port 8080
	// The second argument is the handler, which we will come to later on, but for now it is left as nil,
	// and the handler defined above (in "HandleFunc") is used
	// http.ListenAndServe(":"+port, nil)

	r := mux.NewRouter()
	r.HandleFunc("/match", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(matchInfo)
	})

	// Use default options
	handler := cors.Default().Handler(r)
	http.ListenAndServe(":"+port, handler)

}

// "handler" is our handler function. It has to follow the function signature of a ResponseWriter and Request type
// as the arguments.
func handler(w http.ResponseWriter, r *http.Request) {
	// For this case, we will always pipe "Hello World" into the response writer

	fmt.Fprintf(w, "Hello solver!\n")
	matchList := getMatch().Matches
	for _, m := range matchList {
		fmt.Fprintf(w, m.Letters)
		fmt.Fprintf(w, m.ServerData.Tiles[0].T)
	}
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
