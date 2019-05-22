package ghttpload

import "github.com/supeanut/ghttpload/porter"

/*
	Package ghttpload is used as downloader

	provide apis for using
*/


var defaultPorter = porter.NewPorter()

// set porter path
func SetPath(path string) {
	defaultPorter.SetPath(path)
}

// set porter filename
func SetFilename(filename string) {
	defaultPorter.SetFilename(filename)
}

// set porter url
func SetUrl(url string) {
	defaultPorter.SetUrl(url)
}