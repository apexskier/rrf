package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/fatih/color"
	"github.com/jessevdk/go-flags"
	"golang.org/x/term"
)

// Options describes the command line options this program supports
type Options struct {
	Posix                 bool   `long:"posix" description:"Use POSIX ERE (egrep) syntax"`
	ShowFilteredLineCount bool   `long:"filtered-count" description:"Show count of filtered lines between output"`
	NoHighlight           bool   `long:"no-highlight" description:"Do not highlight matches in output" long-description:"If capturing groups are provided, each group will be highlighted individually. If no capture groups are provided, the entire first match will be highlighted."`
	Version               func() `long:"version" short:"v" description:"Print version information"`
}

func main() {
	var options Options

	options.Version = func() {
		fmt.Printf("%s (%s)\n", Version, GitCommit)
		os.Exit(0)
	}

	flagsParser := flags.NewParser(&options, flags.Default)
	flagsParser.Usage += `[OPTIONS]

  rrf will read from stdin. Usually you'll want to do something like
  ` + "`" + `tail -f my.log | rrf` + "`" + `

  Once running, type a regex. Once a valid regex is entered, the output filter
  will be updated.`
	_, err := flagsParser.Parse()
	if err != nil {
		os.Exit(1)
	}

	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		flagsParser.WriteHelp(os.Stdout)
		os.Exit(1)
	}

	stop := make(chan error, 1)
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
		// Scanner.Scan blocks until the next input. Since Stdin may take a time to
		// produce data, it'll hang
		// Some threads
		// - https://www.reddit.com/r/golang/comments/fsxkqr/cancelling_blocking_read_from_stdin/
		// - https://github.com/golang/go/issues/7990
		// - https://stackoverflow.com/questions/60960288/stopping-a-bufio-scanner-with-a-stop-channel
		scannerIn := bufio.NewScanner(os.Stdin)
		for scannerIn.Scan() {
			inLines <- scannerIn.Bytes()
		}
		stop <- scannerIn.Err()
	}()

	wipeLine := func() {
		fmt.Print("\r\033[K")
	}

	faint := color.New(color.Faint)
	matchHighlight := color.New()
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
		case err := <-stop:
			wipeLine()
			if err != nil {
				panic(err)
			}
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
				if options.NoHighlight {
					fmt.Printf("%s\r\n", line)
				} else {
					fmt.Printf("%s\r\n", highlight(currentRegex, line, matchHighlight, groupHighlight, faint))
				}
				skippedLines = 0
			} else {
				skippedLines++
			}
		case b := <-ttyBytes:
			if err := lineEditor.Consume(b); err != nil {
				if errors.Is(err, ErrCtlC) || errors.Is(err, ErrCtlD) {
					fmt.Printf("\r\nexiting\r\n")
					return
				}
				panic(err)
			}
			re, err := RegexpCompile(lineEditor.GetText())
			if err == nil {
				currentRegex = re
			}
			wipeLine()
		}
		fmt.Fprint(os.Stdout, "\r") // move cursor to start
		suffix := ""
		if options.ShowFilteredLineCount && skippedLines > 0 {
			plural := "s"
			if skippedLines == 1 {
				plural = ""
			}
			suffix = fmt.Sprintf(" (%d line%s filtered since last)", skippedLines, plural)
		}
		lineEditor.Display(os.Stdout, "regex: ", suffix, faint)
	}
}
