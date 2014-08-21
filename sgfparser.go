package main

import (
	"fmt"
	"os"
)

const (
	DELIMITER string = ";" // delimits properties
)

type Property string

type Value string

type Node struct {
	Children   []Node
	Properties map[Property]Value
}

type SgfReader interface {
	ReadNext() interface{}
	ReadNode() *Node
	ReadProperty() Property
	ReadValue() Value
	ReadPropertyAndValue() (Property, Value)
}

func PrintUsage() {
	fmt.Printf("Usage: %s file.sgf\n", os.Args[0])
	os.Exit(1)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	if len(os.Args) < 2 {
		PrintUsage()
	}

	file, err := os.Open(os.Args[1])
	check(err)

	defer func() {
		err := file.Close()
		check(err)
	}()

}
