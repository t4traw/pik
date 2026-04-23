package ui

import (
	"image/color"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

type Token struct {
	Text  string
	Color color.Color
}

// Highlight tokenizes code by filename and returns tokens split per line.
// If no suitable lexer is found, it falls back to one token per line with
// default foreground color.
func Highlight(filename, code string) [][]Token {
	// Normalise: strip CRLF
	code = strings.ReplaceAll(code, "\r\n", "\n")
	lines := strings.Split(code, "\n")

	lexer := lexers.Match(filename)
	if lexer == nil {
		lexer = lexers.Analyse(code)
	}
	if lexer == nil {
		return fallbackLines(lines)
	}
	lexer = chroma.Coalesce(lexer)

	iter, err := lexer.Tokenise(nil, code)
	if err != nil {
		return fallbackLines(lines)
	}

	result := make([][]Token, len(lines))
	li := 0
	for t := iter(); t != chroma.EOF; t = iter() {
		if t.Value == "" {
			continue
		}
		parts := strings.Split(t.Value, "\n")
		for i, p := range parts {
			if i > 0 {
				li++
				if li >= len(result) {
					// Safety: shouldn't happen, but avoid panic on OOB.
					break
				}
			}
			if p == "" {
				continue
			}
			if li >= len(result) {
				break
			}
			result[li] = append(result[li], Token{Text: p, Color: colorForToken(t.Type)})
		}
	}
	return result
}

func fallbackLines(lines []string) [][]Token {
	out := make([][]Token, len(lines))
	for i, ln := range lines {
		if ln == "" {
			continue
		}
		out[i] = []Token{{Text: ln, Color: ColorDefaultFG}}
	}
	return out
}

// colorForToken maps a chroma TokenType to our palette.
func colorForToken(tt chroma.TokenType) color.Color {
	// Walk up the token category if unknown.
	switch tt.Category() {
	case chroma.Keyword:
		switch tt {
		case chroma.KeywordType:
			return ColorSynType
		case chroma.KeywordConstant:
			return ColorSynConst
		}
		return ColorSynKeyword
	case chroma.Name:
		switch tt {
		case chroma.NameFunction, chroma.NameFunctionMagic, chroma.NameBuiltin, chroma.NameBuiltinPseudo:
			return ColorSynFunc
		case chroma.NameClass, chroma.NameException:
			return ColorSynType
		case chroma.NameConstant:
			return ColorSynConst
		case chroma.NameDecorator, chroma.NameAttribute:
			return ColorSynAttr
		case chroma.NameNamespace:
			return ColorSynNamespc
		case chroma.NameTag:
			return ColorSynTag
		case chroma.NameVariable, chroma.NameVariableClass,
			chroma.NameVariableGlobal, chroma.NameVariableInstance,
			chroma.NameVariableMagic:
			return ColorSynVar
		}
		return ColorSynOther
	case chroma.Literal:
		switch tt {
		case chroma.LiteralStringEscape, chroma.LiteralStringRegex:
			return ColorSynEscape
		}
		if isStringish(tt) {
			return ColorSynString
		}
		if isNumberish(tt) {
			return ColorSynNumber
		}
		return ColorSynOther
	case chroma.Comment:
		return ColorSynComment
	case chroma.Operator:
		return ColorSynOp
	case chroma.Punctuation:
		return ColorSynPunct
	case chroma.Error:
		return ColorSynError
	}
	return ColorDefaultFG
}

func isStringish(tt chroma.TokenType) bool {
	return tt >= chroma.LiteralString && tt <= chroma.LiteralStringSymbol
}

func isNumberish(tt chroma.TokenType) bool {
	return tt >= chroma.LiteralNumber && tt <= chroma.LiteralNumberOct
}
