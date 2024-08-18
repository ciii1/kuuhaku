package kuuhaku_parser

import (
	"testing"

	"github.com/ciii1/kuuhaku/pkg/kuuhaku_tokenizer"
)

func TestConsumeMatchRules(t *testing.T) {
	tokenizer := kuuhaku_tokenizer.Init("<.*>hello<[0-9]>hi=");
	matchRulesP, err := consumeMatchRules(&tokenizer)
	if err != nil {
		panic(err)
	}
	matchRules := *matchRulesP
	if len(matchRules) != 4 {
		println("Expected matchRules length to be 4")
		t.Fail()
	}
	node1, ok := matchRules[0].(RegexLiteral)
	if ok == false {
		println("Expected matchRules[0] to be a regex literal")
		t.Fail()
	} else if node1.RegexString != ".*" {
		println("Expected matchRules[0] to contain \".*\"")
		t.Fail()
	}

	node2, ok := matchRules[1].(Identifer)
	if ok == false {
		println("Expected matchRules[1] to be an identifier")
		t.Fail()
	} else if node2.Name != "hello" {
		println("Expected matchRules[1] to contain \"hello\"")
		t.Fail()
	}

	node3, ok := matchRules[2].(RegexLiteral)
	if ok == false {
		println("Expected matchRules[2] to be a regex literal")
		t.Fail()
	} else if node3.RegexString != "[0-9]" {
		println("Expected matchRules[2] to contain \"[0-9]\"")
		t.Fail()
	}

	node4, ok := matchRules[3].(Identifer)
	if ok == false {
		println("Expected matchRules[3] to be an identifier")
		t.Fail()
	} else if node4.Name != "hi" {
		println("Expected matchRules[3] to contain \"hi\"")
		t.Fail()
	}
}

func TestConsumeReplaceRules(t *testing.T) {
	tokenizer := kuuhaku_tokenizer.Init("\"\\t\"len$0$2 \"hi\"");
	replaceRulesP, err := consumeReplaceRules(&tokenizer)
	if err != nil {
		panic(err)
	}
	replaceRules := *replaceRulesP
	if len(replaceRules) != 3 {
		println("Expected replaceRules length to be 3")
		t.Fail()
	}
	node1, ok := replaceRules[0].(Len)
	if ok == false {
		println("Expected matchRules[0] to be len")
		t.Fail()
	} else if firstArg, ok := node1.FirstArgument.(StringLiteral); ok {
		if firstArg.String != "\t" {
			println("Expected Len.FirstArgument to contain \"\\t\"")
			t.Fail()
		}
		if secondArg, ok := node1.SecondArgument.(CaptureGroup); ok {
			if secondArg.Number != 0 {
				println("Expected Len.SecondArgument to contain 0")
				t.Fail()
			}
		} else {
			println("Expected Len.SecondArgument to be a capture group")
			t.Fail()
		}

	} else {
		println("Expected Len.FirstArgument to be a string literal")
		t.Fail()
	}

	node2, ok := replaceRules[1].(CaptureGroup)
	if ok == false {
		println("Expected matchRules[1] to be a capture group")
		t.Fail()
	} else if node2.Number != 2 {
		println("Expected matchRules[1] to contain 2")
		t.Fail()
	}

	node3, ok := replaceRules[2].(StringLiteral)
	if ok == false {
		println("Expected matchRules[2] to be a string literal")
		t.Fail()
	} else if node3.String != "hi" {
		println("Expected matchRules[2] to contain \"hi\"")
		t.Fail()
	}
}
