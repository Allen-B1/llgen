package main

import (
	"fmt"
	"github.com/allen-b1/llgen/parser"
	"strings"
)

func transform(old string) string {
	return strings.Replace(strings.Title(strings.Replace(old, "-", " ", -1)), " ", "", -1)
}

func handleUnit(u parser.NodeUnit) (name string, tag string) {
	if token, ok := u.I.(parser.Token); ok {
		return token.Data, ""
	}
	if unitToken, ok := u.I.(parser.NodeUnitToken); ok {
		return unitToken.I0.Data, unitToken.I2.Data
	}
	panic("invalid tree for unit")
}

func handleUnitEll(u parser.NodeUnitEll) (name string, tag string, ell bool) {
	if unit, ok := u.I.(parser.NodeUnit); ok {
		name, tag := handleUnit(unit)
		return name, tag, false
	}
	if unitFull, ok := u.I.(parser.NodeUnitEllFull); ok {
		name, tag := handleUnit(unitFull.I0)
		return name, tag, true
	}
	panic("invalid tree for unit-ell")
}

func generate(n parser.NodeStatementExpr, symbols map[string]string) (string, error) {
	name := n.I0.Data
	if expr, ok := n.I2.I.(parser.NodeExprOr); ok {
		return generateOr(name, expr, symbols)
	}
	if expr, ok := n.I2.I.(parser.NodeExprAnd); ok {
		return generateAnd(name, expr, symbols)
	}
	return "", fmt.Errorf("invalid expression")
}

func generateAll(ns parser.NodeStatements) (string, error) {
	statements := ns.I0

	symbols := make(map[string]string)
	for _, statement := range statements {
		if token, ok := statement.I.(parser.NodeStatementToken); ok {
			if t, ok := token.I.(parser.NodeStatementTokenBasic); ok {
				symbols[t.I1.Data] = "token"
			}
			if t, ok := token.I.(parser.NodeStatementTokenExpanded); ok {
				symbols[t.I1.Data] = "token"
			}
		}
		if expr, ok := statement.I.(parser.NodeStatementExpr); ok {
			symbols[expr.I0.Data] = "expr"
		}
	}

	str := `
package parser

import (
	"fmt"
)

type Error struct {
	Message string
	Line int
}

func (e Error) Error() string {
	return fmt.Sprintf("%s (%v)", e.Message, e.Line)
}

func newError(msg string, line int) error {
	if line == 0 {
		msg += ": unexpected EOF"
	}
	return Error{msg, line}
}

func getLineOr0(tokens []Token, index int) int {
	if len(tokens) > index {
		return tokens[index].Line
	}
	return 0
}

func wrap(err error, msg string) error {
	if e, ok := err.(Error); ok {
		return Error{msg + ": " + e.Message, e.Line}
	} else {
		return Error{msg + ": " + err.Error(), 0}
	}
}

type Token struct {
	Type string
	Data string
	Line int
}
`
	for _, statement := range statements {
		if expr, ok := statement.I.(parser.NodeStatementExpr); ok {
			generated, err := generate(expr, symbols)
			if err != nil {
				return "", err
			}
			str += generated
		}
	}
	return str, nil
}

