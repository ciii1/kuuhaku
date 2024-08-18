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

func consumeReplaceRules(tokenizer *kuuhaku_tokenizer.Tokenizer, errors *[]error) *[]ReplaceRule {
	var output []ReplaceRule

	ok := consumeToReplaceRuleArray(tokenizer, &output, errors)
	if !ok {
		return nil
	}

	for ok {
		ok = consumeToReplaceRuleArray(tokenizer, &output, errors)	
	}

	return &output
}

func consumeToReplaceRuleArray(tokenizer *kuuhaku_tokenizer.Tokenizer, replaceRuleArray *[]ReplaceRule, errors *[]error) bool {
	stringLiteral := consumeStringLiteral(tokenizer, errors)
	if stringLiteral != nil {
		lenParsed := consumeLen(tokenizer, errors)
		if lenParsed != nil {
			lenParsed.FirstArgument = *stringLiteral
			*replaceRuleArray = append(*replaceRuleArray, *lenParsed)
			return true		
		} else {
			*replaceRuleArray = append(*replaceRuleArray, *stringLiteral)
			return true		
		}
	}	
	captureGroup := consumeCaptureGroup(tokenizer, errors)
	if captureGroup != nil {
		lenParsed := consumeLen(tokenizer, errors)
		if lenParsed != nil {
			lenParsed.FirstArgument = *captureGroup
			*replaceRuleArray = append(*replaceRuleArray, *lenParsed)
			return true
		} else {
			*replaceRuleArray = append(*replaceRuleArray, *captureGroup)
			return true
		}
	}

	lenParsed := consumeLen(tokenizer, errors)
	if lenParsed != nil {
		*errors = append(*errors, ErrUnexpectedLen(tokenizer))
		return true
	}

	return false
}

func consumeStringLiteral(tokenizer *kuuhaku_tokenizer.Tokenizer, errors *[]error) *StringLiteral {
	token, err := tokenizer.Peek()
	if err != nil {
		tokenizer.Next()
		*errors = append(*errors, err)
		return nil
	}
	if token.Type == kuuhaku_tokenizer.STRING_LITERAL {
		tokenizer.Next()
		return &StringLiteral {
			String: token.Content,
			Position: token.Position,
		};
	} else {
		return nil
	}
}

func consumeCaptureGroup(tokenizer *kuuhaku_tokenizer.Tokenizer, errors *[]error) *CaptureGroup {
	token, err := tokenizer.Peek()
	if err != nil {
		tokenizer.Next()
		*errors = append(*errors, err)
		return nil
	}
	if token.Type == kuuhaku_tokenizer.CAPTURE_GROUP {
		tokenizer.Next()
		number, err := strconv.Atoi(token.Content)
		if err != nil {
			*errors = append(*errors, err)
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

func consumeLen(tokenizer *kuuhaku_tokenizer.Tokenizer, errors *[]error) *Len {
	token, err := tokenizer.Peek()
	if err != nil {
		tokenizer.Next()
		*errors = append(*errors, err)
		return nil
	}
	if token.Type == kuuhaku_tokenizer.LEN_KEYWORD {
		tokenizer.Next()
		var argument StringStmt
		ok := consumeStringStmt(tokenizer, &argument, errors)
		if !ok {
			tokenizer.Next()
			*errors = append(*errors, ErrLenArgumentInvalid(tokenizer))
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

func consumeStringStmt(tokenizer *kuuhaku_tokenizer.Tokenizer, stringStmt *StringStmt, errors *[]error) bool {
	stringLiteral := consumeStringLiteral(tokenizer, errors)
	if stringLiteral != nil {
		*stringStmt = *stringLiteral
		return true
	}
	captureGroup := consumeCaptureGroup(tokenizer, errors)
	if captureGroup != nil {
		*stringStmt = *captureGroup
		return true
	}
	return false
}

func consumeMatchRules(tokenizer *kuuhaku_tokenizer.Tokenizer, errors *[]error) *[]MatchRule {
	var output []MatchRule;
	
	ok := consumeToMatchRuleArray(tokenizer, &output, errors)
	if !ok {
		return nil
	}

	for ok {
		ok = consumeToMatchRuleArray(tokenizer, &output, errors)
	}

	return &output
}

func consumeToMatchRuleArray(tokenizer *kuuhaku_tokenizer.Tokenizer, matchRuleArray *[]MatchRule, errors *[]error) bool {
	identifier := consumeIdentifier(tokenizer, errors)
	if identifier != nil {
		*matchRuleArray = append(*matchRuleArray, *identifier)
		return true
	}
	regexLit := consumeRegexLiteral(tokenizer, errors)
	if regexLit != nil {
		*matchRuleArray = append(*matchRuleArray, *regexLit)
		return true
	}
	return false
}

func consumeIdentifier(tokenizer *kuuhaku_tokenizer.Tokenizer, errors *[]error) *Identifer {
	token, err := tokenizer.Peek()
	if err != nil {
		tokenizer.Next()
		*errors = append(*errors, err)
		return nil
	}
	if token.Type == kuuhaku_tokenizer.IDENTIFIER {
		tokenizer.Next()
		return &Identifer {
			Name: token.Content,
			Position: token.Position,
		};
	} else {
		return nil
	}
}

func consumeRegexLiteral(tokenizer *kuuhaku_tokenizer.Tokenizer, errors *[]error) *RegexLiteral {
	token, err := tokenizer.Peek()
	if err != nil {
		tokenizer.Next()
		*errors = append(*errors, err)
		return nil
	}
	if token.Type == kuuhaku_tokenizer.REGEX_LITERAL {
		tokenizer.Next()
		return &RegexLiteral {
			RegexString: token.Content,
			Position: token.Position,
		};
	} else {
		return nil
	}
}
