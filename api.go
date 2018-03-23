package solver

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	googleUrl = "https://www.google.com/search?"
	baidu_url = "http://www.baidu.com/s?"
)

func GetFromBaidu(quiz string, options []string) map[string]int {
	values := url.Values{}
	values.Add("wd", quiz)
	req, _ := http.NewRequest("GET", baidu_url+values.Encode(), nil)
	return GetFromApi(req, quiz, options)
}

func GetFromApi(req *http.Request, quiz string, options []string) (res map[string]int) {
	res = make(map[string]int, len(options))
	for _, option := range options {
		res[option] = 0
	}
	resp, _ := http.DefaultClient.Do(req)
	if resp == nil {
		return
	}
	doc, _ := goquery.NewDocumentFromReader(resp.Body)
	defer resp.Body.Close()
	str := doc.Find("#content_left .result").Text()
	for _, option := range options {
		res[option] = strings.Count(str, option)
	}
	return
}

func GetFromGoogle(quiz string, options []string) (res map[string]int) {
	values := url.Values{}
	values.Add("q", quiz)

	res = make(map[string]int, len(options))
	for _, option := range options {
		res[option] = 0
	}
	resp, _ := http.Get(googleUrl + values.Encode())
	if resp == nil {
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	println(body)

	doc, _ := goquery.NewDocumentFromReader(resp.Body)
	println(doc.Get(0).Data)
	defer resp.Body.Close()
	str := doc.Find(".st").Text()
	for _, option := range options {
		res[option] = strings.Count(str, option)
	}
	return
}
