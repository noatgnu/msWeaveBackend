package dispatcher

import (
	"github.com/gorilla/websocket"
	"github.com/noatgnu/msWeaveBackend/msmsbrowser"
	"github.com/noatgnu/msWeaveBackend/msreformat"
	"log"
	"strconv"
)

type Job struct {
	Name string `json:"name"`
	Data interface{} `json:"data"`
}

type SocketEvent struct {
	Event   string         `json:"event"`
	Message Job `json:"msg"`
}

type DispatchFactory struct {
	Factory map[string]func(interface{})
}

func (f DispatchFactory) InitDispatchFactory() {
	//m := make(map[string]func(interface{}))
}

func Dispatch(job Job, c *websocket.Conn) chan string  {
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
		msChan := msmsbrowser.ReadIonFile(detail["filePath"].(string), detail["fileType"].(string), progressOutput)
		for i := range msChan {
			c.WriteJSON(SocketEvent{Event: "csv", Message: Job{"msmsfile", i}})
		}
	}
	return progressOutput
}
