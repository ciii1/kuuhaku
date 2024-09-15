package kuuhaku_analyzer

import (
	"fmt"
	"strconv"

	"github.com/ciii1/kuuhaku/internal/helper"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_parser"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_tokenizer"
)

type AnalyzeErrorType int

const (
	UNDEFINED_VARIABLE = iota
	CONFLICT
)

type AnalyzeError struct {
	Position kuuhaku_tokenizer.Position	
	Message string
	Type AnalyzeErrorType
}

type ConflictError struct {
	Position1 kuuhaku_tokenizer.Position	
	Position2 kuuhaku_tokenizer.Position	
	Symbol1 *Symbol
	Symbol2 *Symbol
	Message string
}

func (e AnalyzeError) Error() string {
	return fmt.Sprintf("Analyze error (%d, %d): %s", e.Position.Line, e.Position.Column, e.Message)
}

func (e ConflictError) Error() string {
	return fmt.Sprintf("Analyze error (%d, %d): %s", e.Position1.Line, e.Position1.Column, e.Message)
}

func ErrUndefinedVariable(position kuuhaku_tokenizer.Position, variableName string) *AnalyzeError {
	return &AnalyzeError {
		Message: "Variable " + variableName +  " is undefined",
		Position: position,
		Type: UNDEFINED_VARIABLE,
	}
}

func ErrConflict(symbol1 *Symbol, symbol2 *Symbol) *ConflictError {
	var position1 kuuhaku_tokenizer.Position
	if symbol1.Position < len(symbol1.Rule.MatchRules) {
		position1 = symbol1.Rule.MatchRules[symbol1.Position].GetPosition()
	} else {
		position1 = symbol1.Rule.MatchRules[len(symbol1.Rule.MatchRules) - 1].GetPosition()
	}

	var position2 kuuhaku_tokenizer.Position
	if symbol2.Position < len(symbol2.Rule.MatchRules) {
		position2 = symbol2.Rule.MatchRules[symbol2.Position].GetPosition()
	} else {
		position2 = symbol2.Rule.MatchRules[len(symbol2.Rule.MatchRules) - 1].GetPosition()
	}

	return &ConflictError {
		Message: "Detected conflict at rule " + strconv.Itoa(symbol1.Rule.Order) + " and rule " + strconv.Itoa(symbol2.Rule.Order) + " with position (" + strconv.Itoa(position2.Line) + ", " + strconv.Itoa(position2.Column) + ")\n--- Lookaheads are: " + symbol1.Lookeahead.String + " and " + symbol2.Lookeahead.String,
		Position1: position1,
		Position2: position2,
		Symbol1: symbol1,
		Symbol2: symbol2,
	}
}

type Analyzer struct {
	input *kuuhaku_parser.Ast
	Errors []error
	stateNumber int
	parseTable ParseTable
	stateTransitionMap map[Symbol]int
	stateTransitionMapBool map[Symbol]bool
}

func Analyze() {

}

func initAnalyzer(input *kuuhaku_parser.Ast) Analyzer {
	terminalsMap := make(map[string]bool)
	for _, rules := range input.Rules {
		for _, rule := range rules {
			for _, matchRule := range (*rule).MatchRules {
				regexCurr, ok := matchRule.(kuuhaku_parser.RegexLiteral)
				if ok {
					terminalsMap[regexCurr.RegexString] = true
				}
			}
		}
	}

	var terminals []string
	for regexString := range terminalsMap {
		terminals = append(terminals, regexString)	
	}

	var lhss []string
	for lhs := range input.Rules {
		lhss = append(lhss, lhs)
	}

	return Analyzer {
		input: input,
		Errors: []error{},
		stateNumber: 1,
		parseTable: ParseTable {
			States: []ParseTableState{},
			Terminals: terminals,
			Lhss: lhss,
		},
		stateTransitionMap: make(map[Symbol]int),
		stateTransitionMapBool: make(map[Symbol]bool),
	}
}

