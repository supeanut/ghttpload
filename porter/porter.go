package porter

import (
	"github.com/supeanut/ghttpload/httplib"
	"strings"
	"github.com/supeanut/ghttpload/pkg/util"
	"github.com/supeanut/ghttpload/request"
	"github.com/cheggaaa/pb"
	"time"
	"fmt"
	"os"
	"io"
)

type Porter struct {
	// path for download
	Path string
	// filename for download
	Filename string
	// rename for download
	Rename bool
	// steam for download
	Stream Stream
	// Err is used to record whether an error occurred when extracting data.
	Err error
	// Request is used to request remote
	Request httplib.HttpRequest
	// if set to -1 means will retry forever
	Retries int
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

func (p *Porter) SetRetries(n int) {
	p.Retries = n
}

func (p *Porter) SetPath(path string) {
	p.Path = strings.TrimSpace(path)
}

func (p *Porter) SetFilename(filename string) {
	p.Filename = strings.TrimSpace(filename)
	p.Rename = false
}

func (p *Porter) SetUrl(url string) {
	p.Stream.URL.Url = strings.TrimSpace(url)
}

func (p *Porter) extract() error {
	filename, ext, err := util.GetNameAndExt(p.Stream.URL.Url)
	if err != nil {
		return err
	}
	size, err := request.GetContentSize(p.Stream.URL.Url)
	if err != nil {
		return err
	}
	if p.Filename == "" {
		p.Filename = filename
	}
	p.Stream.URL.Ext = ext
	p.Stream.URL.Size = size
	return nil
}

func (p *Porter) Extract() error {
	err := p.extract()
	if err != nil {
		return err
	}
	return nil
}

func (p *Porter) Download() error{

	// check filename
	p.Filename = util.FileName(p.Filename)

	bar := progressBar(p.Stream.URL.Size)
	bar.Start()
	err := p.save(bar)
	if err != nil {
		return err
	}
	bar.Finish()
	return nil
}


func (p *Porter) writeFile(file *os.File, headers map[string]string, bar *pb.ProgressBar) (int64, error) {
	resp, err := request.GetFile(p.Stream.URL.Url, headers)
	if err != nil {
		return 0, err
	}
	if resp == nil {
		return 0, fmt.Errorf("nil response")
	}
	if resp.Body == nil {
		return 0, nil
	}
	defer resp.Body.Close()
	writer := io.MultiWriter(file, bar)
	// Note that io.Copy reads 32kb(maximum) from input and writes them to output
	// So don't worry about memory.
	written, copyErr := io.Copy(writer, resp.Body)
	if copyErr != nil {
		return written, fmt.Errorf("file copy error: %s", copyErr)
	}
	return written, nil
}

func (p *Porter) GetFileSize() (int64, error) {
	// check path
	filePath, err := util.FilePath(p.Filename, p.Stream.URL.Ext, p.Path,false, p.Rename)
	if err != nil {
		return 0, err
	}
	// check file
	fileSize, exists, err := util.FileSize(filePath)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, nil
	}
	return fileSize, nil
}

func (p *Porter) save(bar *pb.ProgressBar) (err error) {
	// check path
	filePath, err := util.FilePath(p.Filename, p.Stream.URL.Ext, p.Path,false, p.Rename)
	if err != nil {
		return err
	}
	// check file
	fileSize, exists, err := util.FileSize(filePath)
	if err != nil {
		return err
	}

	if bar == nil {
		bar := progressBar(p.Stream.URL.Size)
		bar.Start()
	}

	if exists && fileSize == p.Stream.URL.Size {
		bar.Add64(fileSize)
		return nil
	}

	tempFilePath := filePath
	tempFileSize, _, err := util.FileSize(tempFilePath)
	if err != nil {
		return err
	}

	headers := map[string]string{}
	var (
		file    *os.File
		fileError error
	)
	if tempFileSize > 0 {
		// range start from 0, 0-1023 means the first 1024 bytes of the file
		headers["Range"] = fmt.Sprintf("bytes=%d-", tempFileSize)
		file, fileError = os.OpenFile(tempFilePath, os.O_APPEND|os.O_WRONLY, 0644)
		bar.Add64(tempFileSize)
	} else {
		file, fileError = os.Create(tempFilePath)
	}
	if fileError != nil {
		return fileError
	}
	// begin download
	temp := tempFileSize
	for i := 0; p.Retries == -1 || i <= p.Retries; i++ {
		written, err := p.writeFile(file, headers, bar)
		if err == nil {
			break
		}
		temp += written
		headers["Range"] = fmt.Sprintf("bytes=%d-", temp)
		time.Sleep(1 * time.Second)
	}

	// close file
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	return nil
}


func progressBar(size int64) *pb.ProgressBar {
	bar := pb.New64(size).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10)
	bar.ShowSpeed = true
	bar.ShowFinalTime = true
	bar.SetMaxWidth(1000)
	return bar
}