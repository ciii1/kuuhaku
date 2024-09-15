package kuuhaku_analyzer

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/ciii1/kuuhaku/internal/helper"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_parser"
	"github.com/kr/pretty"
)

func TestErrorUndefinedVariable(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("identifier{test<\\.>}\ntest2{identifier}\ntest34{test4}");
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")	
		t.Fatal()
	}
	analyzer := initAnalyzer(&ast)
	_ = analyzer.analyzeStart()	
	if len(analyzer.Errors) != 2 {
		println("Expected analyzer Errors length to be 2")
		t.Fatal()
	}

	println("TestErrorUndefinedVariable - Errors:")
	helper.DisplayAllErrors(analyzer.Errors)

	var analyzeError *AnalyzeError
	if errors.As(analyzer.Errors[0], &analyzeError) {
		if analyzeError.Type != UNDEFINED_VARIABLE {
			println("Expected UndefinedVariableError error")
			t.Fail()
		}
		if (analyzeError.Position.Column != 12 || analyzeError.Position.Line != 1) && (analyzeError.Position.Column != 8 || analyzeError.Position.Line != 3) {
			col := strconv.Itoa(analyzeError.Position.Column)
			line := strconv.Itoa(analyzeError.Position.Line)
			println("Expected UndefinedVariableError error with column 12 and line 1, got (" + col + ", " + line + ")")
			t.Fail()
		}
	} else {
		println("Expected AnalyzeError")
		t.Fail()
	}

	if errors.As(analyzer.Errors[1], &analyzeError) {
		if analyzeError.Type != UNDEFINED_VARIABLE {
			println("Expected UndefinedVariableError error")
			t.Fail()
		}
		if (analyzeError.Position.Column != 8 || analyzeError.Position.Line != 3) && (analyzeError.Position.Column != 12 || analyzeError.Position.Line != 1) {
			col := strconv.Itoa(analyzeError.Position.Column)
			line := strconv.Itoa(analyzeError.Position.Line)
			println("Expected UndefinedVariableError error with column 8 and line 3, got (" + col + ", " + line + ")")
			t.Fail()
		}
	} else {
		println("Expected AnalyzeError")
		t.Fail()
	}

}

func TestStartSymbols(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("identifier{test<\\.>}\ntest{<\\.>}\nidentifier{<\\.>}\ntest3{<\\.>}");
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")	
		t.Fatal()
	}
	analyzer := initAnalyzer(&ast)
	startSymbols := analyzer.analyzeStart()	
	if len(analyzer.Errors) != 0 {
		println("Expected analyzer Errors length to be 0")
		t.Fatal()
	}

	if len(startSymbols) != 2 {
		println("Expected startSymbols length to be 2")
		t.Fatal()
	}
	
	if startSymbols[0] == "identifier" {
		if startSymbols[1] != "test3" {
			println("Expected startSymbols[1] to be \"test3\", got" + startSymbols[1])
			t.Fail()
		}
	} else if startSymbols[1] == "identifier" {
		if startSymbols[0] != "test3" {
			println("Expected startSymbols[1] to be \"test3\", got" + startSymbols[1])
			t.Fail()
		}
	} else {
		println("Expected startSymbols[0] or [1] to be \"identifier\"")
		t.Fail()
	}
}

