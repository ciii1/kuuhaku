package kuuhaku_parser

import (
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_tokenizer"
)

type Head struct {
	Rules map[string]Rule //name - Rule pair
	Position kuuhaku_tokenizer.Position
}

type Rule struct {
	Name string
	Match []MatchRule
	Replace []ReplaceRule
	Position kuuhaku_tokenizer.Position
}

type MatchRule interface {
	matchRule()	
}

type ReplaceRule interface {
	replaceRule()
}

type StringStmt interface {
	ReplaceRule
	stringStmt()	
}

type Identifer struct {
	Name string	
	Position kuuhaku_tokenizer.Position
}
func (i Identifer) matchRule() {}

type RegexLiteral struct {
	RegexString string	
	Position kuuhaku_tokenizer.Position
}
func (r RegexLiteral) matchRule() {}

type CaptureGroup struct {
	Number int
	Position kuuhaku_tokenizer.Position
}
func (c CaptureGroup) replaceRule() {}
func (c CaptureGroup) stringStmt() {}

type StringLiteral struct {
	String string	
	Position kuuhaku_tokenizer.Position
}
func (s StringLiteral) replaceRule() {}
func (s StringLiteral) stringStmt() {}

type Len struct {
	FirstArgument StringStmt
	SecondArgument StringStmt	
	Position kuuhaku_tokenizer.Position
}
func (l Len) replaceRule() {}
