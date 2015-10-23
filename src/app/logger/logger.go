package logger

import "fmt"

func Print(format string, v ...interface{}) {
	fmt.Printf(format+"\n", v...)
}

func Error(format string, v ...interface{}) {
	fmt.Printf("ERROR: "+format+"\n", v...)
}
