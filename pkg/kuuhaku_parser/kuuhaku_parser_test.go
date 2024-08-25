package kuuhaku_parser

import (
	"errors"
	"testing"

	"github.com/ciii1/kuuhaku/pkg/kuuhaku_tokenizer"
	"github.com/ciii1/kuuhaku/internal/helper"
)

func TestConsumeMatchRules(t *testing.T) {
	parser := Init("<.*>hello<[0-9]>hi=");
	matchRulesP := parser.consumeMatchRules()
	if len(parser.Errors) != 0 {
		panic(parser.Errors)
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

	node2, ok := matchRules[1].(Identifer)
	if ok == false {
		println("Expected matchRules[1] to be an identifier")
		t.Fail()
	} else if node2.Name != "hello" {
		println("Expected matchRules[1] to contain \"hello\"")
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

	node4, ok := matchRules[3].(Identifer)
	if ok == false {
		println("Expected matchRules[3] to be an identifier")
		t.Fail()
	} else if node4.Name != "hi" {
		println("Expected matchRules[3] to contain \"hi\"")
		t.Fail()
	}
}

func TestConsumeReplaceRules(t *testing.T) {
	parser := Init("\"\\t\"len$0$2 \"hi\"");
	replaceRulesP := parser.consumeReplaceRules()
	if len(parser.Errors) != 0 {
		panic(parser.Errors)
	}
	replaceRules := *replaceRulesP
	if len(replaceRules) != 3 {
		println("Expected replaceRules length to be 3")
		t.Fail()
	}
	node1, ok := replaceRules[0].(Len)
	if ok == false {
		println("Expected matchRules[0] to be len")
		t.Fail()
	} else if firstArg, ok := node1.FirstArgument.(StringLiteral); ok {
		if firstArg.String != "\t" {
			println("Expected Len.FirstArgument to contain \"\\t\"")
			t.Fail()
		}
		if secondArg, ok := node1.SecondArgument.(CaptureGroup); ok {
			if secondArg.Number != 0 {
				println("Expected Len.SecondArgument to contain 0")
				t.Fail()
			}
		} else {
			println("Expected Len.SecondArgument to be a capture group")
			t.Fail()
		}
	} else {
		println("Expected Len.FirstArgument to be a string literal")
		t.Fail()
	}

	node2, ok := replaceRules[1].(CaptureGroup)
	if ok == false {
		println("Expected matchRules[1] to be a capture group")
		t.Fail()
	} else if node2.Number != 2 {
		println("Expected matchRules[1] to contain 2")
		t.Fail()
	}

	node3, ok := replaceRules[2].(StringLiteral)
	if ok == false {
		println("Expected matchRules[2] to be a string literal")
		t.Fail()
	} else if node3.String != "hi" {
		println("Expected matchRules[2] to contain \"hi\"")
		t.Fail()
	}
}

func TestErrorConsumeReplaceRules(t *testing.T) {
	parser := Init("\"test\nlen test\nlen$1");
	parser.consumeReplaceRules()
	token, _ := parser.tokenizer.Next()
	if token.Type != kuuhaku_tokenizer.EOF {
		println("Expected the parser to reach EOF, got token with content " + token.Content)
		token, _ := parser.tokenizer.Next() 
		println("Next content is " + token.Content)
		t.Fatal()
	}
	println("TestErrorConsumeReplaceRules - All errors:")
	helper.DisplayAllErrors(parser.Errors)

	var tokenizeError *kuuhaku_tokenizer.TokenizeError
	if errors.As(parser.Errors[0], &tokenizeError) {
		if tokenizeError.Type != kuuhaku_tokenizer.STRING_LITERAL_UNTERMINATED {
			println("Expected ErrStringLiteralUnterminated error")
			t.Fail()
		}
	} else {
		println("Expected TokenizeError")
		t.Fail()
	}

	var parseError *ParseError
	if errors.As(parser.Errors[1], &parseError) {
		if parseError.Type != LEN_ARGUMENT_INVALID {
			println("Expected ErrLenArgumentInvalid error")
			t.Fail()
		}
		if parseError.Position.Column != 5 || parseError.Position.Line != 2 {
			println("Expected ErrLenArgumentInvalid error with column 5, line 2")
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}

	if errors.As(parser.Errors[2], &parseError) {
		if parseError.Type != UNEXPECTED_LEN {
			println("Expected ErrUnexpectedLen error")
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}

	if errors.As(parser.Errors[3], &parseError) {
		if parseError.Type != UNEXPECTED_LEN {
			println("Expected ErrUnexpectedLen error at the last len")
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}
}

func TestConsumeRule(t *testing.T) {
	parser := Init("test{\nidentifier\n=\n\"\\t\"$0}");
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
	if len(rule.ReplaceRules) != 2 {
		println("Expected len(ReplaceRules) to be 2")
		t.Fail()
	}
	node2, ok := rule.ReplaceRules[0].(StringLiteral)
	if !ok {
		println("Expected ReplaceRules[0] to be a string literal")
		t.Fail()
	}
	if node2.String != "\t" {
		println("Expected ReplaceRules[0].String to be \"\t\"")
		t.Fail()
	}
	node3, ok := rule.ReplaceRules[1].(CaptureGroup)
	if !ok {
		println("Expected ReplaceRules[1] to be a capture group")
		t.Fail()
	}
	if node3.Number != 0 {
		println("Expected ReplaceRules[1].Number to be 0")
		t.Fail()
	}
}

func TestErrorConsumeRule(t *testing.T) {
	parser := Init("test{\"test2\"=len$1}\nanotherTest{}");
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
	parser := Init("test\ntest{\"test2\"=len$1}\ntest{test\"hello\"}\ntest{test=len$1}test{\"test}");
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

	if errors.As(parser.Errors[3], &parseError) {
		if parseError.Type != UNEXPECTED_LEN {
			println("Expected UnexpectedLenError error")
			t.Fail()
		}
		if parseError.Position.Column != 11 || parseError.Position.Line != 4 {
			println("Expected UnexpectedLenError error with column 11 and line 4")
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}

	var tokenizeError *kuuhaku_tokenizer.TokenizeError
	if errors.As(parser.Errors[4], &tokenizeError) {
		if tokenizeError.Type != kuuhaku_tokenizer.STRING_LITERAL_UNTERMINATED {
			println("Expected StringLiteralUnterminatedError error")
			t.Fail()
		}
		if tokenizeError.Position.Column != 28 || tokenizeError.Position.Line != 4 {
			println("Expected StringLiteralUnterminatedError error with column 28 and line 4")
			t.Fail()
		}
	} else {
		println("Expected TokenizeError")
		t.Fail()
	}

	if errors.As(parser.Errors[5], &parseError) {
		if parseError.Type != EXPECTED_MATCH_RULE {
			println("Expected ExpectedMatchRuleError error")
			t.Fail()
		}
		if parseError.Position.Column != 22 || parseError.Position.Line != 4 {
			println("Expected ExpectedMatchRuleError error with column 22 and line 4")
			t.Fail()
		}
	} else {
		println("Expected ParseError")
		t.Fail()
	}

}

func TestConsumeInput(t *testing.T) {
	parser := Init("test{identifier=\"\\t\"len$0}\nidentifier{<[a-zA-Z]>}\nidentifier{<[a-zA-Z][0-9]>}");
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
	if len(ast.Rules["test"]) != 1 {
		println("Expected len(ast.Rules[\"test\"]) to be 1")
		t.Fatal()
	}
}

func TestConsumeSearchMode(t *testing.T) {
	parser := Init("SEARCH_MODE test{identifier=\"\\t\"len$0}\nidentifier{<[a-zA-Z]>}\nidentifier{<[a-zA-Z][0-9]>}");
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

	if (ast.IsSearchMode != true) {
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
	parser := Init("1 SEARCH_MODE test{identifier=\"\\t\"len$0}\nidentifier{<[a-zA-Z]>}\nidentifier{<[a-zA-Z][0-9]>}");
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
		println("TestConsumeInput - All errors:")
		helper.DisplayAllErrors(parser.Errors)
		t.Fatal()
	}

	if (ast.IsSearchMode != false) {
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
	parser := Init("test{\"test2\"=len$1}\n\"test\"test\nidentifier<test>");
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
	parser := Init("test{\"test2\"\\=len$1}\n\"test\"@test\nidentifier<test>");
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
