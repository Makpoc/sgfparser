package parser_test

import (
	"bufio"
	"fmt"
	"testing"

	"github.com/makpoc/sgfparser/parser"
	"github.com/makpoc/sgfparser/structures"

	"strings"

	"github.com/makpoc/sgfparser/logger"
)

func init() {
	logger.SetLogLevel(logger.INFO)
}

func TestCompareProperties(t *testing.T) {
	type testMap struct {
		p1       structures.Property
		p2       structures.Property
		areEqual bool
	}

	var test = []testMap{
		// Positive
		{
			structures.Property{
				Ident:  structures.PropIdent("AB"),
				Values: []structures.PropValue{structures.PropValue("ac"), structures.PropValue("bc")},
			},
			structures.Property{
				Ident:  structures.PropIdent("AB"),
				Values: []structures.PropValue{structures.PropValue("ac"), structures.PropValue("bc")},
			},
			true,
		},
		{
			structures.Property{
				Ident:  structures.PropIdent("AB"),
				Values: []structures.PropValue{structures.PropValue("bc"), structures.PropValue("ac")},
			},
			structures.Property{
				Ident:  structures.PropIdent("AB"),
				Values: []structures.PropValue{structures.PropValue("ac"), structures.PropValue("bc")},
			},
			true,
		},
		{
			structures.Property{
				Ident:  structures.PropIdent("AB"),
				Values: []structures.PropValue{structures.PropValue("bc"), structures.PropValue("ac")},
			},
			structures.Property{
				Ident:  structures.PropIdent("AB"),
				Values: []structures.PropValue{structures.PropValue("ac"), structures.PropValue("bc")},
			},
			true,
		},
		{
			structures.Property{
				Ident:  structures.PropIdent("AB"),
				Values: []structures.PropValue{structures.PropValue("bc"), structures.PropValue("ac"), structures.PropValue("ac")},
			},
			structures.Property{
				Ident:  structures.PropIdent("AB"),
				Values: []structures.PropValue{structures.PropValue("ac"), structures.PropValue("bc"), structures.PropValue("ac")},
			},
			true,
		},
		// Negative
		{
			structures.Property{
				Ident:  structures.PropIdent("AB"),
				Values: []structures.PropValue{structures.PropValue("bc"), structures.PropValue("ac")},
			},
			structures.Property{
				Ident:  structures.PropIdent("AB"),
				Values: []structures.PropValue{structures.PropValue("ac")},
			},
			false,
		},
		{
			structures.Property{
				Ident:  structures.PropIdent("AB"),
				Values: []structures.PropValue{structures.PropValue("cc"), structures.PropValue("ac")},
			},
			structures.Property{
				Ident:  structures.PropIdent("AB"),
				Values: []structures.PropValue{structures.PropValue("ac"), structures.PropValue("bc")},
			},
			false,
		},
		{
			structures.Property{
				Ident:  structures.PropIdent("AB"),
				Values: []structures.PropValue{structures.PropValue("bc"), structures.PropValue("ac"), structures.PropValue("ac")},
			},
			structures.Property{
				Ident:  structures.PropIdent("AB"),
				Values: []structures.PropValue{structures.PropValue("ac"), structures.PropValue("bc")},
			},
			false,
		},
		{
			structures.Property{
				Ident:  structures.PropIdent("AB"),
				Values: []structures.PropValue{structures.PropValue("bc"), structures.PropValue("ac"), structures.PropValue("ac")},
			},
			structures.Property{
				Ident:  structures.PropIdent("AB"),
				Values: []structures.PropValue{structures.PropValue("ac"), structures.PropValue("bc"), structures.PropValue("bc")},
			},
			false,
		},
	}

	for i, current := range test {
		equal := equalProperties(current.p1, current.p2)
		if equal != current.areEqual {
			t.Fatalf("FATAL!!!: compare function failed for test %d!!!", i)
		}

	}
}

type propIdentMap struct {
	raw    string
	parsed structures.PropIdent
}

var propIdentMatrixPos = []propIdentMap{
	{"FF[", structures.PropIdent("FF")},
	{"F[", structures.PropIdent("F")},
	{" FF	[", structures.PropIdent("FF")},
	{"FF[AA]", structures.PropIdent("FF")},
}

