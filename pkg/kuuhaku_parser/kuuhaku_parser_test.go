package kuuhaku_parser

import (
	"errors"
	"strconv"
	"testing"

	"github.com/ciii1/kuuhaku/internal/helper"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_tokenizer"
)

func TestConsumeMatchRules(t *testing.T) {
	parser := initParser("<.*><[0-9]><[0-9]><10>=")
	matchRulesP := parser.consumeMatchRules()
	if len(parser.Errors) != 0 {
		println("TestConsumeMatchRules - All errors:")
		helper.DisplayAllErrors(parser.Errors)
	}
	matchRules := *matchRulesP
	if len(matchRules) != 4 {
		println("Expected matchRules length to be 4")
		t.Fail()
	}
	node1, ok := matchRules[0].(RegexLiteral)
	if ok == false {
		println("Expected matchRules[0] to be a regex literal")
		t.Fail()
	} else if node1.RegexString != ".*" {
		println("Expected matchRules[0] to contain \".*\"")
		t.Fail()
	}

	node2, ok := matchRules[1].(RegexLiteral)
	if ok == false {
		println("Expected matchRules[1] to be a regex literal")
		t.Fail()
	} else if node2.RegexString != "[0-9]" {
		println("Expected matchRules[1] to contain \"[0-9]\"")
		t.Fail()
	}

	node3, ok := matchRules[2].(RegexLiteral)
	if ok == false {
		println("Expected matchRules[2] to be a regex literal")
		t.Fail()
	} else if node3.RegexString != "[0-9]" {
		println("Expected matchRules[2] to contain \"[0-9]\"")
		t.Fail()
	}

	node4, ok := matchRules[3].(RegexLiteral)
	if ok == false {
		println("Expected matchRules[3] to be a regex literal")
		t.Fail()
	} else if node4.RegexString != "10" {
		println("Expected matchRules[3] to contain \"[0-9]\"")
		t.Fail()
	}
}

func TestConsumeMatchRules2(t *testing.T) {
	parser := initParser("hello hello2=")
	matchRulesP := parser.consumeMatchRules()
	if len(parser.Errors) != 0 {
		panic(parser.Errors)
	}
	matchRules := *matchRulesP
	if len(matchRules) != 2 {
		println("Expected matchRules length to be 2")
		t.Fail()
	}
	node1, ok := matchRules[0].(Identifer)
	if ok == false {
		println("Expected matchRules[0] to be an identifier")
		t.Fail()
	} else if node1.Name != "hello" {
		println("Expected matchRules[0] name's to be \"hello\"")
		t.Fail()
	}

	node2, ok := matchRules[1].(Identifer)
	if ok == false {
		println("Expected matchRules[1] to be an identifier")
		t.Fail()
	} else if node2.Name != "hello2" {
		println("Expected matchRules[1] name's to be \"hello2\"")
		t.Fail()
	}
}

