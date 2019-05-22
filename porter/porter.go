package porter

import (
	"github.com/supeanut/ghttpload/httplib"
	"strings"
	"github.com/supeanut/ghttpload/pkg/util"
	"github.com/supeanut/ghttpload/request"
)

type Porter struct {
	// path for download
	Path string
	// filename for download
	Filename string
	// steam for download
	Stream Stream
	// Err is used to record whether an error occurred when extracting data.
	Err error
	// Request is used to request remote
	Request httplib.HttpRequest
}

type Stream struct {
	URL URL
	// total size of stream
	Size int64
	// name used in storedStream
	name string
}

// URL for single URL information
type URL struct {
	Url string
	Size int64
	Ext  string
}

func NewPorter() *Porter {
	return &Porter{}
}

func (p *Porter) SetPath(path string) {
	p.Path = strings.TrimSpace(path)
}

func (p *Porter) SetFilename(filename string) {
	p.Filename = strings.TrimSpace(filename)
}

func (p *Porter) SetUrl(url string) {
	p.Stream.URL.Url = strings.TrimSpace(url)
}

func (p *Porter) Extract() error {
	filename, ext, err := util.GetNameAndExt(p.Stream.URL.Url)
	if err != nil {
		return err
	}
	size, err := request.GetContentSize(p.Stream.URL.Url)
	if err != nil {
		return err
	}
	p.Filename = filename
	p.Stream.URL.Ext = ext
	p.Stream.URL.Size = size
	return nil
}

func (p *Porter) Download() error{
	if p.Filename == "" {
		err := p.Extract()
		if err != nil {
			return err
		}
	}
}