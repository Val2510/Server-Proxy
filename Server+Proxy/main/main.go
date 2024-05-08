package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const proxyAddr string = "localhost:9000"

var (
	counter            int    = 0
	firstInstanceHost  string = "localhost:8080"
	secondInstanceHost string = "localhost:8081"
)

func main3() {
	http.HandleFunc("/", handleProxy)
	log.Fatalln(http.ListenAndServe("localhost:9000", nil))
}

func handleProxy(w http.ResponseWriter, r *http.Request) {
	textBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalln(err)
	}
	file, err := os.Create("somefile.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	_, err = file.Write(textBytes)
	if err != nil {
		log.Fatalln(err)
	}

	text := string(textBytes)

	if counter == 0 {
		if _, err := http.Post("http://"+firstInstanceHost, "text/plain", bytes.NewBuffer([]byte(text))); err != nil {
			log.Fatalln(err)
		}
		counter++
		return
	}
	if _, err := http.Post("http://"+secondInstanceHost, "text/plain", bytes.NewBuffer([]byte(text))); err != nil {
		log.Fatalln(err)
	}
	counter--
}
