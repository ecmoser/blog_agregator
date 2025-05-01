package main

import (
	"fmt"
	"log"
	"os"

	cfg "github.com/ecmoser/blog_aggregator/internal/config"
)

func main() {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	config, err := cfg.Read(workingDir)
	if err != nil {
		log.Fatal(err)
	}

	config.SetUser("Eli", workingDir)

	config, err = cfg.Read(workingDir)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(config)
}
