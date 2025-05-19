package ecdata

import (
	"errors"
	"regexp"
)

// LoIDPrefix is the prefix for all LoIDs.
const LoIDPrefix = "t2_"

// LocaleRegex validates that locale codes are correctly formatted. They can contain
// either a language, or a language and region specifier separated by an underscore.
// e.g. en, en_US
var LocaleRegex = regexp.MustCompile(`^[a-z]{2,}([_|\-][\da-zA-Z]{2,})*$`)

var (
	// ErrLoIDWrongPrefix is an error could be returned by New() when passed in LoID
	// does not have the correct prefix.
	ErrLoIDWrongPrefix = errors.New("edgecontext: loid should have " + LoIDPrefix + " prefix")

	// ErrInvalidLocaleCode is returned by New() when an invalid locale code is passed in.
	ErrInvalidLocaleCode = errors.New("edgecontext: locale code should match format: en, en_US")
)