func TestExpandSymbol(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("identifier{test<\\.>}\ntest{<\\.>}\nidentifier{<\\.>}\ntest3{<\\.>}");
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")	
		t.Fatal()
	}
	analyzer := initAnalyzer(&ast)
	rules := ast.Rules["identifier"]
	expandedSymbols := analyzer.expandSymbol(&rules, 0, &[]*Symbol{}, SymbolTitle{Type:EMPTY_TITLE})
	if len(analyzer.Errors) != 0 {
		println("Expected analyzer Errors length to be 0")
		t.Fatal()
	}

	if len(*expandedSymbols) != 3 {
		println("Expected expandedSymbols length to be 3, got " + strconv.Itoa(len(*expandedSymbols)))
		fmt.Printf("%# v\n", pretty.Formatter(*expandedSymbols))
		t.Fatal()
	}

	firstSymbol := (*(*expandedSymbols)[0])
	title1 := firstSymbol.Title
	if title1.Type != IDENTIFIER_TITLE {
		println("Expected expandedSymbols[0].Title to be an identifier")
		t.Fail()
	}
	if title1.String != "test" {
		println("Expected expandedSymbols[0].Title.String to be \"test\"")
		t.Fail()
	}
	if firstSymbol.Lookeahead.Type != EMPTY_TITLE {
		println("Expected expandedSymbols[0].Lookahead.Type to be EMPTY_TITLE")
		t.Fail()
	}

	if !reflect.DeepEqual(*firstSymbol.Rule, *ast.Rules["identifier"][0]) {
		println("The first symbol's rule is not matching")
	}

	if firstSymbol.Position != 0 {
		println("Expected expandedSymbols[0].Position to be 0")
	}

	secondSymbol := (*(*expandedSymbols)[1])
	title2 := secondSymbol.Title
	if title2.Type != REGEX_LITERAL_TITLE {
		println("Expected expandedSymbols[1].Title to be a regex literal")
		t.Fail()
	}
	if title2.String != "\\." {
		println("Expected expandedSymbols[1].Title.String to be \"\\.\"")
		t.Fail()
	}

	if !reflect.DeepEqual(*secondSymbol.Rule, *ast.Rules["test"][0]) {
		println("The second symbol's rule is not matching")
	}

	if secondSymbol.Position != 0 {
		println("Expected expandedSymbols[1].Position to be 0")
	}

	if secondSymbol.Lookeahead.Type != REGEX_LITERAL_TITLE {
		println("Expected expandedSymbols[1].Lookahead.Type to be REGEX_LITERAL_TITLE")
		t.Fail()
	}

	if secondSymbol.Lookeahead.String != "\\." {
		println("Expected expandedSymbols[1].Lookahead.String to be \"\\.\"")
		t.Fail()
	}

	thirdSymbol := (*(*expandedSymbols)[2])
	title3 := thirdSymbol.Title
	if title3.Type != REGEX_LITERAL_TITLE {
		println("Expected expandedSymbols[2].Title to be a regex literal")
		t.Fail()
	}

	if title3.String != "\\." {
		println("Expected expandedSymbols[2].Title.String to be \"\\.\"")
		t.Fail()
	}

	if !reflect.DeepEqual(*thirdSymbol.Rule, *ast.Rules["identifier"][1]) {
		println("The third symbol's rule is not matching")
	}

	if thirdSymbol.Lookeahead.Type != EMPTY_TITLE {
		println("Expected expandedSymbols[2].Lookahead.Type to be EMPTY_TITLE")
		t.Fail()
	}

	if thirdSymbol.Position != 0 {
		println("Expected expandedSymbols[2].Position to be 0")
	}
}

func TestExpandSymbol2(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("identifier{<\\.>test}\ntest{<\\.>}\nidentifier{<\\.>}\ntest3{<\\.>}");
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")	
		t.Fatal()
	}
	analyzer := initAnalyzer(&ast)
	rules := ast.Rules["identifier"]
	expandedSymbols := analyzer.expandSymbol(&rules, 1, &[]*Symbol{},SymbolTitle{Type:EMPTY_TITLE})
	if len(analyzer.Errors) != 0 {
		println("Expected analyzer Errors length to be 0")
		t.Fatal()
	}

	if len(*expandedSymbols) != 3 {
		println("Expected expandedSymbols length to be 3, got " + strconv.Itoa(len(*expandedSymbols)))
		fmt.Printf("%# v\n", pretty.Formatter(*expandedSymbols))
		t.Fatal()
	}

	firstSymbol := (*(*expandedSymbols)[0])
	title1 := firstSymbol.Title
	if title1.Type != IDENTIFIER_TITLE {
		println("Expected expandedSymbols[0].Title to be an identifier")
		t.Fail()
	}

	if title1.String != "test" {
		println("Expected expandedSymbols[0].Title.String to be \"test\"")
		t.Fail()
	}
	

	if !reflect.DeepEqual(*firstSymbol.Rule, *ast.Rules["identifier"][0]) {
		println("The first symbol's rule is not matching")
	}

	if firstSymbol.Position != 1 {
		println("Expected expandedSymbols[0].Position to be 1")
	}

	secondSymbol := (*(*expandedSymbols)[1])
	title2 := secondSymbol.Title
	if title2.Type != REGEX_LITERAL_TITLE{
		println("Expected expandedSymbols[1].Title to be a regex literal")
		t.Fail()
	}
	if title2.String != "\\." {
		println("Expected expandedSymbols[1].Title.String to be \"\\.\"")
		t.Fail()
	}

	if !reflect.DeepEqual(*secondSymbol.Rule, *ast.Rules["test"][0]) {
		println("The second symbol's rule is not matching")
		t.Fail()
	}

	if secondSymbol.Position != 0 {
		println("Expected expandedSymbols[1].Position to be 0")
		t.Fail()
	}

	if secondSymbol.Lookeahead.Type != EMPTY_TITLE {
		println("Expected expandedSymbols[1].Lookahead.Type to be EMPTY TITLE")
		t.Fail()
	}

	thirdSymbol := (*(*expandedSymbols)[2])
	title3 := thirdSymbol.Title
	if title3.Type != EMPTY_TITLE {
		println("Expected expandedSymbols[2].Title to be an EMPTY_TITLE")
		t.Fail()
	}

	if !reflect.DeepEqual(*thirdSymbol.Rule, *ast.Rules["identifier"][1]) {
		println("The third symbol's rule is not matching")
	}

	if thirdSymbol.Lookeahead.Type != EMPTY_TITLE {
		println("Expected expandedSymbols[2].Lookahead.Type to be EMPTY_TITLE")
		t.Fail()
	}

	if thirdSymbol.Position != 1 {
		println("Expected expandedSymbols[2].Position to be 0")
	}
}

