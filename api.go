package solver

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	google_URL = "http://www.google.com/search?"
	baidu_URL  = "http://www.baidu.com/s?"
	so360_URL  = "http://www.so.com/s?"
)

//GetFromAPI searh the quiz via popular search engins
func GetFromAPI(quiz string, options []string) map[string]int {
	res := make(map[string]int, len(options))
	for _, option := range options {
		res[option] = 0
	}

	search := make(chan string, 4)
	done := make(chan bool, 1)
	tx := time.Now()

	go searchFeelingLucky(quiz, options, search)
	go searchGoogle(quiz, options, search)
	go searchBaidu(quiz, options, search)
	go searchGoogleWithOptions(quiz, options, search)

	println("\n.......................searching..............................\n")
	rawStr := "                                        "
	count := cap(search)
	go func() {
		for {
			s, more := <-search
			if more {
				log.Println("search received...", s[:7])
				rawStr += s[7:]
				count--
				if count == 0 {
					done <- true
					return
				}
			}
		}
	}()
	select {
	case <-done:
		fmt.Println("search done")
	case <-time.After(2 * time.Second):
		fmt.Println("search timeout")
	}
	rawStr += "                                        "
	tx2 := time.Now()
	log.Printf("Searching time: %d ms\n", tx2.Sub(tx).Nanoseconds()/1e6)

	// sliding window, count the common chars between [neighbor of the option in search text] and [quiz]
	CountMatches(quiz, options, rawStr, res)

	// if all option got 0 match, search the each option.trimLastChar (xx省 -> xx)
	// if totalCount == 0 {
	// 	for _, option := range options {
	// 		res[option] = strings.Count(str, option[:len(option)-1])
	// 	}
	// }

	// For no-number option, add count to its superstring option count （米波 add to 毫米波)
	re := regexp.MustCompile("[\\d]+")
	for _, opt := range options {
		if !re.MatchString(opt) {
			for _, subopt := range options {
				if opt != subopt && strings.Contains(opt, subopt) {
					res[opt] += res[subopt]
				}
			}
		}
	}

	// For negative quiz, flip the count to negative number (dont flip quoted negative word)
	re = regexp.MustCompile("「[^」]*[不][^」]*」")
	nonegreg := regexp.MustCompile("不[同充分对称足够断停止得太值锈]")
	if (strings.Contains(quiz, "不") || strings.Contains(quiz, "没有") || strings.Contains(quiz, "未在")) &&
		!(nonegreg.MatchString(quiz) || re.MatchString(quiz)) {
		for _, option := range options {
			res[option] = -res[option] - 1
		}
	}

	tx3 := time.Now()
	log.Printf("Processing time %d ms\n", tx3.Sub(tx2).Nanoseconds()/1e6)

	return res
}

//CountMatches sliding window, count the common chars between [neighbor of the option in search text] and [quiz]
func CountMatches(quiz string, options []string, rawStr string, res map[string]int) {
	hasQuote := strings.Contains(quiz, "「")
	quoted := ""
	if hasQuote {
		quoted = quiz[strings.Index(quiz, "「"):strings.Index(quiz, "」")]
	}
	// filter out non alphanumeric/chinese/space
	re := regexp.MustCompile("[^\\w\\p{Han} ]+")
	str := re.ReplaceAllString(rawStr, "")
	println(str)
	qz := re.ReplaceAllString(quiz, "")

	// width := len([]rune(qz))
	// if width > 40 {
	width := 30 //max window size
	// }

	// Only match the important part of quiz, in neighbors of options
	if strings.Index(qz, "最") > 0 && strings.Index(qz, "最") < len(qz)-1 {
		qz = qz[strings.Index(qz, "最"):]
	} else if strings.Index(qz, "属于") > 0 && strings.Index(qz, "属于") < len(qz)-1 {
		qz = qz[strings.Index(qz, "属于"):]
	} else if strings.Index(qz, "中") > 0 && strings.Index(qz, "中") < len(qz)-1 {
		qz = qz[strings.Index(qz, "中"):]
	} else if strings.Index(qz, "的") > 0 && strings.Index(qz, "的") < len(qz)-1 && !hasQuote {
		qz = qz[strings.Index(qz, "的"):]
	}

	for _, option := range options {
		opti := option
		if strings.Index(option, "·") > 0 {
			opti = option[strings.Index(option, "·")+1:] //only match last name
		}
		opti = re.ReplaceAllString(opti, "")
		opt := []rune(opti)
		optLen := len(opt)
		strs := []rune(str)
		for i := range strs[0 : len(strs)-40] {
			// find the index of option in the search text
			if string(strs[i:i+optLen]) == opti {
				windowR := strs[i+len(opt) : i+len(opt)+width]
				windowL := strs[i-width : i]
				// Reverse windowL
				windowLr := func(s []rune) []rune {
					for l, r := 0, len(s)-1; l < r; l, r = l+1, r-1 {
						s[l], s[r] = s[r], s[l]
					}
					return s
				}(windowL)
				// Evaluate pts of each window. Quiz the closer to option, the high points (gaussian distribution)
				if !(strings.Contains(qz, "上一") || strings.Contains(qz, "之前")) {
					for j, ch := range windowLr {
						if ch == 'A' || ch == 'B' || ch == 'C' || ch == 'D' {
							// stop match ABCD choices
							break
						} else if ch == '的' {
							continue
						}
						if strings.ContainsRune(qz, ch) {
							res[option] += int(200 * math.Exp(-math.Pow(float64(j)/float64(width), 2)/0.1)) //e^(-x^2), sigma=0.1, factor=200
						}
						if hasQuote && strings.ContainsRune(quoted, ch) {
							res[option] += 200
						}
					}
				}
				if !(strings.Contains(qz, "下一") || strings.Contains(qz, "之后")) {
					for j, ch := range windowR {
						if ch == '的' {
							continue
						}
						if strings.ContainsRune(qz, ch) {
							res[option] += int(100 * math.Exp(-math.Pow(float64(j)/float64(width), 2)/0.2)) //e^(-x^2), sigma=0.2, factor=100
						}
						if hasQuote && strings.ContainsRune(quoted, ch) {
							res[option] += 200
						}
					}
				}
				fmt.Printf("%s%6d\t%40s %40s\n", option, res[option], string(windowL), string(windowR))
			}
		}
	}
}

