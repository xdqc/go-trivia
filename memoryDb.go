package solver

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/boltdb/bolt"
)

var (
	memoryDb            *bolt.DB
	QuestionBucket      = "Question"
	WholeQuestionBucket = "WholeQuestion"
)

func init() {
	var err error
	memoryDb, err = bolt.Open("questions.data", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	memoryDb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(QuestionBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	memoryDb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(WholeQuestionBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}

func StoreQuestion(question *Question) error {
	if question.CalData.TrueAnswer != "" {
		return memoryDb.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(QuestionBucket))
			v := NewQuestionCols(question.CalData.TrueAnswer)
			err := b.Put([]byte(question.Data.Quiz), v.GetData())
			return err
		})
	}
	return nil
}

//StoreWholeQuestion store the whole question to db
func StoreWholeQuestion(question *Question) error {
	if question.CalData.TrueAnswer != "" {
		return memoryDb.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(WholeQuestionBucket))
			v := NewWholeQuestionCols(question.Data.Num, question.Data.School, question.Data.Type,
				question.Data.Options, question.CalData.TrueAnswer)
			err := b.Put([]byte(question.Data.Quiz), v.GetData())
			return err
		})
	}
	return nil
}

func FetchQuestion(question *Question) (str string) {
	memoryDb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(QuestionBucket))
		v := b.Get([]byte(question.Data.Quiz))
		if len(v) == 0 {
			return nil
		}
		q := DecodeQuestionCols(v, time.Now().Unix())
		str = q.Answer
		return nil
	})
	return
}

func ShowAllQuestions() {
	var kv = map[string]string{}
	memoryDb.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(QuestionBucket))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
			kv[string(k)] = string(v)
		}
		return nil
	})

}

type QuestionCols struct {
	Answer string `json:"a"`
	Update int64  `json:"ts"`
}

type WholeQuestionCols struct {
	Number  int      `json:"n"`
	School  string   `json:"sch"`
	Type    string   `json:"typ"`
	Options []string `json:"opt"`
	Answer  string   `json:"a"`
	EndTime int      `json:"te"`
	CurTime int      `json:"tc"`
	Update  int64    `json:"ts"`
}

func NewQuestionCols(answer string) *QuestionCols {
	return &QuestionCols{
		Answer: answer,
		Update: time.Now().Unix(),
	}
}

func NewWholeQuestionCols(num int, school string, typ string, options []string,
	answer string) *WholeQuestionCols {
	return &WholeQuestionCols{
		Number:  num,
		School:  school,
		Type:    typ,
		Options: options,
		Answer:  answer,
		Update:  time.Now().Unix(),
	}
}

func DecodeQuestionCols(bs []byte, update int64) *QuestionCols {
	var q = &QuestionCols{}
	err := json.Unmarshal(bs, q)
	if err == nil {
		return q
	} else {
		q = NewQuestionCols(string(bs))
		q.Update = update
	}
	return q
}

func (q *QuestionCols) GetData() []byte {
	bs, _ := json.Marshal(q)
	return bs
}

func (q *WholeQuestionCols) GetData() []byte {
	bs, _ := json.Marshal(q)
	return bs
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
