package httpClient

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
)

var client *http.Client
var cookiesReg *regexp.Regexp

func init() {
	client = &http.Client{}
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	client.Jar = jar

	cookiesReg, err = regexp.Compile(`document.cookie=".*?"`)
	if err != nil {
		panic(err)
	}
}

func Get(link string) (res *http.Response, err error) {
	res, err = client.Get(link)
	if err != nil {
		return
	}

	if res.StatusCode == 403 {
		// 虾米有时会返回 403 ，并在body中写入js脚本来设置 cookis
		var buf []byte
		buf, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return
		}
		var matched bool
		matched = cookiesReg.Match(buf)

		if matched {
			var cookie *http.Cookie
			cookie, err = getCookieFromJS(buf)
			if err != nil {
				return
			}
			var _url *url.URL
			_url, err = url.Parse(link)
			if err != nil {
				return
			}
			var cookies []*http.Cookie
			cookies = append(cookies, cookie)
			client.Jar.SetCookies(_url, cookies)
			return Get(link)
		}
	} else if res.StatusCode >= 400 {
		var buf []byte
		buf, err = ioutil.ReadAll(res.Body)
		if err == nil {
			fmt.Println(string(buf))
		}
		err = errors.New(fmt.Sprintf("response return %d", res.StatusCode))
	}

	return
}

// getCookieFromJS 从 js 代码中提取 cookie。 例如 `document.cookie="123456"`
func getCookieFromJS(js []byte) (cookie *http.Cookie, err error) {
	cookieStr := cookiesReg.Find(js)
	if len(cookieStr) == 0 {
		err = errors.New("403 but no cookie")
		return
	}

	cookieStr = bytes.Replace(cookieStr, []byte(`"`), []byte(``), -1)
	cookieStr = bytes.Replace(cookieStr, []byte(`document.cookie=`), []byte(``), -1)
	strs := bytes.Split(cookieStr, []byte(`;`))
	var m = make(map[string]string)

	if len(strs) == 0 {
		err = errors.New("403 but no cookie")
		return
	}

	for _, v := range strs {
		strs := bytes.Split(v, []byte(`=`))
		if len(strs) != 2 {
			continue
		}

		m[string(strs[0])] = string(strs[1])
	}

	cookie = &http.Cookie{}

	if age := m["max-age"]; age != "" {
		var a int64
		a, err = strconv.ParseInt(string(age), 10, 0)
		if err != nil {
			return
		}
		cookie.MaxAge = int(a)
		delete(m, "max-age")
	}

	if path := m["path"]; path != "" {
		cookie.Path = path
		delete(m, "path")
	}

	for k, v := range m {
		cookie.Name = k
		cookie.Value = v
	}
	cookie.Domain = "www.xiami.com"
	return
}
