package parser

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/makpoc/sgfparser/logger"
	"github.com/makpoc/sgfparser/structures"
)

const (
	uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var EmptyNodeError = errors.New("EmptyNode")
var ElementEndError = errors.New("Element's end reached")
var ParseError = errors.New("Parsing failed")

func ParseCollection(reader io.RuneScanner) (*structures.Collection, error) {
	collection := new(structures.Collection)

	for {
		// check if we are at the end of the stream
		_, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// if the end is not reached - unread and let the gametree parser handle the next tree if any
		err = reader.UnreadRune()
		if err != nil {
			return nil, err
		}

		gTree, err := ParseGameTree(reader)

		if err != nil {
			// maybe we should let this be configurable - fail the entire parsing process or just skip the current game tree
			logger.LogError("Failed to parse game tree. Skipping it!")
			continue
		}
		collection.GameTrees = append(collection.GameTrees, gTree)
	}

	return collection, nil
}

// ParseGameTree parses a game tree. This function is recursive - if there are sub trees in
// the current game tree - it will parse them as well and attach them as children to the current tree
//
// GameTree = "(" Sequence { GameTree } ")"
func ParseGameTree(reader io.RuneScanner) (*structures.GameTree, error) {
	gTree := new(structures.GameTree)

	// spin to the current tree start
	for {
		currRune, _, err := reader.ReadRune()
		if err != nil {
			return nil, err
		}
		if currRune == structures.GameTreeStart {
			break
		}
	}

	// The sequence for the current tree
	seq, err := ParseSequence(reader)
	if err != nil {
		return nil, err
	}
	gTree.Sequence = *seq

	for {
		currRune, _, err := reader.ReadRune()
		if err != nil {
			return nil, err
		}

		if currRune == structures.GameTreeEnd {
			break
		}

		if currRune == structures.GameTreeStart {
			// subtree start
			err = reader.UnreadRune()
			if err != nil {
				return nil, err
			}

			subTree, err := ParseGameTree(reader)
			if err != nil {
				return nil, err
			}

			gTree.Children = append(gTree.Children, subTree)
			subTree.Parent = gTree
			continue
			//break
		}

	}

	return gTree, nil

}

// ParseSequence parses a sequence of one or more nodes within a GameTree.
// Sequence = Node { Node }
func ParseSequence(reader io.RuneScanner) (*structures.Sequence, error) {
	seq := new(structures.Sequence)

	for {

		currRune, _, err := reader.ReadRune()
		if err != nil {
			return nil, err
		}

		if currRune == structures.GameTreeEnd {
			err = reader.UnreadRune()
			if err != nil {
				return nil, err
			}
			break
		}

		// for subtrees
		if currRune == structures.GameTreeStart {
			err = reader.UnreadRune()
			if err != nil {
				return nil, err
			}
			break
		}

		if currRune == structures.NodeSeparator {
			err = reader.UnreadRune()
			node, err := ParseNode(reader)
			if err != nil {
				return nil, err
			}

			seq.Nodes = append(seq.Nodes, *node)
			continue
		}
	}

	if len(seq.Nodes) < 1 {
		return nil, errors.New("Sequence must contain at least one node!")
	}

	return seq, nil
}

// ParseNode parses an entire Node with all its properties. The function will search for the first
// node separator and parse 1 node. It will NOT consume the next node separator (if any) or the game tree end
// Node = ";" { Property }
func ParseNode(reader io.RuneScanner) (*structures.Node, error) {

	var node structures.Node

	for {
		currRune, _, err := reader.ReadRune()
		if err != nil {
			return nil, err
		}

		if currRune == structures.NodeSeparator {
			// read the next rune to check whether the node contains properties.
			nextRune, _, err := reader.ReadRune()
			if err != nil {
				return nil, err
			}

			// If it ends with either ";",  "(" or ")" this node is empty and we will return it.
			if nextRune == structures.NodeSeparator || nextRune == structures.GameTreeStart || nextRune == structures.GameTreeEnd {
				err = reader.UnreadRune()
				if err != nil {
					return nil, err
				}

				// the node will be empty here.
				return &node, nil
			}

			err = reader.UnreadRune()
			if err != nil {
				return nil, err
			}

			property, err := ParseProperty(reader)
			if err != nil {
				if err == EmptyNodeError {
					return new(structures.Node), nil
				}
				return nil, err
			}

			node.Properties = append(node.Properties, *property)
			break
		}
	}

	return &node, nil
}

// Parses a Property. As per specification a property consist of one PropIdent and one or more unordered PropValues:
// Property = PropIdent PropValue { PropValue }
// TODO: In the future this method will check if the PropValue(s) have a type, suitable for the PropIdent.
func ParseProperty(reader io.RuneScanner) (*structures.Property, error) {
	var prop structures.Property

	ident, err := ParsePropIdent(reader)
	if err != nil {
		if err == EmptyNodeError {
			return nil, err
		}
		return nil, errors.New(fmt.Sprintf("Failed to parse Property. %s", err.Error()))
	}

	prop.Ident = *ident

	for {
		err := seekToNextPropValue(reader)
		if err != nil {
			if err == ElementEndError {
				break
			} else {
				return nil, err
			}
		}

		val, err := ParsePropValue(reader)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to parse Property. %s", err.Error()))
		}

		prop.Values = append(prop.Values, *val)
	}

	return &prop, nil
}

