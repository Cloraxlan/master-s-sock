package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sock"
)

var addr = flag.String("addr", ":8080", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/e" {
		http.ServeFile(w, r, "e.html")
		return
	}
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "index.html")
}

var hub = sock.NewHub()
var weather = "sunny"

type jsonMessage struct {
	Command string
	Text    string
}

func weatherRout() {
	for {
		i := <-hub.Input
		jsonRaw := i.Message
		fmt.Println(jsonRaw)
		jMessage := jsonMessage{}
		if err := json.Unmarshal([]byte(jsonRaw), &jMessage); err == io.EOF {

			break
		} else if err != nil {
			fmt.Println("ERROR: ", err)

		}
		fmt.Println(jMessage)

		fmt.Println("yes")
		if jMessage.Command == "weather" {

			i.Client.Send <- ([]byte(weather))
		} else if jMessage.Command == "change" {
			weather = (jMessage.Text)
		}
	}
}
func main() {
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		sock.ServeWs(hub, w, r)
	})
	go hub.Run()
	go weatherRout()
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		fmt.Println("ListenAndServe: ", err)
	}
}