func TestExpandSymbol3(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("identifier{<\\.>test}\ntest{<\\.>}\nidentifier{<\\.>}\ntest3{<\\.>}");
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")	
		t.Fatal()
	}
	analyzer := initAnalyzer(&ast)
	rules := ast.Rules["test"]
	expandedSymbols := analyzer.expandSymbol(&rules, 2, &[]*Symbol{},SymbolTitle{Type:EMPTY_TITLE})
	if len(analyzer.Errors) != 0 {
		println("Expected analyzer Errors length to be 0")
		t.Fatal()
	}

	if len(*expandedSymbols) != 1 {
		println("Expected expandedSymbols length to be 1, got " + strconv.Itoa(len(*expandedSymbols)))
		fmt.Printf("%# v\n", pretty.Formatter(*expandedSymbols))
		t.Fatal()
	}

	firstSymbol := (*(*expandedSymbols)[0])
	title1 := firstSymbol.Title
	if title1.Type != EMPTY_TITLE {
		println("Expected expandedSymbols[0].Title to be empty")
		t.Fail()
	}
}

func TestGroupSymbols(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("identifier{test<\\.>}\ntest{<\\.>}\nidentifier{<\\.>}\ntest3{<\\.>}");
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")
		t.Fatal()
	}
	analyzer := initAnalyzer(&ast)
	rules := ast.Rules["identifier"]
	expandedSymbols := analyzer.expandSymbol(&rules, 0, &[]*Symbol{},SymbolTitle{Type:EMPTY_TITLE})
	if len(analyzer.Errors) != 0 {
		println("Expected analyzer Errors length to be 0")
		t.Fatal()
	}

	if len(*expandedSymbols) != 3 {
		println("Expected expandedSymbols length to be 3, got " + strconv.Itoa(len(*expandedSymbols)))
		fmt.Printf("%# v\n", pretty.Formatter(*expandedSymbols))
		t.Fatal()
	}

	groupedSymbols := analyzer.groupSymbols(expandedSymbols)

	if len(*groupedSymbols) != 2 {
		println("Expected expandedSymbols length to be 2, got " + strconv.Itoa(len(*groupedSymbols)))
		fmt.Printf("%# v\n", pretty.Formatter(*groupedSymbols))
		t.Fatal()
	}
	
	comparedTitle1 := SymbolTitle{
		String: "\\.", 
		Type: REGEX_LITERAL_TITLE,
	}
	comparedTitle2 := SymbolTitle{
		String: "test", 
		Type: IDENTIFIER_TITLE,
	}
	var regexLitGroup *SymbolGroup
	var identifierGroup *SymbolGroup
	if (*groupedSymbols)[0].Title == comparedTitle1 {	
		if (*groupedSymbols)[1].Title != comparedTitle2 {
			println("Expected groupedSymbols[1].Title to be \"test\" with the type identifier")
			fmt.Printf("%# v\n", pretty.Formatter((*groupedSymbols)[1].Title))
		}
		regexLitGroup = (*groupedSymbols)[0]
		identifierGroup = (*groupedSymbols)[1]
	} else if (*groupedSymbols)[1].Title == comparedTitle1 {
		if (*groupedSymbols)[0].Title != comparedTitle2 {
			println("Expected groupedSymbols[0] to be \"test\" with the type identifier")
		}
		regexLitGroup = (*groupedSymbols)[1]
		identifierGroup = (*groupedSymbols)[0]
	} else {
		println("Expected groupedSymbols[0] or [1] to contain the string \"\\.\" with the type regex literal")
		t.Fatal()
	}

	firstSymbol := *(*identifierGroup.Symbols)[0] 
	title1 := firstSymbol.Title
	if title1.Type != IDENTIFIER_TITLE {
		println("Expected expandedSymbols[0].Title to be an identifier")
		t.Fail()
	}
	if title1.String != "test" {
		println("Expected expandedSymbols[0].Title.String to be \"test\"")
		t.Fail()
	}

	if !reflect.DeepEqual(*firstSymbol.Rule, *ast.Rules["identifier"][0]) {
		println("The first symbol's rule is not matching")
	}

	if firstSymbol.Position != 0 {
		println("Expected expandedSymbols[0].Position to be 0")
	}

	secondSymbol := *(*regexLitGroup.Symbols)[0] 
	title2 := secondSymbol.Title
	if title2.Type != REGEX_LITERAL_TITLE {
		println("Expected expandedSymbols[1].Title to be a regex literal")
		t.Fail()
	}
	if title2.String != "\\." {
		println("Expected expandedSymbols[1].Title.String to be \"\\.\"")
		t.Fail()
	}

	if !reflect.DeepEqual(*secondSymbol.Rule, *ast.Rules["test"][0]) {
		println("The second symbol's rule is not matching")
	}

	if secondSymbol.Position != 0 {
		println("Expected expandedSymbols[1].Position to be 0")
	}

	thirdSymbol := *(*regexLitGroup.Symbols)[1]
	title3 := thirdSymbol.Title
	if title3.Type != REGEX_LITERAL_TITLE {
		println("Expected expandedSymbols[2].Title to be a regex literal")
		t.Fail()
	}

	if title3.String != "\\." {
		println("Expected expandedSymbols[2].Title.String to be \"\\.\"")
		t.Fail()
	}

	if !reflect.DeepEqual(*thirdSymbol.Rule, *ast.Rules["identifier"][1]) {
		println("The third symbol's rule is not matching")
	}

	if thirdSymbol.Position != 0 {
		println("Expected expandedSymbols[2].Position to be 0")
	}
}

