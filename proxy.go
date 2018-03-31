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
	"crypto/md5"
	"encoding/binary"
	"io/ioutil"
	"sort"
	"strconv"
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

//store the nodeid->ip //should be hashv/name -> ip
var server_addr = make(map[int]string)

var ring = make([]int, 0, 10) //default len=0 cap=10

//hash->ip
var hash2ip = make(map[int]string)

//ip->count
var server_data_cnt = make(map[string]int)

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

//return a int hash value
func md5hash(str string) int {
	hash := md5.Sum([]byte(str))
	//res := hex.EncodeToString(hash[:])
	res := binary.BigEndian.Uint32(hash[1:5]) //0-7Byte 64bit
	return int(res)
}

func md5_test() {
	for i := 8001; i < 8051; i++ {
		s := "http://localhost:" + strconv.Itoa(i)
		h := md5hash(s)
		fmt.Println(i-8000, h)
	}
}

//input a hash,return next hash
func insertNodeIntoRing(h int) int {
	if server_amount == 0 {
		fmt.Println("First node insert into Ring")
		ring = append(ring, h) // directly append
		return -1              //no next node
	}

	i := sort.Search(len(ring), func(i int) bool { return ring[i] >= h })
	nextindex := i % server_amount
	nextNodeh := ring[nextindex]
	ring = append(ring[:i], append([]int{h}, ring[i:]...)...) //insert at i
	return nextNodeh
	//get new node hash from nodename/id
	//find the place to insert on the ring
	//insert into slice
	//hash2node[hash]=nodename/id
	//amount++
}

func nextNodeIndex(h int) int {
	index := sort.Search(len(ring), func(i int) bool { return ring[i] >= h }) % server_amount //cound be the last one
	return index
}

//input a hash(node or item hash), return next node hash
func nextNodeHash(h int) int {
	index := sort.Search(len(ring), func(i int) bool { return ring[i] >= h }) % server_amount //cound be the last one
	return ring[index]
}

func deleteNodeFromRing() {
	//get node hash from nodename
	//find the node hash in the ring
	//delete from slice
	//delete hash2node[hash]
	//amount--
}

func moveDataRequest(receiveip string, newip string) {
	//send signal to receiveip to move data to newip
	res := sendRequest(receiveip, tojson([]Kv{Kv{Key: "1", Value: newip}}), "POST", "/movedata")
	fmt.Println(res)
	movednum, _ := strconv.Atoi(res)
	server_data_cnt[receiveip] -= movednum
	server_data_cnt[newip] += movednum
}

//input a kv, return the next node index
// func getnodeNum(kv Kv) int {
// 	h := md5hash(kv.Key)
// 	//next node index in ring
// 	index := sort.Search(len(ring), func(i int) bool { return ring[i] >= h }) % server_amount //cound be the last one
// 	// server_data_cnt[hash2ip[ring[index]]] += 1

// }

//input kev-value pairs list, return which server the key-value belongs matrix
func hash2(kv []Kv) map[string][]Kv {
	server_data := make(map[string][]Kv)
	for _, v := range kv {
		targetnodeh := nextNodeHash(md5hash(v.Key))
		ip := hash2ip[targetnodeh]
		server_data[ip] = append(server_data[ip], v)
	}
	fmt.Println(server_data, ring, hash2ip)
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

func sendRequest_test() {
	fmt.Println(sendRequest("http://www.google.com", json_test(), "GET", ""))
}

func postHandler(w http.ResponseWriter, r *http.Request, request_method string, endpoint string) {
	// jsonVal := json.NewDecoder(r.Body)
	// err := jsonVal.Decode(&msg)
	var data KvArray
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		panic(err)
	}

	dataDisro := hash2(data.KvPair)

	var allrespData []Kv
	fmt.Println(dataDisro)
	// fmt.Println(server_addr)
	fmt.Println(hash2ip)

	for nodeip, list := range dataDisro {
		if len(list) == 0 {
			continue //no data on this server
		}

		if endpoint == "/set" { //maintain the data_cnt
			server_data_cnt[nodeip] += len(list)
		} else if endpoint == "/delete" {
			server_data_cnt[nodeip] -= len(list)
		}

		jsonData := tojson(list)
		resp_content := sendRequest(nodeip, jsonData, request_method, endpoint) //send req to server with that nodeid

		var resp_data KvArray
		err := json.Unmarshal([]byte(resp_content), &resp_data)
		if err != nil {
			panic(err)
		}
		allrespData = append(allrespData, resp_data.KvPair...) //add the allData
	}
	fmt.Fprint(w, string(tojson(allrespData))) //response

}

//get all data from all servers
func getHandler(w http.ResponseWriter, r *http.Request, endpoint string) {

	var allrespData []Kv

	for _, ip := range hash2ip {
		jsonData := []byte("") //empty json
		resp_content := sendRequest(ip, jsonData, "GET", endpoint)

		if server_data_cnt[ip] == 0 { //dont't fetch from empty node, it will return empty
			continue
		}
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
	kv := data.KvPair[0] //only one item

	ip := kv.Value
	h := md5hash(ip)
	hash2ip[h] = ip
	nextNodeh := insertNodeIntoRing(h)
	if server_amount != 0 {
		nextip := hash2ip[nextNodeh]
		if server_data_cnt[nextip] != 0 { //default is 0
			go moveDataRequest(nextip, ip) // send the signal to hash2i[[nextNodeh] with (ip) shouldn't be blocked
		}
	}
	// server_addr[server_amount] = kv.Value //add new server ip
	server_amount++
	fmt.Printf("New Node Added: %v\n", ip)
}

func deletenode(w http.ResponseWriter, r *http.Request) {
	var data KvArray
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		panic(err)
	}
	kv := data.KvPair[0] //only one item

	ip := kv.Value
	h := md5hash(ip)
	//delete the hash from the ring
	index := nextNodeIndex(h)                      //self index
	ring = append(ring[:index], ring[index+1:]...) //delete at index

	server_amount--
	nextNodeh := nextNodeHash(h)
	nextip := hash2ip[nextNodeh] //nextip

	delete(hash2ip, h) //delete from ip list

	server_data_cnt[nextip] += server_data_cnt[ip]
	delete(server_data_cnt, ip)
	//change server data cnt, delete cnt

	fmt.Printf("Node Deleted: %v\n", ip)
	fmt.Fprint(w, nextip)
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
	router.HandleFunc("/deletenode", deletenode)
	fmt.Printf("sever running at localhost:8080\n")
	log.Fatal(http.ListenAndServe(":8080", router))
}
func main() {
	runserver()
	// json_test()
	// md5_test()
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
