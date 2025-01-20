package kuuhaku_parser

import (
	"fmt"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_tokenizer"
)

type ParseErrorType int

const (
	EXPECTED_OPENING_CURLY_BRACKET ParseErrorType = iota
	EXPECTED_CLOSING_CURLY_BRACKET
	EXPECTED_CLOSING_BRACKET_OR_COMMA
	EXPECTED_ARG
	EXPECTED_ARG_LIST
	EXPECTED_EQUAL_SIGN
	EXPECTED_REPLACE_RULE
	EXPECTED_MATCH_RULE
	EXPECTED_RULE
	MIXED_TYPE_MATCH_RULE
	MULTIPLE_GLOBAL_LUA
)

type ParseError struct {
	Position kuuhaku_tokenizer.Position
	Message  string
	Type     ParseErrorType
}

func (e ParseError) Error() string {
	return fmt.Sprintf("Parse error (%d, %d): %s", e.Position.Line, e.Position.Column, e.Message)
}

func ErrMixedTypeMatchRule(tokenizer *kuuhaku_tokenizer.Tokenizer) *ParseError {
	return &ParseError{
		Message:  "Mixing regex literals and variables inside one rule is not allowed",
		Position: tokenizer.PrevPosition,
		Type:     MIXED_TYPE_MATCH_RULE,
	}
}

func ErrExpectedOpeningCurlyBracket(tokenizer *kuuhaku_tokenizer.Tokenizer) *ParseError {
	return &ParseError{
		Message:  "Expected an opening curly bracket",
		Position: tokenizer.PrevPosition,
		Type:     EXPECTED_OPENING_CURLY_BRACKET,
	}
}

func ErrExpectedClosingBracketOrComma(tokenizer *kuuhaku_tokenizer.Tokenizer) *ParseError {
	return &ParseError{
		Message:  "Expected a closing bracket or a comma",
		Position: tokenizer.PrevPosition,
		Type:     EXPECTED_CLOSING_BRACKET_OR_COMMA,
	}
}

func ErrExpectedClosingCurlyBracket(tokenizer *kuuhaku_tokenizer.Tokenizer) *ParseError {
	return &ParseError{
		Message:  "Expected a closing curly bracket",
		Position: tokenizer.PrevPosition,
		Type:     EXPECTED_CLOSING_CURLY_BRACKET,
	}
}

func ErrExpectedArg(tokenizer *kuuhaku_tokenizer.Tokenizer) *ParseError {
	return &ParseError{
		Message:  "Expected an argument",
		Position: tokenizer.PrevPosition,
		Type:     EXPECTED_ARG,
	}
}

func ErrExpectedArgList(tokenizer *kuuhaku_tokenizer.Tokenizer) *ParseError {
	return &ParseError{
		Message:  "Expected an argument list",
		Position: tokenizer.PrevPosition,
		Type:     EXPECTED_ARG_LIST,
	}
}

func ErrExpectedEqualSign(tokenizer *kuuhaku_tokenizer.Tokenizer) *ParseError {
	return &ParseError{
		Message:  "Expected equal sign",
		Position: tokenizer.PrevPosition,
		Type:     EXPECTED_EQUAL_SIGN,
	}
}

func ErrExpectedMatchRules(tokenizer *kuuhaku_tokenizer.Tokenizer) *ParseError {
	return &ParseError{
		Message:  "Expected match rules",
		Position: tokenizer.PrevPosition,
		Type:     EXPECTED_MATCH_RULE,
	}
}

func ErrExpectedReplaceRule(tokenizer *kuuhaku_tokenizer.Tokenizer) *ParseError {
	return &ParseError{
		Message:  "Expected a replace rule",
		Position: tokenizer.PrevPosition,
		Type:     EXPECTED_REPLACE_RULE,
	}
}

func ErrExpectedGlobal(tokenizer *kuuhaku_tokenizer.Tokenizer) *ParseError {
	return &ParseError{
		Message:  "Expected a rule definition or a global lua literal",
		Position: tokenizer.PrevPosition,
		Type:     EXPECTED_RULE,
	}
}

func ErrMultipleGlobalLua(tokenizer *kuuhaku_tokenizer.Tokenizer) *ParseError {
	return &ParseError{
		Message:  "Found multiple global lua literal",
		Position: tokenizer.PrevPosition,
		Type:     MULTIPLE_GLOBAL_LUA,
	}
}

type Parser struct {
	tokenizer kuuhaku_tokenizer.Tokenizer
	GlobalLua *LuaLiteral
	Errors    []error
}

func Parse(input string) (Ast, []error) {
	parser := initParser(input)
	ast := parser.consumeInput()
	return *ast, parser.Errors
}

func initParser(input string) Parser {
	return Parser{
		tokenizer: kuuhaku_tokenizer.Init(input),
		Errors:    []error{},
	}
}

