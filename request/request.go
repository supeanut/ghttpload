package request

import (
	"net/http"
	"github.com/supeanut/ghttpload/httplib"
	"strings"
	"strconv"
)

func getHeader(url string) (http.Header, error) {
	resp, err := httplib.Head(url).Response()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp.Header, nil
}

func ContentType(url string) (string, error) {
	h, err := getHeader(url)
	if err != nil {
		return "", err
	}
	s := h.Get("Content-Type")
	// handle Content-Type like this: "text/html; charset=utf-8"
	return strings.Split(s, ";")[0], nil
}

func GetContentSize(url string) (int64, error) {
	h, err := getHeader(url)
	if err != nil {
		return 0, err
	}
	s := h.Get("Content-Length")
	size, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return size, nil
}
