package msmsbrowser

import (
	"github.com/noatgnu/reformatMS/fileHandler"
	"strconv"
)

type MSElement struct {
	Intensity float64
	MZ float64
	RT float64
	Protein string
	FragmentCharge int64
	IonType string
	Residue int64
	PrecursorMZ float64
	Peptide string
	PrecursorCharge int64
}

func ReadIonFile(filename string, fileType string, progressOutput chan string) (msChan chan MSElement) {
	ionFile := fileHandler.ReadFile(filename, 1)
	switch fileType {
	case "swath":
		msChan = Swath(ionFile, progressOutput)
	}
	return msChan
}

func Swath(fileO fileHandler.FileObject, progressOutput chan string) chan MSElement {
	msChan := make(chan MSElement)
	samples := len(fileO.Header) - 9
	go func() {
		for r := range fileO.OutputChan {
			for i := 0; i < samples; i++ {
				intensity, err := strconv.ParseFloat(r[9+i], 64)
				if err != nil {
					 progressOutput <- err.Error()
				}
				mz, err := strconv.ParseFloat(r[5], 64)
				if err != nil {
					progressOutput <- err.Error()
				}
				precursor, err := strconv.ParseFloat(r[2], 64)
				if err != nil {
					progressOutput <- err.Error()
				}
				rt, err := strconv.ParseFloat(r[4], 64)
				if err != nil {
					progressOutput <- err.Error()
				}
				pcharge, err := strconv.ParseInt(r[3], 10, 64)
				if err != nil {
					progressOutput <- err.Error()
				}
				fcharge, err := strconv.ParseInt(r[6], 10, 64)
				if err != nil {
					progressOutput <- err.Error()
				}
				res, err := strconv.ParseInt(r[8], 10, 64)
				if err != nil {
					progressOutput <- err.Error()
				}
				msChan <- MSElement{
					Intensity:      intensity,
					MZ:             mz,
					RT:             rt,
					Protein:        r[0],
					FragmentCharge: fcharge,
					IonType:        r[7],
					Residue:        res,
					PrecursorMZ:      precursor,
					Peptide:        r[1],
					PrecursorCharge: pcharge,
				}
			}
		}
		progressOutput <- "Completed"
		close(msChan)
	}()

	return msChan
}