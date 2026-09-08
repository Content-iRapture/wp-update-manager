package main

import (
	"os/exec"
	"fmt"
)

func main() {
	cmd := exec.Command("plesk", "bin", "pleskbackup", "--domains-name", "elliott.irapture.com", "--incremental")	
	out, err := cmd.Output()
	
	if (err != nil) {
		fmt.Printf("Error: %v", err)
	} else {
		fmt.Printf("Out: %v", out)
	}
}
