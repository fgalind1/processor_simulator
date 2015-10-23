package translator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"app/logger"
	"app/simulator/consts"
	"app/simulator/models/instruction/set"
	"app/utils"
)

func TranslateFromFile(filename string) (string, error) {

	// Read lines from file
	logger.Print(" => Reading assembly file: %s", filename)
	lines, err := utils.ReadLines(filename)
	if err != nil {
		return "", err
	}

	// Create output file
	outputFilename := getOutputFilename(filename)
	f, err := os.Create(outputFilename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Translate instructions
	instructionSet := set.Init()
	for i, line := range lines {
		instruction, err := instructionSet.GetInstructionFromString(line)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Failed translating: %s. %s", line, err.Error()))
		}
		hex := fmt.Sprintf("%08X", instruction.ToUint32())
		f.WriteString(fmt.Sprintf("%s ; 0x%04X => %s\n", hex, i*consts.BYTES_PER_WORD, strings.Replace(line, "\t", " ", -1)))
	}
	logger.Print(" => Output hex file: %s", outputFilename)
	return outputFilename, nil
}

func getOutputFilename(filename string) string {
	extension := filepath.Ext(filename)
	return filename[0:len(filename)-len(extension)] + ".hex"
}