func searchBaidu(quiz string, options []string, c chan string) {
	values := url.Values{}
	values.Add("wd", quiz+" site:baidu.com")
	req, _ := http.NewRequest("GET", baidu_URL+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	text := "baidu  "
	if resp != nil {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		text += doc.Find("#content_left .t").Text() + doc.Find("#content_left .c-abstract").Text() + doc.Find("#content_left .m").Text() //.m ~zhidao
	}
	c <- text // 2x weight
}

func searchBaiduWithOptions(quiz string, options []string, c chan string) {
	values := url.Values{}
	values.Add("wd", quiz+" "+strings.Join(options, " "))
	req, _ := http.NewRequest("GET", baidu_URL+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	text := "baiOpt "
	if resp != nil {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		text += doc.Find("#content_left .t").Text() + doc.Find("#content_left .c-abstract").Text()
	}
	c <- text
}

func searchGoogle(quiz string, options []string, c chan string) {
	values := url.Values{}
	values.Add("q", quiz)
	req, _ := http.NewRequest("GET", google_URL+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	text := "google "
	if resp != nil {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		text += doc.Find(".r").Text() + doc.Find(".st").Text() + doc.Find(".P1usbc").Text() //.P1usbc ~wiki
	}
	c <- text
}

func searchGoogleWithOptions(quiz string, options []string, c chan string) {
	values := url.Values{}
	values.Add("q", quiz+" \""+strings.Join(options, "\" OR \"")+"\"")
	req, _ := http.NewRequest("GET", google_URL+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	text := "gooOpt "
	if resp != nil {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		text += doc.Find(".r").Text() + doc.Find(".st").Text() + doc.Find(".P1usbc").Text() //.P1usbc ~wiki
	}
	c <- text // 2x weight
}

func searchFeelingLucky(quiz string, options []string, c chan string) {
	values := url.Values{}
	values.Add("q", quiz)
	values.Add("btnI", "") //click I'm feeling lucky! button
	req, _ := http.NewRequest("GET", google_URL+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	log.Println("-------- luck url:  " + resp.Request.URL.Host + resp.Request.URL.Path + " /// " + resp.Request.Host)
	text := "Lucky  "
	if resp == nil || resp.Request.Host == "www.google.com" {

	} else if resp.Request.URL.Host == "zh.wikipedia.org" {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		text += doc.Find(".mw-parser-output").Text()
	} else if resp.Request.URL.Host == "baike.baidu.com" {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		text += doc.Find(".para").Text() + doc.Find(".basicInfo-item").Text()
	} else if resp.Request.URL.Host == "wiki.mbalib.com" {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		text += doc.Find("#bodyContent").Text()
	} else {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		text += doc.Find("body").Text()
		// log.Println(text)
	}
	if len(text) > 5000 {
		text = text[:5000]
	}
	c <- text
}
