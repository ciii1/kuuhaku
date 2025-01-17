package kuuhaku_analyzer

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"

	"github.com/ciii1/kuuhaku/internal/helper"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_parser"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_tokenizer"
	"github.com/yuin/gopher-lua"
)

type AnalyzeErrorType int

const (
	UNDEFINED_VARIABLE = iota
	MULTIPLE_START_SYMBOLS
	OUT_OF_BOUND_CAPTURE_GROUP
	INVALID_REGEX
	INVALID_ARG_LENGTH
	INVALID_LUA_LITERAL
)

type AnalyzeError struct {
	Position kuuhaku_tokenizer.Position
	Message  string
	Type     AnalyzeErrorType
}

type ConflictError struct {
	Position1 kuuhaku_tokenizer.Position
	Position2 kuuhaku_tokenizer.Position
	Symbol1   *Symbol
	Symbol2   *Symbol
	Message   string
}

func (e AnalyzeError) Error() string {
	return fmt.Sprintf("Analyze error (%d, %d): %s", e.Position.Line, e.Position.Column, e.Message)
}

func (e ConflictError) Error() string {
	return fmt.Sprintf("Analyze error (%d, %d): %s", e.Position1.Line, e.Position1.Column, e.Message)
}

func ErrUndefinedVariable(position kuuhaku_tokenizer.Position, variableName string) *AnalyzeError {
	return &AnalyzeError{
		Message:  "Variable " + variableName + " is undefined",
		Position: position,
		Type:     UNDEFINED_VARIABLE,
	}
}

func ErrOutOfBoundCaptureGroup(position kuuhaku_tokenizer.Position, max int) *AnalyzeError {
	return &AnalyzeError{
		Message:  "The capture group exceeds the index of the last element in the match rule which is " + strconv.Itoa(max),
		Position: position,
		Type:     OUT_OF_BOUND_CAPTURE_GROUP,
	}
}

func ErrMultipleStartSymbols(position kuuhaku_tokenizer.Position, startSymbol1 string, startSymbol2 string) *AnalyzeError {
	return &AnalyzeError{
		Message:  "Found multiple start symbols while not in search mode: " + startSymbol1 + ", " + startSymbol2,
		Position: position,
		Type:     MULTIPLE_START_SYMBOLS,
	}
}

func ErrInvalidRegex(position kuuhaku_tokenizer.Position, regex string, regexError error) *AnalyzeError {
	return &AnalyzeError{
		Message:  "Invalid regex: <" + regex + "> (" + regexError.Error() + ")",
		Position: position,
		Type:     INVALID_REGEX,
	}
}

func ErrInvalidArgLength(position kuuhaku_tokenizer.Position, identifierName string, argLength int) *AnalyzeError {
	return &AnalyzeError{
		Message:  strconv.Itoa(argLength) + " is an invalid argument length when using the rule " + identifierName,
		Position: position,
		Type:     INVALID_ARG_LENGTH,
	}
}

func ErrInvalidLuaLiteral(position kuuhaku_tokenizer.Position, luaError string) *AnalyzeError {
	return &AnalyzeError{
		Message:  "Invalid Lua literal. Error:\n\t " + luaError,
		Position: position,
		Type:     INVALID_LUA_LITERAL,
	}
}

func ErrConflict(symbol1 *Symbol, symbol2 *Symbol) *ConflictError {
	var position1 kuuhaku_tokenizer.Position
	if symbol1.Position < len(symbol1.Rule.MatchRules) {
		position1 = symbol1.Rule.MatchRules[symbol1.Position].GetPosition()
	} else {
		position1 = symbol1.Rule.MatchRules[len(symbol1.Rule.MatchRules)-1].GetPosition()
	}

	var position2 kuuhaku_tokenizer.Position
	if symbol2.Position < len(symbol2.Rule.MatchRules) {
		position2 = symbol2.Rule.MatchRules[symbol2.Position].GetPosition()
	} else {
		position2 = symbol2.Rule.MatchRules[len(symbol2.Rule.MatchRules)-1].GetPosition()
	}

	return &ConflictError{
		Message:   "Detected conflict at rule " + strconv.Itoa(symbol1.Rule.Order+1) + " and rule " + strconv.Itoa(symbol2.Rule.Order+1) + " with position (" + strconv.Itoa(position2.Line) + ", " + strconv.Itoa(position2.Column) + ")\n--- Lookaheads are: " + symbol1.Lookeahead.String + " and " + symbol2.Lookeahead.String,
		Position1: position1,
		Position2: position2,
		Symbol1:   symbol1,
		Symbol2:   symbol2,
	}
}

