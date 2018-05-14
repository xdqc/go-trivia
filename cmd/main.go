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
	flag.IntVar(&automatic, "a", 0, "-a 1 adb for android")
	flag.Parse()
}

func main() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		solver.Run("8998", automatic)
	}()
	go func() {
		solver.RunWeb("8080")
	}()
	<-c
	solver.Close()
	return
}
