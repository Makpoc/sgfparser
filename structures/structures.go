package structures

import "fmt"

const (
	// SequenceStart starts new variation sequence
	GameTreeStart rune = '('
	// SequenceEnd ends a variation sequence
	GameTreeEnd rune = ')'
	// NodeSeparator delimits properties
	NodeSeparator rune = ';'
	// PropertyValueStart starts new value sequence
	PropertyValueStart rune = '['
	// PropertyValueEnd starts new value sequence
	PropertyValueEnd rune = ']'
)

type Collection struct {
	GameTrees []GameTree
}

func (collection Collection) String() string {
	var output string

	for _, tree := range collection.GameTrees {
		output += fmt.Sprintf("%s", tree)
	}

	return output
}

type GameTree struct {
	Parent   *GameTree
	Children []GameTree
	Sequence Sequence
}

func (tree GameTree) String() string {
	output := fmt.Sprintf("%s", tree.Sequence)

	for _, subTree := range tree.Children {
		output += fmt.Sprintf("%s", subTree)
	}
	return string(GameTreeStart) + output + string(GameTreeEnd)
}

// Sequence is the structure, holding all nodes in the current variation.
type Sequence struct {
	Nodes []Node
}

func (sequence Sequence) String() string {
	var output string
	for _, node := range sequence.Nodes {
		output += fmt.Sprintf("%s", node)
	}
	return output
}

// Node is the container for properties with their keys and values
type Node struct {
	Id         int
	Properties []Property
}

func (node Node) String() string {
	var output string
	for _, prop := range node.Properties {
		output += fmt.Sprint(prop)
	}

	return string(NodeSeparator) + output
}

// Property is the container for Property and value in SGF files. This means that property is B[xx][x] and not just B
type Property struct {
	Ident  PropIdent
	Values []PropValue
}

func (prop Property) String() string {
	output := string(prop.Ident)
	for _, value := range prop.Values {
		output += fmt.Sprintf("[%s]", value)
	}
	return output
}

type PropIdent string
type PropValue string
