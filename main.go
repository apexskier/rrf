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

func highlight(re *regexp.Regexp, in []byte, highlightMatch *color.Color, highlightGroup *color.Color, dimWith *color.Color) string {
	if re.String() == "" {
		return fmt.Sprintf("%s", in)
	}
	matches := re.FindAllSubmatchIndex(in, -1)
	// check if the entire string is matched
	// if len(matches) == 1 && len(matches[0]) == 1 && matches[0][0] == 0 && matches[0][1] == len(matches)-1 {
	// 	return fmt.Sprintf("%s", in)
	// }
	out := dimWith.Sprintf("%s", in[:matches[0][0]])
	for i, match := range matches {
		matchStr := ""
		if len(match) == 2 {
			matchStr = fmt.Sprintf("%s", in[match[0]:match[1]])
		}
		for j := 2; j < len(match); j += 2 {
			groupStartIndex := match[j]
			groupEndIndex := match[j+1]
			// highlight this group
			matchStr += highlightGroup.Sprintf("%s", in[groupStartIndex:groupEndIndex])
			// add text between this and next group
			if j == len(match)-2 {
				// add end of last group to end of match
				matchStr += fmt.Sprintf("%s", in[groupEndIndex:match[1]])
			} else {
				// add end of last group to start of next one
				matchStr += fmt.Sprintf("%s", in[groupEndIndex:match[j+2]])
			}
		}
		// highlight this match
		out += highlightMatch.Sprint(matchStr)
		// add text between this and next match
		if i == len(matches)-1 {
			// dim end of last match to end of string
			out += dimWith.Sprintf("%s", in[match[1]:])
		} else {
			// dim end of last match to start of next one
			nextMatch := matches[i+1]
			out += dimWith.Sprintf("%s", in[match[1]:nextMatch[0]])
		}
	}
	return out
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
			b := make([]byte, 1)
			_, err := tty.Read(b)
			if err != nil {
				panic(err)
			}
			ttyBytes <- b[0]
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
		w, _, err := term.GetSize(ttyFd)
		if err != nil {
			panic(err)
		}
		fmt.Printf("\r%"+fmt.Sprint(w)+"s\r", "")
	}

	faint := color.New(color.Faint)
	matchHighlight := color.New(color.Bold)
	groupHighlight := color.New(color.Underline)

	RegexpCompile := regexp.Compile
	if options.Posix {
		RegexpCompile = regexp.CompilePOSIX
	}

	currentPattern := []byte{}
	currentRegex, err := RegexpCompile("")
	if err != nil {
		// I'm not using MustCompile here to deduplicate the posix logic
		panic(err)
	}
	skippedLines := 0

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
			switch b {
			case 3: // ctl+c
				fmt.Printf("\r\nexiting\r\n")
				return
			case 4: // ctl+d
				if len(currentPattern) == 0 {
					fmt.Printf("\r\nexiting\r\n")
					return
				}
				currentPattern = []byte{}
				currentRegex = regexp.MustCompile("")
			case 127: // backspace
				if len(currentPattern) > 0 {
					currentPattern = currentPattern[:len(currentPattern)-1]
				}
			default:
				currentPattern = append(currentPattern, b)
			}
			re, err := RegexpCompile(string(currentPattern))
			if err == nil {
				currentRegex = re
			}
			wipeLine()
		}
		if skippedLines == 0 {
			fmt.Printf("\r%s %s", faint.Sprint("regex:"), currentPattern)
		} else {
			fmt.Printf("\r%d %s %s", skippedLines, faint.Sprint("regex:"), currentPattern)
		}
	}
}
