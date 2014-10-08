package main

import "encoding/json"
import "fmt"
import "os"
import "strings"

import "github.com/3ofcoins/shlike"

func main() {
	cfg := shlike.NewConfig()
	for _, arg := range os.Args[1:] {
		if splut := strings.SplitN(arg, "=", 2); len(splut) == 1 {
			if err := cfg.Load(arg); err != nil {
				panic(err)
			}
		} else {
			cfg.Append(splut[0], splut[1])
		}
	}
	if json_bb, err := json.MarshalIndent(cfg, "", "  "); err != nil {
		panic(err)
	} else {
		fmt.Println(string(json_bb))
	}
}
