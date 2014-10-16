package main

import (
	"fmt"
	"os"

	"github.com/makpoc/sgfparser/logger"
)

func printUsage() {
	fmt.Printf("Usage: %s file.sgf", os.Args[0])
	os.Exit(1)
}
func main() {

	if len(os.Args) < 2 {
		printUsage()
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		logger.LogError("Failed to open file!")
		os.Exit(-1)
	}

	defer func() {
		if err := file.Close(); err != nil {
			logger.LogError("Failed to close file!")
		}
	}()

	/*
		LOG_LEVEL = DEBUG
		gameTree := "(;FF[3];BB[asd](;SU[3])(;))"
		fmt.Println(ParseGameTree([]rune(gameTree)))
	*/
}