// Parses a PropIdent. As per specification PropIdent a word, containing 1 or 2 upper case letter(s). Space, tab, new line etc are also allowed.
// Validation whether the PropIdent is known or not will not be made here!
func ParsePropIdent(reader io.RuneScanner) (*structures.PropIdent, error) {
	var propIdent structures.PropIdent

	for {
		currRune, _, err := reader.ReadRune()
		if err != nil {
			return nil, err
		}

		if currRune == unicode.ReplacementChar {
			logger.LogDebug("Invalid unicode character! Skipping..")
			continue
		}

		if currRune == structures.NodeSeparator {
			return nil, EmptyNodeError
		}

		if currRune == structures.PropertyValueStart {
			// Unread the last rune so that ParsePropValue can start parsing
			err = reader.UnreadRune()
			if err != nil {
				return nil, err
			}

			break
		}

		propIdent += structures.PropIdent(currRune)
	}

	propIdent = structures.PropIdent(strings.Trim(string(propIdent), " \t\n"))

	if !isValid(propIdent) {
		return nil, errors.New(fmt.Sprintf("PropIdent %s is invalid!", propIdent))
	}
	return &propIdent, nil
}

func isValid(propIdent structures.PropIdent) bool {

	if len(propIdent) == 0 || len(propIdent) > 2 {
		return false
	}

	for _, r := range propIdent {

		if !strings.ContainsRune(uppercase, r) {
			return false
		}
	}

	return true
}

// Parses a PropValue. As per specs
//
// UcLetter   = "A".."Z"
// Digit      = "0".."9"
// None       = ""
// Number     = [("+"|"-")] Digit { Digit }
// Real       = Number ["." Digit { Digit }]
// Double     = ("1" | "2")
// Color      = ("B" | "W")
// SimpleText = { any character (handling see below) }
// Text       = { any character (handling see below) }
// Point      = game-specific
// Move       = game-specific
// Stone      = game-specific
// Compose    = ValueType ":" ValueType
//
// This Parser will not recognize the Value Type, but will strip some symbols, which are common for all types (e.g. tabs will become spaces).
func ParsePropValue(reader io.RuneScanner) (*structures.PropValue, error) {
	var propValue structures.PropValue

	// seek to the first PropertyValueStart rune
	for {
		currRune, _, err := reader.ReadRune()
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Could not find PropertyValueStart rune. %s", err.Error()))
		}
		if currRune == structures.PropertyValueStart {
			break
		}
	}

	escapeChar := '\\'
	doEscape := false

	for {
		currRune, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, ParseError
			}
		}

		// enter escape only if we are not escaping already
		if currRune == escapeChar && !doEscape {
			doEscape = true
			continue
		}

		// invalid char - ignore
		if currRune == unicode.ReplacementChar {
			logger.LogDebug("Invalid unicode character! Skipping..")
			continue
		}

		// replace tabs with spaces (as per spec)
		if currRune == '\t' {
			currRune = ' '
		}

		// if we are in escape mode and we encounter CR or LF
		if doEscape && (currRune == '\n' || currRune == '\r') {
			nextRune, _, err := reader.ReadRune()
			if err != nil {
				return nil, err
			}

			// if the current + next rune do not make a CRLF or LFCR sequence - unread it and discard only the single CR or LF
			if (currRune == '\n' && nextRune != '\r') || (currRune == '\r' && nextRune != '\n') {
				reader.UnreadRune()
			}

			// remove the new line if it's a "soft line break"
			doEscape = false
			continue
		}

		// if we've reached the end of the property
		if !doEscape && currRune == structures.PropertyValueEnd {
			// end parsing the current propValue only if ] is not escaped
			// Unread the last rune so that the caller knows that we've reached the propvalue end
			err = reader.UnreadRune()
			if err != nil {
				return nil, err
			}
			break
		}

		propValue += structures.PropValue(currRune)

		doEscape = false
	}

	lastRune, _, err := reader.ReadRune()
	if err != nil {
		return nil, err
	}

	if lastRune != structures.PropertyValueEnd {
		// we've exited the loop for some unusual reason. Return error
		return nil, errors.New("Property Value seems invalid")
	}

	return &propValue, nil
}

// This method will advance the reader to the next occurence of PropertyValueStart within the current Node/GameTree
// The reader is supposed to be pointing either to the end of a Property or to a place between two PropValues.
// If it's pointing to the end of a property, this method will return ElementEndError.
// Otherwise it will either return nil (seek successful) or another error
func seekToNextPropValue(reader io.RuneScanner) error {
	for {
		currRune, _, err := reader.ReadRune()
		if err != nil {
			return err
		}

		// stop if next variation or node is reached and return ElementEndError
		if currRune == structures.NodeSeparator || currRune == structures.GameTreeStart || currRune == structures.GameTreeEnd {
			err = reader.UnreadRune()
			if err != nil {
				return err
			}
			return ElementEndError
		}

		if currRune == structures.PropertyValueStart {
			err = reader.UnreadRune()
			if err != nil {
				return err
			}
			return nil
		}
	}

	return errors.New("Could not find PropValue end")
}
