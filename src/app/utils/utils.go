package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func ReadLines(filename string) ([]string, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("File %s does not exist", filename))
		os.Exit(1)
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Some error ocurred while reading file %s. %s", filename, err.Error()))
	}

	text := string(bytes)
	text = strings.Replace(text, "\r", "", -1)
	return strings.Split(text, "\n"), nil
}
