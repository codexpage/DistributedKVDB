package main

import (
	"encoding/json"
	"fmt"
	// "github.com/gorilla/mux"
	// "html"
	"io/ioutil"
	// "log"
	"net/http"
	// "time"
	"strings"
)

func main() {
	// // Make a get request
	// rs, err := http.Get("https://google.com")
	// // Process response
	// if err != nil {
	// 	panic(err) // More idiomatic way would be to print the error and die unless it's a serious error
	// }
	// defer rs.Body.Close()

	// bodyBytes, err := ioutil.ReadAll(rs.Body)
	// if err != nil {
	// 	panic(err)
	// }

	// bodyString := string(bodyBytes)
	// fmt.Println(bodyString)

	// client := &http.Client{}
	// req, err := http.NewRequest("GET", "http://example.com", nil)
	// req.Header.Add("If-None-Match", `some value`)
	// resp, err := client.Do(req)
	httpRequest()
}

//get all data??
func get() {

}

func put() {

}
func post() {

}

func sendRequest(nodeid int, jsonData []byte, request_type string) string {
	url := server_addr[nodeid]
	body := bytes.NewBuffer(jsonData)

	if request_type == "PUT" || request_type == "POST" || request_type == "DELETE" {
		req, err := http.NewRequest(request_type, url, body)
		client := &http.Client{}
		resp, err := client.Do(req) //resp is http response
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		content, _ := ioutil.ReadAll(resp.Body)
		return string(content)
	} else { // GET operation
		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		content, _ := ioutil.ReadAll(resp.Body)
		return string(content)
	}
}
