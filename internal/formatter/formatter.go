package formatter

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/ciii1/kuuhaku/internal/config_reader"
	"github.com/ciii1/kuuhaku/internal/helper"
	"github.com/ciii1/kuuhaku/pkg/kuuhaku_runtime"
)

type FormattedFile struct {
	Content string
	Filename string
}

func Format(filename string, specFormatConfig string, isRecursive bool, isDebug bool) error {
	file, err := os.Stat(filename)
	helper.Check(err)
	var files []FormattedFile
	
	if file.IsDir() {
		files = getFilesRecursive(filename)
	} else {
		targetFile, err := os.ReadFile(filename)
		helper.Check(err)
		files = append(files, FormattedFile{
			Content: string(targetFile),
			Filename: filename,
		})
	}

	for _, formattedFile := range files {
		if isDebug {
			fmt.Println("Format(), content:\n", formattedFile.Content)
		}
		println("Formatting " + formattedFile.Filename + "...")
		formatConfig := specFormatConfig
		if len(formatConfig) == 0 {
			formatConfig = filepath.Ext(formattedFile.Filename)
		}
		res, errs := config_reader.ReadConfig(formatConfig)
		if len(errs) != 0 {
			println("Error while reading configuration, file " + filepath.Ext(formattedFile.Filename) + ":")
			helper.DisplayAllErrors(errs)
		}
		strRes, err := kuuhaku_runtime.Format(formattedFile.Content, res, true, isDebug)
		if err != nil {
			println("Error while formatting the code, file " + formattedFile.Filename + ":")
			print(err.Error())
		}

		f, err := os.OpenFile(formattedFile.Filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		defer f.Close()
		helper.Check(err)

		_, err = f.WriteString(strRes)
		helper.Check(err)
	}
	return nil
}

func getFilesRecursive(filename string) []FormattedFile {
	entries, err := os.ReadDir(filename)
	var files []FormattedFile
	helper.Check(err)
	for _, e := range entries {
		file, err := os.Stat(e.Name())
		helper.Check(err)
		if file.IsDir() {
			files = append(files, getFilesRecursive(e.Name())...)
		}
		if isTextFile(e.Name()) {
			content, err := os.ReadFile(e.Name())
			helper.Check(err)
			files = append(files, FormattedFile{
				Content: string(content),
				Filename: e.Name(),
			})
		}
	}
	return files
}

func isTextFile(filename string) bool {
	readFile, err := os.Open(filename)
	defer readFile.Close()
	helper.Check(err)
    fileScanner := bufio.NewScanner(readFile)
    fileScanner.Split(bufio.ScanLines)
    fileScanner.Scan()
    return utf8.ValidString(string(fileScanner.Text()))
}
