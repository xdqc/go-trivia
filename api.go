package solver

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/yanyiwu/gojieba"
)

var (
	google_URL = "http://www.google.com/search?"
	baidu_URL  = "http://www.baidu.com/s?"
	so360_URL  = "http://www.so.com/s?"
	JB         *gojieba.Jieba
)

func preProcessQuiz(quiz string) (shortQuiz string, quoted string) {
	re := regexp.MustCompile("[^\\w\\p{Han}\\p{Greek} ]+")
	qz := re.ReplaceAllString(quiz, "")
	words := JB.Cut(qz, true)
	var keywords []string
	for _, w := range words {
		if !(strings.ContainsAny(w, " 的哪是了于") || w == "下列" || w == "以下" || w == "可以" || w == "什么" || w == "多少" || w == "选项" || w == "属于" || w == "中") {
			keywords = append(keywords, w)
		}
	}
	shortQuiz = strings.Join(keywords, "")
	hasQuote := strings.ContainsRune(quiz, '「')
	quoted = ""
	if hasQuote {
		quoted = quiz[strings.IndexRune(quiz, '「'):strings.IndexRune(quiz, '」')]
	}
	quoted = re.ReplaceAllString(quoted, "")
	return
}

//GetFromAPI searh the quiz via popular search engins
func GetFromAPI(quiz string, options []string) map[string]int {

	res := make(map[string]int, len(options))
	for _, option := range options {
		res[option] = 0
	}

	search := make(chan string, 2)
	done := make(chan bool, 1)
	tx := time.Now()

	shortquiz, quote := preProcessQuiz(quiz)

	// go searchFeelingLucky(shortquiz, options, 0, search)
	// go searchGoogle(shortquiz, options, search)
	go searchBaidu(quiz, quote, options, search)
	// go searchGoogleWithOptions(shortquiz, options, search)
	go searchBaiduWithOptions(shortquiz, quote, options, search)

	println("\n.......................searching..............................\n")
	rawStr := "                                            "
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
	reg := regexp.MustCompile("[\\d]+")
	for _, opt := range options {
		if !reg.MatchString(opt) {
			for _, subopt := range options {
				if opt != subopt && strings.Contains(opt, subopt) {
					res[opt] += res[subopt]
				}
			}
		}
	}

	// For negative quiz, flip the count to negative number (dont flip quoted negative word)
	negreg := regexp.MustCompile("「[^」]*[不][^」]*」")
	nonegreg := regexp.MustCompile("不[能同充分对称足够断停止得太值敢锈]")
	if (strings.Contains(quiz, "不") || strings.Contains(quiz, "没有") || strings.Contains(quiz, "未在") || strings.Contains(quiz, "未曾") ||
		strings.Contains(quiz, "错字") || strings.Contains(quiz, "无关")) &&
		!(nonegreg.MatchString(quiz) || negreg.MatchString(quiz)) {
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
	// filter out non alphanumeric/chinese/space
	re := regexp.MustCompile("[^\\w\\p{Han}\\p{Greek} ]+")
	str := re.ReplaceAllString(rawStr, "")
	strs := []rune(str)
	qz := re.ReplaceAllString(quiz, "")

	hasQuote := strings.ContainsRune(quiz, '「')
	quoted := ""
	if hasQuote {
		quoted = quiz[strings.IndexRune(quiz, '「'):strings.IndexRune(quiz, '」')]
		log.Println("quoted part of quiz: ", quoted)
	}

	var qkeywords []string
	// Only match the important part of quiz, in neighbors of options
	if !hasQuote {
		// if strings.IndexRune(qz, '最') >= 0 && strings.IndexRune(qz, '最') < len(qz)-4 {
		// 	qz = qz[strings.IndexRune(qz, '最'):]
		// } else if strings.Index(qz, "属于") >= 0 && strings.Index(qz, "属于") < len(qz)-4 {
		// 	qz = qz[strings.Index(qz, "属于"):]
		// } else if strings.IndexRune(qz, '中') >= 0 && strings.IndexRune(qz, '中') < len(qz)-4 {
		// 	qz = qz[strings.IndexRune(qz, '中')+3:]
		// } else if strings.Index(qz, "的") > 0 && strings.Index(qz, "的") < len(qz)-4 && !hasQuote {
		// 	qz = qz[strings.Index(qz, "的"):]
		// }
	} else {
		qkeywords = JB.Cut(quoted, true)
	}

	words := JB.CutForSearch(qz, true) //allow overlapped words for search
	var keywords []string
	stopwords := [...]string{"下列", "以下", "可以", "什么", "多少", "选项", "属于", "未曾", "中"}
	for _, w := range words {
		if !(strings.ContainsAny(w, " 的哪是了于")) {
			stop := false
			for _, sw := range stopwords {
				if w == sw {
					stop = true
					break
				}
			}
			if !stop {
				keywords = append(keywords, w)
			}
		}
	}

	// Evaluate the match points of each keywords for each option
	kwMap := make(map[string][]int)
	for _, kw := range keywords {
		kwMap[kw] = make([]int, 4)
	}

	width := 40 //Evaluate matching points window size

	for k, option := range options {
		opti := option
		if strings.IndexRune(option, '·') > 0 {
			opti = option[strings.IndexRune(option, '·')+2:] //only match last name
		}
		opti = re.ReplaceAllString(opti, "")
		opt := []rune(opti)
		optLen := len(opt)
		for i, r := range strs {
			if r == ' ' {
				continue
			}
			if string(strs[i:i+optLen]) == opti {
				windowR := strs[i+len(opt) : i+len(opt)+width]
				windowL := strs[i-width : i]
				wordsL := JB.Cut(string(windowL), true)
				wordsR := JB.Cut(string(windowR), true)
				// Keyword the closer to option, the high points (gaussian distribution)
				if !(strings.Contains(qz, "上一") || strings.Contains(qz, "之前")) {
					for _, w := range wordsL {
						for _, word := range keywords {
							if w == word {
								kwMap[w][k]++
								//kwMap[w][k] += int(100 * math.Exp(-math.Pow(float64(len(wordsL)-1-j)/float64(width), 2)/0.1)) //e^(-x^2), sigma=0.1, factor=100
							}
						}
					}
				}
				if !(strings.Contains(qz, "下一") || strings.Contains(qz, "之后")) {
					for _, w := range wordsR {
						for _, word := range keywords {
							if w == word {
								kwMap[w][k]++
								//kwMap[w][k] += int(50 * math.Exp(-math.Pow(float64(j)/float64(width), 2)/0.15)) //e^(-x^2), sigma=0.1, factor=100
							}
						}
					}
				}
			}
		}
	}
	// Calculate standard derivation of keywords of options, give each keyword a weight
	kwWeight := make(map[string]float64)
	for kw, vect := range kwMap {
		sum := 0
		sqSum := 0
		for _, v := range vect {
			sum += v
			sqSum += v * v
		}
		mean := float64(sum) / 4.0
		variance := float64(sqSum)/4.0 - mean*mean
		std := math.Sqrt(variance)
		kwWeight[kw] = std
		log.Printf("%s weights:\t%6.4f", kw, std)
	}

	var optCounts [4]int
	plainQuizCount := 0

	for k, option := range options {
		opti := option
		if strings.IndexRune(option, '·') > 0 {
			opti = option[strings.IndexRune(option, '·')+2:] //only match last name
			log.Println("last name of ", option, " : ", opti)
		}
		opti = re.ReplaceAllString(opti, "")
		opt := []rune(opti)
		optLen := len(opt)
		optCount := 1
		var optMatches []int
		optMatches = append(optMatches, 0)

		// calculate matching keywords in slinding window around each option
		for i, r := range strs {
			if r == ' ' {
				continue
			}
			// find the index of option in the search text
			if string(strs[i:i+optLen]) == opti {
				optCount += 2
				optMatch := 0
				windowR := strs[i+len(opt) : i+len(opt)+width]
				windowL := strs[i-width : i]
				wordsL := JB.Cut(string(windowL), true)
				wordsR := JB.Cut(string(windowR), true)
				// Evaluate match-points of each window. Quiz the closer to option, the high points (gaussian distribution)
				if !(strings.Contains(qz, "上一") || strings.Contains(qz, "之前")) {
					quizMark := 0
					for _, w := range wordsL {
						if strings.ContainsAny(w, "ABCDabcd") && len([]rune(w)) == 1 {
							quizMark++
						}
					}
					plainQuizCount += quizMark
					if quizMark > 1 {
						optCount--
					}
					for j, w := range wordsL {
						if w == "答案" {
							plainQuizCount -= quizMark
							// if the option comes after "答案", gives very high match
							optMatch += int(1000 * math.Exp(-math.Pow(float64(len(wordsL)-1-j)/float64(width), 2)/0.025))
						}

						for _, word := range keywords {
							if w == word {
								optMatch += int(100 * kwWeight[w] * math.Exp(-math.Pow(float64(len(wordsL)-1-j)/float64(width), 2)/0.1)) //e^(-x^2), sigma=0.1, factor=100
							}
						}
						if hasQuote {
							for _, word := range qkeywords {
								if w == word {
									optMatch += int(200 * math.Exp(-math.Pow(float64(len(wordsL)-1-j)/float64(width), 2)/0.5))
								}
							}
						}
					}
				}
				if !(strings.Contains(qz, "下一") || strings.Contains(qz, "之后")) {
					quizMark := 0
					for _, w := range wordsR {
						if strings.ContainsAny(w, "ABCDabcd") && len([]rune(w)) == 1 {
							quizMark++
						}
					}
					plainQuizCount += quizMark
					if quizMark > 1 {
						optCount--
					}
					for j, w := range wordsR {
						if w == "答案" {
							plainQuizCount -= quizMark
						}

						for _, word := range keywords {
							if w == word {
								optMatch += int(50 * kwWeight[w] * math.Exp(-math.Pow(float64(j)/float64(width), 2)/0.15)) //e^(-x^2), sigma=0.1, factor=100
							}
						}
						if hasQuote {
							for _, word := range qkeywords {
								if w == word {
									optMatch += int(100 * math.Exp(-math.Pow(float64(j)/float64(width), 2)/0.5))
								}
							}
						}
					}
				}
				res[option] += optMatch
				optMatches = append(optMatches, optMatch)
				fmt.Printf("%s%4d%6d\t%v\n\t\t\t%v\n", option, optMatch, res[option], wordsL, wordsR)
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

	// if majority matches are plain quiz, simplely set matches as the count of each option
	sumCounts := 0
	for i := range optCounts {
		sumCounts += optCounts[i]
	}
	log.Println("Sum Count: ", sumCounts)
	log.Println("PlainQuiz: ", plainQuizCount)
	log.Printf("Key words: %v", keywords)
	if plainQuizCount > sumCounts {
		for i, option := range options {
			res[option] = optCounts[i]
		}
	}

	total := 1
	for _, option := range options {
		total += res[option]
	}
	for i, option := range options {
		odd := float32(res[option]) / float32(total-res[option])
		fmt.Printf("%4d|%8.2f|%s\n", optCounts[i]/2, odd, option)
	}

}

func searchBaidu(quiz string, quoted string, options []string, c chan string) {
	values := url.Values{}
	query := fmt.Sprintf("%q %s", quoted, quiz)
	values.Add("wd", query)
	req, _ := http.NewRequest("GET", baidu_URL+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	text := "baidu  "
	if resp != nil {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		text += doc.Find("#content_left .t").Text() + doc.Find("#content_left .c-abstract").Text() + doc.Find("#content_left .m").Text() //.m ~zhidao
	}
	c <- text + "                                            "
}

func searchBaiduWithOptions(quiz string, quoted string, options []string, c chan string) {
	values := url.Values{}
	query := fmt.Sprintf("%q %s %s", quoted, quiz, strings.Join(options, " "))
	values.Add("wd", query)
	req, _ := http.NewRequest("GET", baidu_URL+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	text := "baiOpt "
	if resp != nil {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		text += doc.Find("#content_left .t").Text() + doc.Find("#content_left .c-abstract").Text()
	}
	c <- text + "                                            "
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
	c <- text + "                                            "
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
	c <- text + "                                            " // 2x weight
}

func searchFeelingLucky(quiz string, options []string, id int, c chan string) {
	values := url.Values{}
	if id == 0 {
		values.Add("q", quiz)
		log.Println(quiz)
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
	c <- fmt.Sprintf("Lucky %d", id) + text + "                                            "
}

//GetFromAPISearchNum search key word of quiz and each option, compare the google results number
func GetFromAPISearchNum(quiz string, options []string) map[string]int {
	res := make(map[string]int, len(options))
	for _, option := range options {
		res[option] = 0
	}

	re := regexp.MustCompile("[^\\w\\p{Han}\\p{Greek} ]+")
	qz := re.ReplaceAllString(quiz, "")

	words := JB.Cut(qz, true)
	var keywords []string
	for _, w := range words {
		if !(strings.ContainsAny(w, " 的哪是了于") || w == "下列" || w == "以下" || w == "可以" || w == "什么" || w == "多少" || w == "选项" || w == "属于" || w == "中") {
			keywords = append(keywords, w)
		}
	}

	search := make(chan string, 4)
	done := make(chan bool, 1)
	tx := time.Now()
	go searchGoogleNum(keywords, options, 1, search)
	go searchGoogleNum(keywords, options, 2, search)
	go searchGoogleNum(keywords, options, 3, search)
	go searchGoogleNum(keywords, options, 4, search)

	println("\n.......................searching..............................\n")
	var optionCounts [4]int
	count := cap(search)
	go func() {
		for {
			r, more := <-search
			if more {
				log.Println("search received...", r)
				rr := strings.Split(r, ":")
				id, _ := strconv.Atoi(rr[0])
				num, _ := strconv.Atoi(rr[1])
				optionCounts[id-1] = num
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
	case <-time.After(3 * time.Second):
		fmt.Println("search timeout")
	}
	tx2 := time.Now()
	log.Printf("Searching time: %d ms\n", tx2.Sub(tx).Nanoseconds()/1e6)

	for i, opt := range options {
		res[opt] = optionCounts[i]
	}

	// For negative quiz, flip the count to negative number (dont flip quoted negative word)
	re = regexp.MustCompile("「[^」]*[不][^」]*」")
	nonegreg := regexp.MustCompile("不[能同充分对称足够断停止得太值敢锈]")
	if (strings.Contains(quiz, "不") || strings.Contains(quiz, "没有") || strings.Contains(quiz, "未在") || strings.Contains(quiz, "未曾") ||
		strings.Contains(quiz, "错字") || strings.Contains(quiz, "无关")) &&
		!(nonegreg.MatchString(quiz) || re.MatchString(quiz)) {
		for _, option := range options {
			res[option] = -res[option] - 1
		}
	}

	tx3 := time.Now()
	log.Printf("Processing time %d ms\n", tx3.Sub(tx2).Nanoseconds()/1e6)

	return res
}

func searchGoogleNum(keywords []string, options []string, id int, c chan string) {
	values := url.Values{}
	query := fmt.Sprintf("%q \"%s\"", options[id-1], strings.Join(keywords, "\" \""))
	log.Println("searching : " + query)
	values.Add("q", query)
	req, _ := http.NewRequest("GET", google_URL+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	text := strconv.Itoa(id) + ":"
	if resp != nil {
		re := regexp.MustCompile("[\\D]+")
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		str := doc.Find("#resultStats").Text()
		str = strings.Split(str, "results")[0]
		str = re.ReplaceAllString(str, "")
		text += str
	}
	c <- text // 2x weight
}
