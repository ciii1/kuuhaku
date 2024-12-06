package kuuhaku_tokenizer

import (
	"fmt"
)

type TokenizeErrorType int

const (
	PATTERN_UNRECOGNIZED = iota
	REGEX_LITERAL_UNTERMINATED
	LUA_LITERAL_UNTERMINATED
	ILLEGAL_CAPTURE_GROUP
)

type TokenizeError struct {
	Position Position
	Message  string
	Type     TokenizeErrorType
}

func (e TokenizeError) Error() string {
	return fmt.Sprintf("Tokenize error (%d, %d): %s", e.Position.Line, e.Position.Column, e.Message)
}

func ErrPatternUnrecognized(tokenizer Tokenizer) *TokenizeError {
	return &TokenizeError{
		Message:  "Pattern is unrecognized",
		Position: tokenizer.Position,
		Type:     PATTERN_UNRECOGNIZED,
	}
}
func ErrRegexLiteralUnterminated(tokenizer Tokenizer) *TokenizeError {
	return &TokenizeError{
		Message:  "Regex literal is not terminated",
		Position: tokenizer.Position,
		Type:     REGEX_LITERAL_UNTERMINATED,
	}
}

// we want the position instead of the tokenizer because the position we have to display is not the current position
func ErrLuaLiteralUnterminated(pos Position) *TokenizeError {
	return &TokenizeError{
		Message:  "Lua literal is not terminated",
		Position: pos,
		Type:     LUA_LITERAL_UNTERMINATED,
	}
}

type TokenType int

const (
	IDENTIFIER = iota
	REGEX_LITERAL
	STRING_LITERAL
	LUA_LITERAL
	OPENING_CURLY_BRACKET
	CLOSING_CURLY_BRACKET
	OPENING_BRACKET
	CLOSING_BRACKET
	COMMA
	EQUAL_SIGN
	SEARCH_MODE_KEYWORD
	EOF
)

type Token struct {
	Type     TokenType
	Content  string
	Position Position
}

type Position struct {
	Column int
	Line   int
	Raw    int
}

type Tokenizer struct {
	Position     Position
	PrevPosition Position
	currToken    *Token
	currError    error
	Input        string
}

func Init(input string) Tokenizer {
	tokenizer := Tokenizer{
		Position: Position{
			Column: 1,
			Line:   1,
			Raw:    0,
		},
		currToken: nil,
		currError: nil,
		Input:     input,
	}
	tokenizer.Next() //Peek() would look at currToken so we need to initialize it first
	return tokenizer
}

func (tokenizer *Tokenizer) Peek() (*Token, error) {
	return tokenizer.currToken, tokenizer.currError
}

func (tokenizer *Tokenizer) Next() (*Token, error) {
	isCurrentTrash := true
	for isCurrentTrash {
		isCurrentTrash = tokenizer.consumeNewline() || tokenizer.consumeWhitespace() || tokenizer.consumeComment()
	}
	if tokenizer.peekChar() == '\003' {
		return tokenizer.returnToken(&Token{
			Position: Position{
				Raw:    tokenizer.Position.Raw,
				Column: tokenizer.Position.Column,
				Line:   tokenizer.Position.Line,
			},
			Type:    EOF,
			Content: "\003",
		}, nil)
	}
	tokenizer.PrevPosition = tokenizer.Position
	token := tokenizer.consumeIdentifierOrKeyword()
	if token != nil {
		return tokenizer.returnToken(token, nil)
	}

	token = tokenizer.consumeOpeningCurlyBracket()
	if token != nil {
		return tokenizer.returnToken(token, nil)
	}

	token = tokenizer.consumeClosingCurlyBracket()
	if token != nil {
		return tokenizer.returnToken(token, nil)
	}

	token = tokenizer.consumeEqualSign()
	if token != nil {
		return tokenizer.returnToken(token, nil)
	}

	token = tokenizer.consumeComma()
	if token != nil {
		return tokenizer.returnToken(token, nil)
	}

	token = tokenizer.consumeClosingBracket()
	if token != nil {
		return tokenizer.returnToken(token, nil)
	}

	token = tokenizer.consumeOpeningBracket()
	if token != nil {
		return tokenizer.returnToken(token, nil)
	}

	token, err := tokenizer.consumeLuaLiteral()
	if err != nil {
		return tokenizer.returnToken(nil, err)
	}
	if token != nil {
		return tokenizer.returnToken(token, nil)
	}

	token, err = tokenizer.consumeRegexLiteral()
	if err != nil {
		return tokenizer.returnToken(nil, err)
	}
	if token != nil {
		return tokenizer.returnToken(token, nil)
	}

	tokenizerState := *tokenizer
	tokenizer.nextChar()
	return tokenizer.returnToken(nil, ErrPatternUnrecognized(tokenizerState))
}

