
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

type NodeUnit struct {
	I interface{}
}

func ParseUnit(in []Token) (NodeUnit, int, error) {
	if node, curr, err := ParseUnitToken(in); err == nil {
		return NodeUnit{node}, curr, nil
	}
		
	if len(in) != 0 && in[0].Type == "ident" {
		return NodeUnit{in[0]}, 1, nil
	}

	return NodeUnit{nil}, 0, newError("failed to parse unit", getLineOr0(in, 0))
}

type NodeUnitToken struct {
	I0 Token // ident
	I1 Token // al
	I2 Token // string
	I3 Token // ar

}

func ParseUnitToken(in []Token) (NodeUnitToken, int, error) {
	var out NodeUnitToken
	curr := 0

	if len(in) <= curr || in[curr].Type != "ident" {
		return NodeUnitToken{}, 0, newError("failed to parse unit-token: ident expected", getLineOr0(in, curr))
	}
	out.I0 = in[curr]
	curr++
	
	if len(in) <= curr || in[curr].Type != "al" {
		return NodeUnitToken{}, 0, newError("failed to parse unit-token: al expected", getLineOr0(in, curr))
	}
	out.I1 = in[curr]
	curr++
	
	if len(in) <= curr || in[curr].Type != "string" {
		return NodeUnitToken{}, 0, newError("failed to parse unit-token: string expected", getLineOr0(in, curr))
	}
	out.I2 = in[curr]
	curr++
	
	if len(in) <= curr || in[curr].Type != "ar" {
		return NodeUnitToken{}, 0, newError("failed to parse unit-token: ar expected", getLineOr0(in, curr))
	}
	out.I3 = in[curr]
	curr++
	
	return out, curr, nil
}

type NodeUnitEll struct {
	I interface{}
}

func ParseUnitEll(in []Token) (NodeUnitEll, int, error) {
	if node, curr, err := ParseUnitEllFull(in); err == nil {
		return NodeUnitEll{node}, curr, nil
	}
		
	if node, curr, err := ParseUnitEllOpt(in); err == nil {
		return NodeUnitEll{node}, curr, nil
	}
		
	if node, curr, err := ParseUnit(in); err == nil {
		return NodeUnitEll{node}, curr, nil
	}
		
	return NodeUnitEll{nil}, 0, newError("failed to parse unit-ell", getLineOr0(in, 0))
}

type NodeUnitEllFull struct {
	I0 NodeUnit
	I1 Token // ell

}

func ParseUnitEllFull(in []Token) (NodeUnitEllFull, int, error) {
	var out NodeUnitEllFull
	curr := 0

	node0, currChange, err := ParseUnit(in[curr:])
	if err != nil {
		return NodeUnitEllFull{}, 0, wrap(err, "failed to parse unit-ell-full")
	}
	out.I0 = node0
	curr += currChange
				
	if len(in) <= curr || in[curr].Type != "ell" {
		return NodeUnitEllFull{}, 0, newError("failed to parse unit-ell-full: ell expected", getLineOr0(in, curr))
	}
	out.I1 = in[curr]
	curr++
	
	return out, curr, nil
}

type NodeUnitEllOpt struct {
	I0 NodeUnit
	I1 Token // opt

}

func ParseUnitEllOpt(in []Token) (NodeUnitEllOpt, int, error) {
	var out NodeUnitEllOpt
	curr := 0

	node0, currChange, err := ParseUnit(in[curr:])
	if err != nil {
		return NodeUnitEllOpt{}, 0, wrap(err, "failed to parse unit-ell-opt")
	}
	out.I0 = node0
	curr += currChange
				
	if len(in) <= curr || in[curr].Type != "opt" {
		return NodeUnitEllOpt{}, 0, newError("failed to parse unit-ell-opt: opt expected", getLineOr0(in, curr))
	}
	out.I1 = in[curr]
	curr++
	
	return out, curr, nil
}

type NodeExprAnd struct {
	I0 []NodeUnitEll

}

func ParseExprAnd(in []Token) (NodeExprAnd, int, error) {
	var out NodeExprAnd
	curr := 0

	for {
		node0, currChange, err := ParseUnitEll(in[curr:])
		if err != nil {
			break
		}
		out.I0 = append(out.I0, node0)
		curr += currChange
				
	}
	return out, curr, nil
}

type NodeExprOr struct {
	I0 NodeUnit
	I1 Token // or
	I2 NodeUnit
	I3 []NodeExprOrExt

}

func ParseExprOr(in []Token) (NodeExprOr, int, error) {
	var out NodeExprOr
	curr := 0

	node0, currChange, err := ParseUnit(in[curr:])
	if err != nil {
		return NodeExprOr{}, 0, wrap(err, "failed to parse expr-or")
	}
	out.I0 = node0
	curr += currChange
				
	if len(in) <= curr || in[curr].Type != "or" {
		return NodeExprOr{}, 0, newError("failed to parse expr-or: or expected", getLineOr0(in, curr))
	}
	out.I1 = in[curr]
	curr++
	
	node2, currChange, err := ParseUnit(in[curr:])
	if err != nil {
		return NodeExprOr{}, 0, wrap(err, "failed to parse expr-or")
	}
	out.I2 = node2
	curr += currChange
				
	for {
		node3, currChange, err := ParseExprOrExt(in[curr:])
		if err != nil {
			break
		}
		out.I3 = append(out.I3, node3)
		curr += currChange
				
	}
	return out, curr, nil
}

type NodeExprOrExt struct {
	I0 Token // or
	I1 NodeUnit

}

