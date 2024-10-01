package main

import (
	qecache "QECache"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	qecache.NewController("scores", 2<<10, qecache.FetcherFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9998"
	peers := qecache.NewHTTPServer(qecache.HTTPServerConfig{
		SelfIP: addr,
	})
	log.Println("Cache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
