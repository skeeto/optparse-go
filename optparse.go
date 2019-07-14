// This is free and unencumbered software released into the public domain.

// Package optparse parses command line arguments very similarly to GNU
// getopt_long(). It supports long options and optional arguments, but
// does not permute arguments. It is intended as a replacement for Go's
// flag package.
//
// To use, define your options as an Option slice and pass it, along
// with the argument slice, to the Next() method of a zero-initialized
// Parser. Each call to Next() will return the next argument, or the
// error if something went wrong.
package optparse // import "github.com/skeeto/optparse-go"

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// KindNone means the option takes no argument
	KindNone = iota
	// KindRequired means the argument requires an option
	KindRequired
	// KindOptional means the argument is optional
	KindOptional
)

// Done is the error returned when parsing is complete. It is analogous
// to io.EOF.
var Done = errors.New("end of arguments")

// Option represents a single argument. Unicode is fully supported, so a
// short option may be any character. Using the zero value for Long
// or Short means the option has form of that size. Kind must be one of
// the constants.
type Option struct {
	Long  string
	Short rune
	Kind  int
}

// Parser represents the option parsing state between calls to Parse().
// The zero value for Parser is ready to use.
type Parser struct {
	optind int
	subopt int
}

// Result is an individual successfully-parsed option. It embeds the
// original Option plus any argument. For options with optional
// arguments (KindOptional), it is not possible determine the difference
// between an empty supplied argument or no argument supplied.
type Result struct {
	Option
	Optarg string
}

func (p *Parser) short(options []Option, args []string) (*Result, error) {
	runes := []rune(args[p.optind])
	c := runes[p.subopt]
	option := findShort(options, c)
	if option == nil {
		return nil, fmt.Errorf("invalid option, %q", c)
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
			if p.optind == len(args) {
				return nil, fmt.Errorf("option requires an argument, %q", c)
			}
			optarg = args[p.optind]
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

func (p *Parser) long(options []Option, args []string) (*Result, error) {
	long := args[p.optind][2:]

	eq := strings.IndexByte(long, '=')
	var optarg string
	var attached bool
	if eq != -1 {
		optarg = long[eq+1:]
		long = long[:eq]
		attached = true
	}

	option := findLong(options, long)
	if option == nil {
		return nil, fmt.Errorf("invalid option, %q", long)
	}
	p.optind++

	switch option.Kind {

	case KindNone:
		if attached {
			return nil, fmt.Errorf("option takes no arguments, %q", long)
		}
		return &Result{*option, ""}, nil

	case KindRequired:
		if p.optind == len(args) {
			return nil, fmt.Errorf("option requires an argument, %q", long)
		}
		if !attached {
			optarg = args[p.optind]
			p.optind++
		}
		return &Result{*option, optarg}, nil

	case KindOptional:
		return &Result{*option, optarg}, nil

	}
	panic("invalid Kind")
}

// Next returns the next option in the argument slice. When no more
// arguments are left, returns Done as the error, like io.EOF. The first
// argument, args[0], is skipped. Arguments are not permuted and parsing
// stops at the first non-option argument, or "--".
//
// If there is an error, the associated argument is not consumed and
// would be returned by the Args() method.
func (p *Parser) Next(options []Option, args []string) (*Result, error) {
	if p.optind == 0 {
		p.optind = 1 // initialize
	}

	if p.optind == len(args) {
		return nil, Done
	}
	arg := args[p.optind]

	if p.subopt > 0 {
		// continue parsing short options
		return p.short(options, args)
	}

	if len(arg) < 2 || arg[0] != '-' {
		return nil, Done
	}

	if arg == "--" {
		p.optind++
		return nil, Done
	}

	if arg[:2] == "--" {
		return p.long(options, args)
	}
	p.subopt = 1
	return p.short(options, args)
}

// Args slices the argument slice to return the arguments that were not
// parsed, excluding the "--".
func (p *Parser) Args(args []string) []string {
	return args[p.optind:]
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
