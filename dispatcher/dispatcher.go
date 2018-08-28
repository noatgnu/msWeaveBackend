package dispatcher

import (
	"github.com/noatgnu/msWeaveBackend/msreformat"
	"log"
	"strconv"
)

type Job struct {
	Name string `json:"name"`
	Data interface{} `json:"data"`
}

type DispatchFactory struct {
	Factory map[string]func(interface{})
}

func (f DispatchFactory) InitDispatchFactory() {
	//m := make(map[string]func(interface{}))
}

func Dispatch(job Job) chan string  {
	detail := job.Data.(map[string]interface{})
	progressOutput := make(chan string)
	switch job.Name {
	case "msreformat":
		cutoff, err := strconv.ParseFloat(detail["p"].(string), 64)
		if err != nil {
			log.Printf("err %v", err)
		}
		go msreformat.Reformat(detail["ion"].(string), detail["fdr"].(string), detail["output"].(string), true, cutoff, progressOutput)
	case "msmsfile":

	}
	return progressOutput
}
