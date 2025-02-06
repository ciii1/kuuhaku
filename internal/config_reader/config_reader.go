package config_reader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ciii1/kuuhaku/internal/helper"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_analyzer"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_parser"
)

var ErrUnrecognizedExtension = fmt.Errorf("Extension is unrecognized")

func ReadConfig(extension string, isDebugAnalyzer bool, isDebugParser bool, isDebugReader bool) (*kuuhaku_analyzer.AnalyzerResult, []error) {
	entries, err := os.ReadDir(ConfigDir())
	helper.Check(err)
	if isDebugReader {
		fmt.Println("ReadConfig(), extension:", extension)
		fmt.Println("ReadConfig(), configs:")
	}

	formatFilePath := ""
	for _, entry := range entries {
		entryName := entry.Name()
		entryNameBase := filepath.Base(strings.TrimSuffix(entryName, filepath.Ext(entryName)))
		if isDebugReader {
			fmt.Println(entryName, entryNameBase)
		}
		if filepath.Base(entryNameBase) == extension[1:] && filepath.Ext(entryName) == ".khk" {
			if isDebugReader {
				fmt.Println(entry.Name())
			}
			formatFilePath = filepath.Join(ConfigDir(), entry.Name())
			break
		}
	}

	if formatFilePath == "" {
		return nil, []error{ErrUnrecognizedExtension}
	}

	formatGrammar, err := os.ReadFile(formatFilePath)
	helper.Check(err)
	if isDebugReader {
		fmt.Println(string(formatGrammar))
	}
	ast, errs := kuuhaku_parser.Parse(string(formatGrammar))
	if len(errs) != 0 {
		return nil, errs
	}
	res, errs := kuuhaku_analyzer.Analyze(&ast, isDebugAnalyzer)
	if len(errs) != 0 {
		return nil, errs
	}
	return &res, []error{}
}

func ConfigDir() string {
	homeDir, err := os.UserHomeDir()
	helper.Check(err)
	return filepath.Join(homeDir, ".config", "kuuhaku")
}
