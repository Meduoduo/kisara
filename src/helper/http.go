package helper

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	urltool "net/url"
)

type HttpOptions struct {
	Type  string
	Value interface{}
}

type ResponseHeader map[string]string

func (r *ResponseHeader) FindKey(key string) string {
	for k, v := range *r {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return ""
}

func (r *ResponseHeader) ToString() string {
	v := *r
	if len(v) == 0 {
		return ""
	}

	var result string
	for k, v := range v {
		result += k + ": " + v + "; "
	}

	return result
}

func NewResponseHeader(header http.Header) *ResponseHeader {
	r := ResponseHeader{}
	for k, v := range header {
		r[k] = strings.Join(v, ";")
		if k == "Content-Length" {
			if len(v) > 0 {
				r[k] = strings.Trim(v[0], " ")
			}
		}
	}
	return &r
}

func (r *ResponseHeader) FindContentType() string {
	return r.FindKey("Content-Type")
}

func (r *ResponseHeader) FindContentLength() int64 {
	length := r.FindKey("Content-Length")
	if length == "" {
		return 0
	}
	n, err := strconv.ParseInt(length, 10, 64)
	if err != nil {
		return 0
	}

	return n
}

func (r *ResponseHeader) FindContentEncoding() string {
	return r.FindKey("Content-Encoding")
}

// milliseconds
func HttpTimeout(timeout int64) HttpOptions {
	return HttpOptions{"timeout", timeout}
}

func HttpHeader(header map[string]string) HttpOptions {
	return HttpOptions{"header", header}
}

func HttpNoRedirect() HttpOptions {
	return HttpOptions{"noRedirect", true}
}

// which is used for params with in url
func HttpParams(params map[string]string) HttpOptions {
	return HttpOptions{"params", params}
}

func HttpProxy(proxy string, user string, password string) HttpOptions {
	return HttpOptions{"proxy", []string{proxy, user, password}}
}

// which is used for POST method only
func HttpPayload(payload map[string]string) HttpOptions {
	return HttpOptions{"payload", payload}
}

// which is used for POST method only
func HttpPayloadText(payload string) HttpOptions {
	return HttpOptions{"payloadText", payload}
}

// which is used for POST method only
func HttpPayloadJson(payload interface{}) HttpOptions {
	return HttpOptions{"payloadJson", payload}
}

func HttpWithRandomUA() HttpOptions {
	return HttpOptions{"randomUA", true}
}

func HttpWithDirectReferer() HttpOptions {
	return HttpOptions{"directReferer", true}
}

func HttpWithRetCode(retCode *int) HttpOptions {
	return HttpOptions{"retCode", retCode}
}

func randUA() string {
	ua := []string{"Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2227.1 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2227.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2227.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.3; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2226.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:40.0) Gecko/20100101 Firefox/40.1",
		"Mozilla/5.0 (Windows NT 6.3; rv:36.0) Gecko/20100101 Firefox/36.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10; rv:33.0) Gecko/20100101 Firefox/33.0",
		"Mozilla/5.0 (X11; Linux i586; rv:31.0) Gecko/20100101 Firefox/31.0",
		"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:31.0) Gecko/20130401 Firefox/31.0",
		"Mozilla/5.0 (Windows NT 5.1; rv:31.0) Gecko/20100101 Firefox/31.0",
		"Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; AS; rv:11.0) like Gecko",
		"Mozilla/5.0 (compatible, MSIE 11, Windows NT 6.3; Trident/7.0; rv:11.0) like Gecko",
		"Mozilla/5.0 (Windows; Intel Windows) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.67"}
	n := rand.Intn(13) + 1
	return ua[n]
}

func buildHttpRequest(method string, url string, options ...HttpOptions) (*http.Client, *http.Request, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, nil, err
	}

	for _, option := range options {
		switch option.Type {
		case "timeout":
			client.Timeout = time.Duration(option.Value.(int64)) * time.Millisecond
		case "header":
			for k, v := range option.Value.(map[string]string) {
				req.Header.Set(k, v)
			}
		case "params":
			q := req.URL.Query()
			for k, v := range option.Value.(map[string]string) {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()
		case "proxy":
			proxy := option.Value.([]string)
			proxyUrl, err := urltool.Parse(proxy[0])
			if err != nil {
				return nil, nil, err
			}

			if len(proxy) > 1 {
				proxyUrl.User = urltool.UserPassword(proxy[1], proxy[2])
			}

			client.Transport = &http.Transport{
				Proxy:           http.ProxyURL(proxyUrl),
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		case "payload":
			q := req.URL.Query()
			for k, v := range option.Value.(map[string]string) {
				q.Add(k, v)
			}
			req.Body = ioutil.NopCloser(strings.NewReader(q.Encode()))
		case "payloadText":
			req.Body = ioutil.NopCloser(strings.NewReader(option.Value.(string)))
		case "payloadJson":
			jsonStr, err := json.Marshal(option.Value)
			if err != nil {
				return nil, nil, err
			}
			req.Body = ioutil.NopCloser(bytes.NewBuffer(jsonStr))
		case "noRedirect":
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
		case "randomUA":
			req.Header.Set("User-Agent", randUA())
		case "directReferer":
			req.Header.Set("Referer", url)
		}
	}

	req.Header.Set("Accept", "*/*;q=0.8")
	req.Header.Set("Connection", "close")

	return client, req, nil
}

func doRequest(method string, url string, options ...HttpOptions) (io.ReadCloser, *ResponseHeader, error) {
	client, req, err := buildHttpRequest(method, url, options...)
	if err != nil {
		return nil, nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	for _, option := range options {
		switch option.Type {
		case "retCode":
			if option.Value != nil {
				*option.Value.(*int) = resp.StatusCode
			}
		}
	}

	return resp.Body, NewResponseHeader(resp.Header), nil
}

func httpGet(url string, options ...HttpOptions) (io.ReadCloser, *ResponseHeader, error) {
	return doRequest("GET", url, options...)
}

func httpPost(url string, options ...HttpOptions) (io.ReadCloser, *ResponseHeader, error) {
	return doRequest("POST", url, options...)
}

func httpPut(url string, options ...HttpOptions) (io.ReadCloser, *ResponseHeader, error) {
	return doRequest("PUT", url, options...)
}

func httpDelete(url string, options ...HttpOptions) (io.ReadCloser, *ResponseHeader, error) {
	return doRequest("DELETE", url, options...)
}

func SendAndParse[T any](method string, url string, options ...HttpOptions) (T, error) {
	var result T
	body, _, err := doRequest(method, url, options...)
	if err != nil {
		return result, err
	}

	body_bytes, err := ioutil.ReadAll(body)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(body_bytes, &result)
	if err != nil {
		return result, fmt.Errorf("json unmarshal error: %s, body: %s", err.Error(), string(body_bytes))
	}

	return result, nil
}

func SendGetAndParse[T any](url string, options ...HttpOptions) (T, error) {
	return SendAndParse[T]("GET", url, options...)
}

func SendPostAndParse[T any](url string, options ...HttpOptions) (T, error) {
	return SendAndParse[T]("POST", url, options...)
}

func SendGet(url string, options ...HttpOptions) ([]byte, *ResponseHeader, error) {
	resp, header, err := httpGet(url, options...)
	if err != nil {
		return nil, nil, err
	}

	body, err := ioutil.ReadAll(resp)
	resp.Close()
	if err != nil {
		return nil, nil, err
	}

	return body, header, nil
}

func SendPost(url string, options ...HttpOptions) ([]byte, *ResponseHeader, error) {
	resp, header, err := httpPost(url, options...)
	if err != nil {
		return nil, nil, err
	}

	body, err := ioutil.ReadAll(resp)
	resp.Close()
	if err != nil {
		return nil, nil, err
	}

	return body, header, nil
}

func SendPut(url string, options ...HttpOptions) ([]byte, *ResponseHeader, error) {
	resp, header, err := httpPut(url, options...)
	if err != nil {
		return nil, nil, err
	}

	body, err := ioutil.ReadAll(resp)
	resp.Close()
	if err != nil {
		return nil, nil, err
	}

	return body, header, nil
}

func SendDelete(url string, options ...HttpOptions) ([]byte, *ResponseHeader, error) {
	resp, header, err := httpDelete(url, options...)
	if err != nil {
		return nil, nil, err
	}

	body, err := ioutil.ReadAll(resp)
	resp.Close()
	if err != nil {
		return nil, nil, err
	}

	return body, header, nil
}

// SendGetAsync is not a real async function, it only open the connection
// and pass the reader to user, to close the connection, the final function
// should be passed, and user will not be nervous about the connection leak
func SendGetAsync(url string, final func(io.ReadCloser, *ResponseHeader, error), options ...HttpOptions) {
	resp, header, err := httpGet(url, options...)
	final(resp, header, err)
	resp.Close()
}

// SendPostAsync is not a real async function, it only open the connection
// and pass the reader to user, to close the connection, the final function
// should be passed, and user will not be nervous about the connection leak
func SendPostAsync(url string, final func(io.ReadCloser, *ResponseHeader, error), options ...HttpOptions) {
	resp, header, err := httpPost(url, options...)
	final(resp, header, err)
	resp.Close()
}
