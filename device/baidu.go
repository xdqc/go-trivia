package device

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	cache "github.com/patrickmn/go-cache"
)

//Baidu baidu ocr api
type Baidu struct {
	apiKey    string
	secretKey string

	sync.RWMutex
}

type accessTokenRes struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int32  `json:"expires_in"`
}

//wordsResults 匹配
type wordsResults struct {
	WordsNum    int32 `json:"words_result_num"`
	WordsResult []struct {
		Words string `json:"words"`
	} `json:"words_result"`
}

//NewBaidu new
func NewBaidu(cfg *Config) *Baidu {
	baidu := new(Baidu)
	baidu.apiKey = cfg.BaiduAPIKey
	baidu.secretKey = cfg.BaiduSecretKey
	return baidu
}

//GetText 识别图片中的文字
func (baidu *Baidu) GetText(imgPath string) (string, error) {
	accessToken, err := baidu.getAccessToken()
	if err != nil {
		return "", err
	}
	base64Data, err := OpenImageToBase64(imgPath)
	if err != nil {
		return "", err
	}
	uri := fmt.Sprintf("https://aip.baidubce.com/rest/2.0/ocr/v1/general_basic?access_token=%s", accessToken)

	postData := url.Values{}
	postData.Add("image", base64Data)
	body, err := PostForm(uri, postData, 6)
	if err != nil {
		return "", err
	}
	wordResults := new(wordsResults)
	err = json.Unmarshal(body, wordResults)
	if err != nil {
		return "", err
	}
	var text string
	for _, words := range wordResults.WordsResult {
		text = fmt.Sprintf("%s\n%s", text, strings.TrimSpace(words.Words))
	}
	text = strings.TrimLeft(text, "\n")
	return text, nil
}

func (baidu *Baidu) getAccessToken() (accessToken string, err error) {
	baidu.Lock()
	defer baidu.Unlock()

	c := GetCache()
	cacheAccessToken, found := c.Get(BaiduAccessTokenKey)
	if found {
		accessToken = cacheAccessToken.(string)
		return
	}
	uri := fmt.Sprintf("https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id=%s&client_secret=%s", baidu.apiKey, baidu.secretKey)
	body, e := PostForm(uri, nil, 5)
	if e != nil {
		err = e
		return
	}
	res := new(accessTokenRes)
	err = json.Unmarshal(body, res)
	if err != nil {
		return
	}
	accessToken = res.AccessToken
	if accessToken != "" {
		//set cache
		c.Set(BaiduAccessTokenKey, accessToken, time.Second*time.Duration((res.ExpiresIn-100)))
	}

	return
}

var c *cache.Cache

//GetCache 获取cache对象
func GetCache() *cache.Cache {
	if c != nil {
		return c
	}
	c = cache.New(5*time.Minute, 10*time.Minute)
	return c
}

//OpenImageToBase64 OpenImageToBase64
func OpenImageToBase64(filename string) (string, error) {
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(f), nil
}

//PostForm PostForm
func PostForm(uri string, data url.Values, timeout int32) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * time.Duration(timeout),
	}
	response, err := client.PostForm(uri, data)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http post error : uri=%v , statusCode=%v", uri, response.StatusCode)
	}
	return ioutil.ReadAll(response.Body)
}
