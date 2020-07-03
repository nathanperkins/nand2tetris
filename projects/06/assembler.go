package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"unicode"
)

var (
	labelRegex        = regexp.MustCompile(`^\((.+)\)$`)
	aInstructionRegex = regexp.MustCompile(`^@(.+)$`)
	cInstructionRegex = regexp.MustCompile(`^([AMD]+)?=?([|&\!\-\w+]+);?(\w+)?$`)

	userVariablesStart uint = 16
	defaultSymbols          = map[string]uint{
		// Virtual registers
		"R0":  0,
		"R1":  1,
		"R2":  2,
		"R3":  3,
		"R4":  4,
		"R5":  5,
		"R6":  6,
		"R7":  7,
		"R8":  8,
		"R9":  9,
		"R10": 10,
		"R11": 11,
		"R12": 12,
		"R13": 13,
		"R14": 14,
		"R15": 15,

		// Predefined pointers
		"SP":   0,
		"LCL":  1,
		"ARG":  2,
		"THIS": 3,
		"THAT": 4,

		// IO pointers
		"SCREEN": 0x4000,
		"KBD":    0x6000,
	}
)

type assemblerState int

const (
	stateNew = iota
	stateFileRead
	stateVariablesProcessed
	stateSymbolsReplaced
	stateTranslated
)

type Assembler struct {
	state       assemblerState
	lines       []string
	symbolTable map[string]uint
}

func New() *Assembler {
	a := &Assembler{
		state:       stateNew,
		symbolTable: make(map[string]uint),
	}
	a.addDefaultSymbols()

	return a
}

func (a *Assembler) addDefaultSymbols() {
	for k, v := range defaultSymbols {
		a.symbolTable[k] = v
	}
}

func (a *Assembler) read(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatalf("Could not read: %s", err)
	}

	// Clean lines and store in assembler.
	for _, line := range strings.Split(string(b), "\n") {
		line = clean(line)

		if line != "" {
			a.lines = append(a.lines, line)
		}
	}

	a.state = stateFileRead
	return nil
}

func (a *Assembler) populateSymbolsTable() error {
	var instructionsOnly []string
	// Look for labels, like "(LOOP)", and add them to the symbol table.
	for _, line := range a.lines {
		matches := labelRegex.FindStringSubmatch(line)
		if len(matches) != 2 {
			instructionsOnly = append(instructionsOnly, line)
			continue
		}

		// This is label, add it to the symbol table if needed.
		symbol := matches[1]
		if _, ok := a.symbolTable[symbol]; !ok {
			value := len(instructionsOnly)
			a.symbolTable[symbol] = uint(value)
		}
	}
	a.lines = instructionsOnly

	// Look for A-instructions, like "@i", and add them to the symbol table.
	var nextVar uint = userVariablesStart
	for _, line := range a.lines {
		matches := aInstructionRegex.FindStringSubmatch(line)
		if len(matches) != 2 {
			continue
		}
		symbol := matches[1]
		if isNumber(symbol) {
			continue
		}

		if _, ok := a.symbolTable[symbol]; !ok {
			a.symbolTable[symbol] = nextVar
			nextVar++
		}
	}

	a.state = stateVariablesProcessed
	return nil
}

func (a *Assembler) replaceSymbols() error {
	for i, line := range a.lines {
		matches := aInstructionRegex.FindStringSubmatch(line)
		if len(matches) != 2 {
			continue
		}

		symbol := matches[1]
		if isNumber(symbol) {
			continue
		}

		value, ok := a.symbolTable[symbol]
		if !ok {
			return fmt.Errorf("could not find symbol: %q", symbol)
		}

		a.lines[i] = fmt.Sprintf("@%d", value)
	}

	a.state = stateSymbolsReplaced
	return nil
}

func (a *Assembler) translate() error {
	for i, line := range a.lines {
		if s, err := instruction(line); err != nil {
			return fmt.Errorf("translation error: %s", err)
		} else {
			a.lines[i] = s
			continue
		}
	}

	a.state = stateTranslated
	return nil
}

func (a *Assembler) Read(r io.Reader) error {
	if err := a.read(r); err != nil {
		return err
	}
	if err := a.populateSymbolsTable(); err != nil {
		return err
	}
	if err := a.replaceSymbols(); err != nil {
		return err
	}
	if err := a.translate(); err != nil {
		return err
	}

	return nil
}

// Write performs all operations to convert assembly language to HACK machine language and writes the resulting executable to w.
func (a *Assembler) Write(w io.Writer) error {
	buf := bufio.NewWriter(w)
	for _, line := range a.lines {
		if _, err := buf.WriteString(line + "\n"); err != nil {
			log.Fatalf("Could not write %q: %s", line, err)
		}
	}

	return buf.Flush()
}

func main() {
	inputFilePath := flag.String("input_file", "", "asm file to be compiled to HACK machine language")
	outputFilePath := flag.String("output_file", "", "output path for compiled program")
	flag.Parse()

	inputFile, err := os.Open(*inputFilePath)
	if err != nil {
		log.Fatalf("--input_file: %s", err)
	}

	outputFile, err := os.Create(*outputFilePath)
	if err != nil {
		log.Fatalf("--output_file: %s", err)
	}

	assembler := New()

	if err := assembler.Read(inputFile); err != nil {
		log.Fatalf("Failed to read file %q: %s", *inputFilePath, err)
	}

	if err := assembler.Write(outputFile); err != nil {
		log.Fatalf("Failed to assemble %q to %q: %s", *inputFilePath, *outputFilePath, err)
	}
}

// clean removes comments and empty lines and trims whitespace.
func clean(s string) string {
	s = strings.SplitN(s, "//", 2)[0]
	s = strings.Join(strings.Fields(s), "")
	return s
}

func isNumber(s string) bool {
	for _, r := range s {
		if !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}
