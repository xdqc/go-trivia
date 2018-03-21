package solver

import (
	"bytes"
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
	// memoryDb.Close()
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
		} else if strings.Contains(ctx.Req.URL.Path, "ad") {
			resp.Body = ioutil.NopCloser(bytes.NewReader(*new([]byte)))
		}
		return resp
	}
	s.proxy.OnResponse().DoFunc(responseHandleFunc)
	s.proxy.OnRequest().DoFunc(requestHandleFunc)
}

func orPanic(err error) {
	if err != nil {
		panic(err)
	}
}
