package optparse

import (
	"strconv"
	"testing"
)

var options = []Option{
	{"amend", 'a', KindNone},
	{"brief", 'b', KindNone},
	{"color", 'c', KindOptional},
	{"delay", 'd', KindRequired},
	{"erase", 'e', KindNone},

	// special cases
	{"pi", 'π', KindNone}, // multibyte short option
	{"long", 0, KindNone}, // long only
	{"", 's', KindNone},   // short only
}

type config struct {
	amend bool
	brief bool
	color string
	delay int
	erase int
	pi    int
}

func parse(args []string) (conf config, rest []string, err error) {
	var results []Result
	results, rest, err = Parse(options, args)
	for _, result := range results {
		switch result.Long {
		case "amend":
			conf.amend = true
		case "brief":
			conf.brief = true
		case "color":
			conf.color = result.Optarg
		case "delay":
			delay, _ := strconv.Atoi(result.Optarg)
			conf.delay = delay
		case "erase":
			conf.erase++
		case "pi":
			conf.pi++
		}
	}
	return
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestParse(t *testing.T) {
	table := []struct {
		args []string
		conf config
		rest []string
		err  error
	}{
		{
			[]string{"", "--", "foobar"},
			config{false, false, "", 0, 0, 0},
			[]string{"foobar"},
			nil,
		},
		{
			[]string{"", "-a", "-b", "-c", "-d", "10", "-e"},
			config{true, true, "", 10, 1, 0},
			[]string{},
			nil,
		},
		{
			[]string{
				"",
				"--amend",
				"--brief",
				"--color",
				"--delay", "10",
				"--erase",
			},
			config{true, true, "", 10, 1, 0},
			[]string{},
			nil,
		},
		{
			[]string{"", "-a", "-b", "-cred", "-d", "10", "-e"},
			config{true, true, "red", 10, 1, 0},
			[]string{},
			nil,
		},
		{
			[]string{"", "-abcblue", "-d10", "foobar"},
			config{true, true, "blue", 10, 0, 0},
			[]string{"foobar"},
			nil,
		},
		{
			[]string{"", "--color=red", "-d", "10", "--", "foobar"},
			config{false, false, "red", 10, 0, 0},
			[]string{"foobar"},
			nil,
		},
		{
			[]string{"", "-eeeeee"},
			config{false, false, "", 0, 6, 0},
			[]string{},
			nil,
		},
		{
			[]string{"", "-πeabπee"},
			config{true, true, "", 0, 3, 2},
			[]string{},
			nil,
		},
		{
			[]string{"", "--delay"},
			config{false, false, "", 0, 0, 0},
			[]string{},
			Error{Option{"delay", 'd', KindRequired}, ErrMissing},
		},
		{
			[]string{"", "--foo", "bar"},
			config{false, false, "", 0, 0, 0},
			[]string{"--foo", "bar"},
			Error{Option{"foo", 0, 0}, ErrInvalid},
		},
		{
			[]string{"", "-x"},
			config{false, false, "", 0, 0, 0},
			[]string{"-x"},
			Error{Option{"", 'x', 0}, ErrInvalid},
		},
		{
			[]string{"", "-"},
			config{false, false, "", 0, 0, 0},
			[]string{"-"},
			nil,
		},
		{
			[]string{"", "-\x00"},
			config{false, false, "", 0, 0, 0},
			[]string{"-\x00"},
			Error{Option{"", 0, 0}, ErrInvalid},
		},
	}

	for _, row := range table {
		conf, rest, err := parse(row.args)
		if conf != row.conf {
			t.Errorf("parse(%q), got %v, want %v", row.args[1:], conf, row.conf)
		}
		if !equal(rest, row.rest) {
			t.Errorf("parse(%q), got %v, want %v", row.args[1:], rest, row.rest)
		}
		if row.err != nil {
			want := row.err.(Error)
			if err == nil {
				t.Errorf("parse(%q), got nil, wanted %#v", row.args[1:], want)
			} else if got := err.(Error); got != want {
				t.Errorf("parse(%q), got %#v, wanted %#v",
					row.args[1:], got, want)
			}
		}
	}
}