func TestConsumeMatchRulesError(t *testing.T) {
	parser := initParser("hello hello2 <2> hello2=")
	parser.consumeMatchRules()
	if len(parser.Errors) != 1 {
		println("Expected parser errors length to be 1")
		t.Fail()
	}
	var parseError *ParseError
	if errors.As(parser.Errors[0], &parseError) {
		if parseError.Type != MIXED_TYPE_MATCH_RULE {
			println("Expected ErrMixedTypeMatchRule error")
			t.Fail()
		}
		if parseError.Position.Column != 14 || parseError.Position.Line != 1 {
			println("Expected error position to be (1, 14), got (" + strconv.Itoa(parseError.Position.Line) + "," + strconv.Itoa(parseError.Position.Column))
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}
}

func TestConsumeMatchRulesError2(t *testing.T) {
	parser := initParser("<2>hello=")
	parser.consumeMatchRules()
	if len(parser.Errors) != 1 {
		println("Expected parser errors length to be 1")
		t.Fail()
	}
	var parseError *ParseError
	if errors.As(parser.Errors[0], &parseError) {
		if parseError.Type != MIXED_TYPE_MATCH_RULE {
			println("Expected ErrMixedTypeMatchRule error")
			t.Fail()
		}
		if parseError.Position.Column != 4 || parseError.Position.Line != 1 {
			println("Expected error position to be (1, 4), got (" + strconv.Itoa(parseError.Position.Line) + "," + strconv.Itoa(parseError.Position.Column))
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}
}

func TestConsumeRule(t *testing.T) {
	parser := initParser("test{\nidentifier\n=\n``test=10; return test``\n}")
	rule := parser.consumeRule()
	if len(parser.Errors) != 0 {
		println("Expected len(parser.Errors) to be 0")
		println("TestConsumeRule - All errors:")
		helper.DisplayAllErrors(parser.Errors)
		t.Fatal()
	}
	if len(rule.MatchRules) != 1 {
		println("Expected len(MatchRules) to be 1")
		t.Fatal()
	}
	node1, ok := rule.MatchRules[0].(Identifer)
	if !ok {
		println("Expected MatchRules[0] to be an identifier")
		t.Fail()
	}
	if node1.Name != "identifier" {
		println("Expected MatchRules[0].Name to be \"identifier\"")
		t.Fail()
	}
	if rule.ReplaceRule.LuaString != "test=10; return test" {
		println("Expected replace rules to contain \"test=10; return test\", got " + rule.ReplaceRule.LuaString)
		t.Fail()
	}
}

func TestErrorConsumeRule(t *testing.T) {
	parser := initParser("test{``test2``=<test>}\nanotherTest{}")
	parser.consumeRule()
	parser.consumeRule()
	token, _ := parser.tokenizer.Next()
	if token.Type != kuuhaku_tokenizer.EOF {
		println("Expected the parser to reach EOF, got token with content " + token.Content)
		token, _ := parser.tokenizer.Next()
		println("Next content is " + token.Content)
		t.Fatal()
	}

	println("TestErrorConsumeRule - All errors:")
	helper.DisplayAllErrors(parser.Errors)

	var parseError *ParseError
	if errors.As(parser.Errors[0], &parseError) {
		if parseError.Type != EXPECTED_MATCH_RULE {
			println("Expected ExpectedMatchRuleError error")
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}

	if errors.As(parser.Errors[1], &parseError) {
		if parseError.Type != EXPECTED_MATCH_RULE {
			println("Expected ExpectedMatchRuleError error")
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}
}

func TestErrorPosition(t *testing.T) {
	parser := initParser("test\ntest{``est``=``n``}\ntest{test``ell``}\ntest{test=``n``}test{``est}")
	parser.consumeRule()
	parser.consumeRule()
	parser.consumeRule()
	parser.consumeRule()
	parser.consumeRule()
	token, _ := parser.tokenizer.Next()
	if token != nil && token.Type != kuuhaku_tokenizer.EOF {
		println("Expected the parser to reach EOF, got token with content " + token.Content)
		token, _ := parser.tokenizer.Next()
		if token != nil {
			println("Next content is " + token.Content)
		} else {
			println("Next content is nil")
		}
		t.Fatal()
	}

	println("TestErrorPosition - All errors:")
	helper.DisplayAllErrors(parser.Errors)

	var parseError *ParseError
	if errors.As(parser.Errors[0], &parseError) {
		if parseError.Type != EXPECTED_OPENING_CURLY_BRACKET {
			println("Expected ExpectedOpeningCurlyBracketError error")
			t.Fail()
		}
		if parseError.Position.Column != 1 || parseError.Position.Line != 2 {
			println("Expected ExpectedOpeningCurlyBracketError error with column 1 and line 2")
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}

	if errors.As(parser.Errors[1], &parseError) {
		if parseError.Type != EXPECTED_MATCH_RULE {
			println("Expected ExpectedMatchRuleError error")
			t.Fail()
		}
		if parseError.Position.Column != 6 || parseError.Position.Line != 2 {
			println("Expected ExpectedMatchRuleError error with column 6 and line 2")
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}

	if errors.As(parser.Errors[2], &parseError) {
		if parseError.Type != EXPECTED_EQUAL_SIGN {
			println("Expected ExpectedEqualSignError error")
			t.Fail()
		}
		if parseError.Position.Column != 10 || parseError.Position.Line != 3 {
			println("Expected ExpectedEqualSignError error with column 10 and line 3")
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}

	var tokenizeError *kuuhaku_tokenizer.TokenizeError
	if errors.As(parser.Errors[3], &tokenizeError) {
		if tokenizeError.Type != kuuhaku_tokenizer.LUA_LITERAL_UNTERMINATED {
			println("Expected LuaLiteralUnterminatedError error")
			t.Fail()
		}
		if tokenizeError.Position.Column != 22 || tokenizeError.Position.Line != 4 {
			println("Expected LuaLiteralUnterminatedError error with column 22 and line 4")
			t.Fail()
		}
	} else {
		println("Expected TokenizeError")
		t.Fail()
	}
}

func TestConsumeInput(t *testing.T) {
	parser := initParser("test{identifier=``hello``}\nidentifier{<[a-zA-Z]>}\nidentifier{<[a-zA-Z][0-9]>}")
	ast := parser.consumeInput()
	token, _ := parser.tokenizer.Next()
	if token.Type != kuuhaku_tokenizer.EOF {
		println("Expected the parser to reach EOF, got token with content " + token.Content)
		token, _ := parser.tokenizer.Next()
		println("Next content is " + token.Content)
		t.Fatal()
	}

	if len(parser.Errors) != 0 {
		println("Expected len(parser.Errors) to be 0")
		println("TestConsumeInput - All errors:")
		helper.DisplayAllErrors(parser.Errors)
		t.Fatal()
	}
	if len(ast.Rules) != 2 {
		println("Expected len(ast.Rules) to be 2")
		t.Fatal()
	}
	if len(ast.Rules["identifier"]) != 2 {
		println("Expected len(ast.Rules[\"identifier\"]) to be 2")
		t.Fatal()
	}
	if ast.Rules["identifier"][0].Order != 1 {
		got := strconv.Itoa(ast.Rules["identifier"][0].Order)
		println("Expected ast.Rules[\"identifier\"][0].Order to be 1, got " + got)
		t.Fail()
	}
	if ast.Rules["identifier"][1].Order != 2 {
		got := strconv.Itoa(ast.Rules["identifier"][1].Order)
		println("Expected ast.Rules[\"identifier\"][1].Order to be 2, got " + got)
		t.Fail()
	}

	if len(ast.Rules["test"]) != 1 {
		println("Expected len(ast.Rules[\"test\"]) to be 1")
		t.Fatal()
	}
	if ast.Rules["test"][0].Order != 0 {
		got := strconv.Itoa(ast.Rules["test"][0].Order)
		println("Expected ast.Rules[\"test\"][0].Order to be 0, got " + got)
		t.Fatal()
	}
}

func TestConsumeSearchMode(t *testing.T) {
	parser := initParser("SEARCH_MODE test{identifier=``allen``}\nidentifier{<[a-zA-Z]>}\nidentifier{<[a-zA-Z][0-9]>}")
	ast := parser.consumeInput()
	token, _ := parser.tokenizer.Next()
	if token.Type != kuuhaku_tokenizer.EOF {
		println("Expected the parser to reach EOF, got token with content " + token.Content)
		token, _ := parser.tokenizer.Next()
		println("Next content is " + token.Content)
		t.Fatal()
	}

	if len(parser.Errors) != 0 {
		println("Expected len(parser.Errors) to be 0")
		println("TestConsumeInput - All errors:")
		helper.DisplayAllErrors(parser.Errors)
		t.Fatal()
	}

	if ast.IsSearchMode != true {
		println("Expected ast.IsSearchMode to be true")
		t.Fatal()
	}

	if len(ast.Rules) != 2 {
		println("Expected len(ast.Rules) to be 2")
		t.Fatal()
	}
	if len(ast.Rules["identifier"]) != 2 {
		println("Expected len(ast.Rules[\"identifier\"]) to be 2")
		t.Fatal()
	}
	if len(ast.Rules["test"]) != 1 {
		println("Expected len(ast.Rules[\"test\"]) to be 1")
		t.Fatal()
	}
}

func TestConsumeSearchModeError(t *testing.T) {
	parser := initParser("1 SEARCH_MODE test{identifier=``allen``}\nidentifier{<[a-zA-Z]>}\nidentifier{<[a-zA-Z][0-9]>}")
	ast := parser.consumeInput()
	token, _ := parser.tokenizer.Next()
	if token.Type != kuuhaku_tokenizer.EOF {
		println("Expected the parser to reach EOF, got token with content " + token.Content)
		token, _ := parser.tokenizer.Next()
		println("Next content is " + token.Content)
		t.Fatal()
	}

	if len(parser.Errors) != 2 {
		println("Expected len(parser.Errors) to be 2")
		println("TestConsumeSearchModeError - All errors:")
		helper.DisplayAllErrors(parser.Errors)
		t.Fatal()
	}

	if ast.IsSearchMode != false {
		println("Expected ast.IsSearchMode to be false")
		t.Fatal()
	}

	if len(ast.Rules) != 2 {
		println("Expected len(ast.Rules) to be 2")
		t.Fatal()
	}
	if len(ast.Rules["identifier"]) != 2 {
		println("Expected len(ast.Rules[\"identifier\"]) to be 2")
		t.Fatal()
	}
	if len(ast.Rules["test"]) != 1 {
		println("Expected len(ast.Rules[\"test\"]) to be 1")
		t.Fatal()
	}
}

func TestErrorConsumeInput(t *testing.T) {
	parser := initParser("test{``est``=``n``}\n``es``test\nidentifier<test>")
	parser.consumeInput()
	token, _ := parser.tokenizer.Next()
	if token.Type != kuuhaku_tokenizer.EOF {
		println("Expected the parser to reach EOF, got token with content " + token.Content)
		token, _ := parser.tokenizer.Next()
		println("Next content is " + token.Content)
		t.Fatal()
	}

	println("TestErrorConsumeInput - All errors:")
	helper.DisplayAllErrors(parser.Errors)

	var parseError *ParseError
	if errors.As(parser.Errors[0], &parseError) {
		if parseError.Type != EXPECTED_MATCH_RULE {
			println("Expected ExpectedMatchRuleError error")
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}

	if errors.As(parser.Errors[1], &parseError) {
		if parseError.Type != EXPECTED_RULE {
			println("Expected ExpectedRuleError error")
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}

	if errors.As(parser.Errors[2], &parseError) {
		if parseError.Type != EXPECTED_OPENING_CURLY_BRACKET {
			println("Expected ExpectedOpeningCurlyBracketError error")
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}

	if errors.As(parser.Errors[3], &parseError) {
		if parseError.Type != EXPECTED_OPENING_CURLY_BRACKET {
			println("Expected ExpectedOpeningCurlyBracketError error")
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}

	if errors.As(parser.Errors[4], &parseError) {
		if parseError.Type != EXPECTED_RULE {
			println("Expected ExpectedRuleError error")
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}
}

func TestErrorTokenizeError(t *testing.T) {
	parser := initParser("test{``est``\\=``n``}\n<test>@test\nidentifier<test>")
	parser.consumeInput()
	token, _ := parser.tokenizer.Next()
	if token.Type != kuuhaku_tokenizer.EOF {
		println("Expected the parser to reach EOF, got token with content " + token.Content)
		token, _ := parser.tokenizer.Next()
		println("Next content is " + token.Content)
		t.Fatal()
	}

	println("TestErrorTokenizeError - All errors:")
	helper.DisplayAllErrors(parser.Errors)

	var parseError *ParseError
	if errors.As(parser.Errors[0], &parseError) {
		if parseError.Type != EXPECTED_MATCH_RULE {
			println("Expected ExpectedMatchRuleError error")
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}

	var tokenizeError *kuuhaku_tokenizer.TokenizeError
	if errors.As(parser.Errors[1], &tokenizeError) {
		if tokenizeError.Type != kuuhaku_tokenizer.PATTERN_UNRECOGNIZED {
			println("Expected PatternUnrecognizedError error")
			t.Fail()
		}
	} else {
		println("Expected TokenizeError")
		t.Fail()
	}

	if errors.As(parser.Errors[2], &parseError) {
		if parseError.Type != EXPECTED_RULE {
			println("Expected ExpectedRuleError error")
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}

	if errors.As(parser.Errors[3], &tokenizeError) {
		if tokenizeError.Type != kuuhaku_tokenizer.PATTERN_UNRECOGNIZED {
			println("Expected PatternUnrecognizedError error")
			t.Fail()
		}
	} else {
		println("Expected TokenizeError")
		t.Fail()
	}

	if errors.As(parser.Errors[4], &parseError) {
		if parseError.Type != EXPECTED_OPENING_CURLY_BRACKET {
			println("Expected ExpectedOpeningCurlyBracketError error")
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}

	if errors.As(parser.Errors[5], &parseError) {
		if parseError.Type != EXPECTED_OPENING_CURLY_BRACKET {
			println("Expected ExpectedOpeningCurlyBracketError error")
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}

	if errors.As(parser.Errors[6], &parseError) {
		if parseError.Type != EXPECTED_RULE {
			println("Expected ExpectedRuleError error")
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}
}
