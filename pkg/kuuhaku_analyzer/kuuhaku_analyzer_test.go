package kuuhaku_analyzer

import (
	"errors"
	"testing"

	"github.com/ciii1/kuuhaku/internal/helper"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_parser"
)

func TestErrorUndefinedVariable(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("identifier{test<\\.>}\ntest2{identifier}\ntest3{test4}");
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
		if analyzeError.Position.Column != 12 || analyzeError.Position.Line != 1 {
			println("Expected UndefinedVariableError error with column 12 and line 1")
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
		if analyzeError.Position.Column != 7 || analyzeError.Position.Line != 3 {
			println("Expected UndefinedVariableError error with column 7 and line 3")
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
	
	if startSymbols[0] != "identifier" {
		println("Expected startSymbols[0] to be \"identifier\"")
		t.Fail()
	}

	if startSymbols[1] != "test3" {
		println("Expected startSymbols[1] to be \"test3\"")
		t.Fail()
	}
}
