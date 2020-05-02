package main

import (
	"html"
)

func clean(in string) string {
	if len(in) > 500 {
		in = in[:500]
	}
	return html.EscapeString(in)
}
