# rrf

`rrf` stands for Realtime Regexp Filter.

It's a small utility for realtime filtering of text streams in your terminal
using regular expressions.

![Demo](https://content.camlittle.com/rrf-demo.mov)

## Usage

```
rrf --help
Usage:
  rrf [OPTIONS]

Application Options:
      --posix           Use POSIX ERE (egrep) syntax
      --filtered-count  Show count of filtered lines between output
      --highlight       Highlight matches in output

Help Options:
  -h, --help            Show this help message
```

## Future plans

- Automatic building/publishing
- More comprehensive tests
- Highlighting customization?
- Versioning
- Find something that will give me readline in the input prompt?
   Nothing I've found so far allows stateful reuse. They all block until a line
   is read, and I want to be able to keep streaming stdin while editing.
