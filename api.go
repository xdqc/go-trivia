package solver

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/yanyiwu/gojieba"
)

var (
	google_URL = "http://www.google.com/search?"
	baidu_URL  = "http://www.baidu.com/s?"
	so360_URL  = "http://www.so.com/s?"
	J          *gojieba.Jieba
)

//GetFromAPI searh the quiz via popular search engins
func GetFromAPI(quiz string, options []string) map[string]int {
	tx21 := time.Now()
	J = gojieba.NewJieba()
	tx22 := time.Now()
	log.Printf("init jieba time: %d ms\n", tx22.Sub(tx21).Nanoseconds()/1e6)

	res := make(map[string]int, len(options))
	for _, option := range options {
		res[option] = 0
	}
	search := make(chan string, 5)
	done := make(chan bool, 1)
	tx := time.Now()

	go searchFeelingLucky(quiz, options, 0, search)
	// go searchFeelingLucky(quiz, options, 1, search)
	// go searchFeelingLucky(quiz, options, 2, search)
	// go searchFeelingLucky(quiz, options, 3, search)
	// go searchFeelingLucky(quiz, options, 4, search)
	go searchGoogle(quiz, options, search)
	go searchBaidu(quiz, options, search)
	go searchGoogleWithOptions(quiz, options, search)
	go searchBaiduWithOptions(quiz, options, search)

	println("\n.......................searching..............................\n")
	rawStr := "                                        "
	count := cap(search)
	go func() {
		for {
			s, more := <-search
			if more {
				// First 7 chars in text is the identifier of the search source
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

	// filter out non alphanumeric/chinese/space
	re := regexp.MustCompile("[^\\w\\p{Han} ]+")
	qz := re.ReplaceAllString(quiz, "")
	words := J.CutForSearch(qz, true)
	var keywords []string
	for _, w := range words {
		if !(strings.ContainsAny(w, " 的哪是") || w == "下列" || w == "可以" || w == "什么" || w == "选项" || w == "属于") {
			println(w)
			keywords = append(keywords, w)
		}
	}
	// sliding window, count the common chars between [neighbor of the option in search text] and [quiz]
	CountMatches(quiz, options, rawStr, res)

	// if all option got 0 match, search the each option.trimLastChar (xx省 -> xx)
	// if totalCount == 0 {
	// 	for _, option := range options {
	// 		res[option] = strings.Count(str, option[:len(option)-1])
	// 	}
	// }

	// For no-number option, add count to its superstring option count （米波 add to 毫米波)
	re = regexp.MustCompile("[\\d]+")
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
	nonegreg := regexp.MustCompile("不[同充分对称足够断停止得太值敢锈]")
	if (strings.Contains(quiz, "不") || strings.Contains(quiz, "没有") || strings.Contains(quiz, "未在") || strings.Contains(quiz, "错字") || strings.Contains(quiz, "无关")) &&
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
	hasQuote := strings.ContainsRune(quiz, '「')
	quoted := ""
	if hasQuote {
		quoted = quiz[strings.IndexRune(quiz, '「'):strings.IndexRune(quiz, '」')]
		log.Println("quoted part of quiz: ", quoted)
	}
	// filter out non alphanumeric/chinese/space
	re := regexp.MustCompile("[^\\w\\p{Han} ]+")
	str := re.ReplaceAllString(rawStr, "")
	println(str)
	strs := []rune(str)
	qz := re.ReplaceAllString(quiz, "")

	// width := len([]rune(qz))
	// if width > 40 {
	width := 30 //max window size
	// }

	// Only match the important part of quiz, in neighbors of options
	if !hasQuote {
		if strings.IndexRune(qz, '最') >= 0 && strings.IndexRune(qz, '最') < len(qz)-4 {
			qz = qz[strings.IndexRune(qz, '最'):]
			// } else if strings.Index(qz, "属于") >= 0 && strings.Index(qz, "属于") < len(qz)-4 {
			// 	qz = qz[strings.Index(qz, "属于"):]
		} else if strings.IndexRune(qz, '中') >= 0 && strings.IndexRune(qz, '中') < len(qz)-4 {
			qz = qz[strings.IndexRune(qz, '中'):]
			// } else if strings.Index(qz, "的") > 0 && strings.Index(qz, "的") < len(qz)-4 && !hasQuote {
			// 	qz = qz[strings.Index(qz, "的"):]
		}
	}
	log.Println("truncated qz: \t" + qz)

	var optCounts [4]int
	plainQuizCount := 0

	for k, option := range options {
		opti := option
		if strings.IndexRune(option, '·') > 0 {
			opti = option[strings.IndexRune(option, '·')+1:] //only match last name
			log.Println("last name of ", option, " : ", opti)
		}
		opti = re.ReplaceAllString(opti, "")
		opt := []rune(opti)
		optLen := len(opt)
		optCount := 1
		var optMatches []int
		optMatches = append(optMatches, 0)
		for i, r := range strs {
			if r == ' ' {
				continue
			}
			// find the index of option in the search text
			if string(strs[i:i+optLen]) == opti {
				optCount++
				optMatch := 0
				// create aother slice of runes, avoiding mess with strs
				wstrs := []rune(str)
				windowR := wstrs[i+len(opt) : i+len(opt)+width]
				windowL := wstrs[i-width : i]
				// Reverse windowL
				func(s []rune) []rune {
					for l, r := 0, len(s)-1; l < r; l, r = l+1, r-1 {
						s[l], s[r] = s[r], s[l]
					}
					return s
				}(windowL)
				// Evaluate match-points of each window. Quiz the closer to option, the high points (gaussian distribution)
				if !(strings.Contains(qz, "上一") || strings.Contains(qz, "之前")) {
					for j, ch := range windowL {
						if ch == 'A' || ch == 'B' || ch == 'C' || ch == 'D' {
							// stop match ABCD choices
							plainQuizCount++
							continue
						} else if ch == '的' {
							continue
						}
						if strings.ContainsRune(qz, ch) {
							optMatch += int(100 * math.Exp(-math.Pow(float64(j)/float64(width), 2)/0.1)) //e^(-x^2), sigma=0.1, factor=100
						}
						if hasQuote && strings.ContainsRune(quoted, ch) {
							optMatch += int(200 * math.Exp(-math.Pow(float64(j)/float64(width), 2)/1)) //e^(-x^2), sigma=1, factor=200
						}
					}
				}
				if !(strings.Contains(qz, "下一") || strings.Contains(qz, "之后")) {
					for j, ch := range windowR {
						if ch == 'A' || ch == 'B' || ch == 'C' || ch == 'D' {
							// stop match ABCD choices
							plainQuizCount++
							continue
						} else if ch == '的' {
							continue
						}
						if strings.ContainsRune(qz, ch) {
							optMatch += int(75 * math.Exp(-math.Pow(float64(j)/float64(width), 2)/0.15)) //e^(-x^2), sigma=0.15, factor=75
						}
						if hasQuote && strings.ContainsRune(quoted, ch) {
							optMatch += int(200 * math.Exp(-math.Pow(float64(j)/float64(width), 2)/1)) //e^(-x^2), sigma=1, factor=200
						}
					}
				}
				res[option] += optMatch
				optMatches = append(optMatches, optMatch)
				fmt.Printf("%s%8d%8d\t%35s %35s\n", option, optMatch, res[option], string(windowL), string(windowR))
			}
		}
		optCounts[k] = optCount
		sort.Sort(sort.Reverse(sort.IntSlice(optMatches)))
		//only take first lg(len) number of top matches, sum up as the result of the option
		optMatches = optMatches[0:int(math.Log2(float64(len(optMatches))))]
		matches := 0
		for _, m := range optMatches {
			matches += m
		}
		res[option] = matches
	}

	// if more than half matches are plain quiz, simplely set matches as the count of each option
	sumCounts := 0
	for i := range optCounts {
		sumCounts += optCounts[i]
	}
	if plainQuizCount > sumCounts {
		for i, option := range options {
			res[option] = optCounts[i]
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

func searchFeelingLucky(quiz string, options []string, id int, c chan string) {
	values := url.Values{}
	if id == 0 {
		values.Add("q", quiz)
	} else {
		values.Add("q", options[id-1])
	}
	values.Add("btnI", "") //click I'm feeling lucky! button
	req, _ := http.NewRequest("GET", google_URL+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	log.Println("------------------------- lucky url:  " + resp.Request.URL.Host + resp.Request.URL.Path + " /// " + resp.Request.Host)
	text := ""
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
		//doc, _ := goquery.NewDocumentFromReader(resp.Body)
		//text += doc.Find("body").Text()
		// log.Println(text)
	}

	// For options wiki, if no significant occurence of quiz in text, drop it
	// if id > 0 {
	// 	text := re.ReplaceAllString(text, "")
	// 	numQzMatch := 0
	// 	for _, w := range keywords {
	// 		if strings.Contains(text, w) {
	// 			numQzMatch++
	// 		}
	// 	}
	// 	if float32(numQzMatch)/float32(len(keywords)) < 0.6 {
	// 		text = ""
	// 		log.Printf("Dropped search result of wiki %d : %s\n", id, options[id-1])
	// 	}
	// }

	if len(text) > 10000 {
		text = text[:10000]
	}
	c <- fmt.Sprintf("Lucky %d", id) + text
}
