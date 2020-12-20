# rrf

`rrf` stands for Realtime Regexp Filter.

It's a small utility for realtime filtering of text streams in your terminal
using regular expressions.

[![asciicast](https://asciinema.org/a/Psv7iVedvswxtGMo4AE8F2AkH.svg)](https://asciinema.org/a/Psv7iVedvswxtGMo4AE8F2AkH)

## Usage

```
rrf --help
Usage:
  rrf [OPTIONS]

  rrf will read from stdin. Usually you'll want to do something like
  `tail -f my.log | rrf`

  Once running, type a regex. Once a valid regex is entered, the output filter
  will be updated.

Application Options:
      --posix           Use POSIX ERE (egrep) syntax
      --filtered-count  Show count of filtered lines between output
      --no-highlight    Do not highlight matches in output

Help Options:
  -h, --help            Show this help message
```

## Future plans

- Highlighting customization?
- Find something that will give me readline in the input prompt?
   Nothing I've found so far allows stateful reuse. They all block until a line
   is read, and I want to be able to keep streaming stdin while editing.
