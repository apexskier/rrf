package main

import (
	"regexp"
	"testing"

	"github.com/fatih/color"
)

func TestHighlightGroups(t *testing.T) {
	dateRE := regexp.MustCompile(`(?P<Year>\d{4})-(?P<Month>\d{2})-(?P<Day>\d{2})`)
	line := highlight(
		dateRE,
		[]byte(`a date is 2015-05-27 end`),
		color.New(),
		color.New(),
		color.New(),
	)
	if string(line) != "a date is 2015-05-27 end" {
		t.Errorf("unexpected output: %s", line)
	}
	line = highlight(
		dateRE,
		[]byte(`a date is 2015-05-27`),
		color.New(),
		color.New(),
		color.New(),
	)
	if string(line) != "a date is 2015-05-27" {
		t.Errorf("unexpected output: %s", line)
	}
	line = highlight(
		dateRE,
		[]byte(`2015-05-27 end`),
		color.New(),
		color.New(),
		color.New(),
	)
	if string(line) != "2015-05-27 end" {
		t.Errorf("unexpected output: %s", line)
	}
}

func TestHighlightNoMatches(t *testing.T) {
	dateRE := regexp.MustCompile("")
	line := highlight(
		dateRE,
		[]byte(`a date is 2015-05-27 end`),
		color.New(),
		color.New(),
		color.New(),
	)
	if string(line) != "a date is 2015-05-27 end" {
		t.Errorf("unexpected output: %s", line)
	}
}

func TestHighlightMultipleMatchesNoGroups(t *testing.T) {
	dateRE := regexp.MustCompile("date")
	line := highlight(
		dateRE,
		[]byte(`a date is 2015-05-27, a date`),
		color.New(),
		color.New(),
		color.New(),
	)
	if string(line) != "a date is 2015-05-27, a date" {
		t.Errorf("unexpected output: %s", line)
	}
}

func TestHighlightMultipleMatchesAndGroups(t *testing.T) {
	dateRE := regexp.MustCompile("(date)")
	line := highlight(
		dateRE,
		[]byte(`a date is 2015-05-27, a date`),
		color.New(),
		color.New(),
		color.New(),
	)
	if string(line) != "a date is 2015-05-27, a date" {
		t.Errorf("unexpected output: %s", line)
	}
}

func TestHighlightPrefixUnGrouped(t *testing.T) {
	dateRE := regexp.MustCompile("da(te)")
	line := highlight(
		dateRE,
		[]byte(`a date is 2015-05-27, a date`),
		color.New(),
		color.New(),
		color.New(),
	)
	if string(line) != "a date is 2015-05-27, a date" {
		t.Errorf("unexpected output: %s", line)
	}
}
