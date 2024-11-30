package kuuhaku_runtime

import (
	"fmt"

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
	Type   ParseStackElementType
	String string
	State  int
}

type RuntimeErrorType int

const (
	PARSE_STACK_IS_NOT_EMPTY RuntimeErrorType = iota
	REDUCE_RULE_IS_NOT_MATCHING
)

type RuntimeError struct {
	Position kuuhaku_tokenizer.Position
	Message  string
	Type     RuntimeErrorType
}

func (e RuntimeError) Error() string {
	return fmt.Sprintf("Runtime error (%d, %d): %s", e.Position.Line, e.Position.Column, e.Message)
}

type RuntimeSyntaxError struct {
	Position kuuhaku_tokenizer.Position
	Message  string
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

	return &RuntimeSyntaxError{
		Message:  "Syntax is invalid. Expected one of the following:" + expectedCombined,
		Expected: expected,
		Position: position,
	}
}

func ErrExpectedEOFError(position kuuhaku_tokenizer.Position) *RuntimeSyntaxError {
	return &RuntimeSyntaxError{
		Message:  "Syntax is invalid. Expected EOF.",
		Position: position,
		Expected: &[]string{"<end>"}, //TODO: make a struct for the expected[]
	}
}

func ErrParseStackIsNotEmpty(position kuuhaku_tokenizer.Position) *RuntimeError {
	return &RuntimeError{
		Message:  "Parse stack is not empty at the end of parsing.",
		Position: position,
		Type:     PARSE_STACK_IS_NOT_EMPTY,
	}
}

func ErrReduceRuleIsNotMatching(position kuuhaku_tokenizer.Position) *RuntimeError {
	return &RuntimeError{
		Message:  "The reduce rule on the parse table doesn't match the rule on the parse stack",
		Position: position,
		Type:     REDUCE_RULE_IS_NOT_MATCHING,
	}
}

func Format(input string, format *kuuhaku_analyzer.AnalyzerResult) (string, error) {
	var currPos kuuhaku_tokenizer.Position
	currPos.Line = 1
	out := ""
	for currPos.Raw < len(input) {
		if input[currPos.Raw] == '\n' {
			currPos.Line++
			currPos.Column = 1
		} else {
			currPos.Column++
		}

		isThereSuccess := false
		//TODO: Investigate. I'm pretty sure we can have only one parse table even if we have multiple start symbols
		//For example, by adding one start symbol as the "unifier"
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
			out += string(input[currPos.Raw])
			currPos.Raw++
		}
		if !format.IsSearchMode && currPos.Raw < len(input)-1 {
			return out, ErrExpectedEOFError(currPos)
		}
	}
	return out, nil
}

func addToPositionFromSlicedString(prevPos kuuhaku_tokenizer.Position, sliced string) kuuhaku_tokenizer.Position {
	raw := prevPos.Raw + len(sliced)

	col := prevPos.Column
	colIfContainsNewLine := 1

	i := len(sliced) - 1
	for i >= 0 {
		if sliced[i] != '\n' {
			break
		}
		col++
		colIfContainsNewLine++
		i--
	}

	//i is always -1 if a \n wasn't found. This is because the condition i >= 0 will not be satisfied untill
	//i == -1, while sliced[i] != '\n' will only be satisfied if i > -1

	if i >= 0 {
		col = colIfContainsNewLine
	}

	line := prevPos.Line
	for _, char := range sliced {
		if char == '\n' {
			line++
		}
	}

	return kuuhaku_tokenizer.Position{
		Raw:    raw,
		Column: col,
		Line:   line,
	}
}

func runParseTable(input string, pos kuuhaku_tokenizer.Position, parseTable *kuuhaku_analyzer.ParseTable) (string, kuuhaku_tokenizer.Position, error) {
	var parseStack []ParseStackElement
	currState := 0
	lookahead := ""
	lookaheadRegex := ""
	lookaheadFound := false

	for true {
		currRow := parseTable.States[currState]
		var expected []string

		if pos.Raw > len(input) {
			for _, terminal := range parseTable.Terminals {
				if currRow.ActionTable[terminal.Terminal] != nil && terminal.Regexp != nil {
					expected = append(expected, terminal.Terminal)
				}
			}
			//TODO: might return all of the strings inside the parse stack combined on error in the future
			return "", pos, ErrSyntaxError(pos, &expected)
		}

		slicedInput := input[pos.Raw:]

		if !lookaheadFound {
			for _, terminal := range parseTable.Terminals {
				if currRow.ActionTable[terminal.Terminal] != nil && terminal.Regexp != nil {
					expected = append(expected, terminal.Terminal)
					loc := terminal.Regexp.FindStringIndex(slicedInput)
					if loc == nil {
						continue
					} else {
						lookahead = slicedInput[0:loc[1]]
						lookaheadRegex = terminal.Terminal
						pos = addToPositionFromSlicedString(pos, lookahead)
						lookaheadFound = true
						break
					}
				}
			}
		}
		if lookaheadFound {
			currActionCell := currRow.ActionTable[lookaheadRegex]
			if currActionCell.Action == kuuhaku_analyzer.SHIFT {
				parseStack = append(parseStack, ParseStackElement{
					Type:   PARSE_STACK_ELEMENT_TYPE_TERMINAL,
					String: lookahead,
					State:  currState,
				})
				currState = currActionCell.ShiftState
				lookaheadFound = false
			} else if currActionCell.Action == kuuhaku_analyzer.REDUCE {
				var err error
				//println(lookahead)
				currState, err = applyRule(parseTable, currActionCell.ReduceRule, &parseStack, pos, false)
				if err != nil {
					return "", pos, err
				}
			}
		} else {
			if currRow.EndReduceRule != nil {
				var err error
				if currRow.EndReduceRule.Action == kuuhaku_analyzer.ACCEPT {
					currState, err = applyRule(parseTable, currRow.EndReduceRule.ReduceRule, &parseStack, pos, true)
					break
				} else if currRow.EndReduceRule.Action == kuuhaku_analyzer.REDUCE {
					currState, err = applyRule(parseTable, currRow.EndReduceRule.ReduceRule, &parseStack, pos, false)
				}
				if err != nil {
					return "", pos, err
				}
			} else {
				//printParseStack(&parseStack)
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
	print("Parse stack: [")
	for i, parseStackElement := range *parseStack {
		if i != 0 {
			print(",")
		}
		print(parseStackElement.String)
	}
	println("]")
}

func applyRule(parseTable *kuuhaku_analyzer.ParseTable, rule *kuuhaku_parser.Rule, parseStack *[]ParseStackElement, pos kuuhaku_tokenizer.Position, isAccept bool) (int, error) {
	nextState := 0
	lhs := rule.Name
	ruleLength := len(rule.MatchRules)

	if len(*parseStack)-ruleLength < 0 {
		return 0, ErrReduceRuleIsNotMatching(pos)
	}
	targetStack := (*parseStack)[len(*parseStack)-ruleLength:]
	*parseStack = (*parseStack)[:len(*parseStack)-ruleLength]

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

	if !isAccept {
		backState := 0
		if len(*parseStack)-1 >= 0 {
			backState = (*parseStack)[len(*parseStack)-1].State
		}
		nextState = parseTable.States[backState].GotoTable[lhs].GotoState
	} else {
		nextState = 0
	}

	*parseStack = append(*parseStack, ParseStackElement{
		Type:   PARSE_STACK_ELEMENT_TYPE_REDUCED,
		String: reducedString,
		State:  nextState,
	})

	return nextState, nil
}
