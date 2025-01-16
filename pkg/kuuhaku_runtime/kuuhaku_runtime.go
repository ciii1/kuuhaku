package kuuhaku_runtime

import (
	"fmt"

	"github.com/ciii1/kuuhaku/pkg/kuuhaku_analyzer"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_parser"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_tokenizer"
)

type ParseStackElementType int

const (
	PARSE_STACK_ELEMENT_TYPE_TREE ParseStackElementType = iota
	PARSE_STACK_ELEMENT_TYPE_TERMINAL
)

type ParseStackElement interface {
	GetType() ParseStackElementType
	GetString() string
	GetState() int
}

type ParseStackTree struct {
	Children *[]ParseStackElement
	Rule *kuuhaku_parser.Rule
	State  int
}

func (_ *ParseStackTree) GetType() ParseStackElementType {
	return PARSE_STACK_ELEMENT_TYPE_TREE;
}

func (p *ParseStackTree) GetState() int {
	return p.State;
}

func (p *ParseStackTree) GetString() string {
	out := "["
	for i, child := range *p.Children {
		if i != 0 {
			out += ","
		}
		out += child.GetString()
	}
	out += "]"
	return out
}

type ParseStackTerminal struct {
	String string
	State  int
}

func (_ *ParseStackTerminal) GetType() ParseStackElementType {
	return PARSE_STACK_ELEMENT_TYPE_TERMINAL;
}

func (p *ParseStackTerminal) GetString() string {
	return p.String
}

func (p *ParseStackTerminal) GetState() int {
	return p.State;
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

func Format(input string, format *kuuhaku_analyzer.AnalyzerResult, isRun bool) (string, error) {
	var currPos kuuhaku_tokenizer.Position
	currPos.Line = 1
	currPos.Column = 1
	out := ""
	for currPos.Raw < len(input) {
		isThereSuccess := false
		// We cannot have only one parse table for multiple start symbols because that'll 
		// prevent us from having the backtracking mechanism
		for _, parseTable := range format.ParseTables {
			res, resPos, err := runParseTable(input, currPos, &parseTable, isRun)
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
		if sliced[i] == '\n' {
			break
		}
		col++
		colIfContainsNewLine++
		i--
	}

	//i is always -1 if a \n wasn't found. This is because the condition i >= 0 will not be satisfied untill
	//i == -1. sliced[i] != '\n' will only be satisfied if i > -1. So it is safe to check if i > -1 to
	//check for newlines

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

func runParseTable(input string, pos kuuhaku_tokenizer.Position, parseTable *kuuhaku_analyzer.ParseTable, isRun bool) (string, kuuhaku_tokenizer.Position, error) {
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
				parseStack = append(parseStack, &ParseStackTerminal {
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
	out := ""
	if isRun {
		out = runParseStack(&parseStack)
	} else {
		out = parseStackToString(&parseStack)
	}
	return out, pos, nil
}

func printParseStack(parseStack *[]ParseStackElement) {
	print("Parse stack: [")
	for i, parseStackElement := range *parseStack {
		if i != 0 {
			print(",")
		}
		print(parseStackElement.GetString())
	}
	println("]")
}

func parseStackToString(parseStack *[]ParseStackElement) string {
	out := "["
	for i, parseStackElement := range *parseStack {
		if i != 0 {
			out += ","
		}
		out += parseStackElement.GetString()
	}
	out += "]"
	return out
}

func runParseStack(parseStack *[]ParseStackElement) string {
	/*if len(rule.ReplaceRules) == 0 {
		for _, targetElement := range targetStack {
			reducedString += targetElement.String
		}
	}*/
	return "" //(*parseStack)[0].GetString()
}

func copyParseStack(parseStack []ParseStackElement) *[]ParseStackElement {
	var newParseStack []ParseStackElement
	for _, e := range parseStack {
		if e.GetType() == PARSE_STACK_ELEMENT_TYPE_TREE {
			newParseStack = append(newParseStack, copyParseStackTreeRecursive(&e))
		} else if  e.GetType() == PARSE_STACK_ELEMENT_TYPE_TERMINAL {
			newParseStack = append(newParseStack, copyParseStackTerminal(&e))
		}
	}
	return &newParseStack
}

func copyParseStackTreeRecursive(e *ParseStackElement) *ParseStackTree {
	parseStackTree, ok := (*e).(*ParseStackTree) 
	var children []ParseStackElement
	if ok {
		for _, child := range *parseStackTree.Children {
			if child.GetType() == PARSE_STACK_ELEMENT_TYPE_TREE {
				children = append(children, copyParseStackTreeRecursive(&child))
			} else if  child.GetType() == PARSE_STACK_ELEMENT_TYPE_TERMINAL {
				newChild := child
				children = append(children, newChild)
			}
		}
		return &ParseStackTree{
			Children: &children,
			Rule: parseStackTree.Rule,
			State: parseStackTree.State,
		}
	}
	return nil
}

func copyParseStackTerminal(e *ParseStackElement) *ParseStackTerminal {
	parseStackTerminal, ok := (*e).(*ParseStackTerminal)
	if ok {
		newTerminal := parseStackTerminal
		return newTerminal
	}
	return nil
}

func applyRule(parseTable *kuuhaku_analyzer.ParseTable, rule *kuuhaku_parser.Rule, parseStack *[]ParseStackElement, pos kuuhaku_tokenizer.Position, isAccept bool) (int, error) {
	nextState := 0
	lhs := rule.Name
	ruleLength := len(rule.MatchRules)

	if len(*parseStack)-ruleLength < 0 {
		return 0, ErrReduceRuleIsNotMatching(pos)
	}
	targetStack := copyParseStack((*parseStack)[len(*parseStack)-ruleLength:])
	*parseStack = (*parseStack)[:len(*parseStack)-ruleLength]	

	if !isAccept {
		backState := 0
		if len(*parseStack)-1 >= 0 {
			backState = (*parseStack)[len(*parseStack)-1].GetState()
		}
		nextState = parseTable.States[backState].GotoTable[lhs].GotoState
	} else {
		nextState = 0
	}

	*parseStack = append(*parseStack, &ParseStackTree{
		Children: targetStack,
		Rule: rule,
		State:  nextState,
	})

	return nextState, nil
}