func TestParsePropIdent(t *testing.T) {

	for i, current := range propIdentMatrixPos {

		logger.LogDebug(fmt.Sprintf("POSITIVE: Testing with %v", current))

		reader := getReader(current.raw)
		result, err := parser.ParsePropIdent(reader)

		if err != nil {
			t.Errorf("Test %d returned error!", i, err.Error())
		}

		if *result != current.parsed {
			t.Errorf("Incorrect result! Expected [%s], found [%s]", current.parsed, *result)
		}
	}

}

func TestParsePropIdentNegative(t *testing.T) {

	var propIdentsNeg = []string{
		"FF",
		"FFF[",
		"aF[",
		"[",
		" [",
		"	[",
		"AA][",
	}

	for i, current := range propIdentsNeg {

		logger.LogDebug(fmt.Sprintf("NEGATIVE: Testing with %v", current))

		reader := getReader(current)
		result, err := parser.ParsePropIdent(reader)

		if err == nil {
			t.Errorf("Test %d was expected but did not return error!", i)
		}

		if result != nil {
			t.Errorf("Expected nil, but found %v!", *result)
		}
	}

}

func TestParsePropValue(t *testing.T) {

	//////////////
	// TEST DATA
	type propValueMap struct {
		raw    string
		parsed structures.PropValue
	}

	var propValueMatrix = []propValueMap{
		// UcLetter and Digit are ignored for now. (there are no examples in the specs)
		// None
		{"[]", structures.PropValue("")},
		// Number
		{"[0]", structures.PropValue("0")},
		{"[-1]", structures.PropValue("-1")},
		{"[-11]", structures.PropValue("-11")},
		{"[+1]", structures.PropValue("+1")},   // TODO - do we need to remove the sign?
		{"[+11]", structures.PropValue("+11")}, // TODO - do we need to remove the sign?
		{"[1]", structures.PropValue("1")},
		// Real
		{"[1.1]", structures.PropValue("1.1")},
		{"[-2.2]", structures.PropValue("-2.2")},
		// Double - already tested as number
		// Color - here - same as (simple)text
		// Text - same as simple text
		{"[something]", structures.PropValue("something")},
		{"[\"some\"thing\"]", structures.PropValue("\"some\"thing\"")},
		{"[some\\]thing]", structures.PropValue("some]thing")},
		{"[somet\\hing]", structures.PropValue("something")},
		{"[some\\\\thing]", structures.PropValue("some\\thing")},
		{"[some\tthing]", structures.PropValue("some thing")},
		{"[some\nthing]", structures.PropValue("some\nthing")},
		{"[some\rthing]", structures.PropValue("some\rthing")},
		{"[some\r\nthing]", structures.PropValue("some\r\nthing")},
		{"[some\n\rthing]", structures.PropValue("some\n\rthing")},
		// Text (from the specs)
		{
			"[Meijin NR: yeah, k4 is won\\\nderful\nsweat NR: thank you! :\\)\ndada NR: yup. I like this move too. It's a move only to be expected from a pro. I really like it :)\njansteen 4d: Can anyone\\\n explain [me\\] k4?]",
			structures.PropValue("Meijin NR: yeah, k4 is wonderful\nsweat NR: thank you! :)\ndada NR: yup. I like this move too. It's a move only to be expected from a pro. I really like it :)\njansteen 4d: Can anyone explain [me] k4?")},
		{"  [something]  ", structures.PropValue("something")},
	}
	//////////////
	// TEST START

	for i, current := range propValueMatrix {

		logger.LogDebug(fmt.Sprintf("POSITIVE: Testing with %v", current))

		reader := getReader(current.raw)
		result, err := parser.ParsePropValue(reader)

		if err != nil {
			t.Errorf("Test %d returned error! %s", i, err.Error())
		}

		if result != nil {
			if *result != current.parsed {
				t.Errorf("Incorrect result! Given [%s], expected [%s], found [%s]", current.raw, current.parsed, *result)
			}
		} else {
			t.Errorf("Incorrect result! result is nil!")
		}
	}
}

