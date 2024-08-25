package kuuhaku_parser

import (
	"fmt"
	"strconv"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_tokenizer"
)

type ParseErrorType int

const (
	LEN_ARGUMENT_INVALID = iota
	UNEXPECTED_LEN
	EXPECTED_OPENING_CURLY_BRACKET
	EXPECTED_CLOSING_CURLY_BRACKET
	EXPECTED_EQUAL_SIGN
	EXPECTED_REPLACE_RULE
	EXPECTED_MATCH_RULE
	EXPECTED_RULE
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
		Position: tokenizer.PrevPosition,
		Type: LEN_ARGUMENT_INVALID,
	}
}

func ErrUnexpectedLen(tokenizer *kuuhaku_tokenizer.Tokenizer) *ParseError {
	return &ParseError {
		Message: "Usage of len here is invalid",
		Position: tokenizer.PrevPosition,
		Type: UNEXPECTED_LEN,
	}
}

func ErrExpectedOpeningCurlyBracket(tokenizer *kuuhaku_tokenizer.Tokenizer) *ParseError {
	return &ParseError {
		Message: "Expected an opening curly bracket",
		Position: tokenizer.PrevPosition,
		Type: EXPECTED_OPENING_CURLY_BRACKET,
	}
}

func ErrExpectedClosingCurlyBracket(tokenizer *kuuhaku_tokenizer.Tokenizer) *ParseError {
	return &ParseError {
		Message: "Expected a closing curly bracket",
		Position: tokenizer.PrevPosition,
		Type: EXPECTED_CLOSING_CURLY_BRACKET,
	}
}

func ErrExpectedEqualSign(tokenizer *kuuhaku_tokenizer.Tokenizer) *ParseError {
	return &ParseError {
		Message: "Expected equal sign",
		Position: tokenizer.PrevPosition,
		Type: EXPECTED_EQUAL_SIGN,
	}
}

func ErrExpectedMatchRules(tokenizer *kuuhaku_tokenizer.Tokenizer) *ParseError {
	return &ParseError {
		Message: "Expected match rules",
		Position: tokenizer.PrevPosition,
		Type: EXPECTED_MATCH_RULE,
	}
}

func ErrExpectedReplaceRules(tokenizer *kuuhaku_tokenizer.Tokenizer) *ParseError {
	return &ParseError {
		Message: "Expected replace rules",
		Position: tokenizer.PrevPosition,
		Type: EXPECTED_REPLACE_RULE,
	}
}

func ErrExpectedRule(tokenizer *kuuhaku_tokenizer.Tokenizer) *ParseError {
	return &ParseError {
		Message: "Expected a rule definition",
		Position: tokenizer.PrevPosition,
		Type: EXPECTED_RULE,
	}
}

type Parser struct {
	tokenizer kuuhaku_tokenizer.Tokenizer
	Errors []error
}

func Parse(input string) (Ast, []error) {
	parser := Init(input)
	ast := parser.consumeInput()
	return *ast, parser.Errors
}

func Init(input string) Parser {
	return Parser {
		tokenizer: kuuhaku_tokenizer.Init(input),
		Errors: []error{},
	}
}

func (parser *Parser) consumeInput() *Ast {
	output := Ast {
		Rules: make(map[string][]Rule),
		Position: parser.tokenizer.Position,
	}

	rule := parser.consumeRule()
	if rule != nil {
		output.Rules[rule.Name] = append(output.Rules[rule.Name], *rule)
	} else {
		parser.Errors = append(parser.Errors, ErrExpectedRule(&parser.tokenizer))
		parser.tokenizer.Next()
	}

	token, err := parser.tokenizer.Peek()
	if err != nil {
		parser.Errors = append(parser.Errors, err)
		parser.tokenizer.Next()
	}
	for token == nil || token.Type != kuuhaku_tokenizer.EOF {
		rule := parser.consumeRule()
		if rule != nil {
			output.Rules[rule.Name] = append(output.Rules[rule.Name], *rule)
		} else {
			parser.Errors = append(parser.Errors, ErrExpectedRule(&parser.tokenizer))
			parser.tokenizer.Next()
		}
		token, err = parser.tokenizer.Peek()
		if err != nil {
			parser.Errors = append(parser.Errors, err)
			parser.tokenizer.Next()
		}
	}

	return &output
}

