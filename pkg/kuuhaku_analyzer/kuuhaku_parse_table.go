package kuuhaku_analyzer

import (
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_parser"
)

type StateTransition struct {
	SymbolGroups *[]*SymbolGroup
}

type SymbolGroup struct {
	Title SymbolTitle
	Symbols *[]*Symbol
}

type Symbol struct {
	Position int
	Title SymbolTitle
	Rule *kuuhaku_parser.Rule
}

type SymbolTitleType int
const (
	REGEX_LITERAL_TITLE = iota
	IDENTIFIER_TITLE
	EMPTY_TITLE
)

type SymbolTitle struct {
	String string
	Type SymbolTitleType
}

type ParseTable struct {
	States []ParseTableState
	Terminals []string
	Lhss []string
}

type ParseTableState struct {
	ActionTable map[string]ActionCell //map[kuuhaku_parser.RegexLiteral.Content]ActionCell
	GotoTable map[string]GotoCell //map[kuuhaku_parser.Rule.Name]GotoCell
}

type Action int

const (
	REDUCE = iota
	SHIFT
)

type ActionCell struct {
	LookaheadTerminal string
	Action Action
	ReduceRule *kuuhaku_parser.Rule
	ShiftState int
}

type GotoCell struct {
	Lhs string
	GotoState int
}
