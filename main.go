package main

import (
	"fmt"
	"log"
	"os"

	"github.com/emerikaji/burrow/repository"
)

func main() {
	executable := os.Args[0]

	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Printf("usage: %s [-v | --version] [-h | --help] <command> [<args>]", executable)
		os.Exit(0)
	}

	if os.Args[1] == "-v" || os.Args[1] == "--version" {
		fmt.Println("TODO produce version")
		os.Exit(0)
	}

	switch os.Args[1] {
	case "init":
		target := "."

		if len(os.Args) > 2 {
			target = os.Args[2]
		}

		if reinit, err := repository.Init(target); err != nil {
			log.Fatalf("Failed to initialize Git repository: %v", err)
		} else {
			if reinit {
				fmt.Printf("Reinitialized existing Git repository in %s/.git/", target)
			} else {
				fmt.Printf("Initialized empty Git repository in %s/.git/", target)
			}
		}
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Nothing specified, nothing added.")
			os.Exit(0)
		}

		pwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current working directory: %v", err)
		}

		if err := repository.Add(pwd, os.Args[2:]); err != nil {
			log.Fatalf("Failed to add files: %v", err)
		}
	case "commit":
	default:
		fmt.Printf("unknown command '%s'. See '%s --help'.", os.Args[1], executable)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
