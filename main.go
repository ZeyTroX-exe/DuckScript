package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	running = map[string]bool{}

	lines     []string
	jmpPoints = map[string]int{}

	variables = map[string]Variable{}
	strType   = regexp.MustCompile(`'[^']*'|\S+`)
)

var tokens = map[string]string{
	">":  "GREAT",
	"<":  "LESS",
	">=": "GREATE",
	"<=": "LESSE",
	"!=": "NOT",
	"==": "EQUAL",
	":":  "THEN",
	"=":  "VAL",

	"+": "ADD",
	"-": "SUB",
	"*": "MUL",
	"/": "DIV",

	"case":   "COND",
	"set":    "VAR",
	"goto":   "JMP",
	"label":  "DEFINE",
	"end":    "END",
	"exit":   "BREAK",
	"print":  "OUT",
	"input":  "IN",
	"invoke": "EXEC",
	"start":  "MAIN",
	"sleep":  "TIMEOUT",
	"thread": "THREAD",
}

type Variable struct {
	Type  string
	Value string
}

func resolveVar(content string) Variable {
	var Type = "null"
	if strings.HasPrefix(content, "'") && strings.HasSuffix(content, "'") {
		content = strings.TrimPrefix(content, "'")
		content = strings.TrimSuffix(content, "'")
		content = strings.ReplaceAll(content, "\\n", "\n")

		Type = "string"
	} else if _, err := strconv.Atoi(content); err == nil {
		Type = "integer"
	} else if variable, exists := variables[content]; exists && variable.Value != content {
		Type = resolveVar(variable.Value).Type
		content = resolveVar(variable.Value).Value
	}
	newVar := Variable{Type: Type, Value: content}

	return newVar
}

func calc(expression []string) int {
	stack := []int{}
	operations := []string{}

	for _, token := range expression {
		if v, err := strconv.Atoi(resolveVar(token).Value); err == nil {
			stack = append(stack, v)
		} else {
			operations = append(operations, token)
		}
	}

	for {
		if slices.Contains(operations, "MUL") || slices.Contains(operations, "DIV") {
			for i, operator := range operations {
				if operator == "MUL" || operator == "DIV" {
					switch operator {
					case "MUL":
						stack[i] *= stack[i+1]

					case "DIV":
						stack[i] /= stack[i+1]
					}

					stack = append(stack[:i+1], stack[i+2:]...)
					operations = append(operations[:i], operations[i+1:]...)
				}
			}
		} else {
			break
		}
	}

	for {
		if slices.Contains(operations, "ADD") || slices.Contains(operations, "SUB") {
			for i, operator := range operations {
				if operator == "ADD" || operator == "SUB" {
					switch operator {
					case "ADD":
						stack[i] += stack[i+1]

					case "SUB":
						stack[i] -= stack[i+1]
					}

					stack = append(stack[:i+1], stack[i+2:]...)
					operations = append(operations[:i], operations[i+1:]...)
				}
			}
		} else {
			break
		}
	}

	if len(stack) != 1 {
		panic(fmt.Sprintf("Invalid expression: >>>%v<<<", expression))
	}

	return stack[0]
}

func cond(expression []string) bool {
	stack := []Variable{}
	operations := []string{}

	for _, statement := range expression {
		variable := resolveVar(statement)
		if variable.Type == "integer" || variable.Type == "string" {
			stack = append(stack, variable)
		} else {
			operations = append(operations, statement)
		}
	}

	for {
		if slices.Contains(operations, "GREAT") ||
			slices.Contains(operations, "LESS") ||
			slices.Contains(operations, "GREATE") ||
			slices.Contains(operations, "LESSE") ||
			slices.Contains(operations, "NOT") ||
			slices.Contains(operations, "EQUAL") {
			for i, operator := range operations {
				if stack[i].Type == "integer" && stack[i+1].Type == "integer" {
					switch operator {
					case "GREAT":
						return stack[i].Value > stack[i+1].Value

					case "LESS":
						return stack[i].Value < stack[i+1].Value

					case "GREATE":
						return stack[i].Value >= stack[i+1].Value

					case "LESSE":
						return stack[i].Value <= stack[i+1].Value
					}
				}

				switch operator {
				case "NOT":
					return stack[i].Value != stack[i+1].Value

				case "EQUAL":
					return stack[i].Value == stack[i+1].Value
				}

				stack = append(stack[:i+1], stack[i+2:]...)
				operations = append(operations[:i], operations[i+1:]...)
			}
		} else {
			break
		}
	}

	return false
}

