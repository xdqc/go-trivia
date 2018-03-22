package solver

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/coreos/goproxy"
)

var (
	_spider = newSpider()
	Mode    int
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
}

func newSpider() *spider {
	sp := &spider{}
	sp.proxy = goproxy.NewProxyHttpServer()
	r, _ := regexp.Compile("solebon.*")
	sp.proxy.OnRequest(goproxy.ReqHostMatches(r)).HandleConnect(goproxy.AlwaysMitm)
	return sp
}

func (s *spider) Run(port string) {
	log.Println("server will at port:" + port)
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
		} else if ctx.Req.URL.Host == "solebonapi.com:443" {

			log.Println(formatRequest(req))

			// bs, _ := ioutil.ReadAll(req.Body)
			// println(string(bs))
		}
		return
	}
	responseHandleFunc := func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if resp == nil {
			return resp
		}
		println(ctx.Req.URL.Host)
		println(ctx.Req.URL.Path)

		if ctx.Req.URL.Path == "/api/1.0/lplist_matches.json" || ctx.Req.URL.Path == "/api/1.0/lpcreate_match.json" {
			//send letterpress match data to webserver
			bs, _ := ioutil.ReadAll(resp.Body)
			println(string(bs))
			setMatch(bs)
			resp.Body = ioutil.NopCloser(bytes.NewReader(bs))
		} else if ctx.Req.URL.Path == "/api/1.0/lp_check_word.json" {
			bs, _ := ioutil.ReadAll(resp.Body)
			println(string(bs))
			println(ctx.Req.URL.RawQuery)
			if strings.Contains(string(bs), "\"found\":false") {
				inValidWord := strings.Split(ctx.Req.URL.RawQuery, "=")[2]
				deleteWord(inValidWord)
			}

			resp.Body = ioutil.NopCloser(bytes.NewReader(bs))
		} else if strings.Contains(ctx.Req.URL.Path, "ad") {
			resp.Body = ioutil.NopCloser(bytes.NewReader(*new([]byte)))
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

func orPanic(err error) {
	if err != nil {
		panic(err)
	}
}
