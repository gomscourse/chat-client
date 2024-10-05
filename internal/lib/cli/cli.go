package cli

import (
	"bufio"
	"fmt"
	"golang.org/x/term"
	"os"
	"strings"
)

type Print interface {
	Info(msg string)
	Warning(msg string)
}

// GetUserInput gets input from user terminal with retrying if input is empty.
func GetUserInput(prompt string, prt Print) string {
	reader := bufio.NewReader(os.Stdin)
	for {
		prt.Info(prompt)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			prt.Warning("value can't be empty")
		} else {
			return input
		}
	}
}

// GetSensitiveUserInput gets input from user terminal with retrying if input is empty. The input is invisible for user.
func GetSensitiveUserInput(prompt string, prt Print) (string, error) {
	for {
		prt.Info(prompt)
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return "", fmt.Errorf("unable to get sensitive input: %w", err)
		}

		fmt.Println() // Print a newline because ReadPassword does not capture the enter key

		if len(bytePassword) == 0 {
			prt.Warning("value can't be empty")
		} else {
			return string(bytePassword), nil
		}
	}
}

// InfinityInput gets input from user terminal and passes it to the inputHandler func
func InfinityInput(inputHandler func(input string), stopWord string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == stopWord {
			fmt.Println("exiting input process")
			return
		}

		inputHandler(input)
	}
}
