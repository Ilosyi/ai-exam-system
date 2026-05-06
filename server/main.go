package main

import (
	"log"

	"week05/homework/server/bootstrap"
)

func main() {
	if err := bootstrap.Start(""); err != nil {
		log.Fatal(err)
	}
}
