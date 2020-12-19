package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"github.com/fatih/color"
	"github.com/jessevdk/go-flags"
	"golang.org/x/term"
)

// Options describes the command line options this program supports
type Options struct {
	Posix                 bool `long:"posix" description:"Use POSIX ERE (egrep) syntax"`
	ShowFilteredLineCount bool `long:"filtered-count" description:"Show count of filtered lines between output"`
	Highlight             bool `long:"highlight" description:"Highlight matches in output" long-description:"If capturing groups are provided, each group will be highlighted individually. If no capture groups are provided, the entire first match will be highlighted."`
}

func main() {
	var options Options
	_, err := flags.Parse(&options)
	if err != nil {
		os.Exit(1)
	}

	inDone := make(chan bool, 1)
	inLines := make(chan []byte)
	ttyBytes := make(chan byte)

	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		color.NoColor = true
	}

	tty, err := os.Open("/dev/tty")
	if err != nil {
		fmt.Printf("can't open /dev/tty: %v", err)
		os.Exit(1)
	}

	ttyFd := int(tty.Fd())
	oldState, err := term.MakeRaw(ttyFd)
	if err != nil {
		panic(err)
	}
	defer term.Restore(ttyFd, oldState)

	go func() {
		for {
			bs := make([]byte, 3)
			n, err := tty.Read(bs)
			if err != nil {
				panic(err)
			}
			for i := 0; i < n; i++ {
				ttyBytes <- bs[i]
			}
		}
	}()

	go func() {
		// on return, this appears to wait until the next line is read before ending,
		// which causes the program to hang until then
		scannerIn := bufio.NewScanner(os.Stdin)
		for scannerIn.Scan() {
			inLines <- scannerIn.Bytes()
		}
		inDone <- true
	}()

	wipeLine := func() {
		fmt.Print("\r\033[K")
	}

	faint := color.New(color.Faint)
	matchHighlight := color.New(color.Bold)
	groupHighlight := color.New(color.Underline)

	RegexpCompile := regexp.Compile
	if options.Posix {
		RegexpCompile = regexp.CompilePOSIX
	}

	currentRegex, err := RegexpCompile("")
	if err != nil {
		// I'm not using MustCompile here to deduplicate the posix logic
		panic(err)
	}
	skippedLines := 0

	lineEditor := LineEditor{}

	for {
		select {
		case <-inDone:
			wipeLine()
			return
		case line := <-inLines:
			wipeLine()
			if currentRegex.Match(line) {
				if options.ShowFilteredLineCount && skippedLines > 0 {
					plural := "s"
					if skippedLines == 1 {
						plural = ""
					}
					fmt.Printf("%s\r\n", faint.Sprintf("%d line%s filtered", skippedLines, plural))
				}
				if options.Highlight {
					fmt.Printf("%s\r\n", highlight(currentRegex, line, matchHighlight, groupHighlight, faint))
				} else {
					fmt.Printf("%s\r\n", line)
				}
				skippedLines = 0
			} else {
				skippedLines++
			}
		case b := <-ttyBytes:
			if done := lineEditor.Consume(b); done {
				fmt.Printf("\r\nexiting\r\n")
				return
			}
			re, err := RegexpCompile(lineEditor.GetText())
			if err == nil {
				currentRegex = re
			}
			wipeLine()
		}
		var prefix string
		if skippedLines == 0 {
			prefix = faint.Sprint("regex: ")
		} else {
			prefix = fmt.Sprintf("%d %s ", skippedLines, faint.Sprint("regex:"))
		}
		fmt.Fprint(os.Stdout, "\r") // move cursor to start
		lineEditor.Display(os.Stdout, prefix)
	}
}
