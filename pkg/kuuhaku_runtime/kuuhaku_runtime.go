package kuuhaku_runtime

import (
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_analyzer"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_parser"
)

const (
	PARSE_STACK_ELEMENT_TYPE_REDUCED = iota
	PARSE_STACK_ELEMENT_TYPE_TERMINAL
)

type ParseStackElementType int

type ParseStackElement struct {
	Type ParseStackElementType
	String string
	State int
}

func Format(input string, format *kuuhaku_analyzer.AnalyzerResult) (string, error) {
	currPos := 0
	out := ""
	for currPos < len(input) {
		isThereSuccess := false
		for _, parseTable := range format.ParseTables {
			res, resPos, err := runParseTable(input, currPos, &parseTable)
			if err == nil {
				isThereSuccess = true
				currPos = resPos
				out += res
				break
			} else {
				if format.IsSearchMode {
					return "", err
				}
			}
		}
		if !isThereSuccess {
			currPos++
		}
		if !format.IsSearchMode && currPos < len(input){
			//return error here: expected eof
		}
	}
	return out, nil
}

func runParseTable(input string, pos int, parseTable *kuuhaku_analyzer.ParseTable) (string, int, error) {
	var parseStack []ParseStackElement
	slicedInput := input[pos:]
	currState := 0
	for true {
		currRow := parseTable.States[currState]
		lookahead := ""
		lookaheadFound := false
		for _, terminal := range parseTable.Terminals {
			if currRow.ActionTable[terminal.Terminal] != nil && terminal.Regexp != nil {
				loc := terminal.Regexp.FindStringIndex(slicedInput)
				if loc == nil {
					continue
				} else {
					pos += loc[1] + 1
					lookahead = terminal.Terminal
					lookaheadFound = true
					break
				}
			}
		}
		if lookaheadFound {
			currActionCell := currRow.ActionTable[lookahead]
			if currActionCell.Action == kuuhaku_analyzer.SHIFT {
				parseStack = append(parseStack, ParseStackElement{
					Type: PARSE_STACK_ELEMENT_TYPE_TERMINAL,
					String: lookahead,
					State: currState,
				})
				currState = currActionCell.ShiftState
			} else if currActionCell.Action == kuuhaku_analyzer.REDUCE {
				var err error
				currState, err = applyRule(parseTable , currActionCell.ReduceRule, &parseStack)	
				if err != nil {
					return "", pos, err
				}
			}
		 } else {
			if currRow.EndReduceRule != nil {
				var err error
				currState, err = applyRule(parseTable , currRow.EndReduceRule.ReduceRule, &parseStack)	
				if err != nil {
					return "", pos, err
				}
			} else {
				//report syntax error
			}
		 }
	}
	if len(parseStack) != 1 {
		//TODO: add unknown error here
		return "", pos, nil
	}
	return parseStack[0].String, pos, nil
}

func applyRule(parseTable *kuuhaku_analyzer.ParseTable, rule *kuuhaku_parser.Rule, parseStack *[]ParseStackElement) (int, error) {
	nextState := 0
	lhs := rule.Name
	ruleLength := len(rule.MatchRules)
	*parseStack = (*parseStack)[:len(*parseStack)-ruleLength]
	targetStack := (*parseStack)[len(*parseStack)-ruleLength-1:]

	reducedString := ""	
	for _, replaceRule := range rule.ReplaceRules {
		captureGroup, okCaptureGroup := replaceRule.(kuuhaku_parser.CaptureGroup)
		stringLit, okStringLit := replaceRule.(kuuhaku_parser.StringLiteral)
		lenFunc, okLenFunc := replaceRule.(kuuhaku_parser.Len)
		if okCaptureGroup {
			reducedString += targetStack[captureGroup.Number].String
		} else if okStringLit {
			reducedString += stringLit.String
		} else if okLenFunc {
			i := 0
			secondArgumentLength := 0
			captureGroupSecArg, okCaptureGroupSecArg := lenFunc.SecondArgument.(kuuhaku_parser.CaptureGroup)
			stringLitSecArg, okStringLitSecArg := lenFunc.SecondArgument.(kuuhaku_parser.StringLiteral)
			if okCaptureGroupSecArg {
				secondArgumentLength = len(targetStack[captureGroupSecArg.Number].String)
			} else if okStringLitSecArg {
				secondArgumentLength = len(stringLitSecArg.String)
			}
			for i < secondArgumentLength {
				captureGroupFirstArg, okCaptureGroupFirstArg := lenFunc.FirstArgument.(kuuhaku_parser.CaptureGroup)
				stringLitFirstArg, okStringLitFirstArg := lenFunc.FirstArgument.(kuuhaku_parser.StringLiteral)
				if okCaptureGroupFirstArg {
					reducedString += targetStack[captureGroupFirstArg.Number].String
				} else if okStringLitFirstArg {
					secondArgumentLength = len(stringLitFirstArg.String)
				}
				i++
			}
		}
	}

	nextState = parseTable.States[(*parseStack)[len(*parseStack)-1].State].GotoTable[lhs].GotoState
	*parseStack = append(*parseStack, ParseStackElement {
		Type: PARSE_STACK_ELEMENT_TYPE_REDUCED,
		String: reducedString,
		State: nextState,
	})
	return nextState, nil
}
