package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	// "html"
	"log"
	"net/http"
	// "time"
	"bytes"
	"io/ioutil"
	"unicode/utf8"
)

//package requre:
//go get "github.com/gorilla/mux"

// type Todo struct {
// 	Name      string    `json:"name"`
// 	Completed bool      `json:"completed"`
// 	Due       time.Time `json:"due"`
// }

type Kv struct {
	Key   string `json:"key"`
	Value string `json:value,omitempty"`
}

//used to pack and unpack json
type KvArray struct {
	KvPair []Kv `json:"kvpair"`
}

//store the nodeid->ip
var server_addr = make(map[int]string, 50)

// var server_addr = map[int]string{0: "http://www.google.com"}
var server_amount = 0

//struct tag ensure json has lowercase name

func hash(kv []Kv) [][]Kv {
	server_data_amount := make([]int, server_amount)
	server_data := make([][]Kv, server_amount)

	for _, v := range kv {
		sum := 0
		key := v.Key
		for len(key) > 0 { //traverse the string, get the sum
			ch, size := utf8.DecodeRuneInString(key)
			sum += int(ch) //ascii value?
			key = key[size:]
		}
		hash := sum % server_amount
		server_data[hash] = append(server_data[hash], Kv{v.Key, v.Value})
		server_data_amount[hash]++
	}

	return server_data
}

// type Todos []Todo

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

func json_test() []byte {
	kvlist := []Kv{Kv{Key: "1", Value: "apple"}, Kv{Key: "2", Value: "orange"}}
	jsonData := tojson(kvlist)
	fmt.Println(string(jsonData))
	return jsonData //for test
}

func sendRequest(nodeid int, jsonData []byte, request_type string, endpoint string) string {
	url := server_addr[nodeid] + endpoint
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

func sendRequest_test() {
	fmt.Println(sendRequest(0, json_test(), "GET", ""))
}

func postHandler(w http.ResponseWriter, r *http.Request, request_method string, endpoint string) {
	// jsonVal := json.NewDecoder(r.Body)
	// err := jsonVal.Decode(&msg)
	var data KvArray
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		panic(err)
	}

	dataDisro := hash(data.KvPair)

	var allrespData []Kv
	fmt.Println(dataDisro)
	fmt.Println(server_addr)

	for nodeid, list := range dataDisro {
		if len(list) == 0 {
			continue //no data on this server
		}

		jsonData := tojson(list)
		resp_content := sendRequest(nodeid, jsonData, request_method, endpoint) //send req to server with that nodeid

		var resp_data KvArray
		err := json.Unmarshal([]byte(resp_content), &resp_data)
		if err != nil {
			panic(err)
		}
		allrespData = append(allrespData, resp_data.KvPair...) //add the allData
	}
	fmt.Fprint(w, string(tojson(allrespData))) //response

}

func getHandler(w http.ResponseWriter, r *http.Request, endpoint string) {

	var allrespData []Kv

	//get to ALL servers
	for nodeid, _ := range server_addr {
		jsonData := []byte("") //empty json
		resp_content := sendRequest(nodeid, jsonData, "GET", endpoint)

		var resp_data KvArray
		err := json.Unmarshal([]byte(resp_content), &resp_data)
		if err != nil {
			panic(err)
		}
		allrespData = append(allrespData, resp_data.KvPair...)
	}
	fmt.Fprint(w, string(tojson(allrespData)))
}

func get(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" { //get specific data
		postHandler(w, r, "POST", "/get")
	} else if r.Method == "GET" { //get all data
		getHandler(w, r, "/get")
	} else {
		fmt.Fprint(w, "Wrong url") //should be json!!
	}
}

func set(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" { //get specific data
		postHandler(w, r, "POST", "/set")
	} else {
		fmt.Fprint(w, "Wrong url") //should be json!!
	}
}

func deleteEntry(w http.ResponseWriter, r *http.Request) {
	fmt.Println("delete...")
	if r.Method == "DELETE" { //get specific data
		postHandler(w, r, "DELETE", "/delete")
	} else {
		fmt.Fprint(w, "Wrong url") //should be json!!
	}
}

func newnode(w http.ResponseWriter, r *http.Request) {
	var data KvArray
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		panic(err)
	}
	kv := data.KvPair[0]                  //only one item info
	server_addr[server_amount] = kv.Value //add new server ip
	server_amount++
	fmt.Printf("New Node Added: %v\n", kv.Value)
}

func runserver() {
	// http.HandleFunc("/", rootHandler) //register the handler
	// log.Fatal(http.ListenAndServe(":8080", nil))
	router := mux.NewRouter().StrictSlash(true)
	// router.HandleFunc("/query", query)
	router.HandleFunc("/get", get)
	router.HandleFunc("/set", set)
	router.HandleFunc("/delete", deleteEntry)
	router.HandleFunc("/newnode", newnode)
	log.Fatal(http.ListenAndServe(":8080", router))
}
func main() {
	runserver()
	// json_test()
}

// func Index(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path)) //write the response to writer
// }

// func TodoIndex(w http.ResponseWriter, r *http.Request) {
// 	// fmt.Fprintln(w, "Todo Index!")
// 	todos := Todos{
// 		Todo{Name: "Write presentation"},
// 		Todo{Name: "Host meetup"},
// 	}

// 	json.NewEncoder(w).Encode(todos)
// }

// func TodoShow(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r) //map [string] string
// 	todoId := vars["todoId"]
// 	fmt.Fprintln(w, "Todo show:", todoId)
// }
