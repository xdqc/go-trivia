package solver

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/coreos/goproxy"
	"github.com/xdqc/letterpress-solver/device"
	"github.com/yanyiwu/gojieba"
)

var (
	_spider = newSpider()
	Mode    int
	magic   = "string" // the identifier of a quiz
)

type spider struct {
	proxy *goproxy.ProxyHttpServer
}

func Run(port string, mode int) {
	Mode = mode
	_spider.Init()
	_spider.Run(port)
}

func Close() {
	db.Close()
	memoryDb.Close()
	JB.Free()
}

func newSpider() *spider {
	sp := &spider{}
	sp.proxy = goproxy.NewProxyHttpServer()
	sp.proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)

	//Initialize jieba segmentor
	JB = gojieba.NewJieba()

	//Initialize corpus
	csvFile, _ := os.Open("CorpusWordPOSlist.csv")
	reader := csv.NewReader(bufio.NewReader(csvFile))
	CorpusWord = make(map[string]CVocab)
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		count, _ := strconv.Atoi(line[3])
		freq, _ := strconv.ParseFloat(line[4], 32)
		CorpusWord[line[0]] = CVocab{
			Category: line[1],
			Count:    count,
			Frequncy: float32(freq),
		}
	}

	//Initialize brainID
	brainID = device.GetConfig().BrainID
	return sp
}

func (s *spider) Run(port string) {
	log.Println("proxy server at port:" + port)
	log.Fatal(http.ListenAndServe(":"+port, s.proxy))
}