func makeEndSymbolTitle () SymbolTitle {
	return SymbolTitle{String: "<end>", Type:EMPTY_TITLE}	
}

func getSymbolTitleFromMatchRule(matchRule kuuhaku_parser.MatchRule) SymbolTitle {
	currIdentifier, ok := matchRule.(kuuhaku_parser.Identifer)
	if ok {
		return SymbolTitle {
			String: currIdentifier.Name,
			Type: IDENTIFIER_TITLE,
		}
	} else {
		currRegexLit, ok := matchRule.(kuuhaku_parser.RegexLiteral);
		if ok {
			return SymbolTitle {
				String: currRegexLit.RegexString,
				Type: REGEX_LITERAL_TITLE,
			}
		}
	}
	return SymbolTitle {Type:EMPTY_TITLE}
}

func (analyzer *Analyzer) expandSymbol(rules *[]*kuuhaku_parser.Rule, position int, previousSymbols *[]*Symbol, lookahead SymbolTitle) *[]*Symbol {
	output := previousSymbols
	for _, currRule := range *rules {
		if position >= len(currRule.MatchRules) {
			*output = append(*output, &Symbol{
				Rule: currRule,
				Position: position,
				Title: makeEndSymbolTitle(),
				Lookeahead: lookahead,
			})
			continue
		}

		currMatchRule := currRule.MatchRules[position]
		nextLookahead := makeEndSymbolTitle()
		if position + 1 < len(currRule.MatchRules) {
			nextMatchRule := currRule.MatchRules[position+1]
			nextLookahead = getSymbolTitleFromMatchRule(nextMatchRule)
		}

		*output = append(*output, &Symbol{
			Rule: currRule,
			Position: position,
			Title: getSymbolTitleFromMatchRule(currMatchRule),
			Lookeahead: lookahead,
		})

		currIdentifier, ok := currMatchRule.(kuuhaku_parser.Identifer);
		if ok {
			is_included := false
			for _, e := range *output {
				if e.Rule.Name == currIdentifier.Name {
					is_included = true
					break
				}
			}
			if !is_included {
				rules := analyzer.input.Rules[currIdentifier.Name]
				output = analyzer.expandSymbol(&rules, 0, output, nextLookahead)
			}
		}	
	}
	return output
}

func (analyzer *Analyzer) buildParseTable(startSymbolString string) *[]*StateTransition {
	if len(analyzer.Errors) != 0 {
		return nil
	}
	startRules := analyzer.input.Rules[startSymbolString]
	expandedStartSymbols := analyzer.expandSymbol(&startRules, 0, &[]*Symbol{}, makeEndSymbolTitle())

	var stateTransitions []*StateTransition
	grouped := analyzer.groupSymbols(expandedStartSymbols)
	grouped = analyzer.buildParseTableState(grouped)	
	stateTransitions = append(stateTransitions, &StateTransition {
		SymbolGroups: grouped,
	})

	i := 0
	for true {
		if i >= len(stateTransitions) {
			break
		}
		state := stateTransitions[i] 
		for _, group := range *state.SymbolGroups {
			var expandedSymbolsAll []*Symbol
			for _, symbol := range *group.Symbols {
				expandedSymbols := analyzer.expandSymbol(&[]*kuuhaku_parser.Rule{symbol.Rule}, symbol.Position + 1, &[]*Symbol{}, symbol.Lookeahead)
				for _, expandedSymbol := range *expandedSymbols {
					expandedSymbolsAll = append(expandedSymbolsAll, expandedSymbol)
				}
			}
			
			grouped := analyzer.groupSymbols(&expandedSymbolsAll)
			grouped = analyzer.buildParseTableState(grouped)	
			if len(*grouped) != 0 {
				stateTransitions = append(stateTransitions, &StateTransition {
					SymbolGroups: grouped,
				})
			}
		}
		i++
	}
	return &stateTransitions
}

