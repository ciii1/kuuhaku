package config_reader

import (
	"os"
	"fmt"
	"path/filepath"
	"strings"
	"github.com/ciii1/kuuhaku/internal/helper"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_parser"
	"github.com/kr/pretty"
)

var ErrUnrecognizedExtension = fmt.Errorf("Extension is unrecognized")

func ReadFormat(extension string) error {
	entries, err := os.ReadDir(FormatsDir())
	helper.Check(err)
	fmt.Println("ReadFormat(), extension:", extension)
	fmt.Println("ReadFormat(), configs:")

	formatFilePath := ""
	for _, entry := range entries {
		entryName := entry.Name()
		entryNameBase := filepath.Base(strings.TrimSuffix(entryName, filepath.Ext(entryName)))
		fmt.Println(entryName, entryNameBase)
		if filepath.Base(entryNameBase) == extension[1:] && filepath.Ext(entryName) == ".khk" {
			fmt.Println(entry.Name())
			formatFilePath = filepath.Join(FormatsDir(), entry.Name())
			break
		}
	}

	if formatFilePath == "" {
		return ErrUnrecognizedExtension
	} else {
		formatGrammar, err := os.ReadFile(formatFilePath) 
		helper.Check(err)
		fmt.Println(string(formatGrammar))
		ast, errs := kuuhaku_parser.Parse(string(formatGrammar))
		fmt.Printf("%# v\n", pretty.Formatter(ast))
		helper.DisplayAllErrors(errs)		
	}

	return nil
}

func FormatsDir() string {
	homeDir, err := os.UserHomeDir()
	helper.Check(err)
	return filepath.Join(homeDir, ".config", "kuuhaku", "formats")
}
