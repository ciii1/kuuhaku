package main

import (
	"flag"
    "os"
    "fmt"
	"github.com/ciii1/kuuhaku/internal/formatter"
)

func main() {
	flag.Usage = PrintHelp;
	var isRecursive = flag.Bool("recursive", false, "Process files recursively");
	var tabNum = flag.Int("tab", 0, "Don't format the file but replace indents to tab times the specified integer")
	var whitespaceNum = flag.Int("whitespace", 0, "Same as -tab but replace with whitespace times the specified integer")

	if (len(os.Args) > 1) {
		flag.Parse();
		fmt.Println("-recursive=", *isRecursive)
		fmt.Println("-tab=", *tabNum)
		fmt.Println("-whitespace=", *whitespaceNum)
		filename := flag.Arg(0)
		fmt.Println("Filename=", filename)
		formatter.Format(filename, *isRecursive, *tabNum, *whitespaceNum)
	} else {
		println("Expected at least 1 argument")
		PrintHelp()
	}
}

func PrintHelp() {
	println("Kuuhaku - A highly costumizable code formatter")
	println("")
	println("Usage:")
	println("kuuhaku <flags> <filename>")
	println("If filename is a directory, kuuhaku will process all of the files inside the directory")
	println("")
	println("Flags:")
	println("-recursive\t\tProcess files recursively")
	println("-tab=<int>\t\tDon't format the file but replace indents to tab times the specified integer")
	println("-whitespace=<int>\tSame as -tab but replace with whitespace times the specified integer")
	println("")
	println("Exiting...")
}
