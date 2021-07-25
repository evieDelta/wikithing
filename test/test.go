package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/blarg", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Query())
		fmt.Println(r.URL.Query()["frag"])
		fmt.Println(r.URL.Query()["other"])
		fmt.Println(r.URL.Query().Encode())
	})

	go http.ListenAndServe(":5555", nil)

	resp, err := http.Get("http://localhost:5555/blarg?frag=1&frag=2&frag=3&other=1,2,3")
	fmt.Println(err)
	fmt.Println(resp.Status)

	time.Sleep(time.Second)
}