func (s *spider) Init() {
	requestHandleFunc := func(request *http.Request, ctx *goproxy.ProxyCtx) (req *http.Request, resp *http.Response) {
		req = request
		if ctx.Req.URL.Host == `abc.com` {
			resp = new(http.Response)
			resp.StatusCode = 200
			resp.Header = make(http.Header)
			resp.Header.Add("Content-Disposition", "attachment; filename=ca.crt")
			resp.Header.Add("Content-Type", "application/octet-stream")
			resp.Body = ioutil.NopCloser(bytes.NewReader(goproxy.CA_CERT))
			ShowAllQuestions()

		} else if false && ctx.Req.URL.Host == "question-zh.hortor.net:443" && ctx.Req.URL.Path == "/question/bat/choose" {
			//fmt.Println(formatRequest(request))
			bs, _ := ioutil.ReadAll(req.Body)

			query := string(bs)
			// Parse query string
			values, keys, err := parseURLquery(query)
			if err == nil {
				// modify the selected option
				selectedOpt := ""
				if opt, ok := values["option"]; ok {
					selectedOpt = opt[0]
					if storedAnsPos > 0 && storedAnsPos <= 4 && values["magic"][0] != magic {
						selectedOpt = strconv.Itoa(storedAnsPos)
					}
					log.Println("selected opt:", selectedOpt, request.URL)
					values["option"][0] = selectedOpt

					magic = values["magic"][0]
				}
				// encode the values
				query = encodeURLquery(values, keys)
				log.Println(query)
			} else {
				println("parse req url query error:", err.Error())
			}
			bs = []byte(query)
			req.Body = ioutil.NopCloser(bytes.NewReader(bs))
		}
		return
	}
	responseHandleFunc := func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if resp == nil {
			return resp
		}
		// log.Println(ctx.Req.URL.Host + ctx.Req.URL.Path)

		if ctx.Req.URL.Path == "/api/1.0/lplist_matches.json" || ctx.Req.URL.Path == "/api/1.0/lpcreate_match.json" || ctx.Req.URL.Path == "/api/1.0/lpmatch_detail.json" {
			//send letterpress match data to webserver
			//
			bs, _ := ioutil.ReadAll(resp.Body)
			println(string(bs))
			go setMatch(bs)
			resp.Body = ioutil.NopCloser(bytes.NewReader(bs))
		} else if ctx.Req.URL.Path == "/api/1.0/lp_check_word.json" {
			bs, _ := ioutil.ReadAll(resp.Body)
			if strings.Contains(string(bs), "\"found\":false") {
				inValidWord := strings.Split(ctx.Req.URL.RawQuery, "=")[2]
				go deleteWordDb(inValidWord)
			}
			resp.Body = ioutil.NopCloser(bytes.NewReader(bs))
		} else if ctx.Req.URL.Path == "/question/bat/findQuiz" {
			bs, _ := ioutil.ReadAll(resp.Body)
			//bsNew, ansPos := handleQuestionResp(bs)
			// println("\nquiz\n" + string(bs))
			go handleQuestionResp(bs)
			resp.Body = ioutil.NopCloser(bytes.NewReader(bs))
		} else if ctx.Req.URL.Path == "/question/bat/choose" {
			bs, _ := ioutil.ReadAll(resp.Body)
			// println("\nchoose:\n" + string(bs))
			go handleChooseResponse(bs)
			resp.Body = ioutil.NopCloser(bytes.NewReader(bs))
		} else if ctx.Req.URL.Path == "/question/bat/fightResult" {
			bs, _ := ioutil.ReadAll(resp.Body)
			question := &Question{}
			if err := json.Unmarshal(bs, question); err != nil {
				log.Println("spider fightResult ", err.Error())
			} else {
				question.Data.Quiz = "game over"
				questionInfo, _ = json.Marshal(question)

				re := regexp.MustCompile("\"gold\":\\d{8,},") // account that has 8+ digits gold
				if Mode == 1 && re.Match(bs) {
					go clickProcess(-1, question)
				} // swipe back, start new game
			}
			resp.Body = ioutil.NopCloser(bytes.NewReader(bs))
		} else if ctx.Req.URL.Host == "question-zh.hortor.net:443" {
			bs, _ := ioutil.ReadAll(resp.Body)
			println(string(bs))
			resp.Body = ioutil.NopCloser(bytes.NewReader(bs))
		} else if ctx.Req.URL.Host == "mp.weixin.qq.com:443" && ctx.Req.URL.Path == "/s" {
			bs, _ := ioutil.ReadAll(resp.Body)

			ioutil.WriteFile("lpsolver/dist/assets/wxmp.html", bs, 0600)

			resp.Body = ioutil.NopCloser(bytes.NewReader(bs))
		}
		return resp
	}
	s.proxy.OnResponse().DoFunc(responseHandleFunc)
	s.proxy.OnRequest().DoFunc(requestHandleFunc)
}

// formatRequest generates ascii representation of a request
func formatRequest(r *http.Request) string {
	// Create return string
	var request []string
	// Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)
	// Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	// Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		r.ParseForm()
		request = append(request, "\n")
		request = append(request, r.Form.Encode())
	}
	// Return the request as a string
	return strings.Join(request, "\n")
}

// Parse query string
func parseURLquery(query string) (m map[string][]string, mk []string, err error) {
	m = make(map[string][]string)
	mk = make([]string, 0)
	for query != "" {
		key := query
		if i := strings.IndexAny(key, "&;"); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}
		if key == "" {
			continue
		}
		value := ""
		if i := strings.Index(key, "="); i >= 0 {
			key, value = key[:i], key[i+1:]
		}
		key, err1 := url.QueryUnescape(key)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		value, err1 = url.QueryUnescape(value)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		m[key] = append(m[key], value)
		mk = append(mk, key)
	}
	return
}

// Encode the values
func encodeURLquery(m map[string][]string, mk []string) string {
	var buf bytes.Buffer
	for _, k := range mk {
		vs := m[k]
		prefix := url.QueryEscape(k) + "="
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(prefix)
			buf.WriteString(url.QueryEscape(v))
		}
	}
	return buf.String()
}

func orPanic(err error) {
	if err != nil {
		panic(err)
	}
}
