package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	// "html"
	"log"
	"net/http"
	// "time"
	// "flag"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

type Kv struct {
	Key   string `json:"key"`
	Value string `json:value,omitempty"`
}

//used to pack and unpack json
type KvArray struct {
	KvPair []Kv `json:"kvpair"`
}

//use a map to store data
var Data = make(map[string]string)

var proxyip = "http://localhost:8080"
var selfip string //selfip

var sigs = make(chan os.Signal, 1)

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

//return a int hash value
func md5hash(str string) int {
	hash := md5.Sum([]byte(str))
	//res := hex.EncodeToString(hash[:])
	res := binary.BigEndian.Uint32(hash[1:5]) //0-7Byte 64bit
	return int(res)
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

func fetchHandler(w http.ResponseWriter, r *http.Request, request_method string) {
	fmt.Println("fetchHandler...")
	var allrespData []Kv

	if request_method == "POST" { //get specific
		var data KvArray
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			panic(err)
		}
		// fmt.Println("%v fetching\n", data)

		for _, kv := range data.KvPair {
			// fmt.Println("%v fetching\n", kv)
			if v, ok := Data[kv.Key]; ok {
				allrespData = append(allrespData, Kv{kv.Key, v})
				// fmt.Println("%v Added\n", Kv{kv.Key, v})
			} else {
				allrespData = append(allrespData, Kv{kv.Key, "None"})
			}
		}
	} else if request_method == "GET" { //get all
		fmt.Println(Data)
		for k, v := range Data {
			allrespData = append(allrespData, Kv{k, v})
		}
	}
	fmt.Println(string(tojson(allrespData)))
	fmt.Fprint(w, string(tojson(allrespData)))
}

func setHandler(w http.ResponseWriter, r *http.Request, request_method string) {
	// fmt.Println("setHandler...")
	var data KvArray
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		fmt.Println("error")
		panic(err)
	}
	var allrespData []Kv

	fmt.Println(data)

	for _, kv := range data.KvPair {
		// fmt.Println("%v Adding\n", kv)
		if _, ok := Data[kv.Key]; ok {
			allrespData = append(allrespData, Kv{kv.Key, "Updated"})
		} else {
			allrespData = append(allrespData, Kv{kv.Key, "Added"})
			// fmt.Println("%v Added\n", kv)
		}
		Data[kv.Key] = kv.Value
	}

	fmt.Fprint(w, string(tojson(allrespData)))
}

func deleteHandler(w http.ResponseWriter, r *http.Request, request_method string) {
	var data KvArray
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		fmt.Println("error")
		panic(err)
	}
	var allrespData []Kv

	fmt.Println(data)

	for _, kv := range data.KvPair {
		// fmt.Println("%v Adding\n", kv)
		if _, ok := Data[kv.Key]; ok {
			allrespData = append(allrespData, Kv{kv.Key, "Deleted"})
		} else {
			allrespData = append(allrespData, Kv{kv.Key, "Not exist"})
			// fmt.Println("%v Added\n", kv)
		}
		delete(Data, kv.Key) //doesn't return anything, and will do nothing if the specified key doesn't exist

	}

	fmt.Fprint(w, string(tojson(allrespData)))

}

func moveDataHandler(w http.ResponseWriter, r *http.Request, request_method string) {
	var data KvArray
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		fmt.Println("error")
		panic(err)
	}
	kv := data.KvPair[0]
	ip := kv.Value //get ip to move
	h := md5hash(ip)
	var dataToMove []Kv
	for k, v := range Data {
		itemh := md5hash(k)
		if itemh <= h || (itemh > h && itemh > md5hash(selfip)) { //two situation
			dataToMove = append(dataToMove, Kv{k, v})
		}
	}
	jsonData := tojson(dataToMove)
	sendRequest(ip, jsonData, request_method, "/set") //send data
	//detele moved data after moved
	movednum := len(dataToMove)
	for _, kv := range dataToMove {
		delete(Data, kv.Key)
	}
	fmt.Fprint(w, movednum)

}

func get(w http.ResponseWriter, r *http.Request) {
	fmt.Println("get...")
	if r.Method == "POST" { //get specific data
		fetchHandler(w, r, "POST")
	} else if r.Method == "GET" { //get all data
		fetchHandler(w, r, "GET")
	} else {
		fmt.Fprint(w, "Wrong url") //should be json!!
	}
}

func set(w http.ResponseWriter, r *http.Request) {
	fmt.Println("set...")
	if r.Method == "POST" { //get specific data
		setHandler(w, r, "POST")
	} else {
		fmt.Fprint(w, "Wrong url") //should be json!!
	}
}

func deleteEntry(w http.ResponseWriter, r *http.Request) {
	fmt.Println("delete...")
	if r.Method == "DELETE" { //get specific data
		deleteHandler(w, r, "DELETE")
	} else {
		fmt.Fprint(w, "Wrong url") //should be json!!
	}
}

func moveData(w http.ResponseWriter, r *http.Request) {
	fmt.Println("moveData...")
	if r.Method == "POST" { //get specific data
		moveDataHandler(w, r, "POST")
	} else {
		fmt.Fprint(w, "Wrong url") //should be json!!
	}
}

func add_to_proxy(ip string) {
	sendRequest(proxyip, tojson([]Kv{Kv{Key: "1", Value: ip}}), "POST", "/newnode")
}

func transferAllData() {
	//send selfip to proxy
	nextip := sendRequest(proxyip, tojson([]Kv{Kv{Key: "1", Value: selfip}}), "POST", "/deletenode")
	var dataToMove []Kv
	for k, v := range Data {
		dataToMove = append(dataToMove, Kv{k, v})
	}
	jsonData := tojson(dataToMove)
	//send all data set to the next ip
	sendRequest(nextip, jsonData, "POST", "/set") //send all data
}

func exitHandler() {
	sig := <-sigs
	fmt.Println()
	fmt.Println(sig)
	transferAllData()
	fmt.Println("Exiting..")
	os.Exit(0)
}

func runserver(port string) {
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go exitHandler()
	// http.HandleFunc("/", rootHandler) //register the handler
	// log.Fatal(http.ListenAndServe(":8080", nil))
	router := mux.NewRouter().StrictSlash(true)
	// router.HandleFunc("/query", query)
	router.HandleFunc("/get", get)
	router.HandleFunc("/set", set)
	router.HandleFunc("/delete", deleteEntry)
	router.HandleFunc("/movedata", moveData)
	fmt.Printf("Server Listen at localhost: %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Not enough arguments")
		return
	}
	port := os.Args[1]
	selfip = "http://localhost:" + port
	add_to_proxy(selfip)
	runserver(port)
	// port := flag.String("port", "8000", "Server Port")
}
