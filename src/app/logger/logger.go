package logger

import (
	"fmt"
	"strings"
	"io/ioutil"
)

var buffer []string

func Print(format string, v ...interface{}) {
	fmt.Printf(format+"\n", v...)
}

func Error(format string, v ...interface{}) {
	fmt.Printf("ERROR: "+format+"\n", v...)
}

func Collect(format string, v ...interface{}) {
	buffer = append(buffer, fmt.Sprintf(format, v...))
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