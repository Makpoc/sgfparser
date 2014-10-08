package parser

import (
	"bufio"
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

// Parses a Property. As per specification a property consist of one PropIdent and one or more unordered PropValues:
// Property = PropIdent PropValue { PropValue }
// TODO: In the future this method will check if the PropValue(s) have a type, suitable for the PropIdent.
func ParseProperty(reader *bufio.Reader) (*structures.Property, error) {
	return nil, nil
}

// Parses a PropIdent. As per specification PropIdent a word, containing 1 or 2 upper case letter(s). Space, tab, new line etc are also allowed.
// Validation whether the PropIdent is known or not will not be made here!
func ParsePropIdent(reader *bufio.Reader) (*structures.PropIdent, error) {
	logger.LogDebug("Parsing PropIdent..")
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

	if isValid(propIdent) {
		return &propIdent, nil
	} else {
		return nil, errors.New(fmt.Sprintf("PropIdent %s is invalid!", propIdent))
	}
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
func ParsePropValue(reader *bufio.Reader) (*structures.PropValue, error) {
	logger.LogDebug("Parsing PropValue..")
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
				return nil, errors.New(fmt.Sprintf("Error while parsing propIdent: %s", err.Error()))
			}
		}

		if currRune == unicode.ReplacementChar {
			logger.LogDebug("Invalid unicode character! Skipping..")
			continue
		}

		if currRune == '\t' {
			currRune = ' ' // replace tabs with spaces
		}

		if doEscape && (currRune == '\n' || currRune == '\r') {
			nextRune, _, err := reader.ReadRune()
			if err != nil {
				return nil, err
			}

			// if the current+next rune do not make a CRLF or LFCR sequence - unread it and discard only the single CR or LF
			if (currRune == '\n' && nextRune != '\r') || (currRune == '\r' && nextRune != '\n') {
				reader.UnreadRune()
			}

			// remove the new line if it's a "soft line break"
			doEscape = false
			continue
		}

		if currRune == structures.PropertyValueEnd && !doEscape {
			// end parsing the current propValue only if ] is not escaped
			// Unread the last rune so that the caller knows what has happend
			err = reader.UnreadRune()
			if err != nil {
				return nil, err
			}
			break
		}

		// enter escape only if we are not escaping already
		if currRune == escapeChar && !doEscape {
			doEscape = true
			continue
		}

		propValue += structures.PropValue(currRune)

		doEscape = false
	}

	return &propValue, nil
}
