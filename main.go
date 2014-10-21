package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/makpoc/sgfparser/logger"
	"github.com/makpoc/sgfparser/parser"
	"github.com/makpoc/sgfparser/structures"
)

func printUsage() {
	fmt.Printf("Usage: %s file.sgf", os.Args[0])
	os.Exit(1)
}

func dumpTree(gTree *structures.GameTree, identLevel int) {
	fmt.Printf("%s %s\n", strings.Repeat("-", identLevel), gTree.Sequence)

	if len(gTree.Children) > 0 {
		for _, child := range gTree.Children {
			dumpTree(child, identLevel+1)
		}
	}

}

func main() {

	if len(os.Args) < 2 {
		printUsage()
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		logger.LogError("Failed to open file!")
		os.Exit(1)
	}

	defer func() {
		if err := file.Close(); err != nil {
			logger.LogError("Failed to close file!")
		}
	}()

	collection, err := parser.ParseCollection(bufio.NewReader(file))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	for _, tree := range collection.GameTrees {
		dumpTree(tree, 0)
	}

}