func Execute(Instructions []string, line int) {
	if len(Instructions) > 0 {
		switch Instructions[0] {
		case "OUT":
			if variable, exists := variables[Instructions[1]]; exists {
				variable = resolveVar(variable.Value)
				if variable.Type == "integer" {
					exepression := Instructions[1:]
					result := calc(exepression)
					fmt.Print(result)
				} else {
					fmt.Print(variable.Value)
				}

			} else {
				variable = resolveVar(Instructions[1])
				if variable.Type == "integer" {
					exepression := Instructions[1:]
					result := calc(exepression)
					fmt.Print(result)
				} else {
					fmt.Print(variable.Value)
				}
			}

		case "VAR":
			if Instructions[2] == "VAL" {
				variable := resolveVar(Instructions[3])
				if variable.Type == "integer" {
					variable.Value = strconv.Itoa(calc(Instructions[3:]))
				}
				variables[Instructions[1]] = variable
			} else {
				panic(fmt.Sprintf("Invalid operator for setting variables: >>>%v<<<", Instructions[2]))
			}

		case "COND":
			for i, split := range Instructions[1:] {
				if split == "THEN" {
					if cond(Instructions[1 : i+1]) {
						Execute(Instructions[i+2:], line)
					}
				}
			}

		case "TIMEOUT":
			if v, err := strconv.Atoi(Instructions[1]); err == nil {
				time.Sleep(time.Millisecond * time.Duration(v))
			}

		case "BREAK":
			running["EXECUTING"] = false

		case "THREAD":
			go Execute(Instructions[1:], line)

		case "JMP":
			if targetLine, exists := jmpPoints[Instructions[1]]; exists {
				for ln := range lines {
					if ln == targetLine {
						running[Instructions[1]] = true
						counter := targetLine
						for running[Instructions[1]] {
							counter++
							lexed := LexLine(lines[counter])
							if lexed[0] == "MAIN" || lexed[0] == "DEFINE" {
								running[Instructions[1]] = false
							}
							Execute(lexed, ln)
						}
					}
				}
			}

		case "END":
			running[Instructions[1]] = false

		case "IN":
			fmt.Print(resolveVar(Instructions[1]).Value)
			input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			input = strings.TrimSpace(input)
			if len(Instructions) > 2 {
				if Instructions[2] == "VAL" {
					Execute(LexLine(fmt.Sprintf("set %v = '%v';", Instructions[3], resolveVar(input).Value)), line)
				}
			}

		case "EXEC":
			args := strings.Split(resolveVar(Instructions[1]).Value, " ")
			cmd := exec.Command(args[0], args[1:]...)
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

			if len(Instructions) > 2 {
				if Instructions[2] == "VAL" {
					out, _ := cmd.Output()
					Execute(LexLine(fmt.Sprintf("set %v = '%v';", Instructions[3], resolveVar(strings.TrimSpace(string(out))).Value)), line)
				} else {
					panic(fmt.Sprintf("Invalid operator for setting variables: >>>%v<<<", Instructions[2]))
				}
			} else {
				cmd.Start()
			}
		}
	}
}

func LexLine(line string) []string {
	parsedInstructions := []string{}

	line = strings.TrimSpace(line)

	for _, instruction := range strType.FindAllString(line, -1) {
		if value, exists := tokens[instruction]; exists {
			parsedInstructions = append(parsedInstructions, value)
		} else {
			parsedInstructions = append(parsedInstructions, instruction)
		}
	}

	return parsedInstructions
}

func main() {
	path, _ := filepath.Abs(os.Args[1])
	if filepath.Ext(path) != ".dk" {
		panic("Only duck files '.dk' can be interpreted!")
	}
	code, _ := os.ReadFile(path)
	lines = strings.Split(string(code), ";")

	for i, line := range lines {
		lexed := LexLine(line)
		if len(lexed) > 0 {
			if lexed[0] == "DEFINE" {
				jmpPoints[lexed[1]] = i
			}
		}
	}

	for i, line := range lines {
		lexed := LexLine(line)
		if len(lexed) > 0 {
			if running["EXECUTING"] {
				Execute(lexed, i)
			} else if lexed[0] == "MAIN" {
				running["EXECUTING"] = true
			}
		}
	}
}
