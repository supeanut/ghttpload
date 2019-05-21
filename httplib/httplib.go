package httplib

import (
	"time"
	"crypto/tls"
	"net/http"
	"net/url"
	"sync"
	"net/http/cookiejar"
	"log"
	"bytes"
	"io/ioutil"
	"encoding/xml"
	"gopkg.in/yaml.v2"
	"encoding/json"
	"strings"
	"io"
	"mime/multipart"
	"os"
	"net"
	"net/http/httputil"
	"compress/gzip"
)

var defaultSetting = HttpSettings {
	UserAgent: "ghttpload",
	ConnectTimeout:     60 * time.Second,
	ReadWriteTimeout:   60 * time.Second,
	Gzip: 				true,
	DumpBody:           true,
}

var defaultCookieJar http.CookieJar
var settingMutex sync.Mutex

// createDefaultCookie creates a global cookiejar to store cookies.
func createDefaultCookie() {
	settingMutex.Lock()
	defer settingMutex.Unlock()
	defaultCookieJar, _ = cookiejar.New(nil)
}

// SetDefaultSetting Overwrite default settings
func SetDefaultSetting(setting HttpSettings) {
	settingMutex.Lock()
	defer settingMutex.Unlock()
	defaultSetting = setting
}

// NewHttpRequest return *HttpRequest with specific method
func NewHttpRequest(rawurl, method string) *HttpRequest {
	var resp http.Response
	u, err := url.Parse(rawurl)
	if err != nil {
		log.Println("Httplib:", err)
	}
	req := http.Request{
		URL: 		u,
		Method: 	method,
		Header: 	make(http.Header),
		Proto: 		"HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	return &HttpRequest{
		url: 		rawurl,
		req: 		&req,
		params: 	map[string][]string{},
		files:		map[string]string{},
		setting:	defaultSetting,
		resp:       &resp,
	}
}

// Get returns *HttpRequest with GET method.
func Get(url string) *HttpRequest {
	return NewHttpRequest(url, "GET")
}

// Post returns *HttpRequest with POST method.
func Post(url string) *HttpRequest {
	return NewHttpRequest(url, "POST")
}

// Put returns *BeegoHttpRequest with PUT method.
func Put(url string) *HttpRequest {
	return NewHttpRequest(url, "PUT")
}

// Delete returns *BeegoHttpRequest DELETE method.
func Delete(url string) *HttpRequest {
	return NewHttpRequest(url, "DELETE")
}

// Head returns *BeegoHttpRequest with HEAD method.
func Head(url string) *HttpRequest {
	return NewHttpRequest(url, "HEAD")
}

// HttpSettings the http.Client setting
type HttpSettings struct {
	ShowDebug			bool
	UserAgent			string
	ConnectTimeout 		time.Duration
	ReadWriteTimeout	time.Duration
	TLSClientConfig		*tls.Config
	Proxy 				func(*http.Request) (*url.URL, error)
	Transport 			http.RoundTripper
	CheckRedirect       func(req *http.Request, via []*http.Request) error
	EnableCookie		bool
	Gzip 				bool
	DumpBody 			bool
	Retries 			int   // if set to -1 means will retry forever
}

// HttpRequest provides more useful methods for requesting one url than http.Request.
type HttpRequest struct {
	url		string
	req 	*http.Request
	params 	map[string][]string
	files 	map[string]string
	setting HttpSettings
	resp 	*http.Response
	body 	[]byte
	dump    []byte
}

// GetRequest return the request object
func (r *HttpRequest) GetRequest() *http.Request {
	return r.req
}

// Setting Change request settings
func (r *HttpRequest) Setting(setting HttpSettings) *HttpRequest {
	r.setting = setting
	return r
}

// SetBasicAuth sets the reqeust's Authorization header to use HTTP Basic Authentication with the provided username and password.
func (r *HttpRequest) SetBasicAuth(username, password string) *HttpRequest {
	r.req.SetBasicAuth(username, password)
	return r
}

