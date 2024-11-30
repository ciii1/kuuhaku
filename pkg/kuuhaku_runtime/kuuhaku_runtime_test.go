package kuuhaku_runtime

import (
	"strconv"
	"testing"

	"github.com/ciii1/kuuhaku/internal/helper"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_analyzer"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_parser"
)

func TestRuntime1(t *testing.T) {
	ast, errs := kuuhaku_parser.Parse("SEARCH_MODE E{C D = $0 \" \" $1 \" \"} C{<a>} D{<b>}")
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	res, errs := kuuhaku_analyzer.Analyze(&ast)
	if len(errs) != 0 {
		println("Expected analyzer errors length to be 0, got " + strconv.Itoa(len(errs)))
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	strRes, err := Format("abababaa", &res)

	if len(errs) != 0 {
		println("Expected runtime errors length to be 0, got " + strconv.Itoa(len(errs)))
		println(err.Error())
		t.Fatal()
	}

	if strRes != "a b a b a b aa" {
		println("Expected the string to be \"a b a b a b aa\", got " + strRes)
		t.Fatal()
	}
}

func TestRuntime2(t *testing.T) {
	println("TestRuntime2:")
	ast, errs := kuuhaku_parser.Parse("SEARCH_MODE E{C D C = $0 $1 \"\\n\" $2} C{<[A-Za-z0-9]+>} D{<\\.>}")
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	res, errs := kuuhaku_analyzer.Analyze(&ast)
	if len(errs) != 0 {
		println("Expected analyzer errors length to be 0, got " + strconv.Itoa(len(errs)))
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	strRes, err := Format("test.Hello test2.Hello3", &res)

	if err != nil {
		println("Expected runtime errors length to be 0, got " + strconv.Itoa(len(errs)))
		println(err.Error())
		t.Fatal()
	}

	if strRes != "test.\nHello test2.\nHello3" {
		println("Expected the string to be \"test.\nHello test2.\nHello3\", got \"" + strRes + "\"")
		t.Fatal()
	}
}

func TestRuntime3(t *testing.T) {
	println("TestRuntime3:")
	ast, errs := kuuhaku_parser.Parse("SEARCH_MODE E{E l nl = $0 $1 $2 \"\\t\"} E{l nl = $0 $1 \"\\t\"} l{<test>} nl{<\\n>}")
	if len(errs) != 0 {
		println("Expected parser errors length to be 0")
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	res, errs := kuuhaku_analyzer.Analyze(&ast)
	if len(errs) != 0 {
		println("Expected analyzer errors length to be 0, got " + strconv.Itoa(len(errs)))
		helper.DisplayAllErrors(errs)
		t.Fatal()
	}
	strRes, err := Format("test\ntest\ntest\ntest", &res)

	if err != nil {
		println("Expected runtime errors length to be 0, got " + strconv.Itoa(len(errs)))
		println(err.Error())
		t.Fatal()
	}

	if strRes != "test.\nHello test2.\nHello3" {
		println("Expected the string to be \"test.\nHello test2.\nHello3\", got \"" + strRes + "\"")
		t.Fatal()
	}
}
