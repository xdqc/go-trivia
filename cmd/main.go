package main

import (
	"flag"
	"os"
	"os/signal"

	solver "github.com/xdqc/letterpress-solver"
)

var (
	mode      int
	automatic int
	hashquiz  int
)

func init() {
	flag.IntVar(&automatic, "a", 0, "-a 1 adb for android")
	flag.IntVar(&hashquiz, "h", 0, "-h 1 store hashed image for quiz and answer")
	flag.Parse()
}

func main() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		solver.Run("8998", automatic, hashquiz)
	}()
	go func() {
		solver.RunWeb("8080")
	}()
	<-c
	solver.Close()
	return
}