func TestBuildParseTableStateTransition(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("identifier{test<\\.>}\ntest{<hello>}\nidentifier{<\\.>}\ntest3{<\\.>}");
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")
		t.Fatal()
	}
	analyzer := initAnalyzer(&ast)
	stateTransitions := analyzer.buildParseTable("identifier")

	if len(*stateTransitions) != 5 {
		println("Expected stateTransitions length to be 5, got " + strconv.Itoa(len(*stateTransitions)))
		fmt.Printf("%# v\n", pretty.Formatter(*stateTransitions))
		t.Fatal()
	}

	if len(*(*stateTransitions)[0].SymbolGroups) != 3 {
		println("Expected the first state transition to contain exactly three groups")
		t.Fail()
	}

	titles := []SymbolTitle{
		{
			String: "test",
			Type: IDENTIFIER_TITLE,
		},
		{
			String: "\\.",
			Type: REGEX_LITERAL_TITLE,
		},
		{
			String: "hello",
			Type: REGEX_LITERAL_TITLE,
		},
	}

	for _, title := range titles {
		isExist := false
		for _, group := range *(*stateTransitions)[0].SymbolGroups {
			if title == group.Title {
				isExist = true	
			}
		}
		if !isExist {
			println("Expected the first state transition to contain groups with title \"test\", \"\\.\", and \"hello\"")	
			t.Fail()
		}
	}

	if len(*(*stateTransitions)[1].SymbolGroups) != 1 {
		println("Expected the second state transition to contain exactly one group")
		t.Fail()
	}

	if len(*(*stateTransitions)[2].SymbolGroups) != 1 {
		println("Expected the third state transition to contain exactly one group")
		t.Fail()
	}

	if len(*(*stateTransitions)[3].SymbolGroups) != 1 {
		println("Expected the fourth state transition to contain exactly one group")
		t.Fail()
	}

	middleTransitions := []*StateTransition{
		(*stateTransitions)[1],
		(*stateTransitions)[2],
		(*stateTransitions)[3],
	}
	titles2 := []SymbolTitle{
		{
			String: "\\.",
			Type: REGEX_LITERAL_TITLE,
		},
		{
			String: "",
			Type: EMPTY_TITLE,
		},
		{
			String: "",
			Type: EMPTY_TITLE,
		},
	}
	for _, title := range titles2 {
		isExist := false
		for _, transition := range middleTransitions {
			if title == (*(*transition).SymbolGroups)[0].Title {
				isExist = true	
			}
		}
		if !isExist {
			println("Expected the second, third, and fourth state transition to contain groups with title \"\\.\", and two empty titles")	
			t.Fail()
		}
	}

	lastSymbol := SymbolTitle {
		String: "",
		Type: EMPTY_TITLE,
	}
	if (*(*stateTransitions)[4].SymbolGroups)[0].Title != lastSymbol {
		println("Expected the fifth state transition to contain group with an empty title")	
		t.Fail()
	}

	if len(*(*stateTransitions)[4].SymbolGroups) != 1 {
		println("Expected the fifth state transition to contain exactly one group")
		t.Fail()
	}
}

func TestBuildParseTable(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("identifier{test<\\.>}\ntest{<\\.>}\nidentifier{<\\.>}");
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")
		t.Fatal()
	}
	analyzer := initAnalyzer(&ast)
	stateTransitions := analyzer.buildParseTable("identifier")

	if len(*stateTransitions) != 4 {
		println("Expected stateTransitions length to be 4, got " + strconv.Itoa(len(*stateTransitions)))
		fmt.Printf("%# v\n", pretty.Formatter(*stateTransitions))
		t.Fatal()
	}
}
