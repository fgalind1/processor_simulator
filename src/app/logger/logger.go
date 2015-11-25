package logger

import (
	"fmt"
	"strings"
	"io/ioutil"
)

var buffer []string
var verboseDebug bool
var verboseQuiet bool

func SetVerboseQuiet(value bool) {
	verboseQuiet = value
}

func SetVerboseDebug(value bool) {
	verboseDebug = value
}

func Print(format string, v ...interface{}) {
	if !verboseQuiet {
	fmt.Printf(format+"\n", v...)
	} else {
		buffer = append(buffer, fmt.Sprintf(format, v...))
	}
}

func Error(format string, v ...interface{}) {
	fmt.Printf("ERROR: "+format+"\n", v...)
}

func Collect(format string, v ...interface{}) {
	buffer = append(buffer, fmt.Sprintf(format, v...))
	if verboseDebug {
		fmt.Printf(format+"\n", v...)
	}
}

func CleanBuffer() {
	buffer = []string{}
}

func WriteBuffer(filename string) error {
	err := ioutil.WriteFile(filename, []byte(strings.Join(buffer, "\n")), 0644)
	if err != nil {
		return err
	}
	CleanBuffer()
	return nil
}