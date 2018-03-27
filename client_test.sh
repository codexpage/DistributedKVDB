
#tests for node 
curl -d '{"kvpair":[{"key":"1","Value":"apple"},{"key":"2","Value":"orange"}]}' -X POST -H "Content-Type: application/json" localhost:8000/set

curl -d '{"kvpair":[{"key":"3","Value":"banana"},{"key":"4","Value":"pear"}]}' -X POST -H "Content-Type: application/json" localhost:8000/set


curl -d '{"kvpair":[{"key":"2","Value":"orange"}]}' -X POST -H "Content-Type: application/json" localhost:8000/get

curl -H "Accept: application/json" -X GET localhost:8000/get

#test for proxy

curl -H "Accept: application/json" -X GET localhost:8080/get


#for delete

curl -d '{"kvpair":[{"key":"1","Value":"apple"},{"key":"2","Value":"orange"}]}' -X POST -H "Content-Type: application/json" localhost:8080/set
curl -d '{"kvpair":[{"key":"3","Value":"banana"},{"key":"4","Value":"pear"}]}' -X POST -H "Content-Type: application/json" localhost:8080/set
curl -H "Accept: application/json" -X GET localhost:8080/get
curl -d '{"kvpair":[{"key":"3","Value":"banana"}]}' -X DELETE -H "Content-Type: application/json" localhost:8080/delete
curl -H "Accept: application/json" -X GET localhost:8080/get
