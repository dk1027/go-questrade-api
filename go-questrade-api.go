package go_questrade_api

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func get() {
	resp, err := http.Get("https://www.google.ca")
	if err != nil {
		fmt.Printf("error %s\n", err)
		return
	}
	data, _ := ioutil.ReadAll(resp.Body)
}
