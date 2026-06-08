// +build ignore

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	fmt.Println("Password Hash Generator for users.json")
	fmt.Println("======================================")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter password to hash (or 'quit' to exit): ")
		password, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		password = strings.TrimSpace(password)
		if password == "" {
			continue
		}
		if strings.ToLower(password) == "quit" {
			break
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Println("Error generating hash:", err)
			continue
		}

		fmt.Println()
		fmt.Println("Password hash (copy this to users.json):")
		fmt.Println(string(hash))
		fmt.Println()
	}

	fmt.Println("Goodbye!")
}