// SetEnableCookie sets enable/disable cookiejar
func (r *HttpRequest) SetEnableCookie(enable bool) *HttpRequest {
	r.setting.EnableCookie = enable
	return r
}

// SetUserAgent sets User-Agent header field
func (r *HttpRequest) SetUserAgent(useragent string) *HttpRequest {
	r.setting.UserAgent = useragent
	return r
}

// Debug sets show debug or not when executing request.
func (r *HttpRequest) Debug(isdebug bool) *HttpRequest {
	r.setting.ShowDebug = isdebug
	return r
}

// Retries sets Retries times.
// default is 0 means no retried.
// -1 means retried forever.
// others means retried times.
func (r *HttpRequest)Retries(times int) *HttpRequest {
	r.setting.Retries = times
	return r
}

// DumpBody setting whenther need to Dump the Body.
func (r *HttpRequest) DumpBody(isdump bool) *HttpRequest {
	r.setting.DumpBody = isdump
	return r
}

// DumpRequest return the DumpRequest
func (r *HttpRequest) DumpRequest() []byte {
	return r.dump
}

// SetTimeout sets connect time out and read-write time out for HttpRequest
func (r *HttpRequest) SetTimeout(connectTimeout, readWriteTimeout time.Duration) *HttpRequest {
	r.setting.ConnectTimeout = connectTimeout
	r.setting.ReadWriteTimeout = readWriteTimeout
	return r
}

// SetTLSClientConfig sets tls connection configurations if visiting https url.
func (r *HttpRequest) SetTLSClientConfig(config *tls.Config) *HttpRequest {
	r.setting.TLSClientConfig = config
	return r
}

// Header add header item string in request.
func (r *HttpRequest) Header(key, value string) *HttpRequest {
	r.req.Header.Set(key, value)
	return r
}

// SetHost set the request host
func (r *HttpRequest) SetHost(host string) *HttpRequest {
	r.req.Host = host
	return r
}

// SetProtocolVersion Set the protocol version for incoming requests.
// Client requests always use HTTP/1.1
func (r *HttpRequest) SetProtocolVersion(vers string) *HttpRequest {
	if len(vers) == 0 {
		vers = "HTTP/1.1"
	}

	major, minor, ok := http.ParseHTTPVersion(vers)
	if ok {
		r.req.Proto = vers
		r.req.ProtoMajor = major
		r.req.ProtoMinor = minor
	}

	return r
}

// SetCookie add cookie into request.
func (r *HttpRequest) SetCookie(cookie *http.Cookie) *HttpRequest {
	r.req.Header.Add("Cookie", cookie.String())
	return r
}

// SetProxy set the http proxy
// example:
//
// func(req *http.Request) (*usr.URL, error) {
//     u, _ := url.ParseReqeustURI("http://127.0.0.1:8118")
//     return u, nil
// }

func (r *HttpRequest) SetProxy(proxy func(r2 *http.Request) (*url.URL, error)) *HttpRequest {
	r.setting.Proxy = proxy
	return r
}

// SetCheckRedirect specifies the policy for handling redirects.
//
// If CheckRedirect is nil, the Client uses its default policy.
// which is to stop after 10 consecutive requests.
func (r *HttpRequest) SetCheckRedirect(redirect func(req *http.Request, via []*http.Request) error) *HttpRequest {
	r.setting.CheckRedirect = redirect
	return r
}

// Param adds query param into request.
// Params build query string as ?key1=value&key2=value2...
func (r *HttpRequest) Param(key, value string) *HttpRequest {
	if param, ok := r.params[key]; ok {
		r.params[key] = append(param, value)
	} else {
		r.params[key] = []string{value}
	}
	return r
}

// PostFile add a post file to the request
func (r *HttpRequest) PostFile(formname, filename string) *HttpRequest {
	r.files[formname] = filename
	return r
}

