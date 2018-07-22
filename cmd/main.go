package main

import (
	"encoding/base64"
	"flag"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

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

func shuffle(src []rune) []rune {
	final := make([]rune, len(src))
	rand.Seed(time.Now().UTC().UnixNano())
	perm := rand.Perm(len(src))

	for i, v := range perm {
		final[v] = src[i]
	}
	return final
}

func testBase64Decode() {
	const letterBytes = "0123456789+/abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const encrypted = "fJfUdBWDRyUNhESisKvYSPb529zmGAAYAeIuDBaIxkdw92EJCy/ivmDfU0Ix0gshq2eC57doZLxra+LFZo0dQ5wNdeCcwg86YLhCp2JaaM8Hd+9j1FyPwadP0uGLmHD1o7cEAWWfH2JYD+2g69+Lj1mksuZMmdOjz6cIp6Bp3wJRvMutaghBZJf5W3PTLt//A9HLKlnsvMjmWP8qYP8FLa1eKZRBmCCwTlN7lh2hpsEPlWsjyzt6PTMA7++Z6NJDmOVHNNLOzuilLVGFcr3YG77PnHq9Fk5X/A7aD7RtjJ7zMmZD00pKe1pzi/+vEXi6vr9sYzWVjtcJXu2XvBzkfUXOC+0CBymDindUHo4m40GlDAJwKyleYohsYWqGpz5uzHe1BGpB0YP5553JUGyo1SXzCWcXqTASz4sOi1Rr1Yri/IAAHmr6EQzDFZuquQ7A1qGic532W5qe/Fnma3LWFw==" //《春江花月夜》的作者是？
	var wg sync.WaitGroup
	for index := 0; index < 3000000; index++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			encodeStr := string(shuffle([]rune(letterBytes)))
			encoding := base64.NewEncoding(encodeStr)
			decrypted, _ := encoding.DecodeString(encrypted)
			decryptedStr := string(decrypted)
			if strings.Contains(decryptedStr, "作者") {
				println(encodeStr)
				println(decryptedStr)
			}
		}()
	}
	wg.Wait()
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