type Analyzer struct {
	input                  *kuuhaku_parser.Ast
	Errors                 []error
	stateNumber            int
	parseTables            []ParseTable
	stateTransitionMap     map[Symbol]int
	stateTransitionMapBool map[Symbol]bool
}

func Analyze(input *kuuhaku_parser.Ast) (AnalyzerResult, []error) {
	analyzer := initAnalyzer(input)
	startSymbols := analyzer.analyzeStart()
	if len(startSymbols) > 1 && !input.IsSearchMode {
		analyzer.Errors = append(analyzer.Errors, ErrMultipleStartSymbols(input.Rules[startSymbols[1]][0].Position, startSymbols[0], startSymbols[1]))
	}
	if len(analyzer.Errors) == 0 {
		for _, startSymbol := range startSymbols {
			analyzer.parseTables = append(analyzer.parseTables, analyzer.makeEmptyParseTable(startSymbol))
			analyzer.buildParseTable(startSymbol)
		}
	}

	return AnalyzerResult{
		ParseTables:  analyzer.parseTables,
		IsSearchMode: input.IsSearchMode,
	}, analyzer.Errors
}

func initAnalyzer(input *kuuhaku_parser.Ast) Analyzer {
	return Analyzer{
		input:                  input,
		Errors:                 []error{},
		stateNumber:            1,
		parseTables:            []ParseTable{},
		stateTransitionMap:     make(map[Symbol]int),
		stateTransitionMapBool: make(map[Symbol]bool),
	}
}

func (analyzer *Analyzer) makeEmptyParseTable(startSymbol string) ParseTable {
	terminalsMapInput := make(map[string]*TerminalList)
	var terminalsMap *map[string]*TerminalList
	lhsMapInput := make(map[string]bool)
	var lhsMap *map[string]bool
	terminalsMap, lhsMap = analyzer.getAllTerminalsAndLhs(startSymbol, &terminalsMapInput, &lhsMapInput)

	terminals := sortTerminalsMaptoArray(terminalsMap)

	var lhsArray []string
	for lhs := range *lhsMap {
		lhsArray = append(lhsArray, lhs)
	}

	return ParseTable{
		States:    []ParseTableState{},
		Terminals: *terminals,
		Lhss:      lhsArray,
	}
}

func sortTerminalsMaptoArray(terminalsMap *map[string]*TerminalList) *[]TerminalList {
	var terminals []TerminalList
	for _, terminal := range *terminalsMap {
		terminals = append(terminals, *terminal)
	}
	sort.Slice(terminals, func(i, j int) bool {
		return terminals[i].Precedence < terminals[j].Precedence
	})
	return &terminals
}

func (analyzer *Analyzer) getAllTerminalsAndLhs(startSymbol string, previousTerminalMap *map[string]*TerminalList, previousLhsMap *map[string]bool) (*map[string]*TerminalList, *map[string]bool) {
	terminalsMap := previousTerminalMap
	lhsMap := previousLhsMap
	(*lhsMap)[startSymbol] = true
	for _, rule := range analyzer.input.Rules[startSymbol] {
		for _, matchRule := range (*rule).MatchRules {
			regexCurr, ok := matchRule.(kuuhaku_parser.RegexLiteral)
			if ok {
				if (*terminalsMap)[regexCurr.RegexString] == nil || (*terminalsMap)[regexCurr.RegexString].Precedence > rule.Order {
					regexCompiled, err := regexp.Compile("^" + regexCurr.RegexString)
					if err != nil {
						analyzer.Errors = append(analyzer.Errors, ErrInvalidRegex(regexCurr.Position, regexCurr.RegexString, err))
					}
					(*terminalsMap)[regexCurr.RegexString] = &TerminalList{
						Terminal:   regexCurr.RegexString,
						Precedence: rule.Order,
						Regexp:     regexCompiled,
					}
				}
			} else {
				identifierCurr, ok := matchRule.(kuuhaku_parser.Identifier)
				if ok {
					if !(*lhsMap)[identifierCurr.Name] {
						terminalsMap, lhsMap = analyzer.getAllTerminalsAndLhs(identifierCurr.Name, terminalsMap, lhsMap)
					}
				}
			}
		}
	}
	return terminalsMap, lhsMap
}