func (parser *Parser) consumeInput() *Ast {
	output := Ast{
		Rules:        make(map[string][]*Rule),
		Position:     parser.tokenizer.Position,
		IsSearchMode: false,
	}

	output.IsSearchMode = parser.consumeSearchMode()
	orderCounter := 0

	rule := parser.consumeRule()
	if rule != nil {
		rule.Order = orderCounter
		output.Rules[rule.Name] = append(output.Rules[rule.Name], rule)
	} else {
		globalLua := parser.consumeGlobalLua()
		if globalLua != nil {
			orderCounter -= 1
			output.GlobalLua = globalLua
		} else {
			parser.Errors = append(parser.Errors, ErrExpectedGlobal(&parser.tokenizer))
			parser.tokenizer.Next()
		}
	}

	token, err := parser.tokenizer.Peek()
	if err != nil {
		parser.Errors = append(parser.Errors, err)
		parser.tokenizer.Next()
	}
	for token == nil || token.Type != kuuhaku_tokenizer.EOF {
		orderCounter += 1
		rule := parser.consumeRule()
		if rule != nil {
			rule.Order = orderCounter
			output.Rules[rule.Name] = append(output.Rules[rule.Name], rule)
		} else {
			globalLua := parser.consumeGlobalLua()
			if globalLua != nil {
				orderCounter -= 1
				output.GlobalLua = globalLua
			} else {
				parser.Errors = append(parser.Errors, ErrExpectedGlobal(&parser.tokenizer))
				parser.tokenizer.Next()
			}
		}
		token, err = parser.tokenizer.Peek()
		if err != nil {
			parser.Errors = append(parser.Errors, err)
			parser.tokenizer.Next()
		}
	}

	return &output
}


