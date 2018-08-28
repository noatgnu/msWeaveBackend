package msreformat

import (
	"bufio"
	"fmt"
	"github.com/noatgnu/reformatMS/fileHandler"
	"log"
	"os"
	"strconv"
	"strings"
)

func Reformat(ionFilename string, fdrFilename string, outputFilename string, ignoreBlank bool, cutoff float64, progressOutput chan string) {
	swathFile := fileHandler.ReadFile(ionFilename, 1)
	samples := len(swathFile.Header) - 9
	progressOutput <- fmt.Sprintf("%d Samples", samples)

	fdrFile := fileHandler.ReadFile(fdrFilename, 1)
	fdrMap := ExtractFDRMap(fdrFile, samples, cutoff)

	o, err := os.Create(outputFilename)
	if err != nil {
		progressOutput <- err.Error()
	}

	writer := bufio.NewWriter(o)
	writer.WriteString("ProteinName,PeptideSequence,PrecursorCharge,FragmentIon,ProductCharge,IsotopeLabelType,Condition,BioReplicate,Run,Intensity\n")
	outputChan := make(chan string)
	go ProcessIons(outputChan, swathFile, fdrMap, samples, ignoreBlank, cutoff)
	for r := range outputChan {
		writer.WriteString(r)
	}
	writer.Flush()
	o.Close()
	progressOutput <- "Completed."
}

func ProcessIons(outputChan chan string, swathFile fileHandler.FileObject, fdrMap map[string]map[string][]float64, samples int, ignoreBlank bool, cutoff float64) {

	//log.Println(fdrMap)
	swathSampleMap := make(map[string][]string)
	log.Println("Processing ions using FDR mapped accession IDs.")
	for c := range swathFile.OutputChan {
		count := 0
		temp := ""
		if v, ok := fdrMap[c[0]]; ok {
			if val, ok := v[c[1]]; ok {
				for i := 0; i < samples; i++ {
					//log.Println(swathFile.Header[9+i])
					var sample []string
					if val1, ok := swathSampleMap[swathFile.Header[9+i]]; ok {
						sample = val1
					} else {
						sample = strings.Split(swathFile.Header[9+i], "_")
						swathSampleMap[swathFile.Header[9+i]] = sample[:]
					}

					row := fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v,%v,", c[0], c[1], c[3], c[7]+c[8], c[6], "L",
						sample[0],
						swathFile.Header[9+i],
						strconv.Itoa(i+1))
					if val[i] < cutoff {
						row += c[9+i]
						if c[9+i] == "" {
							count += 1
						}
					} else {
						row += ""
						count += 1
					}
					row += "\n"
					temp += row

				}

				if !ignoreBlank {
					outputChan <- temp
				} else {
					if count < samples {
						outputChan <- temp
					}
				}
			}
		}

	}
	close(outputChan)
}

func ExtractFDRMap(fdrFile fileHandler.FileObject, samples int, cutoff float64) map[string]map[string][]float64 {
	fdrMap := make(map[string]map[string][]float64)
	log.Println("Mapping FDR to accession ID.")
	for c := range fdrFile.OutputChan {
		fdrFail := 0
		if _, ok := fdrMap[c[0]]; !ok {
			fdrMap[c[0]] = make(map[string][]float64)
		}

		var fdrArray []float64
		for i := 0; i < samples; i++ {
			val, err := strconv.ParseFloat(c[7+i], 64)
			if err != nil {
				log.Fatalln(err)
			}

			if val >= cutoff {
				fdrFail++
			}
			fdrArray = append(fdrArray, val)
		}
		if fdrFail < samples {

			fdrMap[c[0]][c[1]] = fdrArray
		}
	}
	return fdrMap
}