func makeEndSymbolTitle() SymbolTitle {
	return SymbolTitle{String: "<end>", Type: EMPTY_TITLE}
}

func getSymbolTitleFromMatchRule(matchRule kuuhaku_parser.MatchRule) SymbolTitle {
	currIdentifier, ok := matchRule.(kuuhaku_parser.Identifier)
	if ok {
		return SymbolTitle{
			String: currIdentifier.Name,
			Type:   IDENTIFIER_TITLE,
		}
	} else {
		currRegexLit, ok := matchRule.(kuuhaku_parser.RegexLiteral)
		if ok {
			return SymbolTitle{
				String: currRegexLit.RegexString,
				Type:   REGEX_LITERAL_TITLE,
			}
		}
	}
	return SymbolTitle{Type: EMPTY_TITLE}
}

func (analyzer *Analyzer) expandSymbol(rules *[]*kuuhaku_parser.Rule, position int, previousSymbols *[]*Symbol, lookahead SymbolTitle) *[]*Symbol {
	output := previousSymbols
	for _, currRule := range *rules {
		if position >= len(currRule.MatchRules) {
			*output = append(*output, &Symbol{
				Rule:       currRule,
				Position:   position,
				Title:      makeEndSymbolTitle(),
				Lookeahead: lookahead,
			})
			continue
		}

		currMatchRule := currRule.MatchRules[position]
		nextLookahead := makeEndSymbolTitle()
		if position+1 < len(currRule.MatchRules) {
			nextMatchRule := currRule.MatchRules[position+1]
			nextLookahead = getSymbolTitleFromMatchRule(nextMatchRule)
		}

		*output = append(*output, &Symbol{
			Rule:       currRule,
			Position:   position,
			Title:      getSymbolTitleFromMatchRule(currMatchRule),
			Lookeahead: lookahead,
		})

		currIdentifier, ok := currMatchRule.(kuuhaku_parser.Identifier)
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
	grouped = analyzer.buildParseTableState(grouped, startSymbolString)
	stateTransitions = append(stateTransitions, &StateTransition{
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
				expandedSymbols := analyzer.expandSymbol(&[]*kuuhaku_parser.Rule{symbol.Rule}, symbol.Position+1, &[]*Symbol{}, symbol.Lookeahead)
				for _, expandedSymbol := range *expandedSymbols {
					expandedSymbolsAll = append(expandedSymbolsAll, expandedSymbol)
				}
			}

			grouped := analyzer.groupSymbols(&expandedSymbolsAll)
			grouped = analyzer.buildParseTableState(grouped, startSymbolString)
			if len(*grouped) != 0 {
				stateTransitions = append(stateTransitions, &StateTransition{
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
				groupsMap[symbol.Title] = SymbolGroup{
					Title:   symbolTitle,
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

func (analyzer *Analyzer) buildParseTableState(symbolGroups *[]*SymbolGroup, startSymbol string) *[]*SymbolGroup {
	if len(*symbolGroups) == 0 {
		return symbolGroups
	}
	actionTable := make(map[string]*ActionCell)
	gotoTable := make(map[string]*GotoCell)
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
						var action Action = REDUCE
						if symbol.Rule.Name == startSymbol {
							action = ACCEPT
						}
						endReduceRule = &ActionCell{
							LookaheadTerminal: "",
							Action:            action,
							ReduceRule:        symbol.Rule,
							ShiftState:        0,
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
						actionTable[terminal] = &ActionCell{
							LookaheadTerminal: terminal,
							Action:            REDUCE,
							ReduceRule:        symbol.Rule,
							ShiftState:        0,
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
			gotoTable[group.Title.String] = &GotoCell{
				Lhs:       group.Title.String,
				GotoState: analyzer.stateNumber,
			}
			analyzer.stateNumber++
			outGroup = append(outGroup, group)
		} else if group.Title.Type == REGEX_LITERAL_TITLE {
			if usedTerminalsWithSymbol[group.Title.String] != nil {
				analyzer.Errors = append(analyzer.Errors, ErrConflict((*group.Symbols)[0], usedTerminalsWithSymbol[group.Title.String]))
			}
			isStateExisted := false
			existedStateNumber := 0
			for _, symbol := range *group.Symbols {
				if isStateExisted {
					if analyzer.stateTransitionMapBool[*symbol] != false {
						if analyzer.stateTransitionMap[*symbol] != existedStateNumber {
							isStateExisted = false
							existedStateNumber = 0
						}
					} else {
						isStateExisted = false
						existedStateNumber = 0
					}
				} else {
					if analyzer.stateTransitionMapBool[*symbol] != false {
						isStateExisted = true
						existedStateNumber = analyzer.stateTransitionMap[*symbol]
					}
				}
			}
			if !isStateExisted {
				actionTable[group.Title.String] = &ActionCell{
					LookaheadTerminal: group.Title.String,
					Action:            SHIFT,
					ReduceRule:        nil,
					ShiftState:        analyzer.stateNumber,
				}
				outGroup = append(outGroup, group)
				for _, symbol := range *group.Symbols {
					analyzer.stateTransitionMapBool[*symbol] = true
					analyzer.stateTransitionMap[*symbol] = analyzer.stateNumber
				}
				analyzer.stateNumber++
			} else {
				actionTable[group.Title.String] = &ActionCell{
					LookaheadTerminal: group.Title.String,
					Action:            SHIFT,
					ReduceRule:        nil,
					ShiftState:        existedStateNumber,
				}
			}
		}
	}
	currParseTable := &analyzer.parseTables[len(analyzer.parseTables)-1]
	currParseTable.States = append(currParseTable.States, ParseTableState{
		ActionTable:   actionTable,
		GotoTable:     gotoTable,
		EndReduceRule: endReduceRule,
	})
	return &outGroup
}

func (analyzer *Analyzer) makeAugmentedGrammar(startSymbol string) *Symbol {
	startRules := analyzer.input.Rules[startSymbol]
	order := startRules[0].Order
	ruleName := "S" + startSymbol
	rule := &kuuhaku_parser.Rule{
		Name:  ruleName,
		Order: order,
		MatchRules: []kuuhaku_parser.MatchRule{
			kuuhaku_parser.Identifier{
				Name:     startSymbol,
				Position: startRules[0].Position,
			},
		},
		Position: startRules[0].Position,
	}
	analyzer.input.Rules[ruleName] = append(analyzer.input.Rules[ruleName], rule)
	return &Symbol{
		Title:    getSymbolTitleFromMatchRule(rule.MatchRules[0]),
		Position: 0,
		Rule:     rule,
	}
}

func (analyzer *Analyzer) analyzeLuaLiteral(source kuuhaku_parser.LuaLiteral) string {
	L := lua.NewState()
	defer L.Close()

	outStr := ""
	if source.Type == kuuhaku_parser.LUA_LITERAL_TYPE_RETURN {
		outStr += "return "	
	}
	outStr += source.LuaString

	_, err := L.LoadString(outStr)
	if err != nil {
		analyzer.Errors = append(analyzer.Errors, ErrInvalidLuaLiteral(source.Position, err.Error()))
	}

	return outStr
}

// return start symbols
func (analyzer *Analyzer) analyzeStart() []string {
	startSymbols := make([]string, len(analyzer.input.Rules))
	i := 0
	for key := range analyzer.input.Rules {
		startSymbols[i] = key
		i++
	}

	for ruleName, ruleArray := range analyzer.input.Rules {
		for _, rule := range ruleArray {
			for _, arg := range rule.ArgList {
				analyzer.analyzeLuaLiteral(arg)
			}
			analyzer.analyzeLuaLiteral(rule.ReplaceRule)
			for _, matchRule := range rule.MatchRules {
				identifier, ok := matchRule.(kuuhaku_parser.Identifier)
				if !ok {
					continue
				}

				if len(analyzer.input.Rules[identifier.Name]) == 0 {
					analyzer.Errors = append(analyzer.Errors, ErrUndefinedVariable(identifier.Position, identifier.Name))
				} else if !analyzer.doesMatchingArgumentNumberRuleExist(identifier) {
					analyzer.Errors = append(analyzer.Errors, ErrInvalidArgLength(identifier.Position, identifier.Name, len(identifier.ArgList)))
				}

				for _, arg := range identifier.ArgList {
					analyzer.analyzeLuaLiteral(arg)
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

func (analyzer *Analyzer) doesMatchingArgumentNumberRuleExist(ruleName kuuhaku_parser.Identifier) bool {
	for _, rule := range analyzer.input.Rules[ruleName.Name] {
		if len(rule.ArgList) == len(ruleName.ArgList) {
			return true
		}
	}
	return false
}

func PrintParseTable(parseTable *ParseTable) {

	maxWidthTerminals := make(map[string]int)
	maxWidthLhss := make(map[string]int)
	for _, terminal := range parseTable.Terminals {
		maxWidthTerminals[terminal.Terminal] = len(terminal.Terminal)
	}
	for _, lhs := range parseTable.Lhss {
		maxWidthLhss[lhs] = len(lhs)
	}
	for _, state := range parseTable.States {
		for _, terminal := range parseTable.Terminals {
			if state.ActionTable[terminal.Terminal] != nil {
				if len(strconv.Itoa(state.ActionTable[terminal.Terminal].ShiftState)) > maxWidthTerminals[terminal.Terminal] {
					maxWidthTerminals[terminal.Terminal] = len(strconv.Itoa(state.ActionTable[terminal.Terminal].ShiftState))
				}
			}
		}
		for _, lhs := range parseTable.Lhss {
			if state.GotoTable[lhs] != nil {
				if len(strconv.Itoa(state.GotoTable[lhs].GotoState)) > maxWidthLhss[lhs] {
					maxWidthLhss[lhs] = len(strconv.Itoa(state.GotoTable[lhs].GotoState))
				}
			}
		}
	}
	maxWidthState := 6
	if maxWidthState < len(strconv.Itoa(len(parseTable.States))) {
		maxWidthState = len(strconv.Itoa(len(parseTable.States)))
	}

	print("| States")
	i := 0
	for i < maxWidthState - 6 {
		print(" ")
		i++
	}
	print(" || $end ||")

	for _, terminal := range parseTable.Terminals {
		print(" " + terminal.Terminal)
		i = 0
		for i < maxWidthTerminals[terminal.Terminal] - len(terminal.Terminal) {
			print(" ")
			i++
		}
		print(" |")
	}
	print("|")
	for _, lhs := range parseTable.Lhss {
		print(" " + lhs)
		i = 0
		for i < maxWidthLhss[lhs] - len(lhs) {
			print(" ")
			i++
		}
		print(" |")
	}
	println("")

	print("+-------")
	i = 0
	for i < maxWidthState - 6 {
		print("-")
		i++
	}
	print("-++------++")
	for _, terminal := range parseTable.Terminals {
		print("-")
		for range terminal.Terminal {
			print("-")
		}
		i = 0
		for i < maxWidthTerminals[terminal.Terminal] - len(terminal.Terminal) {
			print("-")
			i++
		}
		print("-+")
	}
	print("+")
	for _, lhs := range parseTable.Lhss {
		print("-")
		for range lhs {
			print("-")
		}
		i = 0
		for i < maxWidthLhss[lhs] - len(lhs) {
			print("-")
			i++
		}
		print("-+")
	}

	for i, state := range parseTable.States {
		println("")
		print("| ")
		print(strconv.Itoa(i))
		i = 0
		for i < maxWidthState - len(strconv.Itoa(i)) {
			print(" ")
			i++
		}
		print(" ||")

		if state.EndReduceRule != nil {
			print(" acc  ||")
		} else {
			print("      ||")
		}
		
		for _, terminal := range parseTable.Terminals {
			print(" ")
			actionNumberLength := 0
			if state.ActionTable[terminal.Terminal] != nil{
				print(state.ActionTable[terminal.Terminal].ShiftState)
				actionNumberLength = len(strconv.Itoa(state.ActionTable[terminal.Terminal].ShiftState))
			}
			i = 0
			for i < maxWidthTerminals[terminal.Terminal] - actionNumberLength {
				print(" ")
				i++
			}
			print(" |")
		}

		print("|")

		for _, lhs := range parseTable.Lhss {
			print(" ")
			lhsNumberLength := 0
			if state.GotoTable[lhs] != nil{
				print(state.GotoTable[lhs].GotoState)
				lhsNumberLength = len(strconv.Itoa(state.GotoTable[lhs].GotoState))
			}
			i = 0
			for i < maxWidthLhss[lhs] - lhsNumberLength {
				print(" ")
				i++
			}
			print(" |")
		}
	}

	println("")
}
