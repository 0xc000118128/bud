package main

import "fmt"

func main() {
	if err := setupCli(); err != nil {
		fmt.Printf("error: %s", err.Error())
	}
}
