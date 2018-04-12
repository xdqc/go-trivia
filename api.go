package solver

import (
	"bytes"
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
	"golang.org/x/net/html"
)

var (
	google_URL = "http://www.google.com/search?"
	baidu_URL  = "http://www.baidu.com/s?"
	so360_URL  = "http://www.so.com/s?"
	JB         *gojieba.Jieba
)

func preProcessQuiz(quiz string, isForSearch bool) (keywords []string, quoted string) {
	re := regexp.MustCompile("[^\\p{L}\\p{N}\\p{Han} ]+")
	qz := re.ReplaceAllString(quiz, ",")
	var words []string
	if isForSearch {
		words = JB.CutForSearch(qz, true)
	} else {
		words = JB.Cut(qz, true)
	}
	stopwords := [...]string{"下列", "以下", "可以", "什么", "多少", "选项", "属于", "未曾", "不", "在", "上", "和", "与", "为", "被", "中", "其", "及", "将", "会", "指"}
	for _, w := range words {
		if !(strings.ContainsAny(w, ", 的哪是了几而谁")) {
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
	hasQuote := strings.ContainsRune(quiz, '「')
	quoted = ""
	if hasQuote {
		quoted = quiz[strings.IndexRune(quiz, '「'):strings.IndexRune(quiz, '」')]
	}
	quoted = re.ReplaceAllString(quoted, "")
	return
}

func preProcessOptions(options []string) [4][]rune {
	re := regexp.MustCompile("[\\p{N}\\p{Ll}\\p{Lu}\\p{Lt}]+")
	var newOptions [4][]rune
	for i, option := range options {
		newOptions[i] = []rune(option)
	}
	//trim begin/end commons of options
	isSameBegin := true
	isSameEnd := true
	for isSameBegin || isSameEnd {
		begin := newOptions[0][0]
		end := newOptions[0][len(newOptions[0])-1]
		if re.MatchString(string(begin)) || re.MatchString(string(end)) {
			break
		}
		for _, option := range newOptions {
			if option[0] != begin {
				isSameBegin = false
			}
			if option[len(option)-1] != end {
				isSameEnd = false
			}
		}
		if isSameBegin {
			for i, option := range newOptions {
				option = option[1:]
				newOptions[i] = option
				log.Printf("options: %v", string(option))
			}
		}
		if isSameEnd {
			for i, option := range newOptions {
				option = option[:len(option)-1]
				newOptions[i] = option
				log.Printf("options: %v", string(option))
			}
		}
	}
	for i, option := range newOptions {
		idxDot := -1
		for j, r := range option {
			if r == '·' {
				idxDot = j
			}
		}
		if idxDot >= 0 && idxDot < len(option)-1 {
			opti := option[idxDot+1:] //only match last name
			newOptions[i] = opti
		} else if idxDot == len(option)-1 {
			opti := option[:idxDot] //trim end dot
			newOptions[i] = opti
		}
	}
	return newOptions
}

//GetFromAPI searh the quiz via popular search engins
func GetFromAPI(quiz string, options []string) map[string]int {

	res := make(map[string]int, len(options))
	for _, option := range options {
		res[option] = 0
	}

	search := make(chan string, 13)
	done := make(chan bool, 1)
	tx := time.Now()

	keywords, quote := preProcessQuiz(quiz, false)

	go searchFeelingLucky(strings.Join(keywords, ""), options, 0, false, true, search)         // testing
	go searchGoogle(quiz, options, false, true, search)                                        // testing
	go searchGoogleWithOptions(quiz, options, false, true, search)                             // testing
	go searchGoogleWithOptions(strings.Join(keywords, " "), options[0:1], false, true, search) // testing
	go searchGoogleWithOptions(strings.Join(keywords, " "), options[1:2], false, true, search) // testing
	go searchGoogleWithOptions(strings.Join(keywords, " "), options[2:3], false, true, search) // testing
	go searchGoogleWithOptions(strings.Join(keywords, " "), options[3:4], false, true, search) // testing
	go searchBaidu(quiz, quote, options, true, false, search)                                  // training
	go searchBaiduWithOptions(quiz, options, true, false, search)                              // training
	go searchBaiduWithOptions(strings.Join(keywords, " "), options[0:1], true, false, search)  // training
	go searchBaiduWithOptions(strings.Join(keywords, " "), options[1:2], true, false, search)  // training
	go searchBaiduWithOptions(strings.Join(keywords, " "), options[2:3], true, false, search)  // training
	go searchBaiduWithOptions(strings.Join(keywords, " "), options[3:4], true, false, search)  // training

	println("\n.......................searching..............................\n")
	rawStrTraining := "                                                  "
	rawStrTesting := "                                                  "
	count := cap(search)
	go func() {
		for {
			s, more := <-search
			if more {
				// First 8 chars in text is the identifier of the search source
				id := s[:8]
				// log.Println("search received...", id)
				if id[6] == '1' {
					rawStrTraining += s[8:]
				}
				if id[7] == '1' {
					rawStrTesting += s[8:]
				}
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
	CountMatches(quiz, options, rawStrTraining, rawStrTesting, res)

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
	nonegreg := regexp.MustCompile("不[能会同充分超过应该对称足够太具断停止值得敢锈]")
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
func CountMatches(quiz string, options []string, trainingStr string, testingStr string, res map[string]int) {
	// filter out non alphanumeric/chinese/space
	re := regexp.MustCompile("[^\\w\\p{Han}\\p{Greek} ]+")
	trainingStr = re.ReplaceAllString(trainingStr, "")
	testingStr = re.ReplaceAllString(testingStr, "")
	training := []rune(trainingStr)
	testing := []rune(testingStr)
	log.Printf("\t\t\t\tTraining: %d\tTesting: %d", len(training), len(testing))
	qz := re.ReplaceAllString(quiz, "")

	keywords, _ := preProcessQuiz(quiz, true)
	shortOptions := preProcessOptions(options)

	// var qkeywords []string
	// if quoted != "" {
	// 	qkeywords = JB.Cut(quoted, true)
	// }

	// Calculate RSD of keywords of options, give each keyword a weight
	kwWeight := trainKeyWords(training, keywords, shortOptions)

	var optCounts [4]int
	plainQuizCount := 0

	width := 50
	for k, option := range options {
		opti := string(shortOptions[k])
		optLen := len(shortOptions[k])
		var optMatches []int
		optMatches = append(optMatches, 0)

		// calculate matching keywords in slinding window around each option
		for i, r := range testing {
			if r == ' ' {
				continue
			}
			// find the index of option in the search text
			if string(testing[i:i+optLen]) == opti {
				optMatch := 0
				windowR := testing[i+optLen : i+optLen+width]
				windowL := testing[i-width : i]
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
						optMatch = 0
						continue
					}
					for j, w := range wordsL {
						if w == "不是" && j == len(wordsL)-1 {
							optMatch = 0
							break
						}
						for _, word := range keywords {
							if w == word {
								optMatch += int(100 * kwWeight[w] * math.Exp(-math.Pow(float64(len(wordsL)-1-j)/float64(width), 2)/0.2)) //e^(-x^2), sigma=0.1, factor=100
							}
						}
						// if quoted != "" {
						// 	for _, word := range qkeywords {
						// 		if w == word {
						// 			optMatch += int(100 * math.Exp(-math.Pow(float64(len(wordsL)-1-j)/float64(width), 2)/0.5))
						// 		}
						// 	}
						// }
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
						optMatch = 0
						continue
					}
					for j, w := range wordsR {
						if w == "答案" {
							plainQuizCount -= quizMark
						}

						for _, word := range keywords {
							if w == word {
								optMatch += int(50 * kwWeight[w] * math.Exp(-math.Pow(float64(j)/float64(width), 2)/0.2)) //e^(-x^2), sigma=0.1, factor=100
							}
						}
						// if quoted != "" {
						// 	for _, word := range qkeywords {
						// 		if w == word {
						// 			optMatch += int(100 * math.Exp(-math.Pow(float64(j)/float64(width), 2)/0.5))
						// 		}
						// 	}
						// }
					}
				}
				res[option] += optMatch
				optMatches = append(optMatches, optMatch)
				fmt.Printf("%s%4d%6d\t%v\n\t\t\t%v\n", option, optMatch, res[option], wordsL, wordsR)
			}
		}
		optCounts[k] = len(optMatches)
		sort.Sort(sort.Reverse(sort.IntSlice(optMatches)))
		//only take first lg(len) number of top matches, sum up as the result of the option
		logCount := int(math.Log2(float64(len(optMatches))))
		optMatches = optMatches[0:logCount]
		matches := 0
		for _, m := range optMatches {
			matches += m
		}
		res[option] = matches
	}

	sumCounts := 0
	for i := range optCounts {
		sumCounts += optCounts[i]
	}
	log.Println("Sum Count: ", sumCounts)
	log.Println("PlainQuiz: ", plainQuizCount)
	log.Printf("Key words: %v", keywords)
	// if majority matches are plain quiz, simplely set matches as the count of each option
	if sumCounts < 6 || 2*sumCounts < plainQuizCount {
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
		fmt.Printf("%4d|%8.3f| %s\n", optCounts[i]-1, odd, option)
	}
}

func trainKeyWords(training []rune, keywords []string, shortOptions [4][]rune) (kwWeight map[string]float64) {
	// Evaluate the match points of each keywords for each option
	kwMap := make(map[string][]int)
	for _, kw := range keywords {
		kwMap[kw] = make([]int, 4)
	}

	width := 30 //sliding window size

	for k := range shortOptions {
		opti := string(shortOptions[k])
		optLen := len(shortOptions[k])
		for i, r := range training {
			if r == ' ' {
				continue
			}
			if string(training[i:i+optLen]) == opti {
				windowR := training[i+optLen : i+optLen+width]
				windowL := training[i-width : i]
				wordsL := JB.Cut(string(windowL), true)
				wordsR := JB.Cut(string(windowR), true)
				words := append(wordsL, wordsR...)
				for _, w := range words {
					for _, word := range keywords {
						if w == word {
							kwMap[w][k]++
							//kwMap[w][k] += int(100 * math.Exp(-math.Pow(float64(len(wordsL)-1-j)/float64(width), 2)/0.1)) //e^(-x^2), sigma=0.1, factor=100
						}
					}
				}
			}
		}
	}

	// kwWeight := make(map[string]float64)
	for kw, vect := range kwMap {
		sum := 0
		recpSum := 0.0
		sqSum := 0
		for _, v := range vect {
			sum += v
			sqSum += v * v
			recpSum += 1.0 / float64(v+1)
		}
		mean := float64(sum) / 4.0
		variance := float64(sqSum)/4.0 - mean*mean
		hamonicMean := 4.0 / recpSum
		rsd := math.Sqrt(variance) / hamonicMean
		kwWeight[kw] = rsd
		fmt.Printf("W~\t%4.2f%%\t%6s\t%v\n", rsd*100, kw, vect)
	}
	return
}

func searchBaidu(quiz string, quoted string, options []string, isTrain bool, isTest bool, c chan string) {
	values := url.Values{}
	query := fmt.Sprintf("%q %s site:baidu.com", quoted, quiz)
	values.Add("wd", query)
	req, _ := http.NewRequest("GET", baidu_URL+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	text := "baidu "
	if isTrain {
		text += "1"
	} else {
		text += "0"
	}
	if isTest {
		text += "1"
	} else {
		text += "0"
	}
	if resp != nil {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		// text += doc.Find("#content_left .t").Text() + doc.Find("#content_left .c-abstract").Text() + doc.Find("#content_left .m").Text() //.m ~zhidao

		var buf bytes.Buffer
		// Slightly optimized vs calling Each: no single selection object created
		var f func(*html.Node)
		f = func(n *html.Node) {
			if n.Type == html.TextNode {
				// Keep newlines and spaces, like jQuery
				buf.WriteString(n.Data)
			}
			if n.FirstChild != nil {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					f(c)
				}
			}
		}
		for _, n := range doc.Find("#content_left .c-abstract").Nodes {
			f(n)
			text += buf.String() + "                                                  "
			buf.Reset()
		}

	}
	c <- text + "                                                  "
}

func searchBaiduWithOptions(quiz string, options []string, isTrain bool, isTest bool, c chan string) {
	values := url.Values{}
	query := fmt.Sprintf("%s %s", quiz, strings.Join(options, " "))
	values.Add("wd", query)
	req, _ := http.NewRequest("GET", baidu_URL+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	text := "baiOp "
	if isTrain {
		text += "1"
	} else {
		text += "0"
	}
	if isTest {
		text += "1"
	} else {
		text += "0"
	}
	if resp != nil {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		// text += doc.Find("#content_left .t").Text() + doc.Find("#content_left .c-abstract").Text()
		var buf bytes.Buffer
		// Slightly optimized vs calling Each: no single selection object created
		var f func(*html.Node)
		f = func(n *html.Node) {
			if n.Type == html.TextNode {
				// Keep newlines and spaces, like jQuery
				buf.WriteString(n.Data)
			}
			if n.FirstChild != nil {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					f(c)
				}
			}
		}
		for _, n := range doc.Find("#content_left .c-abstract").Nodes {
			f(n)
			text += buf.String() + "                                                  "
			buf.Reset()
		}
	}
	c <- text + "                                                  "
}

func searchGoogle(quiz string, options []string, isTrain bool, isTest bool, c chan string) {
	values := url.Values{}
	values.Add("q", quiz)
	values.Add("lr", "lang_zh-CN")
	values.Add("ie", "utf8")
	values.Add("oe", "utf8")
	req, _ := http.NewRequest("GET", google_URL+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	text := "googl "
	if isTrain {
		text += "1"
	} else {
		text += "0"
	}
	if isTest {
		text += "1"
	} else {
		text += "0"
	}
	if resp != nil {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		// text += doc.Find(".r").Text() + doc.Find(".st").Text() + doc.Find(".P1usbc").Text() //.P1usbc ~wiki
		var buf bytes.Buffer
		// Slightly optimized vs calling Each: no single selection object created
		var f func(*html.Node)
		f = func(n *html.Node) {
			if n.Type == html.TextNode {
				// Keep newlines and spaces, like jQuery
				buf.WriteString(n.Data)
			}
			if n.FirstChild != nil {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					f(c)
				}
			}
		}
		titleNodes := doc.Find(".r").Nodes
		for i, n := range doc.Find(".st").Nodes {
			f(titleNodes[i])
			f(n)
			text += buf.String() + "                                                  "
			buf.Reset()
		}
	}
	c <- text + "                                                  "
}

func searchGoogleWithOptions(quiz string, options []string, isTrain bool, isTest bool, c chan string) {
	values := url.Values{}
	values.Add("q", quiz+" \""+strings.Join(options, "\" OR \"")+"\"")
	values.Add("lr", "lang_zh-CN")
	values.Add("ie", "utf8")
	values.Add("oe", "utf8")
	req, _ := http.NewRequest("GET", google_URL+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	text := "gooOp "
	if isTrain {
		text += "1"
	} else {
		text += "0"
	}
	if isTest {
		text += "1"
	} else {
		text += "0"
	}
	if resp != nil {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		// text += doc.Find(".r").Text() + doc.Find(".st").Text() + doc.Find(".P1usbc").Text() //.P1usbc ~wiki
		var buf bytes.Buffer
		// Slightly optimized vs calling Each: no single selection object created
		var f func(*html.Node)
		f = func(n *html.Node) {
			if n.Type == html.TextNode {
				// Keep newlines and spaces, like jQuery
				buf.WriteString(n.Data)
			}
			if n.FirstChild != nil {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					f(c)
				}
			}
		}
		titleNodes := doc.Find(".r").Nodes
		for i, n := range doc.Find(".st").Nodes {
			f(titleNodes[i])
			f(n)
			text += buf.String() + "                                                  "
			buf.Reset()
		}
	}
	c <- text + "                                                  " // 2x weight
}

func searchFeelingLucky(quiz string, options []string, id int, isTrain bool, isTest bool, c chan string) {
	values := url.Values{}
	if id == 0 {
		values.Add("q", quiz)
		log.Println(quiz)
	} else {
		values.Add("q", options[id-1])
	}
	values.Add("lr", "lang_zh-CN")
	values.Add("btnI", "") //click I'm feeling lucky! button
	req, _ := http.NewRequest("GET", google_URL+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)

	log.Println("                   luck" + strconv.Itoa(id) + " url:  " + resp.Request.URL.Host + resp.Request.URL.Path + " /// " + resp.Request.Host)
	text := "Luck" + strconv.Itoa(id) + " "
	if isTrain {
		text += "1"
	} else {
		text += "0"
	}
	if isTest {
		text += "1"
	} else {
		text += "0"
	}
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
	} else if resp.Request.URL.Host == "www.zhihu.com" {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		text += doc.Find(".QuestionHeader-title").Text()
	} else {
		//doc, _ := goquery.NewDocumentFromReader(resp.Body)
		//text += doc.Find("body").Text()
		// log.Println(text)
	}

	if len(text) > 10000 {
		text = text[:10000]
	}
	c <- text + "                                                  "
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
	nonegreg := regexp.MustCompile("不[能同充分对称足够超过停止断得太值敢锈]")
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
