package kuuhaku_parser

import (
	"fmt"
	"strconv"
	"github.com/ciii1/kuuhaku/internal/helper"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_tokenizer"
)

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
