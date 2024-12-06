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

const (
	dockerImageFlag = "docker-image"
	commandFlag     = "command"
)

func main() {
	cmd := &cli.Command{
		Action: func(ctx context.Context, command *cli.Command) error {
			workingDir, err := os.Getwd()
			if err != nil {
				return errors.Wrap(err, "could not get current directory")
			}

			modPath, err := goModPath()
			if err != nil {
				return errors.Wrap(err, "could not get go mod path")
			}

			moduleName, err := golangModuleName(workingDir)
			if err != nil {
				return errors.WithStack(err)
			}

			cmd := prepareCommand(commandInfo{
				workingDirectory: workingDir,
				moduleName:       moduleName,
				goModPath:        modPath,
				dockerImage:      command.String(dockerImageFlag),
				command:          command.String(commandFlag),
			})

			fmt.Printf("command: %s\n", cmd)

			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     dockerImageFlag,
				Aliases:  []string{"i"},
				Usage:    "Docker image to use for running the command",
				Required: true,
			},
			&cli.StringFlag{
				Name:     commandFlag,
				Aliases:  []string{"c"},
				Usage:    "Command to run inside the container",
				Required: true,
			},
		},
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func golangModuleName(workingDir string) (string, error) {
	file, err := os.Open(path.Join(workingDir, "go.mod"))
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

func goModPath() (string, error) {
	goMod := os.Getenv("GOMODCACHE")
	if goMod == "" {
		goMod = path.Join(os.Getenv("HOME"), "go/pkg/mod")
	}

	if goMod == "" {
		return "", errors.New("could not find go mod path")
	}

	return goMod, nil
}

func prepareCommand(info commandInfo) string {
	moduleContainerPath := fmt.Sprintf("/go/src/%s", info.moduleName)

	return fmt.Sprintf("docker run --rm -t -i -v%s:%s -v%s:/go/pkg/mod -w%s %s %s",
		info.workingDirectory, moduleContainerPath, info.goModPath, moduleContainerPath, info.dockerImage, info.command)
}

type commandInfo struct {
	workingDirectory string
	moduleName       string
	goModPath        string
	dockerImage      string
	command          string
}
