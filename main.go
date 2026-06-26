package main

import (
	"fmt"
	"os"

	"zabbix-domain-expiry/internal/checkdomain"
	"zabbix-domain-expiry/internal/output"
)

func main() {
	opts, err := checkdomain.ParseArgs(os.Args[1:])
	if err != nil {
		output.Exit(output.StateUnknown, 0, "", fmt.Sprintf("State: UNKNOWN ; %s", err))
	}

	checkdomain.Run(opts)
}