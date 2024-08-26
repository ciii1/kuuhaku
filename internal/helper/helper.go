package helper

import (
	"strconv"
)

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func DisplayAllErrors(errs []error) {
	for i, err := range errs {
		if err != nil {
			i_str := strconv.Itoa(i)
			println(i_str + ". " + err.Error())
		} else {
			println("Found nil pointer")
		}
	}
}

func EmptyStringByValue(strings *[]string, value string) {
	for i, s := range *strings {
		if s == value {
			(*strings)[i] = ""
		}
	}
}
