package kuuhaku_parser

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/ciii1/kuuhaku/internal/helper"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_tokenizer"
)

var errTokenUnrecognized = fmt.Errorf("Token is unrecognized")
var ErrLenArgumentInvalid = fmt.Errorf("Len argument is invalid")
var ErrUnexpectedLen = fmt.Errorf("Usage of len here is invalid")

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

func consumeReplaceRules(tokenizer *kuuhaku_tokenizer.Tokenizer) (*[]ReplaceRule, []error) {
	var output []ReplaceRule
	var errors []error

	ok, err := consumeToReplaceRuleArray(tokenizer, &output)
	if !ok {
		return nil, nil
	}
	if err != nil {
		errors = append(errors, err)
	}

	for ok {
		ok, err = consumeToReplaceRuleArray(tokenizer, &output)	
		if err != nil {
			errors = append(errors, err)
		}
	}

	return &output, errors
}

func consumeToReplaceRuleArray(tokenizer *kuuhaku_tokenizer.Tokenizer, replaceRuleArray *[]ReplaceRule) (bool, error) {
	stringLiteral, err := consumeStringLiteral(tokenizer)
	if err != nil {
		return true, err
	}
	if stringLiteral != nil {
		lenParsed, err := consumeLen(tokenizer)
		helper.Check(err)
		if err == nil && lenParsed != nil {
			lenParsed.FirstArgument = *stringLiteral
			*replaceRuleArray = append(*replaceRuleArray, *lenParsed)
			return true, nil
		} else {
			*replaceRuleArray = append(*replaceRuleArray, *stringLiteral)
			return true, nil
		}
	}
	captureGroup, err := consumeCaptureGroup(tokenizer)
	if err != nil {
		return true, err
	}
	if captureGroup != nil {
		lenParsed, err := consumeLen(tokenizer)
		if err == nil && lenParsed != nil {
			lenParsed.FirstArgument = *captureGroup
			*replaceRuleArray = append(*replaceRuleArray, *lenParsed)
			return true, nil
		} else {
			*replaceRuleArray = append(*replaceRuleArray, *captureGroup)
			return true, nil
		}
	}

	lenParsed, err := consumeLen(tokenizer)
	if err != nil {
		return true, errors.Join(err, ErrUnexpectedLen)
	}
	if lenParsed != nil {
		return true, ErrUnexpectedLen
	}

	return false, nil
}

func consumeStringLiteral(tokenizer *kuuhaku_tokenizer.Tokenizer) (*StringLiteral, error) {
	token, err := tokenizer.Peek()
	if err != nil {
		tokenizer.Next()
		return nil, err
	}
	if token.Type == kuuhaku_tokenizer.STRING_LITERAL {
		tokenizer.Next()
		return &StringLiteral {
			String: token.Content,
			Position: token.Position,
		}, nil;
	} else {
		return nil, nil
	}
}

func consumeCaptureGroup(tokenizer *kuuhaku_tokenizer.Tokenizer) (*CaptureGroup, error) {
	token, err := tokenizer.Peek()
	if err != nil {
		tokenizer.Next()
		return nil, err
	}
	if token.Type == kuuhaku_tokenizer.CAPTURE_GROUP {
		tokenizer.Next()
		number, err := strconv.Atoi(token.Content)
		if err != nil {
			return nil, err
		}
		return &CaptureGroup {
			Number: number,
			Position: token.Position,
		}, nil;
	} else {
		return nil, nil
	}
}

func consumeLen(tokenizer *kuuhaku_tokenizer.Tokenizer) (*Len, error) {
	token, err := tokenizer.Peek()
	if err != nil {
		tokenizer.Next()
		return nil, err
	}
	if token.Type == kuuhaku_tokenizer.LEN_KEYWORD {
		tokenizer.Next()
		var argument StringStmt
		ok, err := consumeStringStmt(tokenizer, &argument)
		if !ok {
			if err != nil {
				return nil, errors.Join(err, ErrLenArgumentInvalid)
			} else {
				return nil, ErrLenArgumentInvalid
			}
		}
		return &Len {
			SecondArgument: argument,
			Position: token.Position,
		}, nil;
	} else {
		return nil, nil
	}
}

func consumeStringStmt(tokenizer *kuuhaku_tokenizer.Tokenizer, stringStmt *StringStmt) (bool, error) {
	stringLiteral, err := consumeStringLiteral(tokenizer)
	if err != nil {
		return false, err
	}
	if stringLiteral != nil {
		*stringStmt = *stringLiteral
		return true, nil
	}
	captureGroup, err := consumeCaptureGroup(tokenizer)
	if err != nil {
		return false, err
	}
	if captureGroup != nil {
		*stringStmt = *captureGroup
		return true, nil
	}
	return false, err
}

func consumeMatchRules(tokenizer *kuuhaku_tokenizer.Tokenizer) (*[]MatchRule, []error) {
	var output []MatchRule;
	var errors []error
	
	ok, err := consumeToMatchRuleArray(tokenizer, &output)	
	if !ok {
		return nil, nil
	}
	if err != nil {
		errors = append(errors, err)
	}

	for ok {
		ok, err = consumeToMatchRuleArray(tokenizer, &output)
		if err != nil {
			errors = append(errors, err)
		}
	}

	return &output, errors
}

func consumeToMatchRuleArray(tokenizer *kuuhaku_tokenizer.Tokenizer, matchRuleArray *[]MatchRule) (bool, error) {
	identifier, err := consumeIdentifier(tokenizer)
	if err != nil {
		return true, err
	}
	if identifier != nil {
		*matchRuleArray = append(*matchRuleArray, *identifier)
		return true, nil
	}
	regexLit, err := consumeRegexLiteral(tokenizer)
	if err != nil {
		return true, err
	}
	if regexLit != nil {
		*matchRuleArray = append(*matchRuleArray, *regexLit)
		return true, nil
	}
	return false, nil
}

func consumeIdentifier(tokenizer *kuuhaku_tokenizer.Tokenizer) (*Identifer, error) {
	token, err := tokenizer.Peek()
	if err != nil {
		tokenizer.Next()
		return nil, err
	}
	if token.Type == kuuhaku_tokenizer.IDENTIFIER {
		tokenizer.Next()
		return &Identifer {
			Name: token.Content,
			Position: token.Position,
		}, nil;
	} else {
		return nil, nil
	}
}

func consumeRegexLiteral(tokenizer *kuuhaku_tokenizer.Tokenizer) (*RegexLiteral, error) {
	token, err := tokenizer.Peek()
	if err != nil {
		tokenizer.Next()
		return nil, err
	}
	if token.Type == kuuhaku_tokenizer.REGEX_LITERAL {
		tokenizer.Next()
		return &RegexLiteral {
			RegexString: token.Content,
			Position: token.Position,
		}, nil;
	} else {
		return nil, nil
	}
}
