package httpclient

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"golang-kit/config"
	"golang-kit/log"
	xtime "golang-kit/time"
	"io"
	"net"
	xhttp "net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	_family  = "http_client"
	_minRead = 16 * 1024 // 16kb
)

var (
	// ErrStatusCode error of http status code
	ErrStatusCode = errors.New("http status code 5xx")
)

var (
	_noKickUserAgent = "chenxuewen@laoyuegou.com"
)

func init() {
	n, err := os.Hostname()
	if err == nil {
		_noKickUserAgent = _noKickUserAgent + runtime.Version() + " " + n
	}
}

// Client is http client.
type Client struct {
	conf      *config.HTTPClient
	client    *xhttp.Client
	dialer    *net.Dialer
	transport *xhttp.Transport
}

// NewClient new a http client.
func NewClient(c *config.HTTPClient) *Client {
	client := new(Client)
	client.conf = c
	client.dialer = &net.Dialer{
		Timeout:   time.Duration(c.Dial),
		KeepAlive: time.Duration(c.KeepAlive),
	}
	client.transport = &xhttp.Transport{
		DialContext:     client.dialer.DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client.client = &xhttp.Client{
		Transport: client.transport,
	}
	return client
}

// SetKeepAlive set http client keepalive.
func (client *Client) SetKeepAlive(d time.Duration) {
	client.dialer.KeepAlive = d
	client.conf.KeepAlive = xtime.Duration(d)
}

// SetDialTimeout set http client dial timeout.
func (client *Client) SetDialTimeout(d time.Duration) {
	client.dialer.Timeout = d
	client.conf.Dial = xtime.Duration(d)
}

// SetTimeout set http client timeout.
func (client *Client) SetTimeout(d time.Duration) {
	client.conf.Timeout = xtime.Duration(d)
}

// Get issues a GET to the specified URL.
func (client *Client) Get(c context.Context, uri, ip string, params url.Values, res interface{}) (err error) {
	req, err := newRequest(xhttp.MethodGet, uri, ip, params)
	if err != nil {
		return
	}
	return client.Do(c, req, res)
}

// Post issues a Post to the specified URL.
func (client *Client) Post(c context.Context, uri, ip string, params url.Values, res interface{}) (err error) {
	req, err := newRequest(xhttp.MethodPost, uri, ip, params)
	if err != nil {
		return
	}
	return client.Do(c, req, res)
}

// Do sends an HTTP request and returns an HTTP response.
func (client *Client) Do(c context.Context, req *xhttp.Request, res interface{}) (err error) {
	var (
		bs     []byte
		cancel func()
		resp   *xhttp.Response
	)
	// TODO timeout use full path
	c, cancel = context.WithTimeout(c, time.Duration(client.conf.Timeout))
	defer cancel()
	req = req.WithContext(c)
	// header
	req.Header.Set("User-Agent", _noKickUserAgent)
	if resp, err = client.client.Do(req); err != nil {
		log.Error("httpClient.Do(%s) error(%v)", realURL(req), err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= xhttp.StatusInternalServerError {
		err = ErrStatusCode
		log.Error("readAll(%s) uri(%s) error(%v)", bs, realURL(req), err)
		return
	}
	if bs, err = readAll(resp.Body, _minRead); err != nil {
		log.Error("readAll(%s) uri(%s) error(%v)", bs, realURL(req), err)
		return
	}
	if res != nil {
		if err = json.Unmarshal(bs, res); err != nil {
			log.Error("json.Unmarshal(%s) uri(%s) error(%v)", bs, realURL(req), err)
		}
	}
	return
}

// readAll reads from r until an error or EOF and returns the data it read
// from the internal buffer allocated with a specified capacity.
func readAll(r io.Reader, capacity int64) (b []byte, err error) {
	buf := bytes.NewBuffer(make([]byte, 0, capacity))
	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	_, err = buf.ReadFrom(r)
	return buf.Bytes(), err
}

// Sign calc appkey and appsecret sign.
func Sign(params url.Values) (query string, err error) {
	if len(params) == 0 {
		return
	}
	if params.Get("appkey") == "" {
		err = fmt.Errorf("utils http get must have parameter appkey")
		return
	}
	if params.Get("appsecret") == "" {
		err = fmt.Errorf("utils http get must have parameter appsecret")
		return
	}
	if params.Get("sign") != "" {
		err = fmt.Errorf("utils http get must have not parameter sign")
		return
	}
	// sign
	secret := params.Get("appsecret")
	params.Del("appsecret")
	tmp := params.Encode()
	if strings.IndexByte(tmp, '+') > -1 {
		tmp = strings.Replace(tmp, "+", "%20", -1)
	}
	mh := md5.Sum([]byte(tmp + secret))
	params.Set("sign", hex.EncodeToString(mh[:]))
	query = params.Encode()
	return
}

// newRequest new http request with method, uri, ip and values.
func newRequest(method, uri, realIP string, params url.Values) (req *xhttp.Request, err error) {
	enc, err := Sign(params)
	if err != nil {
		log.Error("http check params or sign error(%v)", err)
		return
	}
	ru := uri
	if enc != "" {
		ru = uri + "?" + enc
	}
	if method == xhttp.MethodGet {
		req, err = xhttp.NewRequest(xhttp.MethodGet, ru, nil)
	} else {
		req, err = xhttp.NewRequest(xhttp.MethodPost, uri, strings.NewReader(enc))
	}
	if err != nil {
		log.Error("http.NewRequest(%s, %s) error(%v)", method, ru, err)
		return
	}
	if realIP != "" {
		req.Header.Set("X-BACKEND-BILI-REAL-IP", realIP)
	}
	if method == xhttp.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("User-Agent", _noKickUserAgent)
	return
}

func realURL(req *xhttp.Request) string {
	if req.Method == xhttp.MethodGet {
		return req.URL.String()
	} else if req.Method == xhttp.MethodPost {
		ru := req.URL.Path
		if req.Body != nil {
			rd, ok := req.Body.(io.Reader)
			if ok {
				buf := bytes.NewBuffer([]byte{})
				buf.ReadFrom(rd)
				ru = ru + "?" + buf.String()
			}
		}
		return ru
	}
	return req.URL.Path
}
