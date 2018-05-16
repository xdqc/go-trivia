package solver

import (
	"log"
	"time"

	cache "github.com/patrickmn/go-cache"
)

var (
	questions    *cache.Cache
	quizContexts *cache.Cache
)

func init() {
	// Create a cache with a default expiration time of 5 minutes, and which
	// purges expired items every 10 minutes
	questions = cache.New(5*time.Minute, 10*time.Minute)
	quizContexts = cache.New(60*time.Second, 600*time.Second)
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
	quizContexts.Set(quizContext.Quiz, quizContext, time.Second*600)

}
