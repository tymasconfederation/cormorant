package main

import (
	"fmt"
	"io"
	"os"

	"github.com/tymasconfederation/cormorant"
)

var enableRecover bool = false
var authToken string

func Fscan(r io.Reader, a ...interface{}) {
	_, err := fmt.Fscan(r, a...)
	if err != nil {
		panic(fmt.Sprintf("Scan call failed to read from input because %s", err))
	}
}

func main() {
	defer func() {
		if enableRecover {
			if r := recover(); r != nil {
				fmt.Println("Error in main(): ", r)
			}
		}
	}()

	appID := os.Getenv("appid")
	if len(appID) == 0 {
		fmt.Println("App ID needs to be specified in the appid environment variable.")
	}

	authToken := os.Getenv("authtoken")
	if len(authToken) == 0 {
		fmt.Println("Auth token needs to be specified in the authtoken environment variable.")
	}

	if len(appID) == 0 || len(authToken) == 0 {
		return
	}

	runningChan := make(chan int)
	disc := cormorant.NewDiscordUI(appID, authToken, runningChan)
	go disc.Run()
	os.Exit(<-runningChan)
}
