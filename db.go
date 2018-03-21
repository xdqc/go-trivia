package solver

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var (
	db  *sql.DB
	err error
)

type Word struct {
	id     int    `json:"id"`
	Word   string `json:"word"`
	Length int    `json:"length"`
	A      int    `json:"A"`
	B      int    `json:"B"`
	C      int    `json:"C"`
	D      int    `json:"D"`
	E      int    `json:"E"`
	F      int    `json:"F"`
	G      int    `json:"G"`
	H      int    `json:"H"`
	I      int    `json:"I"`
	J      int    `json:"J"`
	K      int    `json:"K"`
	L      int    `json:"L"`
	M      int    `json:"M"`
	N      int    `json:"N"`
	O      int    `json:"O"`
	P      int    `json:"P"`
	Q      int    `json:"Q"`
	R      int    `json:"R"`
	S      int    `json:"S"`
	T      int    `json:"T"`
	U      int    `json:"U"`
	V      int    `json:"V"`
	W      int    `json:"W"`
	X      int    `json:"X"`
	Y      int    `json:"Y"`
	Z      int    `json:"Z"`
	Valid  int    `json:"valid"`
}

func init() {
	db, err = sql.Open("mysql", "root:root@tcp(localhost:3306)/qd16")
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("successfully connected to mysql")
}

func selectWords(loBound map[rune]int, hiBound map[rune]int) []string {

	var args []int
	var sqlclause string
	for _, v := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		l, ok := loBound[v]
		if ok {
			args = append(args, l)
		} else {
			args = append(args, 0)
		}
		h, ok := hiBound[v]
		if ok {
			args = append(args, h)
		} else {
			args = append(args, 0)
		}
		sqlclause = sqlclause + "AND " + string(v) + " >= (?) AND " + string(v) + " <= (?) "
	}

	for i, n := range args {
		println(i, n)
	}

	sql := `SELECT word FROM db_english_all_words WHERE valid = 1 ` + sqlclause
	log.Println(sql)
	result, err := db.Query("SELECT word FROM db_english_all_words WHERE valid = 1 AND A >= (?)", 3)
	if err != nil {
		panic(err.Error())
	}

	res := make([]string, 0, 200)
	for result.Next() {
		var word Word
		err = result.Scan(&word.Word)
		if err != nil {
			panic(err.Error())
		}
	}
	return res
}

func readWords() []Word {
	raw, err := ioutil.ReadFile("./english_all_words.json")
	if err != nil {
		fmt.Println("Fatal ", err.Error())
		os.Exit(1)
	}
	var w []Word
	err = json.Unmarshal(raw, &w)
	if err != nil {
		println("umsh err", err.Error())
	}
	fmt.Println("read file done")
	return w
}
