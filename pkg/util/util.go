package util

import (
	"net/url"
	"strings"
	"github.com/supeanut/ghttpload/request"
)

// GetNameAndExt return the name and ext of the URL
func GetNameAndExt(uri string) (string, string, error) {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return "","",err
	}
	s := strings.Split(u.Path, "/")
	filename := strings.Split(s[len(s)-1], ".")
	if len(filename) > 1 {
		return filename[0], filename[1], nil
	}

	contentType, err := request.ContentType(uri)
	if err != nil {
		return "", "", err
	}
	return filename[0], strings.Split(contentType, "/")[1], nil
}
