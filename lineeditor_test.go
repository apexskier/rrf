package main

import (
	"bytes"
	"testing"

	"github.com/fatih/color"
)

func TestLineEditor(t *testing.T) {
	le := LineEditor{}

	enterText := func(s string) {
		for _, b := range []byte(s) {
			err := le.Consume(b)
			if err != nil {
				t.Error(err)
			}
		}
	}

	checkText := func(expected string) {
		if le.GetText() != expected {
			t.Errorf("text is wrong: expected %s, actual %s", expected, le.GetText())
		}
	}

	checkDisplayed := func(expected string) {
		var buf bytes.Buffer
		// flush to "screen"
		le.Display(&buf, "prefix", "suffix", color.New())
		actual := buf.String()
		// check that it was written, with proper cursor position
		if actual != expected {
			t.Errorf("display is wrong: expected %s, actual %s", expected, actual)
		}
		buf.Reset()
	}

	// ctl+d
	checkText("")
	checkDisplayed("prefixsuffix\033[6D")

	// "type" some data
	enterText("hello world")
	checkText("hello world")
	checkDisplayed("prefixhello worldsuffix\033[6D")

	// ctl+c
	err := le.Consume(3)
	if err != ErrCtlC {
		t.Error("ctl-c should return a ctl-c error")
	}

	// move the cursor back three columns
	enterText("\033[D\033[D\033[D")
	checkText("hello world")
	checkDisplayed("prefixhello worldsuffix\033[9D")

	// type some more
	enterText("__")
	checkText("hello wo__rld")
	checkDisplayed("prefixhello wo__rldsuffix\033[9D")

	// move cursor right
	enterText("\033[C")
	checkText("hello wo__rld")
	checkDisplayed("prefixhello wo__rldsuffix\033[8D")

	// ctl+d
	enterText(string(byte(4)))
	checkText("")
	checkDisplayed("prefixsuffix\033[6D")

	// ctl+d
	err = le.Consume(4)
	if err != ErrCtlD {
		t.Error("ctl-d with no content should return a ctl-d error")
	}

	// ctl+c
	err = le.Consume(3)
	if err != ErrCtlC {
		t.Error("ctl-c should return a ctl-c error")
	}
}
