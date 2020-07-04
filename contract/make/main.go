package main

import (
	"log"
	"os"

	"kortho/contract"
	"kortho/contract/motor"
)

func main() {
	name, err := contract.Motor(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	motor.Asm(name)
}
