package parser_test

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"runtime"
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
		p1      structures.Property
		p2      structures.Property
		isError bool
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
			false,
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
			false,
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
			false,
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
			false,
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
			true,
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
			true,
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
			true,
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
			true,
		},
	}

	for i, current := range test {
		compareError := compareProperties(current.p1, current.p2)
		// if it's supposed to be error, but it's not - fail
		// if it's not supposed to be error, but it is - fail
		if (current.isError && compareError == nil) || (!current.isError && compareError != nil) {
			t.Errorf("Test %d failed! Expected error: %v, but it was %v", i, current.isError, compareError != nil)
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
			"FF[3])",
			structures.Property{
				Ident: structures.PropIdent("FF"),
				Values: []structures.PropValue{
					structures.PropValue("3"),
				},
			},
		},
		{
			" FF [ 3 ] )",
			structures.Property{
				Ident: structures.PropIdent("FF"),
				Values: []structures.PropValue{
					structures.PropValue(" 3 "),
				},
			},
		},
		{
			"FF[])",
			structures.Property{
				Ident: structures.PropIdent("FF"),
				Values: []structures.PropValue{
					structures.PropValue(""),
				},
			},
		},
		{
			"AB[ac][bc])",
			structures.Property{
				Ident: structures.PropIdent("AB"),
				Values: []structures.PropValue{
					structures.PropValue("ac"),
					structures.PropValue("bc"),
				},
			},
		},
		{
			" AB	[ac] [bc]	)",
			structures.Property{
				Ident: structures.PropIdent("AB"),
				Values: []structures.PropValue{
					structures.PropValue("ac"),
					structures.PropValue("bc"),
				},
			},
		},
		{
			" AB	[\\]ac] [bc]	)",
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

		if result == nil {
			t.Errorf("Result is nil!")
			return
		}

		if err := compareProperties(*result, current.parsed); err != nil {
			t.Errorf("Test %d Failed! Error was %s!\nGiven %s, \nExpected %#v,\nFound %#v", i, err.Error(), current.raw, current.parsed, *result)
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
			"(;FF[AA])",
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
			"(;FF[AA][qwe])",
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
			"(;C[])",
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
			"(; FF [AA] [qwe] )",
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
			if err := compareProperties(prop, result.Properties[i]); err != nil {
				t.Errorf("Test %d failed. Error was %s!\nGiven %s, \nExpected \n%#v, \nFound \n%#v\n", i, err.Error(), current.raw, current.parsed, *result)
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
		if err := compareSequence(current.parsed, *result); err != nil {
			t.Errorf("Test %d failed! Error is: %s!\nGiven %s, \nExpected \n%#v, \nFound \n%#v", i, err.Error(), current.raw, current.parsed, *result)
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

	logger.SetLogLevel(logger.DEBUG)

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
		treeStruct{
			"(;)",
			structures.GameTree{
				Sequence: structures.Sequence{
					Nodes: []structures.Node{
						structures.Node{
							Properties: []structures.Property{},
						},
					},
				},
			},
		},
		treeStruct{
			"(;;;(;;;;)(;;)(;;;(;;)(;)))",
			//->(;;;(;;;;)(;;)(;;;(;;)(;)))<-
			structures.GameTree{
				Sequence: structures.Sequence{
					Nodes: []structures.Node{
						structures.Node{
							Properties: []structures.Property{},
						},
						structures.Node{
							Properties: []structures.Property{},
						},
						structures.Node{
							Properties: []structures.Property{},
						},
					},
				},
				Children: []*structures.GameTree{
					//(;;;->(;;;;)<-(;;)(;;;(;;)(;)))
					&structures.GameTree{
						Sequence: structures.Sequence{
							Nodes: []structures.Node{
								structures.Node{
									Properties: []structures.Property{},
								},
								structures.Node{
									Properties: []structures.Property{},
								},
								structures.Node{
									Properties: []structures.Property{},
								},
								structures.Node{
									Properties: []structures.Property{},
								},
							},
						},
					},
					//(;;;(;;;;)->(;;)<-(;;;(;;)(;)))
					&structures.GameTree{
						Sequence: structures.Sequence{
							Nodes: []structures.Node{
								structures.Node{
									Properties: []structures.Property{},
								},
								structures.Node{
									Properties: []structures.Property{},
								},
							},
						},
					},
					//(;;;(;;;;)(;;)->(;;;(;;)(;))<-)
					&structures.GameTree{
						Sequence: structures.Sequence{
							Nodes: []structures.Node{
								structures.Node{
									Properties: []structures.Property{},
								},
								structures.Node{
									Properties: []structures.Property{},
								},
								structures.Node{
									Properties: []structures.Property{},
								},
							},
						},
						Children: []*structures.GameTree{
							//(;;;(;;;;)(;;)(;;;->(;;)<-(;)))
							&structures.GameTree{
								Sequence: structures.Sequence{
									Nodes: []structures.Node{
										structures.Node{
											Properties: []structures.Property{},
										},
										structures.Node{
											Properties: []structures.Property{},
										},
									},
								},
							},
							//(;;;(;;;;)(;;)(;;;(;;)->(;)<-))
							&structures.GameTree{
								Sequence: structures.Sequence{
									Nodes: []structures.Node{
										structures.Node{
											Properties: []structures.Property{},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		treeStruct{
			"(;(;))",
			structures.GameTree{
				Sequence: structures.Sequence{
					Nodes: []structures.Node{
						structures.Node{
							Properties: []structures.Property{},
						},
					},
				},
				Children: []*structures.GameTree{
					&structures.GameTree{
						Sequence: structures.Sequence{
							Nodes: []structures.Node{
								structures.Node{
									Properties: []structures.Property{},
								},
							},
						},
					},
				},
			},
		},
		treeStruct{
			"(;(;(;)))",
			structures.GameTree{
				Sequence: structures.Sequence{
					Nodes: []structures.Node{
						structures.Node{
							Properties: []structures.Property{},
						},
					},
				},
				Children: []*structures.GameTree{
					&structures.GameTree{
						Sequence: structures.Sequence{
							Nodes: []structures.Node{
								structures.Node{
									Properties: []structures.Property{},
								},
							},
						},
						Children: []*structures.GameTree{
							&structures.GameTree{
								Sequence: structures.Sequence{
									Nodes: []structures.Node{
										structures.Node{
											Properties: []structures.Property{},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		treeStruct{
			"(;FF[AA])",
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
					},
				},
			},
		},
		treeStruct{
			"(;FF[AA](;C[bbb]))",
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
					},
				},

				Children: []*structures.GameTree{
					&structures.GameTree{
						Sequence: structures.Sequence{
							Nodes: []structures.Node{
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
			},
		},
	}

	//////////////
	// TEST START
	for i, current := range treeMatrix {
		reader := getReader(current.raw)
		result, err := parser.ParseGameTree(reader)

		if err != nil {
			t.Errorf("Test %d returned error! %s; \nGiven %s, \nexpected %s", i, err.Error(), current.raw, current.parsed)
		}

		if result == nil {
			t.Errorf("Result is nil!")
			return
		}

		if err := compareGameTree(current.parsed, *result); err != nil {
			t.Errorf("Test %d failed. Error is %s!\nGiven %s, \nExpected \n%#v, \nfound \n%#v", i, err.Error(), current.raw, current.parsed, *result)
		}
	}
}

func compareProperties(expected, actual structures.Property) error {
	if &expected == &actual {
		// same object
		return nil
	}

	if expected.Ident != actual.Ident {
		return errors.New("Difference in PropIdent")
	}

	expectedValLen, actualValLen := len(expected.Values), len(actual.Values)
	if expectedValLen != actualValLen {
		return errors.New(fmt.Sprintf("Number of PropValues differes. Expected %d, actual %d", expectedValLen, actualValLen))
	}

	expectedMatch := make([]bool, len(expected.Values))

	for _, actualVal := range expected.Values {
		contains := false
		// Values can be unordered
		for i, expectedVal := range actual.Values {

			if !expectedMatch[i] && actualVal == expectedVal {
				expectedMatch[i] = true // will not check this index twice
				contains = true
				break
			}
		}

		// inner loop did not find a match
		if !contains {
			return errors.New("PropValues differ")
		}
	}

	return nil
}

func compareNode(expected, actual structures.Node) error {
	if &expected == &actual {
		// same object
		return nil
	}

	expectedLen, actualLen := len(expected.Properties), len(actual.Properties)
	if expectedLen != actualLen {
		return errors.New(fmt.Sprintf("Number of properties differ. Expected %d, actual %d", expectedLen, actualLen))
	}
	for i, prop := range expected.Properties {
		if err := compareProperties(prop, actual.Properties[i]); err != nil {
			return err
		}
	}
	return nil
}

func compareSequence(expected, actual structures.Sequence) error {
	if &expected == &actual {
		// same object
		return nil
	}

	// Make sure that we have the same number of nodes in the sequences
	expectedNodesLen, actualNodesLen := len(expected.Nodes), len(actual.Nodes)
	if expectedNodesLen != actualNodesLen {
		return errors.New(fmt.Sprintf("Number of Nodes differ. Expected %d, actual %d", expectedNodesLen, actualNodesLen))
	}

	for i, node := range expected.Nodes {
		// Make sure that we have the same number of properties in each node
		if err := compareNode(node, actual.Nodes[i]); err != nil {
			return err
		}
	}
	return nil
}

func compareGameTree(g1, g2 structures.GameTree) error {
	if &g1 == &g2 {
		//same object
		return nil
	}
	if err := compareSequence(g1.Sequence, g2.Sequence); err != nil {
		return err
	}

	expectedChildrenLen, actualChildrenLen := len(g1.Children), len(g2.Children)
	if expectedChildrenLen != actualChildrenLen {
		return errors.New(fmt.Sprintf("Different number of children! Expected %d, actual %d", expectedChildrenLen, actualChildrenLen))
	}

	for i, child := range g1.Children {
		childError := compareGameTree(*child, *g2.Children[i])
		if childError != nil {
			return childError
		}
	}

	return nil

}

func getReader(raw string) io.RuneScanner {
	//	return EchoReader{Reader: bufio.NewReader(strings.NewReader(raw))}
	return bufio.NewReader(strings.NewReader(raw))
}

// --------------------- DEBUGGING STUFF ------------------
type EchoReader struct {
	*bufio.Reader
}

func (r EchoReader) ReadRune() (char rune, size int, err error) {
	char, size, err = r.Reader.ReadRune()
	_, file, line, _ := runtime.Caller(1)
	if err != nil {
		fmt.Printf("EchoReader: error: %s\n", err.Error())
		fmt.Printf("%s: %d\n", file, line)
		return
	}

	fmt.Printf("EchoReader: read: %s\n", string(char))
	fmt.Printf("%s: %d\n", file, line)
	return
}

func (r EchoReader) UnreadRune() (err error) {
	err = r.Reader.UnreadRune()
	if err != nil {
		fmt.Printf("EchoReader: error: %s\n", err.Error())
		return
	}
	fmt.Println("EchoReader: unread successful!")
	return
}