// Body adds request raw body.
// it supports string and []byte.
func (r *HttpRequest) Body(data interface{}) *HttpRequest {
	switch t := data.(type) {
	case string:
		bf := bytes.NewBufferString(t)
		r.req.Body = ioutil.NopCloser(bf)
		r.req.ContentLength = int64(len(t))
	case []byte:
		bf := bytes.NewBuffer(t)
		r.req.Body = ioutil.NopCloser(bf)
		r.req.ContentLength = int64(len(t))
	}
	return r
}

// XMLBody adds request raw body encoding by XML.
func (r *HttpRequest) XMLBody(obj interface{}) (*HttpRequest, error) {
	if r.req.Body == nil && obj != nil {
		byts, err := xml.Marshal(obj)
		if err != nil {
			return r, err
		}
		r.req.Body = ioutil.NopCloser(bytes.NewReader(byts))
		r.req.ContentLength = int64(len(byts))
		r.req.Header.Set("Content-Type", "application/xml")
	}
	return r, nil
}

// YAMLBody adds request raw body encoding by YAML.
func (r *HttpRequest) YAMLBody(obj interface{}) (*HttpRequest, error) {
	if r.req.Body == nil && obj != nil {
		byts, err := yaml.Marshal(obj)
		if err != nil {
			return r, err
		}
		r.req.Body = ioutil.NopCloser(bytes.NewReader(byts))
		r.req.ContentLength = int64(len(byts))
		r.req.Header.Set("Content-Type", "application/x+yaml")
	}
	return r, nil
}

// JSONBody adds request raw body encoding by JSON.
func (r *HttpRequest) JSONBody(obj interface{}) (*HttpRequest, error) {
	if r.req.Body == nil && obj != nil {
		byts, err := json.Marshal(obj)
		if err != nil {
			return r, err
		}
		r.req.Body = ioutil.NopCloser(bytes.NewReader(byts))
		r.req.ContentLength = int64(len(byts))
		r.req.Header.Set("Content-Type", "application/json")
	}
	return r, nil
}

func (r *HttpRequest) buildURL(paramBody string) {
	// build GET url with query string
	if r.req.Method == "GET" && len(paramBody) > 0 {
		if strings.Contains(r.url, "?") {
			r.url += "&" + paramBody
		} else {
			r.url = r.url + "?" + paramBody
		}
		return
	}

	// build POST/PUT/PATCH url and body
	if (r.req.Method == "POST" || r.req.Method == "PUT" || r.req.Method == "PATCH" || r.req.Method == "DELETE") && r.req.Body == nil {
		// with files
		if len(r.files) > 0 {
			pr, pw := io.Pipe()
			bodyWriter := multipart.NewWriter(pw)
			go func() {
				for formname, filename := range r.files {
					fileWriter, err := bodyWriter.CreateFormFile(formname, filename)
					if err != nil {
						log.Println("Httplib:", err)
					}
					fh, err := os.Open(filename)
					if err != nil {
						log.Println("Httplib:", err)
					}
					//iocopy
					_, err = io.Copy(fileWriter, fh)
					fh.Close()
					if err != nil {
						log.Println("Httplib:", err)
					}
				}
				for k, v := range r.params {
					for _, vv := range v {
						bodyWriter.WriteField(k, vv)
					}
				}
				bodyWriter.Close()
				pw.Close()
			}()
			r.Header("Content-Type", bodyWriter.FormDataContentType())
			r.req.Body = ioutil.NopCloser(pr)
			return
		}

		// with params
		if len(paramBody) > 0 {
			r.Header("Content-Type", "application/x-www-form-urlencoded")
			r.Body(paramBody)
		}
	}
}

func (r *HttpRequest) getResponse() (*http.Response, error) {
	if r.resp.StatusCode != 0 {
		return r.resp, nil
	}
	resp, err := r.DoRequest()
	if err != nil {
		return nil, err
	}
	r.resp = resp
	return resp, nil
}

