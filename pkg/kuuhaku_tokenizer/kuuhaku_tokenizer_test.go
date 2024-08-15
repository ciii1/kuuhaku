package kuuhaku_tokenizer

import (
	"strconv"
	"testing"

	"github.com/ciii1/kuuhaku/internal/helper"
)

func TestFullTrash(t *testing.T) {
	tokenizer := Init("  \n  #test\n  ");
	token, err := tokenizer.Next()
	helper.Check(err)
	if token.Type != EOF {
		t.Fail()
	}
}

func TestCommentAndIdentifier(t *testing.T) {
	tokenizer := Init("test #test\ntes ");
	token, err := tokenizer.Next()
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
	tokenizer := Init("test9230\ntest30");
	token, err := tokenizer.Next()
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

func TestComment(t *testing.T) {
	tokenizer := Init("test #test\n#test again\ntest");
	token, err := tokenizer.Next()
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
	tokenizer := Init("\"hello\" 'test'");
	token, err := tokenizer.Next()
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
	tokenizer := Init("\"hello\\n\\ttest\\010\" 'test\\t' \"test2\\\"\" \"test2\\\\\"");
	token, err := tokenizer.Next()
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
	tokenizer := Init("\"hello\ntest'test\n'");
	token, err := tokenizer.Next()
	if err != ErrStringLiteralUnterminated {
		println("Expected ErrStringLiteralUnterminated error")
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "test" || token.Type != IDENTIFIER {
		println("Expected \"test\", got \"" + token.Content + "\"")
		t.Fail()
	}
	token, err = tokenizer.Next()
	if err != ErrStringLiteralUnterminated {
		println("Exptected ErrStringLiteralUnterminated error")
		t.Fail()
	}
}

func TestRegexLiteralBasic(t *testing.T) {
	tokenizer := Init("<hello> <test>");
	token, err := tokenizer.Next()
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
	tokenizer := Init("<hello\\n> <test\\t> <test2\\>> <test2\\\\>");
	token, err := tokenizer.Next()
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
	tokenizer := Init("<hello\ntest<test\n>");
	token, err := tokenizer.Next()
	if err != ErrRegexLiteralUnterminated {
		println("Exptected ErrRegexLiteralUnterminated error")
		t.Fail()
	}
	token, err = tokenizer.Next()
	helper.Check(err)
	if token.Content != "test" || token.Type != IDENTIFIER {
		println("Expected \"test\", got \"" + token.Content + "\"")
		t.Fail()
	}
	token, err = tokenizer.Next()
	if err != ErrRegexLiteralUnterminated {
		println("Exptected ErrRegexLiteralUnterminated error")
		t.Fail()
	}
}

func TestCaptureGroup(t *testing.T) {
	tokenizer := Init("$1$2hello$34test");
	token, err := tokenizer.Next()
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
	tokenizer := Init("$hello$34test");
	token, err := tokenizer.Next()
	if err != ErrIllegalCaptureGroup {
		println("Exptected ErrIllegalCaptureGroup")
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
	tokenizer := Init("{}{{==test");
	token, err := tokenizer.Next()
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
}

func TestPosition(t *testing.T) {
	tokenizer := Init("test #test\n#test again\ntest third");
	token, err := tokenizer.Next()
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
