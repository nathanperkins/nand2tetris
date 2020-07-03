package main

import (
	"fmt"
	"strconv"
)

var (
	comparisons = map[string]string{
		// use A-register
		"0":   "0101010",
		"1":   "0111111",
		"-1":  "0111010",
		"D":   "0001100",
		"A":   "0110000",
		"!D":  "0001101",
		"!A":  "0110001",
		"D+1": "0011111",
		"A+1": "0110111",
		"D-1": "0001110",
		"A-1": "0110010",
		"D+A": "0000010",
		"D-A": "0010011",
		"A-D": "0000111",
		"D&A": "0000000",
		"D|A": "0010101",

		// use RAM[A]
		"M":   "1110000",
		"!M":  "1110001",
		"M+1": "1110111",
		"M-1": "1110010",
		"D+M": "1000010",
		"D-M": "1010011",
		"M-D": "1000111",
		"D&M": "1000000",
		"D|M": "1010101",
	}

	destinations = map[string]string{
		"":    "000",
		"M":   "001",
		"D":   "010",
		"MD":  "011",
		"A":   "100",
		"AM":  "101",
		"AD":  "110",
		"AMD": "111",
	}

	jumps = map[string]string{
		"":    "000",
		"JGT": "001",
		"JEQ": "010",
		"JGE": "011",
		"JLT": "100",
		"JNE": "101",
		"JLE": "110",
		"JMP": "111",
	}
)

func instruction(line string) (string, error) {
	if matches := aInstructionRegex.FindStringSubmatch(line); len(matches) == 2 {
		value, err := strconv.ParseUint(matches[1], 10, 64)
		if err != nil {
			return "", fmt.Errorf("%q has an invalid value: %s", line, err)
		}

		return fmt.Sprintf("%016b", value), nil
	}

	if matches := cInstructionRegex.FindStringSubmatch(line); len(matches) == 4 {
		out := "111"
		dest, comp, jump := matches[1], matches[2], matches[3]

		if s, ok := comparisons[comp]; !ok {
			return "", fmt.Errorf("comparison for %q not recognized: %q", line, comp)
		} else {
			out += s
		}

		if s, ok := destinations[dest]; !ok {
			return "", fmt.Errorf("destination for %q not recognized: %q", line, dest)
		} else {
			out += s
		}

		if s, ok := jumps[jump]; !ok {
			return "", fmt.Errorf("jump for %q not recognized: %q", line, dest)
		} else {
			out += s
		}

		if len(out) != 16 {
			return "", fmt.Errorf("length of instruction for %q should be 16: %q", line, out)
		}

		return out, nil
	}

	return "", fmt.Errorf("%q is not recognized as valid instruction", line)
}
