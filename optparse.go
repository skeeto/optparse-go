// This is free and unencumbered software released into the public domain.

// Package optparse parses command line arguments very similarly to GNU
// getopt_long(). It supports long options and optional arguments, but
// does not permute arguments. It is intended as a replacement for Go's
// flag package.
//
// To use, define your options as an Option slice and pass it, along
// with the arguments string slice, to the Parse() function. It will
// return a slice of parsing results, which is to be iterated over just
// like getopt().
package optparse // import "github.com/skeeto/optparse-go"

import (
	"fmt"
	"strings"
)

const (
	// KindNone means the option takes no argument
	KindNone Kind = iota
	// KindRequired means the argument requires an option
	KindRequired
	// KindOptional means the argument is optional
	KindOptional

	// ErrInvalid is used when an option is not recognized.
	ErrInvalid = "invalid option"
	// ErrMissing is used when a required argument is missing.
	ErrMissing = "option requires an argument"
	// ErrTooMany is used when an unwanted argument is provided.
	ErrTooMany = "option takes no arguments"
)

// Kind is an enumeration indicating how an option is used.
type Kind int

// Option represents a single argument. Unicode is fully supported, so a
// short option may be any character. Using the zero value for Long
// or Short means the option has form of that size. Kind must be one of
// the constants.
type Option struct {
	Long  string
	Short rune
	Kind  Kind
}

// Error represents all possible parsing errors. It embeds the option
// that has been misused, and Message is one of the three error strings.
// Implements error.
type Error struct {
	Option
	Message string
}

func (e Error) Error() string {
	if e.Long != "" && e.Short != 0 {
		return fmt.Sprintf("%s: --%s (-%c)", e.Message, e.Long, e.Short)
	} else if e.Long != "" {
		return fmt.Sprintf("%s: --%s", e.Message, e.Long)
	} else {
		return fmt.Sprintf("%s: -%c", e.Message, e.Short)
	}
}

// Result is an individual successfully-parsed option. It embeds the
// original Option plus any argument. For options with optional
// arguments (KindOptional), it is not possible determine the difference
// between an empty supplied argument or no argument supplied.
type Result struct {
	Option
	Optarg string
}

// Parse results a slice of the parsed results, the remaining arguments,
// and the first parser error. The results slice always contains results
// up until the first error.
//
// The first argument, args[0], is skipped, and arguments are not
// permuted. Parsing stops at the first non-option argument, or "--".
// The latter is not included in the remaining, unparsed arguments.
func Parse(options []Option, args []string) ([]Result, []string, error) {
	parser := parser{options: options, args: args}
	var results []Result
	for {
		result, err := parser.next()
		if err != nil || result == nil {
			return results, parser.rest(), err
		}
		results = append(results, *result)
	}
}

// Parser represents the option parsing state between calls to next().
// The zero value for Parser is ready to use.
type parser struct {
	options []Option
	args    []string
	optind  int
	subopt  int
}

func (p *parser) short() (*Result, error) {
	runes := []rune(p.args[p.optind])
	c := runes[p.subopt]
	option := findShort(p.options, c)
	if option == nil {
		return nil, Error{Option{"", c, 0}, ErrInvalid}
	}
	switch option.Kind {

	case KindNone:
		p.subopt++
		if p.subopt == len(runes) {
			p.subopt = 0
			p.optind++
		}
		return &Result{*option, ""}, nil

	case KindRequired:
		optarg := string(runes[p.subopt+1:])
		p.subopt = 0
		p.optind++
		if optarg == "" {
			if p.optind == len(p.args) {
				return nil, Error{*option, ErrMissing}
			}
			optarg = p.args[p.optind]
			p.optind++
		}
		return &Result{*option, optarg}, nil

	case KindOptional:
		optarg := string(runes[p.subopt+1:])
		p.subopt = 0
		p.optind++
		return &Result{*option, optarg}, nil

	}
	panic("invalid Kind")
}

func (p *parser) long() (*Result, error) {
	long := p.args[p.optind][2:]

	eq := strings.IndexByte(long, '=')
	var optarg string
	var attached bool
	if eq != -1 {
		optarg = long[eq+1:]
		long = long[:eq]
		attached = true
	}

	option := findLong(p.options, long)
	if option == nil {
		return nil, Error{Option{long, 0, 0}, ErrInvalid}
	}
	p.optind++

	switch option.Kind {

	case KindNone:
		if attached {
			return nil, Error{*option, ErrTooMany}
		}
		return &Result{*option, ""}, nil

	case KindRequired:
		if p.optind == len(p.args) {
			return nil, Error{*option, ErrMissing}
		}
		if !attached {
			optarg = p.args[p.optind]
			p.optind++
		}
		return &Result{*option, optarg}, nil

	case KindOptional:
		return &Result{*option, optarg}, nil

	}
	panic("invalid Kind")
}

// Next returns the next option in the argument slice. When no arguments
// remain, returns nil as the result.
//
// If there is an error, the associated argument is not consumed.
func (p *parser) next() (*Result, error) {
	if p.optind == 0 {
		p.optind = 1 // initialize
	}

	if p.optind == len(p.args) {
		return nil, nil
	}
	arg := p.args[p.optind]

	if p.subopt > 0 {
		// continue parsing short options
		return p.short()
	}

	if len(arg) < 2 || arg[0] != '-' {
		return nil, nil
	}

	if arg == "--" {
		p.optind++
		return nil, nil
	}

	if arg[:2] == "--" {
		return p.long()
	}
	p.subopt = 1
	return p.short()
}

// Args slices the argument slice to return the arguments that were not
// parsed, excluding the "--".
func (p *parser) rest() []string {
	return p.args[p.optind:]
}

func findLong(options []Option, long string) *Option {
	for i, option := range options {
		if option.Long == long {
			return &options[i]
		}
	}
	return nil
}

func findShort(options []Option, short rune) *Option {
	for i, option := range options {
		if option.Short != 0 && option.Short == short {
			return &options[i]
		}
	}
	return nil
}
