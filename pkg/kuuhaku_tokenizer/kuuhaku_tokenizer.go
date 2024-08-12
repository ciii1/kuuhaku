package kuuhaku_tokenizer

import (
	"fmt"
)

var ErrTokenUnrecognized = fmt.Errorf("Token is unrecognized")

type TokenType int

const (
	RESERVED = iota
	REGEX_LITERAL
	STRING_LITERAL
	CAPTURE_GROUP
	OPENING_CURLY_BRACKET
	CLOSING_CURLY_BRACKET
	EQUAL_SIGN
)

type Token struct {
	tokenType TokenType
	content string
	position Position
}

type Position struct {
	column int
	line int
	raw int
}

type Tokenizer struct {
	position Position
	input string
}

func Init(input string) Tokenizer {
	return Tokenizer {
		position: Position {
			column: 1,
			line: 1,
			raw: 0,
		},
		input: input,
	}
}


func (tokenizer *Tokenizer) Next() (*Token, error) {
	tokenizer._ConsumeNewline()
	token := tokenizer._ConsumeReserved()
	if token != nil {
		return token, nil
	}
	return nil, ErrTokenUnrecognized
}

func (tokenizer *Tokenizer) _NextChar() byte {
	char := tokenizer._PeekChar()
	tokenizer.position.raw += 1
	tokenizer.position.column += 1

	return char
}

func (tokenizer *Tokenizer) _PeekChar() byte {
	return tokenizer.input[tokenizer.position.raw];
}

func (tokenizer *Tokenizer) _ConsumeNewline() {
	currChar := tokenizer._PeekChar()
	if currChar == '\n' {
		tokenizer.position.column = 1
		tokenizer.position.line += 1
		tokenizer._NextChar()	
	}
}

func (tokenizer *Tokenizer) _ConsumeWhitespace() {
	currChar := tokenizer._PeekChar()
	if currChar == ' ' || currChar == '\t' {
		tokenizer._NextChar()
	}
}

func (tokenizer *Tokenizer) _ConsumeReserved() *Token {
	position_raw := tokenizer.position.raw
	column := tokenizer.position.column
	line := tokenizer.position.line

	currChar := tokenizer._PeekChar();
	if !isRuneReserved(currChar) {
		return nil
	}

	tokenContent := ""

	isCurrCharBetween_0_9 := false
	for isRuneReserved(currChar) || isCurrCharBetween_0_9 {
		currChar = tokenizer._NextChar()
		tokenContent += string(currChar)
		isCurrCharBetween_0_9 = int(currChar) > int('0') && int(currChar) < int('9')
	}

	return &Token {
		position: Position {
			raw: position_raw,
			column: column,
			line: line,
		},
		tokenType: RESERVED,
		content: tokenContent,
	}
}

func isRuneReserved(char byte) bool {
	isCurrCharBetween_a_z := int(char) > int('a') && int(char) < int('z')
	isCurrCharBetween_A_Z := int(char) > int('A') && int(char) < int('Z')
	isCurrCharSpecialCharacters := char == '_' || char == '-'
	return isCurrCharBetween_a_z || isCurrCharBetween_A_Z || isCurrCharSpecialCharacters
}
