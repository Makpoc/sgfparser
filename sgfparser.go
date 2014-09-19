package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	// VariationStart starts new variation sequence
	VariationStart rune = '('
	// VariationEnd ends a variation sequence
	VariationEnd rune = ')'
	// NodeSeparator delimits properties
	NodeSeparator rune = ';'
	// PropertyValueStart starts new value sequence
	PropertyValueStart rune = '['
	// PropertyValueEnd starts new value sequence
	PropertyValueEnd rune = ']'
)

// Property is the container for Property and value in SGF files. This means that property is B[xx][x] and not just B
type Property struct {
	PropIdent string
	PropValue []string
}

func (prop Property) String() string {
	output := prop.PropIdent
	for _, value := range prop.PropValue {
		output += fmt.Sprintf("[%s]", value)
	}
	return output
}

// Node is the container for properties with their keys and values
type Node struct {
	Id         int
	Properties []Property
}

func (node Node) String() string {
	output := string(NodeSeparator)
	for _, prop := range node.Properties {
		output += fmt.Sprint(prop)
	}

	return output
}

// Variation is the structure, holding the game tree. Variations contains references to the Nodes of the current Variation as well as the parent and the child variations. If this is the root variation Parent will point to itself.
type Variation struct {
	Id       int
	Parent   *Variation
	Children []Variation
	Nodes    []Node
}

func getInvalidFormatMsg(details string) string {
	return fmt.Sprintf("Invalid SGF format! %s\n", details)
}

// ParseGameTree is the entry point for parsing
func ParseGameTree(sgf string) (Variation, error) {
	if !strings.HasPrefix(sgf, "(") && !strings.HasSuffix(sgf, ")") {
		errMsg := getInvalidFormatMsg("SGF must start with root variation")
		log.Println(errMsg)
		return Variation{}, errors.New(errMsg)
	}

	gameTree := Variation{Id: 0, Parent: nil}

	sgf = strings.Replace(sgf, "\n", "", -1)

	gameTree, err := parseVariation([]rune(sgf))
	checkFatal(err)

	return gameTree, nil
}

func parseVariation(sgf []rune) (Variation, error) {
	log.Printf("Parsing %s\n", string(sgf))

	currentVariation := Variation{}

	// parse nodes in the current variation
	log.Println("Start parsing nodes...")

	var nodesEndIndex int

	// if the current variation contains subvariations - parse up to the first one found
	if hasSubVariations(sgf) {
		nodesEndIndex = getNextVariationStartIndex(sgf)
	} else {
		// otherwise parse the entire range
		nodesEndIndex = len(sgf)
	}

	// parse
	nodes, err := parseNode(sgf[:nodesEndIndex])
	if err != nil {
		log.Printf("Failed to parse nodes for current variation. Error is: %s\n", err.Error())
	} else {
		currentVariation.Nodes = nodes
	}

	log.Println("Finished parsing nodes...")

	log.Println("Start parsing subvariations...")

	// now parse the remaining variations if any
	remainingSgf := sgf[nodesEndIndex:]
	subVariationId := 0
	for hasSubVariations(remainingSgf) {
		log.Println("Subvariations exist")

		// get the bounderies for the sub variation
		nextVariationStartIndex, nextVariationEndIndex := getNextVariationBounderies(remainingSgf)
		prettyPrintRange(nextVariationStartIndex, nextVariationEndIndex, remainingSgf)

		nextVariation := remainingSgf[nextVariationStartIndex+1 : nextVariationEndIndex]

		childVariation, err := parseVariation(nextVariation)
		if err != nil {
			return currentVariation, errors.New(fmt.Sprintf("Error parsing %s!\n", string(nextVariationEndIndex)))
		}

		childVariation.Parent, childVariation.Id = &currentVariation, subVariationId
		subVariationId++

		currentVariation.Children = append(currentVariation.Children, childVariation)
		remainingSgf = remainingSgf[nextVariationEndIndex+1:]
	}

	log.Println("Finished parsing subvariations...")
	return currentVariation, nil
}

func parseNode(sgf []rune) ([]Node, error) {
	sgfStr := string(sgf)
	log.Printf("Nodes in this variation: %s\n", sgfStr)
	if len(sgf) == 0 {
		return nil, errors.New("No Nodes in this variation")
	}

	var nodes []Node

	remainingNode := sgf
	var nodeId int
	for {
		currNode, err := getNextNode(remainingNode)
		if err != nil {
			log.Printf("Failed to get node from %s. Error is: %s", string(remainingNode), err.Error())
			break
		}

		currNodeStr := string(currNode)
		log.Printf("Found node %s", currNodeStr)

		remainingNode = remainingNode[len(currNode):]

		node := Node{Id: nodeId}
		nodes = append(nodes, node)
		nodeId++

		props, propsErr := parseProperty(currNode)
		if propsErr != nil {
			log.Printf("ERROR: Failed to parse properties from node %s. Error is %s\n", currNodeStr, propsErr.Error())
			continue
		}
		node.Properties = props
		log.Println(props)
	}

	return nodes, nil
}