func generateAnd(name string, expr parser.NodeExprAnd, symbols map[string]string) (string, error) {
	newName := transform(name)

	var units []parser.NodeUnitEll = expr.I0
	fieldsStr := ""
	methodStr := ""
	i := 0
	for _, unitell := range units {
		identName, identTag, hasEll := handleUnitEll(unitell)

		if !hasEll {
			if symbols[identName] == "token" {
				fieldsStr += fmt.Sprintf("\tI%v Token // %s\n", i, identName)
				if identTag == "" {
					methodStr += fmt.Sprintf(`
	if len(in) <= curr || in[curr].Type != "%s" {
		return Node%s{}, 0, newError("failed to parse %s: %s expected", getLineOr0(in, curr))
	}
	out.I%v = in[curr]
	curr++
	`, identName, newName, name, identName, i)
				} else {
					methodStr += fmt.Sprintf(`
	if len(in) <= curr || in[curr].Type != "%s" || in[curr].Data != "%s" {
		return Node%s{}, 0, newError("failed to parse %s: %s expected", getLineOr0(in, curr))
	}
	out.I%v = in[curr]
	curr++
	`, identName, identTag, newName, name, identName, i)
				}
			} else if symbols[identName] == "expr" {
				fieldsStr += fmt.Sprintf("\tI%v Node%s\n", i, transform(identName))
				methodStr += fmt.Sprintf(`
	node%v, currChange, err := Parse%s(in[curr:])
	if err != nil {
		return Node%s{}, 0, wrap(err, "failed to parse %s")
	}
	out.I%v = node%v
	curr += currChange
				`, i, transform(identName), newName, name, i, i)
			} else {
				return "", fmt.Errorf("unknown identifier: %s", identName)
			}
		} else {
			methodStr += `
	for {`
			if symbols[identName] == "token" {
				fieldsStr += fmt.Sprintf("\tI%v []Token // %s\n", i, identName)
				if identTag == "" {
					methodStr += fmt.Sprintf(`
		if len(in) <= curr || in[curr].Type != "%s" {
			break
		}
		out.I%v = append(out.I%v, in[curr])
		curr++
	`, identName, i, i)
				} else {
					methodStr += fmt.Sprintf(`
		if len(in) <= curr || in[curr].Type != "%s" || in[curr].Data != "%s" {
			break
		}
		out.I%v = append(out.I%v, in[curr])
		curr++
	`, identName, identTag, i, i)
				}
			} else if symbols[identName] == "expr" {
				fieldsStr += fmt.Sprintf("\tI%v []Node%s\n", i, transform(identName))
				methodStr += fmt.Sprintf(`
		node%v, currChange, err := Parse%s(in[curr:])
		if err != nil {
			break
		}
		out.I%v = append(out.I%v, node%v)
		curr += currChange
				`, i, transform(identName), i, i, i)
			} else {
				return "", fmt.Errorf("unknown identifier: %s", identName)
			}

			methodStr += "\n\t}"
		}
		i++
	}

	str := fmt.Sprintf(`
type Node%s struct {
%s
}
`, newName, fieldsStr)
	str += fmt.Sprintf(`
func Parse%s(in []Token) (Node%s, int, error) {
	var out Node%s
	curr := 0
`, newName, newName, newName)
	str += methodStr
	str += `
	return out, curr, nil
}
`
	return str, nil
}

func generateOr(name string, expr parser.NodeExprOr, symbols map[string]string) (string, error) {
	newName := transform(name)
	str := fmt.Sprintf(`
type Node%s struct {
	I interface{}
}
`, newName)

	str += fmt.Sprintf(`
func Parse%s(in []Token) (Node%s, int, error) {`, newName, newName)

	var units []parser.NodeUnit
	units = append(units, expr.I0, expr.I2)
	for _, ext := range expr.I3 {
		units = append(units, ext.I1)
	}

	for _, unit := range units {
		identName := ""
		musteq := ""
		if token, ok := unit.I.(parser.Token); ok {
			identName = token.Data
		}
		if unitToken, ok := unit.I.(parser.NodeUnitToken); ok {
			identName = unitToken.I0.Data
			musteq = unitToken.I2.Data
		}

		if symbols[identName] == "token" && musteq == "" {
			str += fmt.Sprintf(`
	if len(in) != 0 && in[0].Type == "%s" {
		return Node%s{in[0]}, 1, nil
	}
`, identName, newName)
		} else if symbols[identName] == "token" && musteq != "" {
			str += fmt.Sprintf(`
	if len(in) != 0 && in[0].Type == "%s" && in[0].Data == "%s" {
		return Node%s{in[0]}, 1, nil
	}
`, identName, musteq, newName)
		} else if symbols[identName] == "expr" {
			str += fmt.Sprintf(`
	if node, curr, err := Parse%s(in); err == nil {
		return Node%s{node}, curr, nil
	}
		`, transform(identName), newName)
		} else {
			return "", fmt.Errorf("unknown identifier: %s", identName)
		}
	}

	str += fmt.Sprintf(`
	return Node%s{nil}, 0, newError("failed to parse %s", getLineOr0(in, 0))
}
`, newName, name)
	return str, nil
}
