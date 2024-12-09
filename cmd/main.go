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
)

func main() {
	cmd := &cli.Command{
		Name:  "docker-go-shell",
		Usage: "Runs a shell command in a Docker container inside of the current directory",
		Action: func(ctx context.Context, command *cli.Command) error {
			cmd, err := prepareCommand(commandInfo{
				dockerImage: command.String(dockerImageFlag),
				args:        command.Args().Slice(),
			})
			if err != nil {
				return errors.Wrap(err, "could not prepare command")
			}

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
		},
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func prepareCommand(info commandInfo) (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "could not get current directory")
	}

	moduleName, err := golangModuleName(workingDir)
	if err != nil {
		return "", errors.WithStack(err)
	}

	moduleContainerPath := fmt.Sprintf("/go/src/%s", moduleName)

	modPath, err := goModPath()
	if err != nil {
		return "", errors.Wrap(err, "could not get go mod path")
	}

	args := strings.Join(info.args, " ")

	return fmt.Sprintf("docker run --rm -t -i -v%s:%s -v%s:/go/pkg/mod -w%s %s %s",
		workingDir, moduleContainerPath, modPath, moduleContainerPath, info.dockerImage, args), nil
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

	var moduleName string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "module ") {
			moduleName = text[7:]
			break
		}
	}

	if err = scanner.Err(); err != nil {
		return "", errors.Wrap(err, "could not read go.mod file")
	}

	if moduleName == "" {
		return "", errors.New("could not determine module name")
	}

	return moduleName, nil
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

type commandInfo struct {
	dockerImage string
	args        []string
}