func (analyzer *Analyzer) groupSymbols(symbols *[]*Symbol) *[]*SymbolGroup {
	groupsMap := make(map[SymbolTitle]SymbolGroup)

	for _, symbol := range *symbols {
		if symbol.Position <= len(symbol.Rule.MatchRules) {
			symbolTitle := (*symbol).Title
			if groupsMap[symbolTitle].Symbols == nil {
				groupsMap[symbol.Title] = SymbolGroup {
					Title: symbolTitle,
					Symbols: &[]*Symbol{},
				}
			}
			*groupsMap[symbolTitle].Symbols = append(*groupsMap[symbolTitle].Symbols, symbol)
		}
	}

	var groups []*SymbolGroup
	for _, group := range groupsMap {
		groupVar := group
		groups = append(groups, &groupVar)	
	}
	return &groups
}

func (analyzer *Analyzer) buildParseTableState(symbolGroups *[]*SymbolGroup) *[]*SymbolGroup {
	if len(*symbolGroups) == 0 {
		return symbolGroups
	}
	actionTable := make(map[string]ActionCell)
	gotoTable := make(map[string]GotoCell)
	var endReduceRule *ActionCell
	endReduceRule = nil

	var outGroup []*SymbolGroup
	
	isThereEndReduce := false
	var endReducedSymbol *Symbol

	usedTerminalsWithSymbol := make(map[string]*Symbol)	

	var emptyTitleGroup *SymbolGroup
	emptyTitleGroup = nil
	for _, group := range *symbolGroups {
		if group.Title.Type == EMPTY_TITLE {
			emptyTitleGroup = group
		}
	}

	if emptyTitleGroup != nil {
		//resolve end reduce actions
		outGroup = append(outGroup, emptyTitleGroup)
		for _, symbol := range *emptyTitleGroup.Symbols {
			if symbol.Position >= len(symbol.Rule.MatchRules) {
				if symbol.Lookeahead.Type == EMPTY_TITLE {
					if isThereEndReduce {
						analyzer.Errors = append(analyzer.Errors, ErrConflict(symbol, endReducedSymbol))
					} else {
						isThereEndReduce = true
						endReducedSymbol = symbol
						endReduceRule = &ActionCell {
							LookaheadTerminal: "",
							Action: REDUCE,
							ReduceRule: symbol.Rule,
							ShiftState: 0,
						}
					}
				}
			}
		}
		//resolve reduce actions
		for _, symbol := range *emptyTitleGroup.Symbols {
			if symbol.Position >= len(symbol.Rule.MatchRules) {
				if symbol.Lookeahead.Type == EMPTY_TITLE {
					continue
				}
				var terminals []string
				if symbol.Lookeahead.Type == IDENTIFIER_TITLE {
					rule := analyzer.input.Rules[symbol.Lookeahead.String]
					symbols := analyzer.expandSymbol(&rule, 0, &[]*Symbol{}, SymbolTitle{})
					for _, symbol := range *symbols {
						if (*symbol).Title.Type == REGEX_LITERAL_TITLE {
							terminals = append(terminals, (*symbol).Title.String)
						}
					}
				} else if symbol.Lookeahead.Type == REGEX_LITERAL_TITLE {
					terminals = append(terminals, symbol.Lookeahead.String)
				}
				for _, terminal := range terminals {
					if usedTerminalsWithSymbol[terminal] == nil {
						usedTerminalsWithSymbol[terminal] = symbol
						actionTable[terminal] = ActionCell {
							LookaheadTerminal: terminal,
							Action: REDUCE,
							ReduceRule: symbol.Rule,
							ShiftState: 0,
						}
					} else {
						analyzer.Errors = append(analyzer.Errors, ErrConflict(symbol, usedTerminalsWithSymbol[terminal]))
					}
				}
			}
		}
	}

	for _, group := range *symbolGroups {
		if group.Title.Type == IDENTIFIER_TITLE {
			gotoTable[group.Title.String] = GotoCell {
				Lhs: group.Title.String,
				GotoState: analyzer.stateNumber,
			}
			analyzer.stateNumber++
			outGroup = append(outGroup, group)
		} else if group.Title.Type == REGEX_LITERAL_TITLE {
			if usedTerminalsWithSymbol[group.Title.String] != nil{
				analyzer.Errors = append(analyzer.Errors, ErrConflict((*group.Symbols)[0], usedTerminalsWithSymbol[group.Title.String]))
			}
			isStateExisted := false
			existedStateNumber := 0
			var existedSymbol *Symbol
			for _, symbol := range *group.Symbols {
				if isStateExisted {
					if analyzer.stateTransitionMap[*symbol] != existedStateNumber {
						analyzer.Errors = append(analyzer.Errors, ErrConflict(symbol, existedSymbol))
					}
				} else if analyzer.stateTransitionMapBool[*symbol] != false {
					isStateExisted = true
					existedStateNumber = analyzer.stateTransitionMap[*symbol]
					existedSymbol = symbol
				}
			}
			if !isStateExisted {
				actionTable[group.Title.String] = ActionCell {
					LookaheadTerminal: group.Title.String,
					Action: SHIFT,
					ReduceRule: nil,
					ShiftState: analyzer.stateNumber,
				}
				outGroup = append(outGroup, group)
				for _, symbol := range *group.Symbols {
					analyzer.stateTransitionMapBool[*symbol] = true
					analyzer.stateTransitionMap[*symbol] = analyzer.stateNumber
				}
				analyzer.stateNumber++
			} else {
				actionTable[group.Title.String] = ActionCell {
					LookaheadTerminal: group.Title.String,
					Action: SHIFT,
					ReduceRule: nil,
					ShiftState: existedStateNumber,
				}
			}
		}
	}
	analyzer.parseTable.States = append(analyzer.parseTable.States, ParseTableState{
		ActionTable: actionTable,
		GotoTable: gotoTable,
		EndReduceRule: endReduceRule,
	})
	return &outGroup
}

