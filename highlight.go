package main

import (
	"fmt"
	"regexp"

	"github.com/fatih/color"
)

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
		if len(match) == 2 {
			// no groups
			out += highlightMatch.Sprintf("%s", in[match[0]:match[1]])
		} else {
			// add ungrouped prefix
			out += highlightMatch.Sprintf("%s", in[match[0]:match[2]])
		}
		for j := 2; j < len(match); j += 2 {
			groupStartIndex := match[j]
			groupEndIndex := match[j+1]
			// highlight this group
			out += highlightGroup.Sprintf("%s", in[groupStartIndex:groupEndIndex])
			// add text between this and next group
			if j == len(match)-2 {
				// add end of last group to end of match
				out += highlightMatch.Sprintf("%s", in[groupEndIndex:match[1]])
			} else {
				// add end of last group to start of next one
				out += highlightMatch.Sprintf("%s", in[groupEndIndex:match[j+2]])
			}
		}
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