func ParseExprOrExt(in []Token) (NodeExprOrExt, int, error) {
	var out NodeExprOrExt
	curr := 0

	if len(in) <= curr || in[curr].Type != "or" {
		return NodeExprOrExt{}, 0, newError("failed to parse expr-or-ext: or expected", getLineOr0(in, curr))
	}
	out.I0 = in[curr]
	curr++
	
	node1, currChange, err := ParseUnit(in[curr:])
	if err != nil {
		return NodeExprOrExt{}, 0, wrap(err, "failed to parse expr-or-ext")
	}
	out.I1 = node1
	curr += currChange
				
	return out, curr, nil
}

type NodeExpr struct {
	I interface{}
}

func ParseExpr(in []Token) (NodeExpr, int, error) {
	if node, curr, err := ParseExprOr(in); err == nil {
		return NodeExpr{node}, curr, nil
	}
		
	if node, curr, err := ParseExprAnd(in); err == nil {
		return NodeExpr{node}, curr, nil
	}
		
	return NodeExpr{nil}, 0, newError("failed to parse expr", getLineOr0(in, 0))
}

type NodeStatementExpr struct {
	I0 Token // ident
	I1 Token // eq
	I2 NodeExpr
	I3 Token // newline

}

func ParseStatementExpr(in []Token) (NodeStatementExpr, int, error) {
	var out NodeStatementExpr
	curr := 0

	if len(in) <= curr || in[curr].Type != "ident" {
		return NodeStatementExpr{}, 0, newError("failed to parse statement-expr: ident expected", getLineOr0(in, curr))
	}
	out.I0 = in[curr]
	curr++
	
	if len(in) <= curr || in[curr].Type != "eq" {
		return NodeStatementExpr{}, 0, newError("failed to parse statement-expr: eq expected", getLineOr0(in, curr))
	}
	out.I1 = in[curr]
	curr++
	
	node2, currChange, err := ParseExpr(in[curr:])
	if err != nil {
		return NodeStatementExpr{}, 0, wrap(err, "failed to parse statement-expr")
	}
	out.I2 = node2
	curr += currChange
				
	if len(in) <= curr || in[curr].Type != "newline" {
		return NodeStatementExpr{}, 0, newError("failed to parse statement-expr: newline expected", getLineOr0(in, curr))
	}
	out.I3 = in[curr]
	curr++
	
	return out, curr, nil
}

type NodeStatementToken struct {
	I0 Token // ident
	I1 Token // ident
	I2 *NodeStatementTokenAnnotation
	I3 Token // newline

}

func ParseStatementToken(in []Token) (NodeStatementToken, int, error) {
	var out NodeStatementToken
	curr := 0

	if len(in) <= curr || in[curr].Type != "ident" || in[curr].Data != "token" {
		return NodeStatementToken{}, 0, newError("failed to parse statement-token: ident expected", getLineOr0(in, curr))
	}
	out.I0 = in[curr]
	curr++
	
	if len(in) <= curr || in[curr].Type != "ident" {
		return NodeStatementToken{}, 0, newError("failed to parse statement-token: ident expected", getLineOr0(in, curr))
	}
	out.I1 = in[curr]
	curr++
	
	node2, currChange, err := ParseStatementTokenAnnotation(in[curr:])
	if err == nil {
		out.I2 = &node2
		curr += currChange
	}
				
	if len(in) <= curr || in[curr].Type != "newline" {
		return NodeStatementToken{}, 0, newError("failed to parse statement-token: newline expected", getLineOr0(in, curr))
	}
	out.I3 = in[curr]
	curr++
	
	return out, curr, nil
}

type NodeStatementTokenAnnotation struct {
	I0 Token // eq
	I1 Token // string

}

func ParseStatementTokenAnnotation(in []Token) (NodeStatementTokenAnnotation, int, error) {
	var out NodeStatementTokenAnnotation
	curr := 0

	if len(in) <= curr || in[curr].Type != "eq" {
		return NodeStatementTokenAnnotation{}, 0, newError("failed to parse statement-token-annotation: eq expected", getLineOr0(in, curr))
	}
	out.I0 = in[curr]
	curr++
	
	if len(in) <= curr || in[curr].Type != "string" {
		return NodeStatementTokenAnnotation{}, 0, newError("failed to parse statement-token-annotation: string expected", getLineOr0(in, curr))
	}
	out.I1 = in[curr]
	curr++
	
	return out, curr, nil
}

type NodeStatementEmpty struct {
	I0 Token // newline

}

func ParseStatementEmpty(in []Token) (NodeStatementEmpty, int, error) {
	var out NodeStatementEmpty
	curr := 0

	if len(in) <= curr || in[curr].Type != "newline" {
		return NodeStatementEmpty{}, 0, newError("failed to parse statement-empty: newline expected", getLineOr0(in, curr))
	}
	out.I0 = in[curr]
	curr++
	
	return out, curr, nil
}

type NodeStatement struct {
	I interface{}
}

func ParseStatement(in []Token) (NodeStatement, int, error) {
	if node, curr, err := ParseStatementToken(in); err == nil {
		return NodeStatement{node}, curr, nil
	}
		
	if node, curr, err := ParseStatementExpr(in); err == nil {
		return NodeStatement{node}, curr, nil
	}
		
	if node, curr, err := ParseStatementEmpty(in); err == nil {
		return NodeStatement{node}, curr, nil
	}
		
	return NodeStatement{nil}, 0, newError("failed to parse statement", getLineOr0(in, 0))
}

type NodeStatements struct {
	I0 []NodeStatement

}

func ParseStatements(in []Token) (NodeStatements, int, error) {
	var out NodeStatements
	curr := 0

	for {
		node0, currChange, err := ParseStatement(in[curr:])
		if err != nil {
			break
		}
		out.I0 = append(out.I0, node0)
		curr += currChange
				
	}
	return out, curr, nil
}

