package solver

import (
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	googleUrl = "http://www.google.com/search?"
	baidu_url = "http://www.baidu.com/s?"
)

func GetFromApi(quiz string, options []string) map[string]int {
	res := make(map[string]int, len(options))
	for _, option := range options {
		res[option] = 0
	}

	sl := make(chan string)
	sg := make(chan string)
	sb := make(chan string)
	sgo := make(chan string)
	sbo := make(chan string)

	go searchFeelingLucky(quiz, options, sl)
	go searchGoogle(quiz, options, sg)
	go searchBaidu(quiz, options, sb)
	go searchGoogleWithOptions(quiz, options, sgo)
	go searchBaiduWithOptions(quiz, options, sbo)

	println("\n.......................searching..............................\n")
	str := <-sg + <-sb + <-sgo + <-sbo + <-sl
	tx := time.Now()
	// println("str:\n" + str)

	for _, option := range options {
		res[option] = strings.Count(str, option)
	}

	// if all option got 0 match, search the each option.trimLastChar (xx省 -> xx)
	// if totalCount == 0 {
	// 	for _, option := range options {
	// 		res[option] = strings.Count(str, option[:len(option)-1])
	// 	}
	// }

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
			res[option] = -res[option] - 1
		}
	}

	tx2 := time.Now()
	log.Printf("process time %d us\n", tx2.Sub(tx).Nanoseconds()/1e3)

	return res
}

func searchBaidu(quiz string, options []string, c chan string) {
	values := url.Values{}
	values.Add("wd", quiz+" site:baidu.com")
	req, _ := http.NewRequest("GET", baidu_url+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if resp == nil {
		c <- ""
	} else {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		text := doc.Find("#content_left .t").Text() + doc.Find("#content_left .c-abstract").Text() + doc.Find("#content_left .m").Text() //.m ~zhidao
		c <- text + text + text                                                                                                          // 3x weight
	}
}

func searchBaiduWithOptions(quiz string, options []string, c chan string) {
	values := url.Values{}
	values.Add("wd", quiz+" "+strings.Join(options, " "))
	req, _ := http.NewRequest("GET", baidu_url+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if resp == nil {
		c <- ""
	} else {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		c <- doc.Find("#content_left .t").Text() + doc.Find("#content_left .c-abstract").Text()
	}
}

func searchGoogle(quiz string, options []string, c chan string) {
	values := url.Values{}
	values.Add("q", quiz)
	req, _ := http.NewRequest("GET", googleUrl+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if resp == nil {
		c <- ""
	} else {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		str := doc.Find(".r").Text() + doc.Find(".st").Text() + doc.Find(".P1usbc").Text() //.P1usbc ~wiki
		c <- str
	}
}

func searchGoogleWithOptions(quiz string, options []string, c chan string) {
	values := url.Values{}
	values.Add("q", quiz+" \""+strings.Join(options, "\" OR \"")+"\"")
	req, _ := http.NewRequest("GET", googleUrl+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if resp == nil {
		c <- ""
	} else {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		text := doc.Find(".r").Text() + doc.Find(".st").Text() + doc.Find(".P1usbc").Text() //.P1usbc ~wiki
		c <- text + text + text                                                             // 3x weight
	}
}

func searchFeelingLucky(quiz string, options []string, c chan string) {
	values := url.Values{}
	values.Add("q", quiz)
	values.Add("btnI", "") //click I'm feeling lucky! button
	req, _ := http.NewRequest("GET", googleUrl+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	// bs, _ := ioutil.ReadAll(resp.Body)
	// log.Println("-------- luck :  \n" + string(bs)[:1000])
	log.Println("-------- luck url:  \n" + resp.Request.URL.Host + resp.Request.URL.RawPath + " /// " + resp.Request.Host)
	if resp == nil || resp.Request.Host == "www.google.com" {
		c <- ""
	} else if resp.Request.URL.Host == "zh.wikipedia.org" {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		text := doc.Find(".mw-parser-output").Text()
		log.Println(text)
		c <- text
	} else if resp.Request.URL.Host == "baike.baidu.com" {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		text := doc.Find(".para").Text() + doc.Find(".basicInfo-item").Text()
		log.Println(text)
		c <- text
	}
}
