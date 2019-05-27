package main

import (
	"github.com/supeanut/ghttpload"
	"fmt"
	"time"
)

func main() {
	url := "http://xxxxxxxxxxx/youku15/85/5eefbfbd40db817a11ef/%25u89C2%25u590D%25u561F%25u561F%25u4E01%25u9149%25u7248/%25u4E09%25u4EBA%25u6210%25u864E%25u4E0D%25u6210%25u55B5%2520180103/XMzI4MzAxODA5Mg==/transcode/1920_1080_h264_aac_8000k_cbr.ts"
	porter := ghttpload.NewPorter()
	porter.SetRetries(100)
	porter.SetUrl(url)
	porter.SetFilename("abc")
	porter.Extract()
	fmt.Println(porter.Stream.URL.Size)
	go func() {
		for {
			s, err := porter.GetFileSize()
			fmt.Println("=========================")
			fmt.Printf("size:%d,err:%v\n",s,err)
			fmt.Println("=========================")
			time.Sleep(2 * time.Second)
		}


	}()
	porter.Download()
}
