package solver

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"runtime"
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
	//JB jieba chinese words segregation
	JB *gojieba.Jieba
	//CorpusWord the chiese vocabulary
	CorpusWord map[string]zhCNvocabulary
	//N_opt number of question options
	N_opt = 4
)

func preProcessQuiz(quiz string, isForSearch bool) (keywords []string, quoted string) {
	// trim pre-adj clause
	adjRegex := regexp.MustCompile("[^，]+中，")
	qz := adjRegex.ReplaceAllString(quiz, "")

	re := regexp.MustCompile("[^\\p{L}\\p{N}\\p{Han} ]+")
	qz = re.ReplaceAllString(qz, " ")
	var words []string
	if isForSearch {
		words = JB.CutForSearch(qz, true)
	} else {
		words = JB.Cut(qz, true)
	}
	stopwords := [...]string{"下列", "以下", "可以", "什么", "多少", "选项", "一项", "属于", "关于", "按照", "有关", "没有", "共有", "无关", "包括", "其中", "未曾", "第几", "称为", "位于", "下面", "英文单词", "缩写", "下一句", "上一句", "几", "不", "有", "在", "上", "以", "和", "种", "或", "与", "为", "于", "被", "由", "用", "过", "中", "其", "及", "至", "们", "将", "会", "指", "叫", "所", "省", "年"}
	for _, w := range words {
		if !(strings.ContainsAny(w, " 的哪是了而谁么者几着")) {
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

	quoted = ""
	if strings.IndexRune(quiz, '「') >= 0 && strings.IndexRune(quiz, '「') < strings.IndexRune(quiz, '」') {
		quoted = quiz[strings.IndexRune(quiz, '「'):strings.IndexRune(quiz, '」')]
	}
	if (len(quoted) == 0 || len(quoted) > 5) && strings.IndexRune(quiz, '《') >= 0 && strings.IndexRune(quiz, '《') < strings.IndexRune(quiz, '》') {
		quoted = quiz[strings.IndexRune(quiz, '《'):strings.IndexRune(quiz, '》')]
		stopwordss := append(stopwords[:], "三国演义", "水浒传", "红楼梦", "西游记")
		for _, sw := range stopwordss {
			if strings.Contains(quoted, sw) {
				quoted = ""
				break
			}
		}
	}
	quoted = re.ReplaceAllString(quoted, "")
	return
}

func preProcessOptions(options []string) [][]rune {
	re := regexp.MustCompile("[\\p{N}\\p{Ll}\\p{Lu}\\p{Lt}]+")
	newOptions := make([][]rune, N_opt)
	for i, option := range options {
		newOptions[i] = []rune(option)
	}
	//trim begin/end commons of options
	isSameBegin := true
	isSameEnd := true
	for isSameBegin || isSameEnd {
		if len(newOptions) == 0 {
			return newOptions
		}
		for _, opt := range newOptions {
			if len(opt) == 0 {
				return newOptions
			}
		}
		begin := newOptions[0][0]
		end := newOptions[0][len(newOptions[0])-1]
		if re.MatchString(string(begin)) || re.MatchString(string(end)) {
			break
		}
		for _, option := range newOptions {
			if option[0] != begin || len(option) < 3 {
				isSameBegin = false
			}
			if option[len(option)-1] != end || len(option) < 3 {
				isSameEnd = false
			}
		}
		if isSameBegin {
			for i, option := range newOptions {
				option = option[1:]
				newOptions[i] = option
				// log.Printf("options: %v", string(option))
			}
		}
		if isSameEnd {
			for i, option := range newOptions {
				option = option[:len(option)-1]
				newOptions[i] = option
				// log.Printf("options: %v", string(option))
			}
		}
	}
	re = regexp.MustCompile("[^\\p{L}\\p{N}\\p{Han} ]+")
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
		newOptions[i] = []rune(re.ReplaceAllString(string(newOptions[i]), ""))
	}
	return newOptions
}

//GetFromAPI searh the quiz via popular search engins
func GetFromAPI(quiz string, options []string) (res map[string]int) {
	N_opt = len(options)

	res = make(map[string]int, N_opt)
	for _, option := range options {
		res[option] = 0
	}
	if N_opt == 0 {
		return res
	}

	search := make(chan string, 4+2*N_opt)
	done := make(chan bool, 1)
	// tx := time.Now()

	keywords, quote := preProcessQuiz(quiz, false)

	go searchFeelingLucky(strings.Join(keywords, ""), options, 0, false, true, search)   // testing
	go searchGoogle(quiz, options, true, true, search)                                   // testing
	go searchGoogleWithOptions(strings.Join(keywords, " "), options, true, true, search) // testing
	go searchBaidu(quiz, quote, options, false, true, search)                            // training
	go searchBaiduWithOptions(quiz, options, false, true, search)                        // training
	for i := range options {
		go searchGoogleWithOptions(strings.Join(keywords, " "), options[i:i+1], false, true, search) // testing
		go searchBaiduWithOptions(strings.Join(keywords, " "), options[i:i+1], false, true, search)  // training
	}

	// startBrowser(keywords)

	// println("\n.......................searching..............................\n")
	rawStrTraining := "                                                  "
	rawStrTesting := "                                                  "
	count := cap(search)
	go func() {
		for {
			s, more := <-search
			if more {
				// The first 8 chars in text is the identifier of the search source
				id := s[:8]
				// log.Println("search received...", id)
				if id[6] == '1' {
					rawStrTraining += (s[8:] + s[8:])
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
		// fmt.Println("search done")
	case <-time.After(2 * time.Second):
		// fmt.Println("search timeout")
	}
	// tx2 := time.Now()
	// log.Printf("Searching time: %d ms\n", tx2.Sub(tx).Nanoseconds()/1e6)

	// sliding window, count the common chars between [neighbor of the option in search text] and [quiz]
	CountMatches(quiz, options, rawStrTraining, rawStrTesting, res)

	// if all option got 0 match, search the each option.trimLastChar (xx省 -> xx)
	// if totalCount == 0 {
	// 	for _, option := range options {
	// 		res[option] = strings.Count(str, option[:len(option)-1])
	// 	}
	// }

	// For no-number option, add count to its superstring option count （米波 add to 毫米波)
	// reg := regexp.MustCompile("[\\d]+")
	// for _, opt := range options {
	// 	if !reg.MatchString(opt) {
	// 		for _, subopt := range options {
	// 			if opt != subopt && strings.Contains(opt, subopt) {
	// 				res[opt] += res[subopt]
	// 			}
	// 		}
	// 	}
	// }

	// For negative quiz, flip the count to negative number (dont flip quoted negative word)
	qtnegreg := regexp.MustCompile("「[^」]*[不][^」]*」")
	negreg := regexp.MustCompile("[不未][是属在包含可会曾参]") //regexp.MustCompile("不[能同变充分超过应该对称足够适合自主知靠太具断停止值得敢锈]")

	if (negreg.MatchString(quiz) || strings.Contains(quiz, "没有") || strings.Contains(quiz, "并非") ||
		strings.Contains(quiz, "错字") || strings.Contains(quiz, "错误的是") || strings.Contains(quiz, "很难") || strings.Contains(quiz, "无关")) &&
		!qtnegreg.MatchString(quiz) {
		for _, option := range options {
			res[option] = -res[option] - 1
		}
	}

	// tx3 := time.Now()
	// log.Printf("Processing time %d ms\n", tx3.Sub(tx2).Nanoseconds()/1e6)
	return res
}

//CountMatches sliding window, count the common chars between [neighbor of the option in search text] and [quiz]
func CountMatches(quiz string, options []string, trainingStr string, testingStr string, res map[string]int) {
	// filter out non alphanumeric/chinese/space
	re := regexp.MustCompile("[^\\p{L}\\p{N}\\p{Han} ]+")
	trainingStr = re.ReplaceAllString(trainingStr, "")
	testingStr = re.ReplaceAllString(testingStr, "")
	training := []rune(trainingStr)
	testing := []rune(testingStr)
	// log.Printf("\t\tTraining: %d\tTesting: %d", len(training), len(testing))

	optCounts, _ := trainKeyWords(append(testing, training...), quiz, options, res)

	sumCounts := 0
	for i := range optCounts {
		sumCounts += optCounts[i]
	}
	// log.Printf("Sum Count: %d\tPlain quiz: %d\n", sumCounts, plainQuizCount)

	// If all counts of options in text less than 2, choose the 1 or nothing
	// Or If majority matches are plain quiz, just use count
	if sumCounts < 6 { //|| sumCounts*3 < plainQuizCount
		for i, option := range options {
			res[option] = optCounts[i] - 1
		}
	}

	// If only one option literally appeared in quiz, probably it won't be the answer, skip the option
	numLiteral := 0
	for _, option := range options {
		if strings.Contains(quiz, option) {
			numLiteral++
		}
	}
	if numLiteral == 1 {
		for _, option := range options {
			if strings.Contains(quiz, option) {
				res[option] = 1
			}
		}
	}

	// Calculte probability odd for each option
	total := 1
	for _, option := range options {
		total += res[option]
	}
	// for i, option := range options {
	// 	odd := float32(res[option]) / float32(total-res[option])
	// 	fmt.Printf("%4d|%8.3f| %s\n", optCounts[i]-1, odd, option)
	// }
}

func trainKeyWords(text []rune, quiz string, options []string, res map[string]int) ([]int, int) {
	keywords, quoted := preProcessQuiz(quiz, false)
	// Evaluate the match points of each keywords for each option
	kwMap := make(map[string][]int)
	for _, kw := range keywords {
		kwMap[kw] = make([]int, N_opt)
	}
	var quotedKeywords []string
	if quoted != "" {
		quotedKeywords = JB.Cut(quoted, true)
	}
	shortOptions := preProcessOptions(options) //:= make([][]rune, 0)
	// for _, opt := range options {
	// 	shortOptions = append(shortOptions, []rune(opt))
	// }

	optCounts := make([]int, N_opt)
	plainQuizCount := 0

	width := 50 //sliding window size

	for k := range shortOptions {
		opti := string(shortOptions[k])
		optLen := len(shortOptions[k])
		optCount := 1

		if optLen == 0 {
			continue
		}

		for i, r := range text {
			if r == ' ' {
				continue
			}
			if string(text[i:i+optLen]) == opti {
				optCount++
				windowR := text[i+optLen : i+optLen+width]
				windowL := text[i-width : i]
				wordsL := JB.Cut(string(windowL), true)
				wordsR := JB.Cut(string(windowR), true)
				wordsLR := append(wordsL, wordsR...)
				quizMark := 0
				for _, w := range wordsLR {
					if strings.ContainsAny(w, "ABCDabcd") && len([]rune(w)) == 1 {
						quizMark++
					}
				}
				plainQuizCount += quizMark
				if quizMark > 1 {
					continue
				}
				/**
				 * According to <i>Advances In Chinese Document And Text Processing</i>, P.142, Figure.7,
				 * GP-TSM (Exponential) Kernal function gives highest accuracy rate for chinese text process.
				 */
				kernel := 0
				if !(strings.Contains(quiz, "上一") || strings.Contains(quiz, "之前")) {
					for j, w := range wordsL {
						for _, word := range keywords {
							if w == word {
								// kwMap[w][k]++
								// Gaussian Kernel
								// kwMap[w][k] += int(10 * math.Exp(-math.Pow(float64(len(wordsL)-1-j)/float64(width), 2)/0.5)) //e^(-x^2), sigma=0.1, factor=100
								// Exponential Kernel
								kernel = int(50 * math.Exp(-math.Abs(float64(len(wordsL)-1-j)/float64(width))/0.5)) //e^(-x^2), sigma=0.5, factor=10					}
								for _, qkw := range quotedKeywords {
									if w == qkw {
										kernel *= 3
									}
								}
								kwMap[w][k] += kernel
							}
						}
					}
				}
				if !(strings.Contains(quiz, "下一") || strings.Contains(quiz, "之后")) {
					for j, w := range wordsR {
						for _, word := range keywords {
							if w == word {
								// kwMap[w][k]++
								// Gaussian Kernel
								// kwMap[w][k] += int(8 * math.Exp(-math.Pow(float64(j)/float64(width), 2)/0.5)) //e^(-x^2), sigma=0.1, factor=100
								// Exponential Kernel
								kernel = int(40 * math.Exp(-math.Abs(float64(j)/float64(width))/0.2)) //e^(-x^2), sigma=0.5, factor=8
								for _, qkw := range quotedKeywords {
									if w == qkw {
										kernel *= 3
									}
								}
								kwMap[w][k] += kernel
							}
						}
					}
				}
				// fmt.Printf("%8s\t%v\n%8d\t%v\n", opti, wordsL, kernel, wordsR)
				// Create stream around the context sliding window around option
				go cacheQuizContext(quiz, string(text[i-width:i+optLen+width]))
			}
		}
		optCounts[k] = optCount
	}

	var kwKeys []string
	for k := range kwMap {
		kwKeys = append(kwKeys, k)
	}
	sort.Strings(kwKeys)

	// Calculate the share of each option (logarithm) on keyword
	kwShare := make(map[string][]float64)
	for _, kw := range kwKeys {
		total := 1.0
		for _, pts := range kwMap[kw] {
			total += float64(pts)
		}
		for _, pts := range kwMap[kw] {
			kwShare[kw] = append(kwShare[kw], math.Log2(float64(pts+1))*float64(pts)/total)
		}
	}

	// Calculate the weight of each keyword, by the RSD of the kw score on options
	kwWeight := make(map[string]float64)
	for _, kw := range kwKeys {
		sum := 0
		sqSum := 0
		for _, v := range kwMap[kw] {
			// v := math.Log(float64(val) + 1)
			sum += v
			sqSum += v * v
		}
		mean := float64(sum) / float64(N_opt)
		variance := float64(sqSum)/float64(N_opt) - mean*mean
		rsd := 0.0
		if mean > 0 {
			rsd = math.Sqrt(variance) / mean
		}
		kwWeight[kw] = rsd

		// Use corpus frequency data to correct keyword weight
		if c, ok := CorpusWord[kw]; ok {
			count := c.Count
			kwWeight[kw] = rsd / math.Log(float64(count)) * 10
		} else {
			// the min count in corpus is 50, use 10 for non-exist rare word
			kwWeight[kw] = rsd / math.Log(30) * 10
		}

		// 10 times important
		if strings.ContainsRune(kw, '第') || strings.ContainsRune(kw, '最') {
			kwWeight[kw] *= 10
		}

		// fmt.Printf("W~\t%4.2f%%\t%6s\t%v\n", kwWeight[kw]*100, kw, kwMap[kw])
	}

	optMatrix := make([][]float64, N_opt)
	for i, option := range options {
		optMatrix[i] = make([]float64, len(kwShare))
		vNorm := 1.0
		for j, kw := range kwKeys {
			val := kwShare[kw][i]
			optMatrix[i][j] = val
			vNorm += val * val
		}
		vNorm = math.Sqrt(vNorm)
		vM := 0.0
		for j, kw := range kwKeys {
			val := kwWeight[kw] * optMatrix[i][j] / vNorm
			vM += val * val * float64(kwMap[kw][i]) //* math.Log(math.Log(optMatrix[i][j]*optMatrix[i][j]+1)+1)
			optMatrix[i][j] = val
		}
		// vM = math.Sqrt(vM)
		res[option] = int(vM * 10000)
		// fmt.Printf("%10s %4.3f\t%1.2f\n", option, vM, optMatrix[i])
	}

	return optCounts, plainQuizCount
}

func searchBaidu(quiz string, quoted string, options []string, isTrain bool, isTest bool, c chan string) {
	values := url.Values{}
	query := fmt.Sprintf("%q %s site:baidu.com", quoted, quiz)
	values.Add("wd", query)
	req, _ := http.NewRequest("GET", baidu_URL+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)
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
		if doc != nil && doc.Find(".c-abstract") != nil {
			for _, n := range doc.Find(".c-abstract").Nodes {
				f(n)
				text += buf.String() + "                                                  "
				buf.Reset()
			}
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
		if doc != nil && doc.Find(".c-abstract") != nil {
			for _, n := range doc.Find(".c-abstract").Nodes {
				f(n)
				text += buf.String() + "                                                  "
				buf.Reset()
			}
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
	c <- text + "                                                  "
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
	values.Add("ie", "utf8")
	values.Add("oe", "utf8")
	values.Add("btnI", "") //click I'm feeling lucky! button
	req, _ := http.NewRequest("GET", google_URL+values.Encode(), nil)
	resp, _ := http.DefaultClient.Do(req)

	// log.Println("                   luck url:  " + resp.Request.URL.Host + resp.Request.URL.Path + " /// " + resp.Request.Host)
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
	if resp == nil {

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

func searchBaiduBaike(options []string, id int, c chan string) {
	req, _ := http.NewRequest("GET", "https://baike.baidu.com/item/"+options[id-1], nil)
	resp, _ := http.DefaultClient.Do(req)

	text := "bdbk" + strconv.Itoa(id) + " 00"

	if resp == nil {

	} else if resp.Request.URL.Host == "baike.baidu.com" {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		text += doc.Find(".para").Text()
	} else {
		//doc, _ := goquery.NewDocumentFromReader(resp.Body)
		//text += doc.Find("body").Text()
		// log.Println(text)
	}

	if len(text) > 10000 {
		text = text[:10000]
	}
	c <- text
}

func startBrowser(keywords []string) {
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	default:
		args = []string{"xdg-open"}
	}
	cmd := exec.Command(args[0], append(args[1:], baidu_URL+"wd="+strings.Join(keywords, ""))...)
	err := cmd.Start()
	if err != nil {
		println("Failed to start chrome:", err)
	}
}

//zhCNvocabulary chinese word
type zhCNvocabulary struct {
	Word     string  `json:"word"`
	Category string  `json:"category"`
	Count    int     `json:"count"`
	Frequncy float32 `json:"frequncy"`
}