func (parser *Parser) consumeRule() *Rule {	
	position := parser.tokenizer.Position
	token, err := parser.tokenizer.Peek()
	if err != nil {
		parser.tokenizer.Next()
		parser.Errors = append(parser.Errors, err)
		return nil
	}
	var name string
	if token.Type == kuuhaku_tokenizer.IDENTIFIER {
		name = token.Content	
	} else {
		return nil
	}
		
	token, err = parser.tokenizer.Next()
	if err != nil {
		parser.Errors = append(parser.Errors, err)
		parser.tokenizer.Next()
		return &Rule {
			Name: name,	
			Position: position,
		}
	}
	if token.Type != kuuhaku_tokenizer.OPENING_CURLY_BRACKET {
		parser.Errors = append(parser.Errors, ErrExpectedOpeningCurlyBracket(&parser.tokenizer))
		return &Rule {
			Name: name,	
			Position: position,
		}
	}
	parser.tokenizer.Next()

	matchRules := parser.consumeMatchRules()
	if matchRules == nil {
		parser.Errors = append(parser.Errors, ErrExpectedMatchRules(&parser.tokenizer))
		parser.panicTillToken(kuuhaku_tokenizer.CLOSING_CURLY_BRACKET)
		return &Rule {
			Name: name,	
			Position: position,
		}
	}

	token, err = parser.tokenizer.Peek()
	if err != nil {
		parser.panicTillToken(kuuhaku_tokenizer.CLOSING_CURLY_BRACKET)
		return &Rule {
			Name: name,	
			MatchRules: *matchRules,
			Position: position,
		}
	}
	if token.Type == kuuhaku_tokenizer.CLOSING_CURLY_BRACKET {
		parser.tokenizer.Next()
		return &Rule {
			Name: name,	
			MatchRules: *matchRules,
			Position: position,
		}
	}
	if token.Type != kuuhaku_tokenizer.EQUAL_SIGN {
		parser.Errors = append(parser.Errors, ErrExpectedEqualSign(&parser.tokenizer))
		parser.panicTillToken(kuuhaku_tokenizer.CLOSING_CURLY_BRACKET)
		return &Rule {
			Name: name,	
			MatchRules: *matchRules,
			Position: position,
		}
	}
	parser.tokenizer.Next()

	replaceRules := parser.consumeReplaceRules()
	if replaceRules == nil {
		parser.Errors = append(parser.Errors, ErrExpectedReplaceRules(&parser.tokenizer))
		parser.panicTillToken(kuuhaku_tokenizer.CLOSING_CURLY_BRACKET)
		return &Rule {
			Name: name,	
			MatchRules: *matchRules,
			Position: position,
		}
	}

	token, err = parser.tokenizer.Peek()
	if err != nil {
		parser.panicTillToken(kuuhaku_tokenizer.CLOSING_CURLY_BRACKET)
		return &Rule {
			Name: name,	
			MatchRules: *matchRules,
			ReplaceRules: *replaceRules,
			Position: position,
		}
	}
	if token.Type != kuuhaku_tokenizer.CLOSING_CURLY_BRACKET {
		parser.Errors = append(parser.Errors, ErrExpectedClosingCurlyBracket(&parser.tokenizer))
		parser.panicTillToken(kuuhaku_tokenizer.CLOSING_CURLY_BRACKET)
		return &Rule {
			Name: name,	
			MatchRules: *matchRules,
			ReplaceRules: *replaceRules,
			Position: position,
		}
	}
	parser.tokenizer.Next()

	return &Rule {
		Name: name,
		MatchRules: *matchRules,
		ReplaceRules: *replaceRules,
		Position: position,
	}
}

func (parser *Parser) panicTillToken(tokenType kuuhaku_tokenizer.TokenType) {
	token, err := parser.tokenizer.Peek()
	if err != nil {
		parser.Errors = append(parser.Errors, err)
	}
	for token == nil || token.Type != tokenType && token.Type != kuuhaku_tokenizer.EOF {
		token, err = parser.tokenizer.Next()
		if err != nil {
			parser.Errors = append(parser.Errors, err)
		}
	}
	parser.tokenizer.Next()
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

	errUnexpectedLen := ErrUnexpectedLen(&parser.tokenizer)
	lenParsed := parser.consumeLen()
	if lenParsed != nil {
		parser.Errors = append(parser.Errors, errUnexpectedLen)
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
