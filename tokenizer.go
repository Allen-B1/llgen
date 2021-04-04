package main

import (
	"fmt"
	"github.com/allen-b1/llgen/parser"
	"strings"
	"unicode"
)

func tokenize(in string) ([]parser.Token, error) {
	in = strings.Replace(in, "\r", "", -1)

	out := make([]parser.Token, 0)
	line := 1
	i := 0
	for i < len(in) {
		if in[i] == '=' {
			out = append(out, parser.Token{"eq", "=", line})
			i += 1
			continue
		}
		if in[i] == '|' {
			out = append(out, parser.Token{"or", "|", line})
			i += 1
			continue
		}
		if in[i] == '<' {
			out = append(out, parser.Token{"al", "<", line})
			i += 1
			continue
		}
		if in[i] == '>' {
			out = append(out, parser.Token{"ar", ">", line})
			i += 1
			continue
		}
		if in[i] == '\n' {
			out = append(out, parser.Token{"newline", "", line})
			line += 1
			i += 1
			continue
		}
		if unicode.IsLetter(rune(in[i])) {
			end := i
			for end < len(in) && (unicode.IsLetter(rune(in[end])) || unicode.IsNumber(rune(in[end])) || in[end] == '-') {
				end += 1
			}
			out = append(out, parser.Token{"ident", in[i:end], line})
			i = end
			continue
		}
		if in[i] == '"' {
			end := i + 1
			for end < len(in) && in[end] != '"' {
				end += 1
			}
			out = append(out, parser.Token{"string", in[i+1 : end], line})
			i = end + 1
			continue
		}
		if in[i] == '.' {
			if len(in) <= i+2 {
				return nil, parser.Error{"expected ., got EOF", line}
			}
			if in[i+1] != '.' || in[i+2] != '.' {
				return nil, parser.Error{"expected .", line}
			}
			out = append(out, parser.Token{"ell", "...", line})
			i += 3
			continue
		}
		if in[i] == ' ' || in[i] == '\t' {
			i++
			continue
		}
		return nil, parser.Error{fmt.Sprintf("invalid token: %c", in[i]), line}
	}
	return out, nil
}
