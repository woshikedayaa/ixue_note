package internal

import (
	"github.com/woshikedayaa/ixue_note/internal/utils"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type PreRequestFunc func(req *http.Request) error
type PostRequestFunc func(req *http.Request, resp *http.Response) error

type HttpClientWrapper struct {
	client *http.Client

	preRequestFunc  []PreRequestFunc
	postRequestFunc []PostRequestFunc
}

func NewHttpClientWrapper(client *http.Client) *HttpClientWrapper {
	h := new(HttpClientWrapper)
	h.client = client
	h.client.Jar, _ = cookiejar.New(nil) // 这个 error 只会返回 nil .....
	return h
}

// randomUA
func randomUA(req *http.Request) error {
	req.Header.Set("User-Agent", utils.RandomString(rand.Int()%20+10))
	return nil
}
func (h *HttpClientWrapper) EnableRandomUA() {
	h.AppendPreRequest(randomUA)
}

// randomReferer
func randomReferer(req *http.Request) error {
	req.Header.Set("Referer", utils.RandomString(rand.Int()%20+10))
	return nil
}
func (h *HttpClientWrapper) EnableRandomReferer() {
	h.AppendPreRequest(randomReferer)
}

func (h *HttpClientWrapper) AppendPreRequest(f PreRequestFunc) {
	h.preRequestFunc = append(h.preRequestFunc, f)
}

func (h *HttpClientWrapper) AppendPostRequest(f PostRequestFunc) {
	h.postRequestFunc = append(h.postRequestFunc, f)
}

func (h *HttpClientWrapper) preRequest(req *http.Request) error {
	var err error
	for i := 0; i < len(h.preRequestFunc) && err == nil; i++ {
		err = h.preRequestFunc[i](req)
	}
	return err
}
func (h *HttpClientWrapper) postRequest(req *http.Request, resp *http.Response) error {
	var err error
	for i := 0; i < len(h.postRequestFunc) && err == nil; i++ {
		err = h.postRequestFunc[i](req, resp)
	}
	return err
}

// cookie

func (h *HttpClientWrapper) SetCookieJar(j *cookiejar.Jar) {
	h.client.Jar = j
}

func (h *HttpClientWrapper) ResetCookie() {
	h.client.Jar, _ = cookiejar.New(nil)
}

func (h *HttpClientWrapper) GetCookie(u *url.URL) Cookies {
	return h.client.Jar.Cookies(u)
}

func addCookies(cookies []*http.Cookie) PreRequestFunc {
	// 这里闭包 把这个 cookie 数组其实是固定了的
	return func(req *http.Request) error {
		for i := 0; i < len(cookies); i++ {
			req.AddCookie(cookies[i])
		}
		return nil
	}
}

func (h *HttpClientWrapper) AddCookie(cookies []*http.Cookie) {
	h.AppendPreRequest(addCookies(cookies))
}

// http method Below

func (h *HttpClientWrapper) Do(req *http.Request) (resp *http.Response, err error) {
	err = h.preRequest(req)
	if err != nil {
		return nil, err
	}
	resp, err = h.client.Do(req)
	if err != nil {
		return nil, err
	}
	err = h.postRequest(req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (h *HttpClientWrapper) GET(u string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	return h.Do(req)
}
func (h *HttpClientWrapper) POST(u string, body io.Reader, contentType string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodPost, u, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return h.Do(req)
}
func (h *HttpClientWrapper) POSTForm(u string, data url.Values) (resp *http.Response, err error) {
	return h.POST(u, strings.NewReader(data.Encode()), "application/x-www-form-urlencoded")
}