func (analyzer *Analyzer) makeAugmentedGrammar(startSymbol string) *Symbol {
	startRules := analyzer.input.Rules[startSymbol]
	order := startRules[0].Order
	ruleName := "S" + startSymbol 
	rule := &kuuhaku_parser.Rule {
		Name: ruleName,
		Order: order,
		MatchRules: []kuuhaku_parser.MatchRule{
			kuuhaku_parser.Identifer{
				Name: startSymbol,
				Position: startRules[0].Position,
			},
		},
		Position: startRules[0].Position,
	}
	analyzer.input.Rules[ruleName] = append(analyzer.input.Rules[ruleName], rule)
	return &Symbol {
		Title: getSymbolTitleFromMatchRule(rule.MatchRules[0]),
		Position: 0,
		Rule: rule,
	}
}

//return start symbols
func (analyzer *Analyzer) analyzeStart () []string {
	startSymbols := make([]string, len(analyzer.input.Rules))
	i := 0
	for key := range analyzer.input.Rules {
   		startSymbols[i] = key
		i++
	}
	
	for ruleName, ruleArray := range analyzer.input.Rules {
		for _, rule := range ruleArray {
			for _, matchRule := range rule.MatchRules {
				identifier, ok := matchRule.(kuuhaku_parser.Identifer)
				if !ok {
					continue	
				}
				if len(analyzer.input.Rules[identifier.Name]) == 0 {
					analyzer.Errors = append(analyzer.Errors, ErrUndefinedVariable(identifier.Position, identifier.Name))
				}

				if identifier.Name != ruleName {
					helper.EmptyStringByValue(&startSymbols, identifier.Name)	
				}
			}
		}
	}
	
	var outputSymbols []string

	for _, startSymbol := range startSymbols {
		if startSymbol != "" {
			outputSymbols = append(outputSymbols, startSymbol)
		}
	}

	return outputSymbols
}

