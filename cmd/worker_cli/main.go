package main

import (
	"fmt"

	"brainyping/pkg/checks/httpcheck"
	"brainyping/pkg/initapp"
	"brainyping/pkg/utilities"
)

func main() {
	initapp.InitApp()
	subType := utilities.ReadUserInput("CHECK SUB TYPE? (leave blank for GET) ")
	if subType == "" {
		subType = "GET"
	}
	url := utilities.ReadUserInput("HOST? (include protocol and port if necessary) ")
	ua := utilities.ReadUserInput("USER AGENT? (leave blank to use default) ")

	checkReponse, err := httpcheck.ProcessCheck(url, subType, ua)
	utilities.FailOnError(err)

	fmt.Printf("%v", checkReponse)

}
