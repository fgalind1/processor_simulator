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

func TranslateFromFile(filename string, outputFilename string) (string, error) {

	// Read lines from file
	logger.Print(" => Reading assembly file: %s", filename)
	lines, err := utils.ReadLines(filename)
	if err != nil {
		return "", err
	}

	// Create output file
	f, err := os.Create(outputFilename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Clean lines, remove labels and get map of labels
	lines, labels := getLinesAndMapLabels(lines)

	// Translate instructions
	instructionSet := set.Init()
	for i, line := range lines {
		instruction, err := instructionSet.GetInstructionFromString(line, uint32(i)*consts.BYTES_PER_WORD, labels)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Failed translating line %d: %s. %s", i, line, err.Error()))
		}
		hex := fmt.Sprintf("%08X", instruction.ToUint32())
		f.WriteString(fmt.Sprintf("%s // 0x%04X => %s\n", hex, i*consts.BYTES_PER_WORD, strings.Replace(line, "\t", " ", -1)))
	}
	logger.Print(" => Output hex file: %s", outputFilename)
	return outputFilename, nil
}

func getLinesAndMapLabels(lines []string) ([]string, map[string]uint32) {
	cleanLines := []string{}
	labels := map[string]uint32{}
	address := uint32(0)
	for _, line := range lines {
		// Remove comments in the right
		line = strings.Split(line, ";")[0]
		// Check if line is only a comment or an empty line
		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0:1] == ";" {
			continue
		}
		// Assert if line is a label
		if line[len(line)-1:] == ":" {
			labels[line[:len(line)-1]] = address
		} else {
			cleanLines = append(cleanLines, line)
			address += consts.BYTES_PER_WORD
		}
	}
	return cleanLines, labels
}

func getOutputFilename(filename string) string {
	extension := filepath.Ext(filename)
	return filename[0:len(filename)-len(extension)] + ".hex"
}
