package main

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"github.com/noatgnu/msWeaveBackend/dispatcher"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

type LogEvent struct {
	Name string
	Detail string
}

type configuration struct {
	name   string //Application Name
	server string
	port   string
}

var done = make(chan bool, 1)      //Wait for Shutdown signal over websocket
var upgrader = websocket.Upgrader{ //Upgrader for websockets
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	Subprotocols:    []string{"p0", "p1"},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func loadConfig() configuration {
	viper.SetConfigName("config")

	// Paths to search for a config file
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		color.Set(color.FgRed)
		fmt.Println("No configuration file loaded - using defaults")
		color.Unset()
	}

	// default values
	viper.SetDefault("name", "MSWeave")
	viper.SetDefault("server", "localhost")
	viper.SetDefault("port", "8888")

	// Write all params to stdout
	color.Set(color.FgGreen)
	fmt.Println("Loaded Configuration:")
	color.Unset()

	// Print config
	keys := viper.AllKeys()
	for i := range keys {
		key := keys[i]
		fmt.Println(key + ":" + viper.GetString(key))
	}
	fmt.Println("---")

	return configuration{
		name:   viper.GetString("name"),
		server: viper.GetString("server"),
		port:   viper.GetString("port")}
}

//Handles msgs to communicate with nodejs electron for rampup & shutdown
func socket(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		var event dispatcher.SocketEvent
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("ElectronSocket: [err]", err)
			break
		}

		//Handle Message
		err = json.Unmarshal(message, &event)
		if err != nil {
			log.Println("Unmashal: ", err)
			break
		}
		log.Printf("ElectronSocket: [received] %+v", event)

		//Shutdown Event
		switch event.Event {
		case "shutdown":
			done <- true
		case "connected":
			log.Println("Connected")
		case "job":
			log.Println(event.Message)
			progressC := dispatcher.Dispatch(event.Message, c)
			for r := range progressC {

				if r == "Completed." {
					c.WriteJSON(event.Message)
				}
			}
		}
	}
}

func socketCSV(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		var event dispatcher.SocketEvent
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("ElectronSocket: [err]", err)
			break
		}

		//Handle Message
		err = json.Unmarshal(message, &event)
		if err != nil {
			log.Println("Unmashal: ", err)
			break
		}
		log.Printf("ElectronSocket: [received] %+v", event)

		//Shutdown Event
		switch event.Event {
		case "connected":
			log.Println("Connected")
		case "job":
			log.Println(event.Message)
			progressC := dispatcher.Dispatch(event.Message, c)
			for r := range progressC {

				if r == "Completed." {
					c.WriteJSON(event.Message)
				}
			}
		}
	}
}

func main() {
	config := loadConfig()

	var addr = config.server + ":" + config.port
	http.HandleFunc("/ui", socket)    //Endpoint for Electron startup/teardown
	http.HandleFunc("/csv", socket)
	go http.ListenAndServe(addr, nil) //Start websockets in goroutine

	color.Set(color.FgGreen)
	log.Printf("%s succesfully started", config.name)
	color.Unset()

	<-done //Wait for shutdown signal
	color.Set(color.FgGreen)
	log.Printf("Shutting down...")
	color.Unset()
}