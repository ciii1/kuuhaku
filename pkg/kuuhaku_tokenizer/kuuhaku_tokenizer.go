package kuuhaku_tokenizer

import (
	"fmt"
	"strconv"
)

var ErrPatternUnrecognized = fmt.Errorf("Pattern is unrecognized")
var ErrStringLiteralUnterminated = fmt.Errorf("String literal is unterminated")
var ErrRegexLiteralUnterminated = fmt.Errorf("Regex literal is unterminated")

type TokenType int

const (
	IDENTIFIER = iota
	REGEX_LITERAL
	STRING_LITERAL
	CAPTURE_GROUP
	OPENING_CURLY_BRACKET
	CLOSING_CURLY_BRACKET
	EQUAL_SIGN
	EOF
)

type Token struct {
	Type TokenType
	Content string
	Position Position
}

type Position struct {
	Column int
	Line int
	Raw int
}

type Tokenizer struct {
	Position Position
	Input string
}

func Init(input string) Tokenizer {
	return Tokenizer {
		Position: Position {
			Column: 1,
			Line: 1,
			Raw: 0,
		},
		Input: input,
	}
}

func (tokenizer *Tokenizer) Next() (*Token, error) {
	isCurrentTrash := true
	for isCurrentTrash {
		isCurrentTrash = tokenizer.consumeNewline() || tokenizer.consumeWhitespace() || tokenizer.consumeComment()
	}
	if tokenizer.peekChar() == '\003' {
		return &Token {
			Position: Position {
				Raw: tokenizer.Position.Raw,
				Column: tokenizer.Position.Column,
				Line: tokenizer.Position.Line,
			},
			Type: EOF,
			Content: "\003",
		}, nil
	}
	token := tokenizer.consumeIdentifier()
	if token != nil {
		return token, nil
	}
	token, err := tokenizer.consumeStringLiteral()
	if err != nil {
		return nil, err
	}
	if token != nil {
		return token, nil
	}
	token, err = tokenizer.consumeRegexLiteral()
	if err != nil {
		return nil, err
	}
	if token != nil {
		return token, nil
	}
	return nil, ErrPatternUnrecognized
}

func (tokenizer *Tokenizer) nextChar() byte {
	tokenizer.Position.Raw += 1
	tokenizer.Position.Column += 1
	return tokenizer.peekChar()
}

func (tokenizer *Tokenizer) peekChar() byte {
	if tokenizer.Position.Raw >= len(tokenizer.Input) {
		return '\003' //ETX (end of text) https://wikipedia.org/wiki/ASCII
	}

	return tokenizer.Input[tokenizer.Position.Raw];
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

func (tokenizer *Tokenizer) consumeStringLiteral() (*Token, error) {
	positionRaw := tokenizer.Position.Raw
	column := tokenizer.Position.Column
	line := tokenizer.Position.Line
	content := ""

	startChar := tokenizer.peekChar()
	if startChar != '"' && startChar != '\'' {
		return nil, nil
	}

	prevPrevChar := byte(0)
	prevChar := tokenizer.peekChar()
	currChar := tokenizer.nextChar()
	for currChar != startChar || (prevChar == '\\' && prevPrevChar != '\\') {
		content += string(currChar)
		prevPrevChar = prevChar
		prevChar = tokenizer.peekChar()
		currChar = tokenizer.nextChar()	
		if currChar == '\n' || currChar == '\003' {
			return nil, ErrStringLiteralUnterminated
		}
	}
	tokenizer.nextChar()

	content, err := strconv.Unquote("\"" + content + "\"")
	if err != nil {
		return nil, err
	}

	return &Token {
		Position: Position {
			Raw: positionRaw,
			Column: column,
			Line: line,
		},
		Type: STRING_LITERAL,
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
		if currChar == '>' && prevChar == '\\' && prevPrevChar != '\\' {

		} else {
			content += string(prevChar)
		}
		if currChar == '\n' || currChar == '\003' {
			return nil, ErrRegexLiteralUnterminated
		}
	}
	tokenizer.nextChar()

	return &Token {
		Position: Position {
			Raw: positionRaw,
			Column: column,
			Line: line,
		},
		Type: REGEX_LITERAL,
		Content: content,
	}, nil
}

func (tokenizer *Tokenizer) consumeIdentifier() *Token {
	positionRaw := tokenizer.Position.Raw
	column := tokenizer.Position.Column
	line := tokenizer.Position.Line

	currChar := tokenizer.peekChar();
	if !isRuneIdentifier(currChar) {
		return nil
	}

	tokenContent := ""

	isCurrCharBetween_0_9 := false
	for isRuneIdentifier(currChar) || isCurrCharBetween_0_9 {
		tokenContent += string(currChar)
		currChar = tokenizer.nextChar()
		isCurrCharBetween_0_9 = int(currChar) >= int('0') && int(currChar) <= int('9')
	}

	return &Token {
		Position: Position {
			Raw: positionRaw,
			Column: column,
			Line: line,
		},
		Type: IDENTIFIER,
		Content: tokenContent,
	}
}

func isRuneIdentifier(char byte) bool {
	isCurrCharBetween_a_z := int(char) >= int('a') && int(char) <= int('z')
	isCurrCharBetween_A_Z := int(char) >= int('A') && int(char) <= int('Z')
	isCurrCharSpecialCharacters := char == '_' || char == '-'
	return isCurrCharBetween_a_z || isCurrCharBetween_A_Z || isCurrCharSpecialCharacters
}
