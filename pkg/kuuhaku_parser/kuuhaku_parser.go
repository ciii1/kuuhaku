package kuuhaku_parser

import (
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_tokenizer"
)

func Parse(input string) error {
	tokenizer := kuuhaku_tokenizer.Init(input);
	tokenizer.Next();
	return nil
}
