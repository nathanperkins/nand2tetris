package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLabelRegex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "Label",
			input: "(LOOP)",
			want:  []string{"(LOOP)", "LOOP"},
		},
		{
			name:  "A instruction",
			input: "@LOOP",
			want:  nil,
		},
		{
			name:  "C instruction",
			input: "D=M+1;JMP",
			want:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := labelRegex.FindStringSubmatch(test.input)
			if diff := cmp.Diff(got, test.want); diff != "" {
				t.Errorf("Label submatches for %q didn't match expected (-got,+want):\n%s", test.input, diff)
			}
		})
	}
}

func TestAInstructionRegex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "Label",
			input: "(LOOP)",
			want:  nil,
		},
		{
			name:  "A instruction",
			input: "@LOOP",
			want:  []string{"@LOOP", "LOOP"},
		},
		{
			name:  "A instruction/With Underscore",
			input: "@USER_NAME",
			want:  []string{"@USER_NAME", "USER_NAME"},
		},
		{
			name:  "C instruction",
			input: "D=M+1;JMP",
			want:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := aInstructionRegex.FindStringSubmatch(test.input)
			if diff := cmp.Diff(got, test.want); diff != "" {
				t.Errorf("Label submatches for %q didn't match expected (-got,+want):\n%s", test.input, diff)
			}
		})
	}
}

func TestCInstructionRegex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "Label",
			input: "(LOOP)",
			want:  nil,
		},
		{
			name:  "A instruction",
			input: "@LOOP",
			want:  nil,
		},
		{
			name:  "C instruction/Complete",
			input: "D=M+1;JMP",
			want:  []string{"D=M+1;JMP", "D", "M+1", "JMP"},
		},
		{
			name:  "C instruction/No Assignment",
			input: "D;JMP",
			want:  []string{"D;JMP", "", "D", "JMP"},
		},
		{
			name:  "C instruction/No Jump",
			input: "D=M+1",
			want:  []string{"D=M+1", "D", "M+1", ""},
		},
		{
			name:  "C instruction/With hyphen",
			input: "D=M-1",
			want:  []string{"D=M-1", "D", "M-1", ""},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := cInstructionRegex.FindStringSubmatch(test.input)
			if diff := cmp.Diff(got, test.want); diff != "" {
				t.Errorf("Label submatches for %q didn't match expected (-got,+want):\n%s", test.input, diff)
			}
		})
	}
}

func TestClean(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "comment",
			input: "// blah",
			want:  "",
		},
		{
			name:  "inline comment",
			input: "@LOOP // loop var",
			want:  "@LOOP",
		},
		{
			name:  "extra whitespace",
			input: "     D = M +     1;   JMP    // comment",
			want:  "D=M+1;JMP",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := clean(test.input)
			if got != test.want {
				t.Errorf("Clean(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestRead(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  *Assembler
	}{
		{
			name: "empty",
			// no input
			want: &Assembler{
				state:       stateVariablesProcessed,
				lines:       nil,
				symbolTable: map[string]uint{},
			},
		},
		{
			name: "standard usage",
			input: `
// line comment
D = M + 1; JMP // inline comment
(LOOP)
@10
M=1
@LOOP
0;JMP
(END)
	@END
	0;JMP
`,
			want: &Assembler{
				state: stateVariablesProcessed,
				lines: []string{
					"D=M+1;JMP",
					"@10",
					"M=1",
					"@2",
					"0;JMP",
					"@7",
					"0;JMP",
				},
				symbolTable: map[string]uint{
					"LOOP": 2,
					"END":  7,
				},
			},
		},
		{
			name: "with variables",
			input: `
@i
D=A
`,
			want: &Assembler{
				state: stateVariablesProcessed,
				lines: []string{
					"@16",
					"D=A",
				},
				symbolTable: map[string]uint{
					"i": 16,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := New()
			r := strings.NewReader(test.input)
			if err := got.Read(r); err != nil {
				t.Fatalf("Read failed: %v", err)
			}

			test.want.addDefaultSymbols()
			allowAllUnexported := cmp.Exporter(func(reflect.Type) bool { return true })
			if diff := cmp.Diff(got, test.want, allowAllUnexported); diff != "" {
				t.Errorf("New *Assembler differed from expected (-got,+want):\n%s", diff)
			}
		})
	}
}
