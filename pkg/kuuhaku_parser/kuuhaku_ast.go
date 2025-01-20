package kuuhaku_parser

import (
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_tokenizer"
)

type Ast struct {
	Rules        map[string][]*Rule //name - []Rule pair
	Position     kuuhaku_tokenizer.Position
	GlobalLua	 *LuaLiteral
	IsSearchMode bool
}

type Rule struct {
	Name        string
	Order       int
	MatchRules  []MatchRule
	ReplaceRule *LuaLiteral
	Position    kuuhaku_tokenizer.Position
	ArgList     []Identifier
}

type MatchRule interface {
	matchRule()
	GetPosition() kuuhaku_tokenizer.Position
}

type Identifier struct {
	Name     string
	ArgList  []LuaLiteral
	Position kuuhaku_tokenizer.Position
}

func (i Identifier) matchRule() {}
func (i Identifier) GetPosition() kuuhaku_tokenizer.Position {
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

type LuaLiteral struct {
	LuaString string
	Position  kuuhaku_tokenizer.Position
	Type LuaLiteralType
}

const (
	LUA_LITERAL_TYPE_MULTI_STMT = iota
	LUA_LITERAL_TYPE_RETURN
)

type LuaLiteralType int
