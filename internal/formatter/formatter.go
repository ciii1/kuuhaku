package formatter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ciii1/kuuhaku/internal/config_reader"
	"github.com/ciii1/kuuhaku/internal/helper"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_runtime"
)

func Format(filename string, specFormatConfig string, isRecursive bool, tabNum int, whitespaceNum int) error {
	content, err := os.ReadFile(filename)
	helper.Check(err)
	fmt.Println("Format(), content:\n", string(content))
	formatConfig := specFormatConfig
	if len(formatConfig) == 0 {
		formatConfig = filepath.Ext(filename)
	}
	res, errs := config_reader.ReadFormat(formatConfig)
	if len(errs) != 0 {
		println("Error while reading configuration, file " + filepath.Ext(filename) + ":")
		helper.DisplayAllErrors(errs)
	}
	kuuhaku_runtime.Format(string(content), res)
	return nil
}
