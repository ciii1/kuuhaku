package kuuhaku_analyzer

import (
	"fmt"

	"github.com/ciii1/kuuhaku/internal/helper"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_parser"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_tokenizer"
)

type AnalyzeErrorType int

const (
	UNDEFINED_VARIABLE = iota
)

type AnalyzeError struct {
	Position kuuhaku_tokenizer.Position	
	Message string
	Type AnalyzeErrorType
}

func (e AnalyzeError) Error() string {
	return fmt.Sprintf("Analyze error (%d, %d): %s", e.Position.Line, e.Position.Column, e.Message)
}

func ErrUndefinedVariable(position kuuhaku_tokenizer.Position, variableName string) *AnalyzeError {
	return &AnalyzeError {
		Message: "Variable " + variableName +  " is undefined",
		Position: position,
		Type: UNDEFINED_VARIABLE,
	}
}

type Analyzer struct {
	input *kuuhaku_parser.Ast
	Errors []error
	stateNumber int
}

func Analyze() {

}

func initAnalyzer(input *kuuhaku_parser.Ast) Analyzer {
	return Analyzer {
		input: input,
		Errors: []error{},
		stateNumber: 0,
	}
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
	return SymbolTitle {}
}

func (analyzer *Analyzer) expandSymbol(rules *[]*kuuhaku_parser.Rule, position int, previousSymbols *[]*Symbol) *[]*Symbol {
	output := previousSymbols
	for _, currRule := range *rules {
		if position >= len(currRule.MatchRules) {
			*output = append(*output, &Symbol{
				Rule: currRule,
				Position: position,
				Title: SymbolTitle {},
			})
			continue
		}

		currMatchRule := currRule.MatchRules[position]

		*output = append(*output, &Symbol{
			Rule: currRule,
			Position: position,
			Title: getSymbolTitleFromMatchRule(currMatchRule),
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
				output = analyzer.expandSymbol(&rules, 0, output)
			}
		}	
	}
	return output
}

func (analyzer *Analyzer) buildParseTable(startSymbolString string) *[]*StateTransition {
	if len(analyzer.Errors) != 0 {
		return nil
	}
	println("---")
	//var states []StateTransition
	startRules := analyzer.input.Rules[startSymbolString]
	expandedStartSymbols := analyzer.expandSymbol(&startRules, 0, &[]*Symbol{})

	var stateTransitions []*StateTransition
	stateTransitions = append(stateTransitions, &StateTransition {
		SymbolGroups: analyzer.groupSymbols(expandedStartSymbols),
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
				symbol.Position += 1
				symbol.Title = SymbolTitle{}
				expandedSymbols := analyzer.expandSymbol(&[]*kuuhaku_parser.Rule{symbol.Rule}, symbol.Position, &[]*Symbol{})
				for _, expandedSymbol := range *expandedSymbols {
					expandedSymbolsAll = append(expandedSymbolsAll, expandedSymbol)
				}
			}
			
			grouped := analyzer.groupSymbols(&expandedSymbolsAll)
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

