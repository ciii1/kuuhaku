package kuuhaku_parser

import (
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_tokenizer"
)

type Ast struct {
	Rules        map[string][]*Rule //name - []Rule pair
	Position     kuuhaku_tokenizer.Position
	IsSearchMode bool
}

type Rule struct {
	Name         string
	Order        int
	MatchRules   []MatchRule
	ReplaceRules []ReplaceRule
	Position     kuuhaku_tokenizer.Position
}

type MatchRule interface {
	matchRule()
	GetPosition() kuuhaku_tokenizer.Position
}

type ReplaceRule interface {
	replaceRule()
}

type StringStmt interface {
	ReplaceRule
	stringStmt()
}

type Identifer struct {
	Name     string
	Position kuuhaku_tokenizer.Position
}

func (i Identifer) matchRule() {}
func (i Identifer) GetPosition() kuuhaku_tokenizer.Position {
	return i.Position
}

type RegexLiteral struct {
	RegexString string
	Position    kuuhaku_tokenizer.Position
}

func (r RegexLiteral) matchRule() {}
func (r RegexLiteral) GetPosition() kuuhaku_tokenizer.Position {
	return r.Position
}

type CaptureGroup struct {
	Number   int
	Position kuuhaku_tokenizer.Position
}

func (c CaptureGroup) replaceRule() {}
func (c CaptureGroup) stringStmt()  {}

type StringLiteral struct {
	String   string
	Position kuuhaku_tokenizer.Position
}

func (s StringLiteral) replaceRule() {}
func (s StringLiteral) stringStmt()  {}

type Len struct {
	FirstArgument  StringStmt
	SecondArgument StringStmt
	Position       kuuhaku_tokenizer.Position
}

func (l Len) replaceRule() {}