func TestParsePropValueNeg(t *testing.T) {

	//////////////
	// TEST DATA
	var propMatrix = []string{
		"[",
		"a",
		"[\\]",
	}

	//////////////
	// TEST START

	for i, current := range propMatrix {
		result, err := parser.ParsePropValue(getReader(current))

		if err == nil {
			t.Errorf("Test %d Failed for input %s. Expected an error, got nil", i, current)
		}
		if result != nil {
			t.Errorf("Test %d Failed for input %s. Expected nil as result, but got [%#v]", i, current, *result)
		}
	}
}

func TestParseProperty(t *testing.T) {

	//////////////
	// TEST DATA
	type propMap struct {
		raw    string
		parsed structures.Property
	}

	var propMatrix = []propMap{
		{
			"FF[3]",
			structures.Property{
				Ident: structures.PropIdent("FF"),
				Values: []structures.PropValue{
					structures.PropValue("3"),
				},
			},
		},
		{
			" FF [ 3 ] ",
			structures.Property{
				Ident: structures.PropIdent("FF"),
				Values: []structures.PropValue{
					structures.PropValue(" 3 "),
				},
			},
		},
		{
			"FF[]",
			structures.Property{
				Ident: structures.PropIdent("FF"),
				Values: []structures.PropValue{
					structures.PropValue(""),
				},
			},
		},
		{
			"AB[ac][bc]",
			structures.Property{
				Ident: structures.PropIdent("AB"),
				Values: []structures.PropValue{
					structures.PropValue("ac"),
					structures.PropValue("bc"),
				},
			},
		},
		{
			" AB	[ac] [bc]	",
			structures.Property{
				Ident: structures.PropIdent("AB"),
				Values: []structures.PropValue{
					structures.PropValue("ac"),
					structures.PropValue("bc"),
				},
			},
		},
		{
			" AB	[\\]ac] [bc]	",
			structures.Property{
				Ident: structures.PropIdent("AB"),
				Values: []structures.PropValue{
					structures.PropValue("]ac"),
					structures.PropValue("bc"),
				},
			},
		},
	}

	//////////////
	// TEST START

	for i, current := range propMatrix {
		reader := getReader(current.raw)
		result, err := parser.ParseProperty(reader)

		if err != nil {
			t.Errorf("Test %d returned error! %s", i, err.Error())
		}

		if result != nil {
			ok := equalProperties(*result, current.parsed)
			if !ok {
				t.Errorf("Test %d Failed! Given %s, expected %#v, found %#v", i, current.raw, current.parsed, *result)
			}
		} else {
			t.Errorf("Result is nil!")
		}
	}

}

func TestParsePropertyNeg(t *testing.T) {

	//////////////
	// TEST DATA

	var propMatrix = []string{
		"FF)",
		"FF[;",
	}

	//////////////
	// TEST START

	for i, current := range propMatrix {
		result, err := parser.ParseProperty(getReader(current))

		if err == nil {
			t.Errorf("Test %d Failed. Expected an error, got nil", i)
		}
		if result != nil {
			t.Errorf("Test %d Failed. Expected nil as result, but got [%#v]", i, *result)
		}
	}
}

func TestParseNode(t *testing.T) {
	//////////////
	// TEST DATA
	type nodeStruct struct {
		raw    string
		parsed structures.Node
	}

	var nodesMatrix = []nodeStruct{
		nodeStruct{
			";FF[AA]",
			structures.Node{
				Properties: []structures.Property{
					structures.Property{
						Ident: structures.PropIdent("FF"),
						Values: []structures.PropValue{
							structures.PropValue("AA"),
						},
					},
				},
			},
		},
		nodeStruct{
			";FF[AA][qwe]",
			structures.Node{
				Properties: []structures.Property{
					structures.Property{
						Ident: structures.PropIdent("FF"),
						Values: []structures.PropValue{
							structures.PropValue("AA"),
							structures.PropValue("qwe"),
						},
					},
				},
			},
		},
		nodeStruct{
			";C[]",
			structures.Node{
				Properties: []structures.Property{
					structures.Property{
						Ident: structures.PropIdent("C"),
						Values: []structures.PropValue{
							structures.PropValue(""),
						},
					},
				},
			},
		},
		nodeStruct{
			"; FF [AA] [qwe] ",
			structures.Node{
				Properties: []structures.Property{
					structures.Property{
						Ident: structures.PropIdent("FF"),
						Values: []structures.PropValue{
							structures.PropValue("AA"),
							structures.PropValue("qwe"),
						},
					},
				},
			},
		},
		nodeStruct{
			"(;)",
			structures.Node{
				Properties: []structures.Property{},
			},
		},
		nodeStruct{
			"(;;)",
			structures.Node{
				Properties: []structures.Property{},
			},
		},
	}

	//////////////
	// TEST START

	for i, current := range nodesMatrix {
		reader := getReader(current.raw)
		result, err := parser.ParseNode(reader)

		if err != nil {
			t.Errorf("Test %d returned error! %s", i, err.Error())
		}

		if result == nil {
			t.Errorf("Result is nil!")
			return
		}
		expectedLen, actualLen := len(current.parsed.Properties), len(result.Properties)
		if expectedLen != actualLen {
			t.Errorf("Expected number of properties: %d, actual: %d", expectedLen, actualLen)
			return
		}
		for i, prop := range current.parsed.Properties {
			ok := equalProperties(prop, result.Properties[i])
			if !ok {
				t.Errorf("Test %d failed! Given %s, \nexpected \n%#v, \nfound \n%#v", i, current.raw, current.parsed, *result)
			}
		}
	}
}