func (parser *Parser) consumeSearchMode() bool {
	token, err := parser.tokenizer.Peek()
	if err != nil {
		parser.tokenizer.Next()
		parser.Errors = append(parser.Errors, err)
		return false
	}
	if token.Type == kuuhaku_tokenizer.SEARCH_MODE_KEYWORD {
		parser.tokenizer.Next()
		return true
	} else {
		return false
	}
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

	var argList []LuaLiteral
	argListP := parser.consumeArgList()
	if argListP != nil {
		argList = *argListP
	}

	token, err = parser.tokenizer.Peek()
	if err != nil {
		parser.Errors = append(parser.Errors, err)
		parser.tokenizer.Next()
		return &Rule{
			Name:     name,
			Position: position,
			ArgList:  argList,
		}
	}

	if token.Type != kuuhaku_tokenizer.OPENING_CURLY_BRACKET {
		parser.Errors = append(parser.Errors, ErrExpectedOpeningCurlyBracket(&parser.tokenizer))
		return &Rule{
			Name:     name,
			Position: position,
			ArgList:  argList,
		}
	}
	parser.tokenizer.Next()

	matchRules := parser.consumeMatchRules()
	if matchRules == nil {
		parser.Errors = append(parser.Errors, ErrExpectedMatchRules(&parser.tokenizer))
		parser.panicTillToken(kuuhaku_tokenizer.CLOSING_CURLY_BRACKET)
		return &Rule{
			Name:     name,
			Position: position,
			ArgList:  argList,
		}
	}

	token, err = parser.tokenizer.Peek()
	if err != nil {
		parser.panicTillToken(kuuhaku_tokenizer.CLOSING_CURLY_BRACKET)
		return &Rule{
			Name:       name,
			MatchRules: *matchRules,
			Position:   position,
			ArgList:    argList,
		}
	}
	if token.Type == kuuhaku_tokenizer.CLOSING_CURLY_BRACKET {
		parser.tokenizer.Next()
		return &Rule{
			Name:       name,
			MatchRules: *matchRules,
			Position:   position,
			ArgList:    argList,
		}
	}
	if token.Type != kuuhaku_tokenizer.EQUAL_SIGN {
		parser.Errors = append(parser.Errors, ErrExpectedEqualSign(&parser.tokenizer))
		parser.panicTillToken(kuuhaku_tokenizer.CLOSING_CURLY_BRACKET)
		return &Rule{
			Name:       name,
			MatchRules: *matchRules,
			Position:   position,
			ArgList:    argList,
		}
	}
	parser.tokenizer.Next()

	replaceRule := parser.consumeLuaLiteral()
	if replaceRule == nil {
		parser.Errors = append(parser.Errors, ErrExpectedReplaceRule(&parser.tokenizer))
		parser.panicTillToken(kuuhaku_tokenizer.CLOSING_CURLY_BRACKET)
		return &Rule{
			Name:       name,
			MatchRules: *matchRules,
			Position:   position,
			ArgList:    argList,
		}
	}

	token, err = parser.tokenizer.Peek()
	if err != nil {
		parser.panicTillToken(kuuhaku_tokenizer.CLOSING_CURLY_BRACKET)
		return &Rule{
			Name:        name,
			MatchRules:  *matchRules,
			ReplaceRule: *replaceRule,
			Position:    position,
			ArgList:     argList,
		}
	}
	if token.Type != kuuhaku_tokenizer.CLOSING_CURLY_BRACKET {
		parser.Errors = append(parser.Errors, ErrExpectedClosingCurlyBracket(&parser.tokenizer))
		parser.panicTillToken(kuuhaku_tokenizer.CLOSING_CURLY_BRACKET)
		return &Rule{
			Name:        name,
			MatchRules:  *matchRules,
			ReplaceRule: *replaceRule,
			Position:    position,
			ArgList:     argList,
		}
	}
	parser.tokenizer.Next()

	return &Rule{
		Name:        name,
		MatchRules:  *matchRules,
		ReplaceRule: *replaceRule,
		Position:    position,
		ArgList:     argList,
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

func (parser *Parser) consumeMatchRules() *[]MatchRule {
	var output []MatchRule
	doesRegexLiteralExist := false

	ok, isRegexLiteral := parser.consumeToMatchRuleArray(&output)
	if !ok {
		return nil
	}
	if isRegexLiteral {
		doesRegexLiteralExist = true
	}

	for ok {
		//TODO: Debug why this happens
		tokenizer := parser.tokenizer
		ok, isRegexLiteral = parser.consumeToMatchRuleArray(&output)
		if ok {
			if (!isRegexLiteral && doesRegexLiteralExist) || (isRegexLiteral && !doesRegexLiteralExist) {
				parser.Errors = append(parser.Errors, ErrMixedTypeMatchRule(&tokenizer))
			}
		}
	}

	return &output
}

// returns (ok bool, isRegexLiteral bool) 
func (parser *Parser) consumeToMatchRuleArray(matchRuleArray *[]MatchRule) (bool, bool) {
	identifier := parser.consumeIdentifier()
	if identifier != nil {
		*matchRuleArray = append(*matchRuleArray, *identifier)
		return true, false
	}
	regexLit := parser.consumeRegexLiteral()
	if regexLit != nil {
		*matchRuleArray = append(*matchRuleArray, *regexLit)
		return true, true
	}
	return false, false
}

func (parser *Parser) consumeIdentifier() *Identifier {
	token, err := parser.tokenizer.Peek()
	if err != nil {
		parser.tokenizer.Next()
		parser.Errors = append(parser.Errors, err)
		return nil
	}
	if token.Type == kuuhaku_tokenizer.IDENTIFIER {
		parser.tokenizer.Next()

		var argList []LuaLiteral
		argListP := parser.consumeArgList()
		if argListP != nil {
			argList = *argListP
		}

		return &Identifier{
			Name:     token.Content,
			Position: token.Position,
			ArgList:  argList,
		}
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
		return &RegexLiteral{
			RegexString: token.Content,
			Position:    token.Position,
		}
	} else {
		return nil
	}
}

func (parser *Parser) consumeArgList() *[]LuaLiteral {
	var argList []LuaLiteral
	token, err := parser.tokenizer.Peek()
	if err != nil {
		parser.tokenizer.Next()
		parser.Errors = append(parser.Errors, err)
		return nil
	}
	if token.Type != kuuhaku_tokenizer.OPENING_BRACKET {
		return nil
	}

	parser.tokenizer.Next()

	arg := parser.consumeLuaLiteral()
	if arg == nil {
		parser.Errors = append(parser.Errors, ErrExpectedArg(&parser.tokenizer))
		return nil
	}
	argList = append(argList, *arg)

	for true {
		token, err := parser.tokenizer.Peek()
		if err != nil {
			parser.tokenizer.Next()
			parser.Errors = append(parser.Errors, err)
			return nil
		}

		if token.Type == kuuhaku_tokenizer.CLOSING_BRACKET {
			parser.tokenizer.Next()
			break
		} else if token.Type == kuuhaku_tokenizer.COMMA {
			parser.tokenizer.Next()
		} else {
			parser.Errors = append(parser.Errors, ErrExpectedClosingBracketOrComma(&parser.tokenizer))
			return nil
		}

		arg := parser.consumeLuaLiteral()
		if arg == nil {
			parser.Errors = append(parser.Errors, ErrExpectedArg(&parser.tokenizer))
			return nil
		}
		argList = append(argList, *arg)
	}

	return &argList
}

func (parser *Parser) consumeGlobalLua() *LuaLiteral {
	token := parser.consumeLuaLiteral()	
	if token == nil {
		return nil
	}
	if parser.GlobalLua != nil {
		parser.Errors = append(parser.Errors, ErrMultipleGlobalLua(&parser.tokenizer))
	}
	parser.GlobalLua = token
	return token
}

func (parser *Parser) consumeLuaLiteral() *LuaLiteral {
	token, err := parser.tokenizer.Peek()
	if err != nil {
		parser.tokenizer.Next()
		parser.Errors = append(parser.Errors, err)
		return nil
	}

	if token.Type == kuuhaku_tokenizer.LUA_LITERAL || token.Type == kuuhaku_tokenizer.LUA_RETURN_LITERAL {
		parser.tokenizer.Next()
		var luaLiteralType LuaLiteralType = LUA_LITERAL_TYPE_MULTI_STMT
		if token.Type == kuuhaku_tokenizer.LUA_RETURN_LITERAL {
			luaLiteralType = LUA_LITERAL_TYPE_RETURN
		}
		return &LuaLiteral{
			LuaString: token.Content,
			Position:  token.Position,
			Type: luaLiteralType,
		}
	} else {
		return nil
	}
}
