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
)

func init() {
	flag.IntVar(&mode, "m", 0, "run mode 0 : default mode, easy to be detected of cheating; 1 : invisible mode")
	flag.IntVar(&automatic, "a", 0, "run automatic  0 : manual  1 : automatic")
	flag.Parse()
}

func main() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		solver.Run("8998", mode)
	}()
	go func() {
		solver.RunWeb("8080")
	}()
	<-c
	solver.Close()
}
