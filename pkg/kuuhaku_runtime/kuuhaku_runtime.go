package kuuhaku_runtime

import (
	"fmt"
	"strconv"

	"github.com/ciii1/kuuhaku/pkg/kuuhaku_analyzer"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_parser"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_tokenizer"
)

type ParseStackElementType int

const (
	PARSE_STACK_ELEMENT_TYPE_REDUCED ParseStackElementType = iota
	PARSE_STACK_ELEMENT_TYPE_TERMINAL
)

type ParseStackElement struct {
	Type ParseStackElementType
	String string
	State int
}

type RuntimeErrorType int

const (
	PARSE_STACK_IS_NOT_EMPTY RuntimeErrorType = iota
	REDUCE_RULE_IS_NOT_MATCHING
)

type RuntimeError struct {
	Position kuuhaku_tokenizer.Position	
	Message string
	Type RuntimeErrorType
}
func (e RuntimeError) Error() string {
	return fmt.Sprintf("Runtime error (%d, %d): %s", e.Position.Line, e.Position.Column, e.Message)
}

type RuntimeSyntaxError struct {
	Position kuuhaku_tokenizer.Position	
	Message string
	Expected *[]string
}
func (e RuntimeSyntaxError) Error() string {
	return fmt.Sprintf("Syntax formatting error (%d, %d): %s", e.Position.Line, e.Position.Column, e.Message)
}


func ErrSyntaxError(position kuuhaku_tokenizer.Position, expected *[]string) *RuntimeSyntaxError {
	expectedCombined := ""
	for _, expectedE := range *expected {
		expectedCombined += "\n\t" + expectedE
	}

	return &RuntimeSyntaxError {
		Message: "Syntax is invalid. Expected one of the following:" + expectedCombined,
		Expected: expected,
		Position: position,
	}
}

func ErrExpectedEOFError(position kuuhaku_tokenizer.Position) *RuntimeSyntaxError {
	return &RuntimeSyntaxError {
		Message: "Syntax is invalid. Expected EOF.",
		Position: position,
		Expected: &[]string{"<end>"}, //TODO: make a struct for the expected[]
	}
}

func ErrParseStackIsNotEmpty(position kuuhaku_tokenizer.Position) *RuntimeError {
	return &RuntimeError {
		Message: "Parse stack is not empty at the end of parsing.",
		Position: position,
		Type:PARSE_STACK_IS_NOT_EMPTY,
	}
}

func ErrReduceRuleIsNotMatching(position kuuhaku_tokenizer.Position) *RuntimeError {
	return &RuntimeError {
		Message: "The reduce rule on the parse table doesn't match the rule on the parse stack",
		Position: position,
		Type:REDUCE_RULE_IS_NOT_MATCHING,
	}
}

func Format(input string, format *kuuhaku_analyzer.AnalyzerResult) (string, error) {
	var currPos kuuhaku_tokenizer.Position
	out := ""
	for currPos.Raw < len(input) {
		if input[currPos.Raw] == '\n' {
			currPos.Line++
			currPos.Column = 1
		} else {
			currPos.Column++
		}

		isThereSuccess := false
		for _, parseTable := range format.ParseTables {
			res, resPos, err := runParseTable(input, currPos, &parseTable)
			if err == nil {
				isThereSuccess = true
				currPos = resPos
				out += res
				break
			} else {
				if !format.IsSearchMode {
					return "", err
				}
			}
		}
		if !isThereSuccess {
			currPos.Raw++
		}
		if !format.IsSearchMode && currPos.Raw < len(input)-1 {
			//return error here: expected eof
		}
	}
	return out, nil
}

func addToPositionFromSlicedString(prevPos kuuhaku_tokenizer.Position, sliced string) kuuhaku_tokenizer.Position {
	raw := prevPos.Raw + len(sliced)	
	
	col := prevPos.Column
	i := len(sliced) - 1
	for i >= 0 && sliced[i] != '\n' {
		col++
		i--
	}
	col++

	line := prevPos.Line
	for _, char := range sliced {
		if char == '\n' {
			line++
		}
	}

	return kuuhaku_tokenizer.Position {
		Raw: raw,
		Column: col,
		Line: line,
	}
}

func runParseTable(input string, pos kuuhaku_tokenizer.Position, parseTable *kuuhaku_analyzer.ParseTable) (string, kuuhaku_tokenizer.Position, error) {
	var parseStack []ParseStackElement
	currState := 0
	for pos.Raw < len(input){
		currRow := parseTable.States[currState]
		slicedInput := input[pos.Raw:]
		lookahead := ""
		lookaheadFound := false
		var expected []string
		for _, terminal := range parseTable.Terminals {
			if currRow.ActionTable[terminal.Terminal] != nil && terminal.Regexp != nil {
				expected = append(expected, terminal.Terminal)
				loc := terminal.Regexp.FindStringIndex(slicedInput)
				if loc == nil {
					continue
				} else {
					lookahead = slicedInput[0:loc[1]]
					pos = addToPositionFromSlicedString(pos, lookahead)
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
				currState, err = applyRule(parseTable, currActionCell.ReduceRule, &parseStack)	
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
				break
			} else {
				return "", pos, ErrSyntaxError(pos, &expected)
			}
		 }
	}
	if len(parseStack) != 1 {
		return "", pos, ErrParseStackIsNotEmpty(pos)
	}
	return parseStack[0].String, pos, nil
}

func printParseStack(parseStack *[]ParseStackElement) {
	print("[")
	for i, parseStackElement := range *parseStack {
		if i != 0 {
			print(",")
		}
		print(parseStackElement.String)
	}
	println("]")
}

func applyRule(parseTable *kuuhaku_analyzer.ParseTable, rule *kuuhaku_parser.Rule, parseStack *[]ParseStackElement) (int, error) {
	nextState := 0
	lhs := rule.Name
	ruleLength := len(rule.MatchRules)

	targetStack := (*parseStack)[len(*parseStack)-ruleLength:]
	if len(*parseStack) == 1 {
		*parseStack = []ParseStackElement{}
	} else {
		*parseStack = (*parseStack)[:len(*parseStack)-(ruleLength+1)]
	}

	reducedString := ""	
	
	if len(rule.ReplaceRules) == 0 {
		for _, targetElement := range targetStack {
			reducedString += targetElement.String
		}
	} else {
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
	}

	if len(*parseStack)-1 < 0 {
		nextState = 0
	} else {
		nextState = parseTable.States[(*parseStack)[len(*parseStack)-1].State].GotoTable[lhs].GotoState
	}

	*parseStack = append(*parseStack, ParseStackElement {
		Type: PARSE_STACK_ELEMENT_TYPE_REDUCED,
		String: reducedString,
		State: nextState,
	})
	return nextState, nil
}