func (tokenizer *Tokenizer) returnToken(token *Token, err error) (*Token, error) {
	tokenizer.currToken = token
	tokenizer.currError = err
	return token, err
}

func (tokenizer *Tokenizer) nextChar() byte {
	tokenizer.Position.Raw += 1
	tokenizer.Position.Column += 1
	return tokenizer.peekChar()
}

// never use this at the start of a lexing function
func (tokenizer *Tokenizer) prevChar() byte {
	tokenizer.Position.Raw -= 1
	tokenizer.Position.Column -= 1
	return tokenizer.peekChar()
}

func (tokenizer *Tokenizer) peekChar() byte {
	if tokenizer.Position.Raw >= len(tokenizer.Input) {
		return '\003' //ETX (end of text) https://wikipedia.org/wiki/ASCII
	}

	return tokenizer.Input[tokenizer.Position.Raw]
}

func (tokenizer *Tokenizer) consumeNewline() bool {
	currChar := tokenizer.peekChar()
	if currChar == '\n' {
		tokenizer.nextChar()
		tokenizer.Position.Column = 1
		tokenizer.Position.Line += 1
		return true
	} else {
		return false
	}
}

func (tokenizer *Tokenizer) consumeWhitespace() bool {
	currChar := tokenizer.peekChar()
	if currChar == ' ' || currChar == '\t' {
		tokenizer.nextChar()
		return true
	} else {
		return false
	}
}

func (tokenizer *Tokenizer) consumeComment() bool {
	currChar := tokenizer.peekChar()
	if currChar == '#' {
		for currChar != '\n' {
			currChar = tokenizer.nextChar()
		}
		return true
	} else {
		return false
	}
}

func (tokenizer *Tokenizer) consumeOpeningCurlyBracket() *Token {
	positionRaw := tokenizer.Position.Raw
	column := tokenizer.Position.Column
	line := tokenizer.Position.Line

	currChar := tokenizer.peekChar()
	if currChar != '{' {
		return nil
	}

	tokenizer.nextChar()

	return &Token{
		Position: Position{
			Raw:    positionRaw,
			Column: column,
			Line:   line,
		},
		Type:    OPENING_CURLY_BRACKET,
		Content: "{",
	}
}

func (tokenizer *Tokenizer) consumeClosingCurlyBracket() *Token {
	positionRaw := tokenizer.Position.Raw
	column := tokenizer.Position.Column
	line := tokenizer.Position.Line

	currChar := tokenizer.peekChar()
	if currChar != '}' {
		return nil
	}

	tokenizer.nextChar()

	return &Token{
		Position: Position{
			Raw:    positionRaw,
			Column: column,
			Line:   line,
		},
		Type:    CLOSING_CURLY_BRACKET,
		Content: "}",
	}
}

func (tokenizer *Tokenizer) consumeEqualSign() *Token {
	positionRaw := tokenizer.Position.Raw
	column := tokenizer.Position.Column
	line := tokenizer.Position.Line

	currChar := tokenizer.peekChar()
	if currChar != '=' {
		return nil
	}

	tokenizer.nextChar()

	return &Token{
		Position: Position{
			Raw:    positionRaw,
			Column: column,
			Line:   line,
		},
		Type:    EQUAL_SIGN,
		Content: "=",
	}
}

func (tokenizer *Tokenizer) consumeOpeningBracket() *Token {
	positionRaw := tokenizer.Position.Raw
	column := tokenizer.Position.Column
	line := tokenizer.Position.Line

	currChar := tokenizer.peekChar()
	if currChar != '(' {
		return nil
	}

	tokenizer.nextChar()

	return &Token{
		Position: Position{
			Raw:    positionRaw,
			Column: column,
			Line:   line,
		},
		Type:    OPENING_BRACKET,
		Content: "(",
	}
}

func (tokenizer *Tokenizer) consumeClosingBracket() *Token {
	positionRaw := tokenizer.Position.Raw
	column := tokenizer.Position.Column
	line := tokenizer.Position.Line

	currChar := tokenizer.peekChar()
	if currChar != ')' {
		return nil
	}

	tokenizer.nextChar()

	return &Token{
		Position: Position{
			Raw:    positionRaw,
			Column: column,
			Line:   line,
		},
		Type:    CLOSING_BRACKET,
		Content: ")",
	}
}

