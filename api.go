package solver

import (
	"net/http"
	"net/url"
	"regexp"
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
	totalCount := 0
	for _, option := range options {
		res[option] = strings.Count(str, option)
		totalCount += res[option]
	}

	// if all option got 0 match, search the each option.trimLastChar (xx省 -> xx)
	if totalCount == 0 {
		for _, option := range options {
			res[option] = strings.Count(str, option[:len(option)-1])
		}
	}

	// add option count to its superstring option count （红色 add to 红色变无色）
	for _, opt := range options {
		for _, subopt := range options {
			if opt != subopt && strings.Contains(opt, subopt) {
				res[opt] += res[subopt]
			}
		}
	}

	// For negative quiz, flip the count to negative number (dont flip quoted negative word)
	re := regexp.MustCompile("「.*不.*」")
	if strings.Contains(quiz, "不") && !re.MatchString(quiz) {
		for _, option := range options {
			res[option] = -res[option]
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
