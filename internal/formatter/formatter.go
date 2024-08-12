package formatter

import (
	"github.com/ciii1/kuuhaku/internal/config_reader"
	"github.com/ciii1/kuuhaku/internal/helper"
	"path/filepath"
	"errors"
	"os"
	"fmt"
)

func Format(filename string, isRecursive bool, tabNum int, whitespaceNum int) error {
	content, err := os.ReadFile(filename)
	helper.Check(err)
	fmt.Println("Format(), content:\n", string(content))
	err = config_reader.ReadFormat(filepath.Ext(filename))
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
