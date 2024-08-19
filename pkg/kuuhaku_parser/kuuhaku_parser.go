package kuuhaku_parser

import (
	"fmt"
	"strconv"

	"github.com/ciii1/kuuhaku/internal/helper"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_tokenizer"
)

type ParseErrorType int

const (
	LEN_ARGUMENT_INVALID = iota
	UNEXPECTED_LEN
)

type ParseError struct {
	Position kuuhaku_tokenizer.Position	
	Message string
	Type ParseErrorType
}

func (e ParseError) Error() string {
	return fmt.Sprintf("Parse error (%d, %d): %s", e.Position.Line, e.Position.Column, e.Message)
}

func ErrLenArgumentInvalid(tokenizer *kuuhaku_tokenizer.Tokenizer) *ParseError {
	return &ParseError {
		Message: "Len argument is invalid",
		Position: tokenizer.Position,
		Type: LEN_ARGUMENT_INVALID,
	}
}
func ErrUnexpectedLen(tokenizer *kuuhaku_tokenizer.Tokenizer) *ParseError {
	return &ParseError {
		Message: "Usage of len here is invalid",
		Position: tokenizer.Position,
		Type: UNEXPECTED_LEN,
	}
}

type Parser struct {
	tokenizer kuuhaku_tokenizer.Tokenizer
	Errors []error
}

func Parse(input string) error {
	tokenizer := kuuhaku_tokenizer.Init(input);
	token, err := tokenizer.Next()
	helper.Check(err)
	for token.Type != kuuhaku_tokenizer.EOF {
		fmt.Println("\n\"" + token.Content + "\"")
		fmt.Println("\t-Column: " + strconv.Itoa(token.Position.Column))
		fmt.Println("\t-Line: " + strconv.Itoa(token.Position.Line))
		token, err = tokenizer.Next()
		helper.Check(err)
	}
	return nil
}

func Init(input string) Parser {
	return Parser {
		tokenizer: kuuhaku_tokenizer.Init(input),
		Errors: []error{},
	}
}

func (parser *Parser) consumeReplaceRules() *[]ReplaceRule {
	var output []ReplaceRule

	ok := parser.consumeToReplaceRuleArray(&output)
	if !ok {
		return nil
	}

	for ok {
		ok = parser.consumeToReplaceRuleArray(&output)	
	}

	return &output
}

func (parser *Parser) consumeToReplaceRuleArray(replaceRuleArray *[]ReplaceRule) bool {
	stringLiteral := parser.consumeStringLiteral()
	if stringLiteral != nil {
		lenParsed := parser.consumeLen()
		if lenParsed != nil {
			lenParsed.FirstArgument = *stringLiteral
			*replaceRuleArray = append(*replaceRuleArray, *lenParsed)
			return true		
		} else {
			*replaceRuleArray = append(*replaceRuleArray, *stringLiteral)
			return true		
		}
	}	
	captureGroup := parser.consumeCaptureGroup()
	if captureGroup != nil {
		lenParsed := parser.consumeLen()
		if lenParsed != nil {
			lenParsed.FirstArgument = *captureGroup
			*replaceRuleArray = append(*replaceRuleArray, *lenParsed)
			return true
		} else {
			*replaceRuleArray = append(*replaceRuleArray, *captureGroup)
			return true
		}
	}

	lenParsed := parser.consumeLen()
	if lenParsed != nil {
		parser.Errors = append(parser.Errors, ErrUnexpectedLen(&parser.tokenizer))
		return true
	}

	return false
}

func (parser *Parser) consumeStringLiteral() *StringLiteral {
	token, err := parser.tokenizer.Peek()
	if err != nil {
		parser.tokenizer.Next()
		parser.Errors = append(parser.Errors, err)
		return nil
	}
	if token.Type == kuuhaku_tokenizer.STRING_LITERAL {
		parser.tokenizer.Next()
		return &StringLiteral {
			String: token.Content,
			Position: token.Position,
		};
	} else {
		return nil
	}
}

func (parser *Parser) consumeCaptureGroup() *CaptureGroup {
	token, err := parser.tokenizer.Peek()
	if err != nil {
		parser.tokenizer.Next()
		parser.Errors = append(parser.Errors, err)
		return nil
	}
	if token.Type == kuuhaku_tokenizer.CAPTURE_GROUP {
		parser.tokenizer.Next()
		number, err := strconv.Atoi(token.Content)
		if err != nil {
			parser.Errors = append(parser.Errors, err)
			return nil
		}
		return &CaptureGroup {
			Number: number,
			Position: token.Position,
		};
	} else {
		return nil
	}
}

func (parser *Parser) consumeLen() *Len {
	token, err := parser.tokenizer.Peek()
	if err != nil {
		parser.tokenizer.Next()
		parser.Errors = append(parser.Errors, err)
		return nil
	}
	if token.Type == kuuhaku_tokenizer.LEN_KEYWORD {
		parser.tokenizer.Next()
		var argument StringStmt
		ok := parser.consumeStringStmt(&argument)
		if !ok {
			parser.tokenizer.Next()
			parser.Errors = append(parser.Errors, ErrLenArgumentInvalid(&parser.tokenizer))
			return &Len {
				SecondArgument: nil,
				Position: token.Position,
			}
		}
		return &Len {
			SecondArgument: argument,
			Position: token.Position,
		};
	} else {
		return nil
	}
}

func (parser *Parser) consumeStringStmt(stringStmt *StringStmt) bool {
	stringLiteral := parser.consumeStringLiteral()
	if stringLiteral != nil {
		*stringStmt = *stringLiteral
		return true
	}
	captureGroup := parser.consumeCaptureGroup()
	if captureGroup != nil {
		*stringStmt = *captureGroup
		return true
	}
	return false
}

func (parser *Parser) consumeMatchRules() *[]MatchRule {
	var output []MatchRule;
	
	ok := parser.consumeToMatchRuleArray(&output)
	if !ok {
		return nil
	}

	for ok {
		ok = parser.consumeToMatchRuleArray(&output)
	}

	return &output
}

func (parser *Parser) consumeToMatchRuleArray(matchRuleArray *[]MatchRule) bool {
	identifier := parser.consumeIdentifier()
	if identifier != nil {
		*matchRuleArray = append(*matchRuleArray, *identifier)
		return true
	}
	regexLit := parser.consumeRegexLiteral()
	if regexLit != nil {
		*matchRuleArray = append(*matchRuleArray, *regexLit)
		return true
	}
	return false
}

func (parser *Parser) consumeIdentifier() *Identifer {
	token, err := parser.tokenizer.Peek()
	if err != nil {
		parser.tokenizer.Next()
		parser.Errors = append(parser.Errors, err)
		return nil
	}
	if token.Type == kuuhaku_tokenizer.IDENTIFIER {
		parser.tokenizer.Next()
		return &Identifer {
			Name: token.Content,
			Position: token.Position,
		};
	} else {
		return nil
	}
}

func (parser *Parser) consumeRegexLiteral() *RegexLiteral {
	token, err := parser.tokenizer.Peek()
	if err != nil {
		parser.tokenizer.Next()
		parser.Errors = append(parser.Errors, err)
		return nil
	}
	if token.Type == kuuhaku_tokenizer.REGEX_LITERAL {
		parser.tokenizer.Next()
		return &RegexLiteral {
			RegexString: token.Content,
			Position: token.Position,
		};
	} else {
		return nil
	}
}
