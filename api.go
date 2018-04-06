package solver

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	googleUrl = "https://www.google.com/search?"
	baidu_url = "http://www.baidu.com/s?"
)

func GetFromApi(quiz string, options []string) map[string]int {
	res := make(map[string]int, len(options))
	for _, option := range options {
		res[option] = 0
	}

	sg := make(chan string)
	sb := make(chan string)

	go getFromGoogle(quiz, options, sg)
	go getFromBaidu(quiz, options, sb)

	println("\n.......................here................................\n")
	str := <-sg + <-sb
	println("str:\n" + str)
	for _, option := range options {
		res[option] = strings.Count(str, option)
	}

	//add option count to its superstring option count
	for _, opt := range options {
		for _, subopt := range options {
			if opt != subopt && strings.Contains(opt, subopt) {
				res[opt] += res[subopt]
			}
		}
	}
	return res
}

func getFromBaidu(quiz string, options []string, c chan string) {
	values := url.Values{}
	values.Add("wd", quiz)
	req, _ := http.NewRequest("GET", baidu_url+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)
	if resp == nil {
		c <- ""
	} else {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		c <- doc.Find("#content_left .t").Text() + doc.Find(".c-abstract").Text()
	}

}

func getFromGoogle(quiz string, options []string, c chan string) {
	values := url.Values{}
	values.Add("q", quiz)
	req, _ := http.NewRequest("GET", googleUrl+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)
	if resp == nil {
		c <- ""
	} else {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		str := doc.Find(".r").Text() + doc.Find(".st").Text()
		c <- str
	}
}
