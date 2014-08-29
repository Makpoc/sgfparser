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

// Node is the container for properties with their keys and values
type Node struct {
	Properties []Property
	Variations []Variation
	AsString   string
}

// Variation is the structure, holding the game tree. Variations contains references to the Nodes of the current Variation as well as the parent and the child variations. If this is the root variation Parent will point to itself.
type Variation struct {
	Id       int
	Parent   *Variation
	Children []*Variation
	Nodes    []*Node
}

func getInvalidFormatMsg(details string) string {
	return fmt.Sprintf("Invalid SGF format! %s\n", details)
}

// ParseGameTree is the entry point for parsing
func ParseGameTree(sgf string) (*Variation, error) {
	if !strings.HasPrefix(sgf, "(") && !strings.HasSuffix(sgf, ")") {
		errMsg := getInvalidFormatMsg("SGF must start with root variation")
		log.Println(errMsg)
		return &Variation{}, errors.New(errMsg)
	}

	gameTree := new(Variation)
	gameTree.Id = 0
	gameTree.Parent = nil

	sgf = strings.Replace(sgf, "\n", "", -1)

	gameTree, err := parseVariation(gameTree, []rune(sgf[1:len(sgf)-1]))
	checkFatal(err)

	return gameTree, nil
}

var depth int = 0

func parseVariation(parentVariation *Variation, sgf []rune) (*Variation, error) {
	log.Printf("Entering Depth: %d", depth)
	depth++
	log.Printf("Parsing variation for string %s\n", string(sgf))

	if !isValidNode(sgf) {
		return &Variation{}, errors.New("Invalid node definition!")
	}

	currentVar := new(Variation)

	isInsideQuotes := false

	for index := 0; index < len(sgf); index++ {

		currentRune := sgf[index]
		// skip spaces outside comments
		if currentRune == ' ' && !isInsideQuotes {
			log.Printf("Found white space at %d", index)
			continue
		}

		// TODO escapes
		if currentRune == '"' {
			// invert
			isInsideQuotes = !isInsideQuotes
			log.Printf("Found quote character at %d. Inside quotes is %t", index, isInsideQuotes)
			continue
		}

		if currentRune == VariationStart && !isInsideQuotes {
			log.Printf("Found new variation start at %d", index)
			prettyPrintCharArrow(index, sgf)

			index++ // increment to skip the ValidationStart char
			variationEnd, err := seekVariationEndIndex(sgf[index:])
			checkFatal(err)

			variation, err := parseVariation(currentVar, sgf[index:variationEnd])
			checkFatal(err)

			currentVar.Children = append(currentVar.Children, variation)
			for i, child := range currentVar.Children {
				child.Parent = currentVar
				child.Id = currentVar.Id + i + 1
			}

			// skip ahead to the closing bracket
			index = variationEnd

			continue

		}

		if currentRune == VariationEnd && !isInsideQuotes {
			log.Printf("Found closing character at %d. Returning", index)
			break
		}

		currentNodes, consumed, _ := parseNodes(sgf[index:])
		for _, node := range currentNodes {
			currentVar.Nodes = append(currentVar.Nodes, node)
		}
		index += consumed
	}

	depth--
	log.Printf("Exiting Depth: %d", depth)
	return currentVar, nil
}

func parseNodes(sgfRunes []rune) ([]*Node, int, error) {
	log.Println("Parsing nodes...")
	prettyPrintCharArrow(0, sgfRunes)

	if !isValidNode(sgfRunes) {
		return nil, 0, errors.New("Invalid Node format!")
	}

	sgfAsString := string(sgfRunes)
	var toIndex int

	varEnd := float64(strings.IndexRune(sgfAsString, VariationEnd))
	varStart := float64(strings.IndexRune(sgfAsString, VariationStart))

	if varEnd == -1 && varStart == -1 {
		toIndex = len(sgfRunes) - 1
	} else if varEnd == -1 {
		toIndex = int(varStart)
	} else {
		toIndex = int(varEnd)
	}

	var nodes []*Node
	var runesConsumed int

	for _, nodeAsString := range strings.Split(string(sgfRunes[:toIndex]), string(NodeSeparator)) {
		n := &Node{AsString: nodeAsString}
		log.Printf("Found node %s", n.AsString)
		runesConsumed += len(nodeAsString)
		nodes = append(nodes, n)
	}

	log.Printf("Consumed %d runes after parsing %v", runesConsumed, nodes)

	return nodes, runesConsumed, nil
}

func isValidNode(sgfRunes []rune) bool {

	if len(sgfRunes) < 1 || sgfRunes[0] != NodeSeparator {
		return false
	}
	return true
}

func seekVariationEndIndex(sgfRunes []rune) (varEndIndex int, err error) {

	log.Printf("Seeking for the closing character")
	nestLevel := 0

	varEndIndex = -1

	for index := 0; index < len(sgfRunes); index++ {
		if sgfRunes[index] == VariationStart {
			nestLevel++
			log.Printf(">> Found nested variation. Increasing nestLevel to %d\n", nestLevel)
			prettyPrintCharArrow(index, sgfRunes)
			continue
		}
		if sgfRunes[index] == VariationEnd {
			if nestLevel == 0 {
				log.Printf(">> Found closing char at %d\n", index)
				varEndIndex = index
				break
			} else {
				nestLevel--
				log.Printf(">> Found closing char but it's for nested variation. Decreasing nestLevel to %d\n", nestLevel)
				prettyPrintCharArrow(index, sgfRunes)
				continue
			}
		}
	}

	if varEndIndex >= 0 && varEndIndex < len(sgfRunes) {
		prettyPrintCharArrow(varEndIndex, sgfRunes)
	} else {
		err = errors.New("There was a problem while searching for variation end")
	}

	return
}

func prettyPrintCharArrow(index int, sgfRunes []rune) {
	log.Printf("%s\n", string(sgfRunes))
	log.Printf("%s^\n", strings.Repeat(" ", index))
	log.Printf("%s%d\n", strings.Repeat(" ", index), index)

}

func DumpTree(gameTree *Variation, indent int) {
	fmt.Printf("%s%+d", strings.Repeat("-", indent), gameTree.Id)

	var nodesAsString string
	for _, node := range gameTree.Nodes {
		nodesAsString += fmt.Sprintf("%s", node.AsString)
	}

	fmt.Println(nodesAsString)
	if len(gameTree.Children) > 0 {
		for _, variation := range gameTree.Children {
			DumpTree(variation, indent+1)
		}
	}
}

//func (v Variation) String() string {
//	return fmt.Sprintf("%+p\n", v)
//}

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

	//	fmt.Printf("%+v\n", gameTree)
	//	fmt.Printf("%+v\n", gameTree.Children[0])
	//	fmt.Printf("%+v\n", gameTree.Children[0].Parent)

	DumpTree(gameTree, 0)

}
