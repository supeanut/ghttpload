package util

import (
	"net/url"
	"strings"
	"github.com/supeanut/ghttpload/request"
	"runtime"
	"fmt"
	"path/filepath"
	"os"
)

// MAXLENGTH Maximum length of file name
const MAXLENGTH = 80

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

// FileName Converts a string to a valid filename
func FileName(name string) string {
	rep := strings.NewReplacer("\n", " ", "/", " ", "|", "-", ": ", "：", ":", "：", "'", "’")
	if runtime.GOOS == "windows" {
		rep = strings.NewReplacer("\"", " ", "?", " ", "*", " ", "\\", " ", "<", " ", ">", " ")
		name = rep.Replace(name)
	}
	return LimitLength(name, MAXLENGTH)
}

// LimitLength Handle overly long strings
func LimitLength(s string, length int) string {
	const ELLIPSES = "..."
	str := []rune(s)
	if len(str) > length {
		return string(str[:length-len(ELLIPSES)])
	}
	return s
}

// FilePath gen valid file path
func FilePath(name, ext, path string, escape, rename bool) (string, error) {
	var outputPath,fileName string
	if path != "" {
		if _, err := os.Stat(path); err != nil {
			return "", err
		}
	}
	if rename {
		fileName = fmt.Sprintf("%s.%s", name, ext)
	} else {
		fileName = name
	}

	if escape {
		fileName = FileName(fileName)
	}
	outputPath = filepath.Join(path, fileName)
	return outputPath, nil
}

// FileSize return the file size of the specified path file
func FileSize(filePath string) (int64, bool, error) {
	file, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, false, nil
		}
		return 0, false, nil
	}
	return file.Size(), true, nil
}