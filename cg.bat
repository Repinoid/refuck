curl localhost:8080/update/ -H "Content-Type":"application/json" -d "{\"type\":\"gauge\",\"id\":\"nameofGau\",\"value\":1234.4321}"
curl localhost:8080/update/ -H "Content-Type":"application/json" -d "{\"type\":\"counter\",\"id\":\"nameofCou\",\"delta\":1234}"
curl localhost:8080/value/ -H "Content-Type":"application/json" -d "{\"type\":\"gauge\",\"id\":\"nameofGau\"}"
curl localhost:8080/value/ -H "Content-Type":"application/json" -d "{\"type\":\"counter\",\"id\":\"nameofCou\"}"