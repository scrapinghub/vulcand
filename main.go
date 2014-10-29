package main

import (
	"fmt"
	"os"

	"github.com/mailgun/vulcand/service"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	r, err := GetRegistry()
	if err != nil {
		fmt.Printf("Service exited with error: %s\n", err)
		os.Exit(255)
	}
	if err := service.Run(r); err != nil {
		fmt.Printf("Service exited with error: %s\n", err)
		os.Exit(255)
	} else {
		fmt.Println("Service exited gracefully")
	}
}
