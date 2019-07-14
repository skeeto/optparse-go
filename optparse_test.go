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
	{"pi", 'π', KindNone},
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
	var parser Parser
	for {
		result, err := parser.Next(options, args)
		if err != nil {
			if err != Done {
				return conf, parser.Args(args), err
			}
			break
		}
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
	return conf, parser.Args(args), nil
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
		err  bool
	}{
		{
			[]string{"", "--", "foobar"},
			config{false, false, "", 0, 0, 0},
			[]string{"foobar"},
			false,
		},
		{
			[]string{"", "-a", "-b", "-c", "-d", "10", "-e"},
			config{true, true, "", 10, 1, 0},
			[]string{},
			false,
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
			false,
		},
		{
			[]string{"", "-a", "-b", "-cred", "-d", "10", "-e"},
			config{true, true, "red", 10, 1, 0},
			[]string{},
			false,
		},
		{
			[]string{"", "-abcblue", "-d10", "foobar"},
			config{true, true, "blue", 10, 0, 0},
			[]string{"foobar"},
			false,
		},
		{
			[]string{"", "--color=red", "-d", "10", "--", "foobar"},
			config{false, false, "red", 10, 0, 0},
			[]string{"foobar"},
			false,
		},
		{
			[]string{"", "-eeeeee"},
			config{false, false, "", 0, 6, 0},
			[]string{},
			false,
		},
		{
			[]string{"", "-πeabπee"},
			config{true, true, "", 0, 3, 2},
			[]string{},
			false,
		},
		{
			[]string{"", "--delay"},
			config{false, false, "", 0, 0, 0},
			[]string{},
			true,
		},
		{
			[]string{"", "--foo", "bar"},
			config{false, false, "", 0, 0, 0},
			[]string{"--foo", "bar"},
			true,
		},
		{
			[]string{"", "-x"},
			config{false, false, "", 0, 0, 0},
			[]string{"-x"},
			true,
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
		if row.err {
			if err == nil {
				t.Errorf("parse(%q), got nil, wanted nil", row.args[1:])
			}
		} else {
			if err != nil {
				t.Errorf("parse(%q), got %v, wanted nil", row.args[1:], err)
			}
		}
	}
}
