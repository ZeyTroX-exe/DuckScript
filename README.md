# DuckScript Interpreter

DuckScript is a custom scripting language interpreted by this Go-based program. It reads `.dk` script files and executes commands based on a unique syntax.

## Features:

- **Basic Arithmetic** (`+`, `-`, `*`, `/`)
- **Conditional Execution** (`case`, `goto`)
- **Variable Assignment** (`set`)
- **Input & Output** (`print`, `input`)
- **Threading Support** (`thread`)
- **External Command Execution** (`invoke`)
- **Program Flow Control** (`label`, `goto`, `end`)
- **Built-in Sleep Function** (`sleep`)

## Installation:

1. Install Go: [Download Go](https://go.dev/dl/)
2. Clone the repository: `git clone https://github.com/ZeyTroX-exe/DuckScript.git`
3. Build the executable: `go build main.go -o quack.exe`

## Usage:
To run a DuckScript file: `.\quack.exe C:\path\to\script.dk`

## Commands & Syntax:

| Command  | Description                        | Example Usage           |
|----------|------------------------------------|-------------------------|
| `set`    | Assigns a value to a variable     | `set x = 'Hello';`       |
| `print`  | Outputs a value                   | `print x;`               |
| `input`  | Reads user input                  | `input name;`            |
| `goto`   | Jumps to a label                  | `goto start;`            |
| `label`  | Defines a label                   | `label start;`           |
| `thread` | Runs a command in a new thread    | `thread print 'Hi';`     |
| `sleep`  | Pauses execution for a duration   | `sleep 1000;` (1 sec)    |

## Example `test.dk`:

```
label count;
    set counter = counter + 1;
    print counter;
    print '\n';
    sleep 1000;
    case counter == num : end count;
    goto x;

start;
    input 'Enter a number: ' = num;
    set counter = 0;
    goto count;
    exit;
```

