package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ciii1/kuuhaku/internal/formatter"
)

func main() {
	flag.Usage = PrintHelp
	var isRecursive = flag.Bool("recursive", false, "Process files recursively")
	var isDebug = flag.Bool("debug", false, "Print debug messages")

	if len(os.Args) > 1 {
		println("Kuuhaku is still in its experimental state! Make sure to commit your project files using your version control program before running the formatter. The formatter will run in 3 seconds...")
		time.Sleep(3000000000)
		flag.Parse()
		if *isDebug {
			fmt.Println("-recursive=", *isRecursive)
			fmt.Println("-debug=", *isDebug)
		}
		filename := flag.Arg(0)
		configName := flag.Arg(1)
		if *isDebug {
			fmt.Println("Filename=", filename)
			if len(configName) == 0 {
				fmt.Println("no format provided")
			} else {
				fmt.Println("Format=", configName)
			}
		}
		formatter.Format(filename, configName, *isRecursive, *isDebug)
	} else {
		println("Expected at least 1 argument")
		PrintHelp()
	}
}

func PrintHelp() {
	println("Kuuhaku - A highly costumizable code formatter")
	println("")
	println("Usage:")
	println("kuuhaku <flags> <filename> <config_name>")
	println("Filename is the file to be formatted. If filename is a directory, kuuhaku will process all of the files inside the directory")
	println("Config name is the name of the format configuration to be used inside the kuuhaku's config directory ($HOME/.config/kuuhaku), without the .khk extension. If ommitted, the extension of files that are going to be formatted will be used")
	println("")
	println("Flags:")
	println("-recursive\t\tProcess directories recursively")
	println("-debug\t\tPrint debug messages")
	println("")
	println("Exiting...")
}
