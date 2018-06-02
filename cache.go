package solver

import (
	"log"
	"regexp"
	"time"

	cache "github.com/patrickmn/go-cache"
)

var (
	questions       *cache.Cache
	quizContexts    *cache.Cache
	quizContextInfo []byte
)

func init() {
	// Create a cache with a default expiration time of 5 minutes, and which
	// purges expired items every 10 minutes
	questions = cache.New(5*time.Minute, 10*time.Minute)
	quizContexts = cache.New(5*time.Minute, 180*time.Minute)
}

func GetQuestion(roomID, quizNum string) *Question {
	key := roomID + "_" + quizNum
	if entity, ok := questions.Get(key); ok {
		return entity.(*Question)
	}
	return nil
}

func SetQuestion(question *Question) {
	key := question.CalData.RoomID + "_" + question.CalData.quizNum
	questions.Set(key, question, cache.DefaultExpiration)
}

func getQuizContext() *QuizContext {
	log.Println("Cached ctx size: ", quizContexts.ItemCount())
	if quizContexts.ItemCount() > 0 {
		for _, item := range quizContexts.Items() {
			qc := item.Object.(*QuizContext)
			quizContexts.Delete(qc.Quiz)
			return qc
		}
	}
	return &QuizContext{}
}

func setQuizContext(quizContext *QuizContext) {
	if entity, ok := quizContexts.Get(quizContext.Quiz); ok {
		//append words to same quiz
		words := entity.(*QuizContext).Words
		words = append(words, quizContext.Words...)
		quizContext.Words = words
	}
	quizContexts.Set(quizContext.Quiz, quizContext, 30*time.Minute)

}

// Create stream around the context sliding window around option
func cacheQuizContext(quiz string, ctx string) {
	cuts := JB.Cut(ctx, true)
	words := make([]string, 0)
	re := regexp.MustCompile("[^\\p{Han}]+")
	for _, w := range cuts {
		if w == " " || re.MatchString(w) {
			continue
		}
		words = append(words, w)
	}
	if len(words) < 2 {
		return
	}
	// trim start and end word (could be partial)
	words = words[1 : len(words)-1]

	quizContext := &QuizContext{}
	quizContext.Words = words
	quizContext.Quiz = quiz

	setQuizContext(quizContext)
}

//QuizContext ...
type QuizContext struct {
	Words []string `json:"words"`
	Quiz  string   `json:"quiz"`
}