func TestParseNodeNeg(t *testing.T) {

	// TEST DATA
	type nodeStruct struct {
		raw    string
		parsed structures.Node
	}

	var nodesMatrix = []nodeStruct{
		nodeStruct{
			"FF[AA]",
			structures.Node{
				Properties: []structures.Property{
					structures.Property{
						Ident: structures.PropIdent("FF"),
						Values: []structures.PropValue{
							structures.PropValue("AA"),
						},
					},
				},
			},
		},
	}

	for i, current := range nodesMatrix {
		reader := getReader(current.raw)
		result, err := parser.ParseNode(reader)

		if err == nil {
			t.Errorf("%d: Test expected to return error but did not!", i)
			if result != nil {
				t.Errorf(fmt.Sprintf("Instead the returned value was: \n%#v", *result))
			}
		}
	}
}

func TestParseSequence(t *testing.T) {

	/////////////
	// TEST DATA

	type sequenceStruct struct {
		raw    string
		parsed structures.Sequence
	}

	var sequenceMatrix = []sequenceStruct{
		sequenceStruct{
			// single node
			";FF[AA])",
			structures.Sequence{
				Nodes: []structures.Node{
					structures.Node{
						Properties: []structures.Property{
							structures.Property{
								Ident: structures.PropIdent("FF"),
								Values: []structures.PropValue{
									structures.PropValue("AA"),
								},
							},
						},
					},
				},
			},
		},
		sequenceStruct{
			// multiple nodes
			";FF[AA];C[BB][asd])",
			structures.Sequence{
				Nodes: []structures.Node{
					structures.Node{
						Properties: []structures.Property{
							structures.Property{
								Ident: structures.PropIdent("FF"),
								Values: []structures.PropValue{
									structures.PropValue("AA"),
								},
							},
						},
					},
					structures.Node{
						Properties: []structures.Property{
							structures.Property{
								Ident: structures.PropIdent("C"),
								Values: []structures.PropValue{
									structures.PropValue("BB"),
									structures.PropValue("asd"),
								},
							},
						},
					},
				},
			},
		},
		sequenceStruct{
			// handle spaces
			" ; FF[AA] ; C[BB][asd] )",
			structures.Sequence{
				Nodes: []structures.Node{
					structures.Node{
						Properties: []structures.Property{
							structures.Property{
								Ident: structures.PropIdent("FF"),
								Values: []structures.PropValue{
									structures.PropValue("AA"),
								},
							},
						},
					},
					structures.Node{
						Properties: []structures.Property{
							structures.Property{
								Ident: structures.PropIdent("C"),
								Values: []structures.PropValue{
									structures.PropValue("BB"),
									structures.PropValue("asd"),
								},
							},
						},
					},
				},
			},
		},
		sequenceStruct{
			// will parse only one sequence - the first one
			";FF[AA](;C[BB])",
			structures.Sequence{
				Nodes: []structures.Node{
					structures.Node{
						Properties: []structures.Property{
							structures.Property{
								Ident: structures.PropIdent("FF"),
								Values: []structures.PropValue{
									structures.PropValue("AA"),
								},
							},
						},
					},
				},
			},
		},
		sequenceStruct{
			// parse the empty sequence
			";)",
			structures.Sequence{
				Nodes: []structures.Node{
					structures.Node{
						Properties: []structures.Property{},
					},
				},
			},
		},
	}
	/////////////
	// TEST START

	for i, current := range sequenceMatrix {
		reader := getReader(current.raw)
		result, err := parser.ParseSequence(reader)

		if err != nil {
			t.Errorf("Test %d returned error! %s", i, err.Error())
		}

		if result == nil {
			t.Errorf("Test %d failed! Given %s, \nExpected %#v, but\nResult is nil!", i, current.raw, current.parsed)
			return
		}
		if !equalSequence(current.parsed, *result) {
			t.Errorf("Test %d failed! Given %s, \nExpected \n%#v, \nfound \n%#v", i, current.raw, current.parsed, *result)
		}
	}
}

