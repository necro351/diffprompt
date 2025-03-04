package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"github.com/necro351/diffprompt/chat/chatgpt"
	"github.com/pkg/errors"
)

func main() {
	apiKey := flag.String("api-key", "", "API key for authentication")
	message := flag.String("message", "", "Message to complete")
	flag.Parse()

	if *message == "" {
		// Read from stdin
		messageBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}

		*message = string(messageBytes)
	}

	if *apiKey == "" {
		// Read API key from config file in home directory
		cfgPath := path.Join(os.Getenv("HOME"), ".diffprompt")

		apiKeyBytes, err := os.ReadFile(cfgPath)
		if err != nil {
			log.Fatal(err)
		}

		// Strip off any trailing white space from the API key
		*apiKey = strings.TrimSpace(string(apiKeyBytes))
	}

	if *apiKey == "" {
		log.Fatal("API key is required")
	}

	if *message == "" {
		log.Fatal("Message is required")
	}

	prompt, input := parseMessage(*message)

	if prompt == "" {
		// If we are not using a prompt, then search for and apply commands
		fmt.Println(applyCommands(input))

		return
	}

	*message = prompt + "\n\n" + input

	// Remove boilerplate from the message
	*message = fmt.Sprintf(boilerRemover, *message)

	completer := chatgpt.Completer{APIKey: *apiKey}
	result, err := completer.Complete(*message)
	if err != nil {
		log.Fatal(err)
	}

	result = result + "\n"

	sideBySide, err := sideBySideDiff(input, result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(sideBySide)
}

const boilerRemover = `%s

Only write the code, without comments or explanations. Do not use markdown. Preserve indentation of the below input.`

// parseMessage splits the input message into a prompt and input string.
// It looks for a separator line which is a line which contains only 'vvv' and whitespace.
func parseMessage(message string) (prompt, input string) {
	lines := strings.Split(message, "\n")

	separator := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "vvv" {
			separator = i
			break
		}
	}

	if separator == -1 {
		return "", message
	}

	prompt = strings.Join(lines[:separator], "\n")
	input = strings.Join(lines[separator+1:], "\n")

	return prompt, input
}

// applyCommands goes through the input line by line, and groups lines into
// command and diff blocks. For each block, it then applies the command, to
// the diff block. The concatenated result of applying all commands to their
// respective blocks is returned.
//
// A command is a line with nothing but whitespace and a single command word.
// Valid commands are 'apply' and 'reject'. Apply will remove all lines that
// start with the '-' character, and will keep all lines that start with the '+',
// or ' ' characters, but will remove the '+' or ' ' character. Reject will
// do the opposite.
//
// Here is an example:
//
//	apply
//	+line1
//	-line2
//	 line3
//	reject
//	+line4
//	-line5
//	 line6
//
// ...would result in:
//
//	line1
//	line3
//	line5
//	line6
func applyCommands(input string) string {
	command := ""
	commandLines := []string{}
	lines := strings.Split(input, "\n")
	outLines := []string{}

	for len(lines) > 0 {
		command, commandLines, lines = nextCommandBlock(lines)
		outLines = append(outLines, applyCommand(command, commandLines)...)
	}

	if len(outLines) == 1 {
		outLines = append(outLines, "")
	}

	return strings.Join(outLines, "\n")
}

func nextCommandBlock(lines []string) (string, []string, []string) {
	if len(lines) == 0 {
		return "", nil, nil
	}

	command := ""
	firstLine := strings.TrimSpace(lines[0])

	if firstLine == "apply" || firstLine == "reject" {
		command = firstLine
		lines = lines[1:]
	}

	// search for next command
	for i, line := range lines {
		line = strings.TrimSpace(line)

		if line == "apply" || line == "reject" {
			return command, lines[:i], lines[i:]
		}
	}

	return command, lines[:], nil
}

func applyCommand(command string, lines []string) []string {
	outLines := make([]string, 0)

	for _, line := range lines {
		if command == "apply" {
			if strings.HasPrefix(line, "-") {
				continue
			} else if strings.HasPrefix(line, "+") {
				line = strings.TrimPrefix(line, "+")
			} else if strings.HasPrefix(line, " ") {
				line = strings.TrimPrefix(line, " ")
			}
		} else if command == "reject" {
			if strings.HasPrefix(line, "+") {
				continue
			} else if strings.HasPrefix(line, "-") {
				line = strings.TrimPrefix(line, "-")
			} else if strings.HasPrefix(line, " ") {
				line = strings.TrimPrefix(line, " ")
			}
		}

		outLines = append(outLines, line)
	}

	return outLines
}

// sideBySideDiff returns a side-by-side diff of the input and result strings by
// writing both strings to temporary files, then running `diff -y` on them.
func sideBySideDiff(input, result string) (string, error) {
	inputFile, err := os.CreateTemp("", "input-")
	if err != nil {
		return "", errors.Wrap(err, "creating input file failed")
	}
	defer os.Remove(inputFile.Name())

	resultFile, err := os.CreateTemp("", "result-")
	if err != nil {
		return "", errors.Wrap(err, "creating result file failed")
	}
	defer os.Remove(resultFile.Name())

	if _, err := inputFile.WriteString(input); err != nil {
		return "", errors.Wrap(err, "writing input file failed")
	}
	if _, err := resultFile.WriteString(result); err != nil {
		return "", errors.Wrap(err, "writing result file failed")
	}

	if err := inputFile.Close(); err != nil {
		return "", errors.Wrap(err, "closing input file failed")
	}
	if err := resultFile.Close(); err != nil {
		return "", errors.Wrap(err, "closing result file failed")
	}

	// diff --no-prefix -U1000
	cmd := exec.Command("diff", "-U10000000", inputFile.Name(), resultFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws, ok := exitError.Sys().(syscall.WaitStatus)
			if !ok {
				return "", errors.New("failed to get exit status")
			}

			exitCode := ws.ExitStatus()
			if exitCode > 1 {
				return "", errors.New("diff command failed")
			}
		} else {
			return "", errors.New("diff call failed")
		}
	}

	return strip3(string(output)), nil
}

// strip3 removes the first three lines from the input string. If the first line
// starts with '---', the second starts with '+++', and the third starts with '@@'
func strip3(input string) string {
	lines := strings.Split(input, "\n")

	if len(lines) < 3 {
		return input
	}

	if strings.HasPrefix(lines[0], "---") && strings.HasPrefix(lines[1], "+++") && strings.HasPrefix(lines[2], "@@") {
		return strings.Join(lines[3:], "\n")
	}

	return input
}