// DoRequest will do the client.Do
func (r *HttpRequest) DoRequest() (resp *http.Response, err error) {
	var paramBody string
	if len(r.params) > 0 {
		var buf bytes.Buffer
		for k, v := range r.params {
			for _, vv := range v {
				buf.WriteString(url.QueryEscape(k))
				buf.WriteByte('=')
				buf.WriteString(url.QueryEscape(vv))
				buf.WriteByte('&')
			}
		}
		paramBody = buf.String()
		paramBody = paramBody[0 : len(paramBody)-1]
	}

	r.buildURL(paramBody)
	urlParsed, err := url.Parse(r.url)
	if err != nil {
		return nil, err
	}

	r.req.URL = urlParsed

	trans := r.setting.Transport

	if trans == nil {
		// create default transport
		trans = &http.Transport{
			TLSClientConfig:     r.setting.TLSClientConfig,
			Proxy: 			     r.setting.Proxy,
			Dial:			     TimeoutDialer(r.setting.ConnectTimeout, r.setting.ReadWriteTimeout),
			MaxIdleConnsPerHost: 100,
		}
	} else {
		// if b.transport is *http.Transport then set the settings.
		if t, ok := trans.(*http.Transport); ok {
			if t.TLSClientConfig == nil {
				t.TLSClientConfig = r.setting.TLSClientConfig
			}
			if t.Proxy == nil {
				t.Proxy = r.setting.Proxy
			}
			if t.Dial == nil {
				t.Dial = TimeoutDialer(r.setting.ConnectTimeout, r.setting.ReadWriteTimeout)
			}
		}
	}

	var jar http.CookieJar
	if r.setting.EnableCookie {
		if defaultCookieJar == nil {
			createDefaultCookie()
		}
		jar = defaultCookieJar
	}

	client := &http.Client{
		Transport: trans,
		Jar: 	   jar,
	}

	if r.setting.UserAgent != "" && r.req.Header.Get("User-Agent") == "" {
		r.req.Header.Set("User-Agent", r.setting.UserAgent)
	}

	if r.setting.CheckRedirect != nil {
		client.CheckRedirect = r.setting.CheckRedirect
	}

	if r.setting.ShowDebug {
		dump, err := httputil.DumpRequest(r.req, r.setting.DumpBody)
		if err != nil {
			log.Println(err.Error())
		}
		r.dump = dump
	}

	// retries default value is 0, it will run once.
	// retries equal to -1, it will run forever until success
	// retries is setted, it will retries fixed times.
	for i := 0; r.setting.Retries == -1 || i <= r.setting.Retries; i++ {
		resp, err = client.Do(r.req)
		if err == nil {
			break
		}
	}
	return resp, err
}

// String returns the body string in response.
// it calls Response inner.
func (r *HttpRequest) String() (string, error) {
	data, err := r.Bytes()
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Bytes returns the body []byte in response.
// it calls Response inner.
func (r *HttpRequest) Bytes() ([]byte, error) {
	if r.body != nil {
		return r.body, nil
	}
	resp, err := r.getResponse()
	if err != nil {
		return nil, err
	}
	if resp.Body == nil {
		return nil, nil
	}
	defer resp.Body.Close()
	if r.setting.Gzip && resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		r.body, err = ioutil.ReadAll(reader)
		return r.body, err
	}
	r.body, err = ioutil.ReadAll(resp.Body)
	return r.body, err
}

// ToFile saves the body data in response to one file.
// it calls Response inner.
func (r *HttpRequest) ToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := r.getResponse()
	if err != nil {
		return err
	}
	if resp.Body == nil {
		return nil
	}
	defer resp.Body.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

// ToJSON returns the map that marshals from the body bytes as json in response.
// it calls Response inner.
func (r *HttpRequest) ToJSON(v interface{}) error {
	data, err := r.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// ToXML returns the map that marshals from the body bytes as xml in response.
// it calls Response inner.
func (r *HttpRequest) ToYAML(v interface{}) error {
	data, err := r.Bytes()
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, v)
}

// Response executes request client gets response mannually.
func (r *HttpRequest) Response() (*http.Response, error) {
	return r.getResponse()
}


// TimeoutDialer returs functions of connection dialer with timeout settings for http.Transport Dial field.
func TimeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		err = conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, err
	}
}
