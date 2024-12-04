package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "greet",
		Usage: "say a greeting",
		Action: func(ctx context.Context, command *cli.Command) error {
			moduleName, err := golangModuleName()
			if err != nil {
				return errors.WithStack(err)
			}

			fmt.Printf("module name: %s\n", moduleName)

			return nil
		},
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func golangModuleName() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "could not get current directory")
	}

	file, err := os.Open(path.Join(dir, "go.mod"))
	if err != nil {
		return "", errors.Wrap(err, "could not find or open go.mod file for reading")
	}
	defer func() {
		err2 := file.Close()
		if err2 != nil {
			log.Fatal(err2)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "module ") {
			return text[7:], nil
		}
	}

	if err = scanner.Err(); err != nil {
		return "", errors.Wrap(err, "could not read go.mod file")
	}

	return "", errors.New("could not find module name in go.mod file")
}
