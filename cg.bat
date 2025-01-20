curl localhost:8080/update/ -H "Content-Type":"application/json" -d "{\"type\":\"gauge\",\"id\":\"11nameofGau\",\"value\":1234.4321}"
curl localhost:8080/update/ -H "Content-Type":"application/json" -d "{\"type\":\"counter\",\"id\":\"11nameofCou\",\"delta\":1234}"
curl localhost:8080/value/ -H "Content-Type":"application/json" -d "{\"type\":\"gauge\",\"id\":\"11nameofGau\"}"
curl localhost:8080/value/ -H "Content-Type":"application/json" -d "{\"type\":\"counter\",\"id\":\"11nameofCou\"}"