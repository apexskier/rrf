package main

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

// LineEditor manages a "text entry" field that can be re-written to different
// places
type LineEditor struct {
	textBytes  []byte // TODO this should be a list of Runes?
	processing []byte
	// cursorPos stores the cursor position
	cursorPos int
	Done      chan bool
}

// GetText returns the text the user has currently typed
func (le *LineEditor) GetText() string {
	return string(le.textBytes)
}

// Display writes the current text with an appropriate cursor position to the
// given writer.
func (le *LineEditor) Display(to io.Writer, prefix, suffix string, fixColor *color.Color) {
	fmt.Fprint(to, fixColor.Sprint(prefix), string(le.textBytes), fixColor.Sprint(suffix))
	if le.cursorPos < len(le.textBytes)+len(suffix) {
		fmt.Fprintf(to, "\033[%dD", len(le.textBytes)-le.cursorPos+len(suffix))
	}
}

// Consume takes in a byte of date from keyboard entry, and returns true if it's
// the last one expected to be consumed
func (le *LineEditor) Consume(b byte) bool {
	if len(le.processing) == 2 {
		// we're in a control sequence
		switch b {
		case 65: // up
		case 66: // down
		case 67: // right
			if le.cursorPos < len(le.textBytes) {
				le.cursorPos++
			}
		case 68: // left
			if le.cursorPos > 0 {
				le.cursorPos--
			}
		}
		le.processing = []byte{}
		return false
	}
	switch b {
	case 3: // ctl+c
		return true
	case 4: // ctl+d
		// treat as ctl+c if no content
		if len(le.textBytes) == 0 {
			return true
		}
		// otherwise clear input
		le.textBytes = []byte{}
		le.cursorPos = 0
	case 127: // backspace
		if le.cursorPos > 0 {
			le.textBytes = append(le.textBytes[:le.cursorPos-1], le.textBytes[le.cursorPos:]...)
			le.cursorPos--
		}
	case 27: // esc
		le.processing = append(le.processing, b)
	case 91: // [
		if len(le.processing) == 1 {
			le.processing = append(le.processing, b)
			return false
		}
		fallthrough
	default:
		// increase length
		le.textBytes = append(le.textBytes, 0)
		// shift over
		copy(le.textBytes[le.cursorPos+1:], le.textBytes[le.cursorPos:])
		// insert new byte
		le.textBytes[le.cursorPos] = b
		le.cursorPos++
	}
	return false
}
