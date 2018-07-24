package solver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

var (
	memoryDb            *bolt.DB
	QuestionBucket      = "Question"
	WholeQuestionBucket = "WholeQuestion"
	HashQuestionBucket  = "HashQuestion"
)

func initMemoryDb() {
	var err error
	memoryDb, err = bolt.Open("questions.data", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	memoryDb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(HashQuestionBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

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

//StoreHashQuestion store the hash question and hash correct answer to db
func StoreHashQuestion(question *Question) error {
	if question.HashData.TrueAnswer != "" {
		return memoryDb.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(HashQuestionBucket))
			err := b.Put([]byte(question.HashData.Quiz), []byte(question.HashData.TrueAnswer))
			return err
		})
	}
	return nil
}

//FetchQuestion get question answer of a given quiz
func FetchQuestion(question string) (str string) {
	memoryDb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(QuestionBucket))
		v := b.Get([]byte(question))
		if len(v) == 0 {
			return nil
		}
		q := DecodeQuestionCols(v, time.Now().Unix())
		str = q.Answer
		return nil
	})
	return
}

//FetchHashQuestion get question answer hash of a given quiz hash
func FetchHashQuestion(questionHash string) (answerHash string) {
	memoryDb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(HashQuestionBucket))
		v := b.Get([]byte(questionHash))
		if len(v) == 0 {
			return nil
		}
		answerHash = string(v)
		return nil
	})
	return
}

//FetchRandomQuestion get a random whole question
func FetchRandomQuestion(topic string) (question *Question) {
	question = &Question{}
	kv := make(map[string][]byte)
	kvt := make(map[string][]byte)
	memoryDb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(WholeQuestionBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			kv[string(k)] = v
			if topic != "" && (strings.Contains(string(k), topic) || strings.Contains(string(v), "\""+topic+"\"")) {
				kvt[string(k)] = v
			}
		}
		// use topiced keys
		if len(kvt) > 0 {
			kv = kvt
		}

		// get random key of the map
		i := rand.Intn(len(kv))
		n := i
		var k string
		for k = range kv {
			if i == 0 {
				break
			}
			i--
		}
		v := kv[k]

		fmt.Printf("%d/%d	%v\n", n, len(kv), k)

		var wq = &WholeQuestionCols{}
		err := json.Unmarshal(v, wq)
		if err == nil {
			question.Data.Quiz = string(k)
			question.Data.School = wq.School
			question.Data.Type = wq.Type
			question.Data.Options = wq.Options
			question.CalData.Answer = wq.Answer
		} else {
			log.Println(err.Error())
		}
		return nil
	})
	return
}

func ShowAllQuestions() {
	var kv = map[string]string{}
	memoryDb.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(HashQuestionBucket))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			// fmt.Printf("key=%s, value=%s\n", k, v)
			kv[string(k)] = string(v)
		}

		bs := new(bytes.Buffer)
		fmt.Fprintf(bs, "[")
		for key, value := range kv {
			fmt.Fprintf(bs, "{\"%s\":\"%s\"},\n", key, value)
		}
		fmt.Fprintf(bs, "]")
		ioutil.WriteFile("wq.json", bs.Bytes(), 0600)
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
