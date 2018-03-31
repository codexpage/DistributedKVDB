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
	"crypto/md5"
	"encoding/binary"
	"strconv"
	// "strings"
	"bytes"
)

type Kv struct {
	Key   string `json:"key"`
	Value string `json:value,omitempty"`
}

//used to pack and unpack json
type KvArray struct {
	KvPair []Kv `json:"kvpair"`
}

var proxyip = "http://localhost:8080"

//get all data??
func get() {

}

func set_test() {
	var list []Kv
	for i := 0; i < 100; i++ {
		s := strconv.Itoa(i)
		list = append(list, Kv{s, s})
		fmt.Println(md5hash(s))
	}
	jsonData := tojson(list)
	sendRequest(proxyip, jsonData, "POST", "/set")
}

func post() {

}

func tojson(data []Kv) []byte {
	var jsonData []byte
	// kvs_amount := 0

	if len(data) == 0 {
		return jsonData
	}

	jsonData, err_data := json.Marshal(KvArray{data})
	if err_data != nil {
		fmt.Println("Error in JSON formatting !! ")
	}
	return jsonData

}

func md5hash(str string) int {
	hash := md5.Sum([]byte(str))
	//res := hex.EncodeToString(hash[:])
	res := binary.BigEndian.Uint32(hash[1:5]) //0-7Byte 64bit
	return int(res)
}

func md5_test() {
	for i := 8001; i < 8003; i++ {
		s := "http://localhost:" + strconv.Itoa(i)
		h := md5hash(s)
		fmt.Println(i, h)
	}
}

func sendRequest(host string, jsonData []byte, request_type string, endpoint string) string {
	url := host + endpoint
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

func main() {
	md5_test()
	set_test()
}