func (tokenizer *Tokenizer) consumeComma() *Token {
	positionRaw := tokenizer.Position.Raw
	column := tokenizer.Position.Column
	line := tokenizer.Position.Line

	currChar := tokenizer.peekChar()
	if currChar != ',' {
		return nil
	}

	tokenizer.nextChar()

	return &Token{
		Position: Position{
			Raw:    positionRaw,
			Column: column,
			Line:   line,
		},
		Type:    COMMA,
		Content: ",",
	}
}

func (tokenizer *Tokenizer) consumeLuaLiteral() (*Token, error) {
	positionRaw := tokenizer.Position.Raw
	column := tokenizer.Position.Column
	line := tokenizer.Position.Line
	content := ""

	currChar := tokenizer.peekChar()
	if currChar != '`' {
		return nil, nil
	}
	currChar = tokenizer.nextChar()
	if currChar != '`' {
		tokenizer.prevChar()
		return nil, nil
	}

	prevPrevPrevChar := byte(0)
	prevPrevChar := tokenizer.peekChar()
	prevChar := tokenizer.nextChar()
	currChar = tokenizer.nextChar()
	for currChar != '`' || prevChar != '`' || (prevPrevChar == '\\' && prevPrevPrevChar != '\\') {
		prevPrevPrevChar = prevPrevChar
		prevPrevChar = prevChar
		prevChar = tokenizer.peekChar()
		currChar = tokenizer.nextChar()
		if currChar == '`' && prevChar == '`' && prevPrevChar == '\\' && prevPrevPrevChar != '\\' {

		} else {
			content += string(prevPrevChar)
		}
		if currChar == '\003' {
			return nil, ErrLuaLiteralUnterminated(Position{
				Raw:    positionRaw,
				Column: column,
				Line:   line,
			})
		}
	}
	tokenizer.nextChar()

	return &Token{
		Position: Position{
			Raw:    positionRaw,
			Column: column,
			Line:   line,
		},
		Type:    LUA_LITERAL,
		Content: content,
	}, nil
}

func (tokenizer *Tokenizer) consumeRegexLiteral() (*Token, error) {
	positionRaw := tokenizer.Position.Raw
	column := tokenizer.Position.Column
	line := tokenizer.Position.Line
	content := ""

	currChar := tokenizer.peekChar()
	if currChar != '<' {
		return nil, nil
	}

	prevPrevChar := byte(0)
	prevChar := tokenizer.peekChar()
	currChar = tokenizer.nextChar()
	for currChar != '>' || (prevChar == '\\' && prevPrevChar != '\\') {
		prevPrevChar = prevChar
		prevChar = tokenizer.peekChar()
		currChar = tokenizer.nextChar()
		//For escaping characters
		if currChar == '>' && prevChar == '\\' && prevPrevChar != '\\' {

		} else {
			content += string(prevChar)
		}
		if currChar == '\n' || currChar == '\003' {
			return nil, ErrRegexLiteralUnterminated(*tokenizer)
		}
	}
	tokenizer.nextChar()

	return &Token{
		Position: Position{
			Raw:    positionRaw,
			Column: column,
			Line:   line,
		},
		Type:    REGEX_LITERAL,
		Content: content,
	}, nil
}

func (tokenizer *Tokenizer) consumeIdentifierOrKeyword() *Token {
	positionRaw := tokenizer.Position.Raw
	column := tokenizer.Position.Column
	line := tokenizer.Position.Line

	currChar := tokenizer.peekChar()
	if !isRuneIdentifier(currChar) {
		return nil
	}

	tokenContent := ""

	isCurrCharBetween_0_9 := false
	for isRuneIdentifier(currChar) || isCurrCharBetween_0_9 {
		tokenContent += string(currChar)
		currChar = tokenizer.nextChar()
		isCurrCharBetween_0_9 = isRuneNumber(currChar)
	}

	var tokenType TokenType
	if tokenContent == "SEARCH_MODE" {
		tokenType = SEARCH_MODE_KEYWORD
	} else {
		tokenType = IDENTIFIER
	}

	return &Token{
		Position: Position{
			Raw:    positionRaw,
			Column: column,
			Line:   line,
		},
		Type:    tokenType,
		Content: tokenContent,
	}
}

func isRuneIdentifier(char byte) bool {
	isCurrCharBetween_a_z := int(char) >= int('a') && int(char) <= int('z')
	isCurrCharBetween_A_Z := int(char) >= int('A') && int(char) <= int('Z')
	isCurrCharSpecialCharacters := char == '_' || char == '-'
	return isCurrCharBetween_a_z || isCurrCharBetween_A_Z || isCurrCharSpecialCharacters
}

func isRuneNumber(char byte) bool {
	return int(char) >= int('0') && int(char) <= int('9')
}
