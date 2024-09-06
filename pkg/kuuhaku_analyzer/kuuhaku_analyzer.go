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

func (analyzer *Analyzer) expandSymbol(symbol string, position int, previousSymbols *[]*kuuhaku_parser.Rule, first bool) *[]*kuuhaku_parser.Rule {
	output := previousSymbols
	for _, currRule := range analyzer.input.Rules[symbol] {
		if !first {
			*output = append(*output, currRule)
		}
		currMatchRule := currRule.MatchRules[position]
		currIdentifier, ok := currMatchRule.(kuuhaku_parser.Identifer);
		if ok {
			is_included := false
			for _, e := range *output {
				if e.MatchRules[0] == currIdentifier {
					is_included = true
					break
				}
			}
			if !is_included {
				output = analyzer.expandSymbol(symbol, 0, output, true)
			}
		}	
	}
	return output
}

func (analyzer *Analyzer) buildParseTable(startSymbol string) {
	if len(analyzer.Errors) != 0 {
		return
	}
	//var states []StateTransition
	startRules := analyzer.input.Rules[startSymbol]
	analyzer.groupRule(&startRules, 0) //TODO: make all of the rules in ast be a rule pointer
}

func (analyzer *Analyzer) groupRule(rules *[]*kuuhaku_parser.Rule, position int) {

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

