package kuuhaku_tokenizer

import (
	"errors"
	"strconv"
	"testing"

	"github.com/ciii1/kuuhaku/internal/helper"
)

func TestFullTrash(t *testing.T) {
	tokenizer := Init("  \n  #test\n  ")
	token, _ := tokenizer.Peek()
	if token.Type != EOF {
		t.Fail()
	}
}

func TestCommentAndIdentifier(t *testing.T) {
	tokenizer := Init("test #test\ntes ")
	token, err := tokenizer.Peek()
	helper.Check(err)
	if token.Content != "test" || token.Type != IDENTIFIER {
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "tes" || token.Type != IDENTIFIER {
		t.Fail()
	}
}

func TestIdentifierWithNumber(t *testing.T) {
	tokenizer := Init("test9230\ntest30")
	token, err := tokenizer.Peek()
	helper.Check(err)
	if token.Content != "test9230" || token.Type != IDENTIFIER {
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "test30" || token.Type != IDENTIFIER {
		t.Fail()
	}
}

func TestIdentifierWithLen(t *testing.T) {
	tokenizer := Init("test9230\nlen\nlens")
	token, err := tokenizer.Peek()
	helper.Check(err)
	if token.Content != "test9230" || token.Type != IDENTIFIER {
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "len" || token.Type != LEN_KEYWORD {
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "lens" || token.Type != IDENTIFIER {
		t.Fail()
	}
}

func TestSearchMode(t *testing.T) {
	tokenizer := Init("SEARCH_MODE a9230\nSEARCH_MODE2\nSEARCH_MODE")
	token, err := tokenizer.Peek()
	helper.Check(err)
	if token.Content != "SEARCH_MODE" || token.Type != SEARCH_MODE_KEYWORD {
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "a9230" || token.Type != IDENTIFIER {
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "SEARCH_MODE2" || token.Type != IDENTIFIER {
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "SEARCH_MODE" || token.Type != SEARCH_MODE_KEYWORD {
		t.Fail()
	}
}

func TestPatternUnrecognizedError(t *testing.T) {
	tokenizer := Init("test@\nlen%")
	token, err := tokenizer.Peek()
	helper.Check(err)
	if token.Content != "test" || token.Type != IDENTIFIER {
		t.Fail()
	}

	token, err = tokenizer.Next()
	var tokenizeError *TokenizeError
	if errors.As(err, &tokenizeError) {
		if tokenizeError.Type != PATTERN_UNRECOGNIZED {
			println("Expected PatternUnrecognizedErr")
			t.Fail()
		}
	} else {
		println("Expected TokenizeError")
		t.Fail()
	}

	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "len" || token.Type != LEN_KEYWORD {
		t.Fail()
	}

	token, err = tokenizer.Next()
	if errors.As(err, &tokenizeError) {
		if tokenizeError.Type != PATTERN_UNRECOGNIZED {
			println("Expected PatternUnrecognizedErr")
			t.Fail()
		}
	} else {
		println("Expected TokenizeError")
		t.Fail()
	}
}

func TestIdentifierWithCaptureGroup(t *testing.T) {
	tokenizer := Init("len$1\n$32len")
	token, err := tokenizer.Peek()
	helper.Check(err)
	if token.Content != "len" || token.Type != LEN_KEYWORD {
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "1" || token.Type != CAPTURE_GROUP {
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "32" || token.Type != CAPTURE_GROUP {
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "len" || token.Type != LEN_KEYWORD {
		t.Fail()
	}
}

func TestComment(t *testing.T) {
	tokenizer := Init("test #test\n#test again\ntest")
	token, err := tokenizer.Peek()
	helper.Check(err)
	if token.Content != "test" || token.Type != IDENTIFIER {
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "test" || token.Type != IDENTIFIER {
		t.Fail()
	}
}

func TestStringLiteralBasic(t *testing.T) {
	tokenizer := Init("\"hello\" 'test'")
	token, err := tokenizer.Peek()
	helper.Check(err)
	if token.Content != "hello" || token.Type != STRING_LITERAL {
		println("Expected \"hello\", got \"" + token.Content + "\"")
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "test" || token.Type != STRING_LITERAL {
		println("Expected \"test\", got \"" + token.Content + "\"")
		t.Fail()
	}
}

func TestStringLiteralEscapes(t *testing.T) {
	tokenizer := Init("\"hello\\n\\ttest\\010\" 'test\\t' \"test2\\\"\" \"test2\\\\\"")
	token, err := tokenizer.Peek()
	helper.Check(err)
	if token.Content != "hello\n\ttest\010" || token.Type != STRING_LITERAL {
		println("Expected \"hello\n\ttest\010\", got \"" + token.Content + "\"")
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "test\t" || token.Type != STRING_LITERAL {
		println("Expected \"test\t\", got \"" + token.Content + "\"")
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "test2\"" || token.Type != STRING_LITERAL {
		println("Expected \"test2\", got \"" + token.Content + "\"")
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "test2\\" || token.Type != STRING_LITERAL {
		println("Expected \"test2\\\", got \"" + token.Content + "\"")
		t.Fail()
	}
}

func TestStringLiteralUnterminated(t *testing.T) {
	tokenizer := Init("\"hello\ntest'test\n'")

	token, err := tokenizer.Peek()
	var tokenizeError *TokenizeError
	if errors.As(err, &tokenizeError) {
		if tokenizeError.Type != STRING_LITERAL_UNTERMINATED {
			println("Expected StringLiteralUnterminatedError")
			t.Fail()
		}
	} else {
		println("Expected TokenizeError")
		t.Fail()
	}

	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "test" || token.Type != IDENTIFIER {
		println("Expected \"test\", got \"" + token.Content + "\"")
		t.Fail()
	}
	token, err = tokenizer.Next()
	if errors.As(err, &tokenizeError) {
		if tokenizeError.Type != STRING_LITERAL_UNTERMINATED {
			println("Expected StringLiteralUnterminatedError")
			t.Fail()
		}
	} else {
		println("Expected TokenizeError")
		t.Fail()
	}
}

func TestRegexLiteralBasic(t *testing.T) {
	tokenizer := Init("<hello> <test>")
	token, err := tokenizer.Peek()
	helper.Check(err)
	if token.Content != "hello" || token.Type != REGEX_LITERAL {
		println("Expected \"hello\", got \"" + token.Content + "\"")
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "test" || token.Type != REGEX_LITERAL {
		println("Expected \"test\", got \"" + token.Content + "\"")
		t.Fail()
	}
}

func TestRegexLiteralEscapes(t *testing.T) {
	tokenizer := Init("<hello\\n> <test\\t> <test2\\>> <test2\\\\>")
	token, err := tokenizer.Peek()
	helper.Check(err)
	if token.Content != "hello\\n" || token.Type != REGEX_LITERAL {
		println("Expected \"hello\\n\", got \"" + token.Content + "\"")
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "test\\t" || token.Type != REGEX_LITERAL {
		println("Expected \"test\\t\", got \"" + token.Content + "\"")
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "test2>" || token.Type != REGEX_LITERAL {
		println("Expected \"test2>\", got \"" + token.Content + "\"")
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "test2\\\\" || token.Type != REGEX_LITERAL {
		println("Expected \"test2\\\\\", got \"" + token.Content + "\"")
		t.Fail()
	}
}

func TestRegexLiteralUnterminated(t *testing.T) {
	tokenizer := Init("<hello\ntest<test\n>")
	token, err := tokenizer.Peek()
	var tokenizeError *TokenizeError
	if errors.As(err, &tokenizeError) {
		if tokenizeError.Type != REGEX_LITERAL_UNTERMINATED {
			println("Expected RegexLiteralUnterminatedError")
			t.Fail()
		}
	} else {
		println("Expected TokenizeError")
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "test" || token.Type != IDENTIFIER {
		println("Expected \"test\", got \"" + token.Content + "\"")
		t.Fail()
	}
	token, err = tokenizer.Next()
	if errors.As(err, &tokenizeError) {
		if tokenizeError.Type != REGEX_LITERAL_UNTERMINATED {
			println("Expected RegexLiteralUnterminatedError")
			t.Fail()
		}
	} else {
		println("Expected TokenizeError")
		t.Fail()
	}
}

func TestCaptureGroup(t *testing.T) {
	tokenizer := Init("$1$2hello$34test")
	token, err := tokenizer.Peek()
	helper.Check(err)
	if token.Content != "1" || token.Type != CAPTURE_GROUP {
		println("Exptected 1, got " + token.Content)
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "2" || token.Type != CAPTURE_GROUP {
		println("Exptected 2, got " + token.Content)
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "hello" || token.Type != IDENTIFIER {
		println("Exptected hello, got " + token.Content)
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "34" || token.Type != CAPTURE_GROUP {
		println("Exptected 34, got " + token.Content)
		t.Fail()
	}
}

func TestIllegalCaptureGroup(t *testing.T) {
	tokenizer := Init("$hello$34test")
	token, err := tokenizer.Peek()
	var tokenizeError *TokenizeError
	if errors.As(err, &tokenizeError) {
		if tokenizeError.Type != ILLEGAL_CAPTURE_GROUP {
			println("Exptected IllegalCaptureGroupError")
			t.Fail()
		}
	} else {
		println("Expected TokenizeError")
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "hello" || token.Type != IDENTIFIER {
		println("Exptected hello, got " + token.Content)
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "34" || token.Type != CAPTURE_GROUP {
		println("Exptected 34, got " + token.Content)
		t.Fail()
	}
}

func TestSigns(t *testing.T) {
	tokenizer := Init("{}{{==test(),,(")
	token, err := tokenizer.Peek()
	helper.Check(err)
	if token.Content != "{" || token.Type != OPENING_CURLY_BRACKET {
		println("Exptected {, got " + token.Content)
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "}" || token.Type != CLOSING_CURLY_BRACKET {
		println("Exptected }, got " + token.Content)
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "{" || token.Type != OPENING_CURLY_BRACKET {
		println("Exptected {, got " + token.Content)
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "{" || token.Type != OPENING_CURLY_BRACKET {
		println("Exptected {, got " + token.Content)
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "=" || token.Type != EQUAL_SIGN {
		println("Exptected =, got " + token.Content)
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "=" || token.Type != EQUAL_SIGN {
		println("Exptected =, got " + token.Content)
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "test" || token.Type != IDENTIFIER {
		println("Exptected test, got " + token.Content)
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "(" || token.Type != OPENING_BRACKET {
		println("Exptected (, got " + token.Content)
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != ")" || token.Type != CLOSING_BRACKET {
		println("Exptected ), got " + token.Content)
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "," || token.Type != COMMA {
		println("Exptected ',', got " + token.Content)
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "," || token.Type != COMMA {
		println("Exptected ',', got " + token.Content)
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "(" || token.Type != OPENING_BRACKET {
		println("Exptected (, got " + token.Content)
		t.Fail()
	}
}

func TestPosition(t *testing.T) {
	tokenizer := Init("test #test\n#test again\ntest third")
	token, err := tokenizer.Peek()
	helper.Check(err)
	if token.Position.Raw != 0 || token.Position.Column != 1 || token.Position.Line != 1 {
		println("\n" + token.Content)
		println("Raw: " + strconv.Itoa(token.Position.Raw) + ", expected: 0")
		println("Column: " + strconv.Itoa(token.Position.Column) + ", expected: 1")
		println("Line: " + strconv.Itoa(token.Position.Column) + ", expected: 1")
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Position.Raw != 23 || token.Position.Column != 1 || token.Position.Line != 3 {
		println("\n" + token.Content)
		println("Raw: " + strconv.Itoa(token.Position.Raw) + ", expected: 23")
		println("Column: " + strconv.Itoa(token.Position.Column) + ", expected: 1")
		println("Line: " + strconv.Itoa(token.Position.Line) + ", expected: 3")
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Position.Raw != 28 || token.Position.Column != 6 || token.Position.Line != 3 {
		println("\n" + token.Content)
		println("Raw: " + strconv.Itoa(token.Position.Raw) + ", expected: 28")
		println("Column: " + strconv.Itoa(token.Position.Column) + ", expected: 6")
		println("Line: " + strconv.Itoa(token.Position.Line) + ", expected: 3")
		t.Fail()
	}
}
