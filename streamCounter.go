package solver

var quizContextInfo []byte

// Create stream around the context sliding window around option
func cacheQuizContext(quiz string, ctx string) {
	cuts := JB.Cut(ctx, true)
	words := make([]string, 0)
	for _, w := range cuts {
		if w == "" {
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