// TODO: change this to support NodeSeparator inside comments
func getNextNode(sgf []rune) ([]rune, error) {

	log.Printf("Searching for next node in %s", string(sgf))

	if len(sgf) == 0 {
		return nil, errors.New("No more nodes!")
	}

	startIndex := -1

	for currIndex, currRune := range sgf {
		if currRune == NodeSeparator {
			if startIndex == -1 {
				startIndex = currIndex
			} else {
				// startIndex was already found - slice and return
				return sgf[startIndex:currIndex], nil
			}
		}
	}

	if startIndex != -1 {
		// Last node will not end with separator.
		return sgf[startIndex:], nil
	}

	// There was no NodeSeparator, so the slice wasn't a valid Node
	return nil, errors.New("Failed to find node separator")

}

func parseProperty(sgf []rune) ([]Property, error) {

	log.Printf("Parsing properties from %s", string(sgf))

	if (len(sgf) == 0) || (len(sgf) == 1 && sgf[0] == NodeSeparator) {
		return nil, errors.New("No properties found!")
	}

	if sgf[0] == NodeSeparator {
		sgf = sgf[1:]
	}

	props := []Property{}
	currProperty := Property{}

	lastValueStartIndex := -1
	lastIdentStartIndex := 0

	for index, currRune := range sgf {
		if currRune == PropertyValueStart {
			if currProperty.PropIdent == "" {
				// we've just passed the PropIdent - set it and continue parsing the values
				currProperty.PropIdent = string(sgf[lastIdentStartIndex:index])
			}
			lastValueStartIndex = index + 1
		} else if currRune == PropertyValueEnd {
			// TODO - handle ] in comments
			if lastValueStartIndex == -1 {
				// no [ before this closing ]. seems like broken file..
				return nil, errors.New(fmt.Sprintf("Invalid file format - found ']' without '[' at %d in %s", index, string(sgf)))
			} else {
				// end of regular Value - add it
				currProperty.PropValue = append(currProperty.PropValue, string(sgf[lastValueStartIndex:index]))
				lastValueStartIndex = -1
			}

			if (index == len(sgf)-1) || (index != len(sgf)-1 && sgf[index+1] != PropertyValueStart) {
				// we are starting new PropIdent section
				props = append(props, currProperty)
				currProperty = Property{}
				lastIdentStartIndex = index + 1
			}
		}

	}

	return props, nil

}

func hasSubVariations(sgf []rune) bool {
	return getNextVariationStartIndex(sgf) != -1
}

// TODO: Take into account quotes (e.g. inside comments)
func getNextVariationStartIndex(sgf []rune) int {
	return strings.Index(string(sgf), string(VariationStart))
}

// TODO: Take into account quotes (e.g. inside comments)
func getNextVariationEndIndex(sgf []rune) int {
	nesting := 0

	for index, currRune := range sgf {
		if currRune == VariationEnd {
			log.Printf("Found %s at %d...\n", string(currRune), index)
			nesting--
			if nesting == 0 {
				log.Printf("... and it's the one we are searching for!\n")
				prettyPrintCharArrow(index, sgf)
				return index
			}
		}
		if currRune == VariationStart {
			nesting++
			log.Printf("Found %s at %d. Increasing nesting level to %d\n", string(currRune), index, nesting)
			prettyPrintCharArrow(index, sgf)
		}
	}

	return -1
}

func getNextVariationBounderies(sgf []rune) (int, int) {
	return getNextVariationStartIndex(sgf), getNextVariationEndIndex(sgf)
}

func prettyPrintCharArrow(index int, sgf []rune) {
	log.Printf("%s\n", string(sgf))
	log.Printf("%s^\n", strings.Repeat(" ", index))
	log.Printf("%s%d\n", strings.Repeat(" ", index), index)

}

func prettyPrintRange(start int, end int, sgf []rune) {
	sgfStr := string(sgf)
	log.Printf("%s\n", sgfStr)
	fromToArrows := fmt.Sprintf("%s>%s<", strings.Repeat(" ", start), strings.Repeat(" ", end-start-1))
	log.Printf("%s\n", fromToArrows)
}

func DumpTree(gameTree Variation, indent int) {
	fmt.Printf("-%s| %d", strings.Repeat("-", indent), gameTree.Id)

	var row string
	for _, node := range gameTree.Nodes {
		row += fmt.Sprintf("%s", node)
	}

	fmt.Println(row)
	if len(gameTree.Children) > 0 {
		for _, variation := range gameTree.Children {
			DumpTree(variation, indent+1)
		}
	}
}

func printUsage() {
	fmt.Printf("Usage: %s file.sgf", os.Args[0])
	os.Exit(1)
}

func checkFatal(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	//	log.SetOutput(ioutil.Discard)

	if len(os.Args) < 2 {
		printUsage()
	}

	file, err := os.Open(os.Args[1])
	checkFatal(err)

	defer func() {
		err := file.Close()
		checkFatal(err)
	}()

	//	sgfBufReader := bufio.NewReader(file)
	sgf, err := ioutil.ReadAll(file)
	checkFatal(err)

	log.Printf("SGF size: %d\n", len(sgf)-2)
	gameTree, err := ParseGameTree(string(sgf))
	checkFatal(err)

	DumpTree(gameTree, 0)

	fmt.Println()
	fmt.Println(gameTree.Children[0])
	fmt.Println(gameTree.Children[0].Nodes[0])
	prop, err := parseProperty([]rune(";C[z][q]B[c][y]Z[a][x]"))
	fmt.Println(prop)
}
