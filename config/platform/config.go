package platform

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"io"
	"net/http"
	"net/url"
	"os"
)

var (
	Config         = &config{}
	ConfigNotFound = errors.New("this app config is not found")
)

const (
	_minRead     = 16 * 1024
	_defaultHost = "192.168.0.251:8081"
	_apiFileUrl  = "http://%s/api/config/file?%s"
)

// the interface of config
type ConfigInterface interface {
	GetConfig(res interface{}) error
}

// the implement of ConfigInterface
type config struct {
	App     string // app
	Env     string // env
	Version string // version
	Key     string // key
	Host    string // get config url
}

func init() {
	// if os env
	Config.App = os.Getenv("conf_app")
	Config.Env = os.Getenv("conf_env")
	Config.Version = os.Getenv("conf_version")
	Config.Key = os.Getenv("conf_key")
	Config.Host = os.Getenv("conf_host")

	// if flag
	flag.StringVar(&Config.App, "conf_app", "", "the app name of your config")
	flag.StringVar(&Config.Env, "conf_env", "", "the app env of your config")
	flag.StringVar(&Config.Version, "conf_version", "", "the app version of your config")
	flag.StringVar(&Config.Key, "conf_key", "", "the app key of your config")
	flag.StringVar(&Config.Host, "conf_host", _defaultHost, "the app url of your config")
}

// get config according to the key
func (c *config) GetConfig(key string, res interface{}) (err error) {
	var (
		req  *http.Request
		resp *http.Response
		bs   []byte
	)
	if key != "" {
		c.Key = key
	}
	if req, err = http.NewRequest("GET", c.buildUrl(), nil); err != nil {
		return
	}
	client := http.Client{}
	if resp, err = client.Do(req); err != nil {
		return
	}
	if resp.StatusCode == http.StatusNotFound {
		err = ConfigNotFound
		return
	}
	defer resp.Body.Close()
	if bs, err = readAll(resp.Body, _minRead); err != nil {
		return
	}
	_, err = toml.Decode(string(bs), res)
	return
}

func (c *config) buildUrl() (query string) {
	params := url.Values{}
	params.Set("app", c.App)
	params.Set("env", c.Env)
	params.Set("version", c.Version)
	params.Set("key", c.Key)
	return fmt.Sprintf(_apiFileUrl, Config.Host, params.Encode())
}

func readAll(r io.Reader, capacity int64) (b []byte, err error) {
	buf := bytes.NewBuffer(make([]byte, 0, capacity))
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
