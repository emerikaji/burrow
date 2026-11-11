package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/emerikaji/burrow/commit"
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

		if err = repository.Add(pwd, os.Args[2:]); err != nil {
			log.Fatalf("Failed to add files: %v", err)
		}
	case "commit":
		if len(os.Args) < 4 {
			log.Fatalf("Not enough arguments specified.")
		}

		if os.Args[2] != "-m" && os.Args[2] != "--message" {
			log.Fatalf("Unknown switch %s.", os.Args[2])
		}

		msg := strings.Trim(os.Args[3], "\"")

		pwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current working directory: %v", err)
		}

		when := time.Now()

		author := commit.Signature{
			Name:  os.Getenv("GIT_AUTHOR_NAME"),
			Email: os.Getenv("GIT_AUTHOR_EMAIL"),
			When:  when,
		}
		committer := commit.Signature{
			Name:  os.Getenv("GIT_COMMITTER_NAME"),
			Email: os.Getenv("GIT_COMMITTER_EMAIL"),
			When:  when,
		}

		commitHash, branch, isRoot, err := repository.Commit(pwd, author, committer, msg)
		if err != nil {
			log.Fatalf("Failed to commit files: %v", err)
		}

		shortHash := commitHash[:7]

		rootStatus := ""
		if isRoot {
			rootStatus = " (root-commit)"
		}

		// Grab just the first line of the message if it's multi-line
		msgLine, _, _ := strings.Cut(msg, "\n")

		fmt.Printf("[%s%s %s] %s\n", branch, rootStatus, shortHash, msgLine)
	default:
		fmt.Printf("unknown command '%s'. See '%s --help'.", os.Args[1], executable)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
