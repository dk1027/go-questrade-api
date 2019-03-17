package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/dk1027/go-questrade-api/api"
	"github.com/dk1027/go-questrade-api/controlflow"
)

func unpack(s []string, vars ...*string) {
	for i, str := range s {
		*vars[i] = str
	}
}
func main() {
	var cmd, arg1, arg2 string
	unpack(os.Args[1:], &cmd, &arg1, &arg2)

	switch cmd {
	case "redeem":
		Redeem(arg1, arg2)
	case "check":
		Check(arg1)
	default:
		log.Printf("Undefined cmd %s\n", cmd)
	}
}

func Redeem(refreshToken, output string) {
	session, err := api.Redeem(refreshToken)
	if err != nil {
		log.Fatalln(err)
	}

	j, _ := json.Marshal(session)
	_ = ioutil.WriteFile(output, j, 0644)
}

func Check(configFile string) {
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalln(err)
	}
	cf := controlflow.Parse(bytes)
	cf.Execute()
}
