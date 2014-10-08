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

func TestCompare(t *testing.T) {
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
		equal := compare(current.p1, current.p2)
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

	//logger.SetLogLevel(logger.DEBUG)

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

	logger.SetLogLevel(logger.OFF)
}

var propIdentsNeg = []string{
	"FF",
	"FFF[",
	"aF[",
	"[",
	" [",
	"	[",
	"AA][",
}

func TestParsePropIdentNegative(t *testing.T) {

	//	logger.SetLogLevel(logger.DEBUG)

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

	logger.SetLogLevel(logger.OFF)
}

type propValueMap struct {
	raw    string
	parsed structures.PropValue
}

var propValueMatrixPos = []propValueMap{
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

func TestParsePropValue(t *testing.T) {

	//logger.SetLogLevel(logger.DEBUG)

	for i, current := range propValueMatrixPos {

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
	logger.SetLogLevel(logger.OFF)
}

type propMap struct {
	raw    string
	parsed structures.Property
}

var propMatrixPos = []propMap{
	{"FF[3]", structures.Property{Ident: structures.PropIdent("FF"), Values: []structures.PropValue{structures.PropValue("3")}}},
	{" FF [ 3 ] ", structures.Property{Ident: structures.PropIdent("FF"), Values: []structures.PropValue{structures.PropValue("3")}}},
	{"FF[]", structures.Property{Ident: structures.PropIdent("FF"), Values: []structures.PropValue{structures.PropValue("")}}},
	{"AB[ac][bc]", structures.Property{Ident: structures.PropIdent("AB"), Values: []structures.PropValue{structures.PropValue("ac"), structures.PropValue("bc")}}},
}

func TestParseProperty(t *testing.T) {

	logger.SetLogLevel(logger.DEBUG)

	for i, current := range propMatrixPos {
		reader := getReader(current.raw)
		result, err := parser.ParseProperty(reader)

		if err != nil {
			t.Errorf("Test %d returned error! %s", i, err.Error())
		}

		if result != nil {

		}
	}

	logger.SetLogLevel(logger.OFF)
}

func compare(p1, p2 structures.Property) bool {
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

func getReader(raw string) *bufio.Reader {
	return bufio.NewReader(strings.NewReader(raw))
}