func TestParseGameTree(t *testing.T) {
	/////////////
	// TEST DATA
	type treeStruct struct {
		raw    string
		parsed structures.GameTree
	}

	var treeMatrix = []treeStruct{
		treeStruct{
			"(;FF[AA];C[bbb])",
			structures.GameTree{
				Sequence: structures.Sequence{
					Nodes: []structures.Node{
						structures.Node{
							Properties: []structures.Property{
								structures.Property{
									Ident: structures.PropIdent("FF"),
									Values: []structures.PropValue{
										structures.PropValue("AA"),
									},
								},
							},
						},
						structures.Node{
							Properties: []structures.Property{
								structures.Property{
									Ident: structures.PropIdent("C"),
									Values: []structures.PropValue{
										structures.PropValue("bbb"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	//////////////
	// TEST START
	for i, current := range treeMatrix {
		reader := getReader(current.raw)
		result, err := parser.ParseGameTree(reader)

		if err != nil {
			t.Errorf("Test %d returned error! %s", i, err.Error())
		}

		if result == nil {
			t.Errorf("Result is nil!")
			return
		}

		if !equalSequence(current.parsed.Sequence, result.Sequence) {
			t.Errorf("Test %d failed! Given %s, \nexpected \n%#v, \nfound \n%#v", i, current.raw, current.parsed, *result)
		}
	}
}

func equalProperties(p1, p2 structures.Property) bool {
	if &p1 == &p2 {
		// same object
		return true
	}

	if p1.Ident != p2.Ident || len(p1.Values) != len(p2.Values) {
		return false
	}

	p2Matched := make([]bool, len(p2.Values))

	for _, val1 := range p1.Values {
		contains := false
		// Values can be unordered
		for i, val2 := range p2.Values {

			if !p2Matched[i] && val1 == val2 {
				p2Matched[i] = true // will not check this index twice
				contains = true
				break
			}
		}

		// inner loop did not find a match
		if !contains {
			return false
		}
	}

	return true
}

func equalNode(n1, n2 structures.Node) bool {
	if &n1 == &n2 {
		// same object
		return true
	}

	expectedLen, actualLen := len(n1.Properties), len(n2.Properties)
	if expectedLen != actualLen {
		return false
	}
	for i, prop := range n1.Properties {
		if !equalProperties(prop, n2.Properties[i]) {
			return false
		}
	}
	return true
}

func equalSequence(s1, s2 structures.Sequence) bool {
	if &s1 == &s2 {
		// same object
		return true
	}

	// Make sure that we have the same number of nodes in the sequences
	expectedNodesLen, actualNodesLen := len(s1.Nodes), len(s2.Nodes)
	if expectedNodesLen != actualNodesLen {
		return false
	}

	for i, node := range s1.Nodes {
		// Make sure that we have the same number of properties in each node
		expectedPropsLen, actualPropsLen := len(node.Properties), len(s2.Nodes[i].Properties)
		if expectedPropsLen != actualPropsLen {
			return false
		}

		// compare each property in each node
		for j, prop := range node.Properties {
			if !equalProperties(prop, s2.Nodes[i].Properties[j]) {
				return false
			}
		}
	}
	return true
}

func getReader(raw string) *bufio.Reader {
	return bufio.NewReader(strings.NewReader(raw))
}
