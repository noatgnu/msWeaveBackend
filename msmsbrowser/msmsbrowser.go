package msmsbrowser

import "github.com/noatgnu/reformatMS/fileHandler"

func ReadIonFile(filename string) {
	ionFile := fileHandler.ReadFile(filename, 1)
	samples := len(ionFile.Header) - 9
}
