package formatter

import (
	"github.com/ciii1/kuuhaku/internal/config_reader"
	"github.com/ciii1/kuuhaku/internal/helper"
	"path/filepath"
	"errors"
	"os"
	"fmt"
)

func Format(filename string, spec_format_config string, isRecursive bool, tabNum int, whitespaceNum int) error {
	content, err := os.ReadFile(filename)
	helper.Check(err)
	fmt.Println("Format(), content:\n", string(content))
	format_config := spec_format_config
	if len(format_config) == 0 {
		format_config = filepath.Ext(filename)
	}
	err = config_reader.ReadFormat(format_config)
	if err != nil {
		if errors.Is(err, config_reader.ErrUnrecognizedExtension) {
			PrintError(filename, "File extension " + filepath.Ext(filename) + " is unrecognized")
			return config_reader.ErrUnrecognizedExtension
		}
	}
	return nil
}

func PrintError(filename string, message string) {
	fmt.Println("Formatting Error, File " + filename + ":", message)
}
