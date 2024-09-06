package kuuhaku_analyzer

import (
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_parser"
)

type StateTransition struct {
	Kernel SymbolGroup		
	Position int
	ExpandedSymbols []SymbolGroup
}

type SymbolGroup struct {
	Title kuuhaku_parser.MatchRule
	Rules []*kuuhaku_parser.Rule
}

type ExpandedSymbol struct {
	Position int
	Title kuuhaku_parser.MatchRule
	Rule *kuuhaku_parser.Rule
}

type ParseTable struct {
	States []ParseTableState
	TerminalSymbols []string 
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
	ReduceRule *kuuhaku_parser.ReplaceRule
	ShiftState *ParseTableState
}

type GotoCell struct {
	LhsRule string
	GotoState *ParseTableState
}
