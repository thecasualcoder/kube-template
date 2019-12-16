package cmd

import (
	"fmt"
	"github.com/spf13/afero"
	"os"
	"strings"
)

type templateArg struct {
	source string
	target afero.File
}

func newTemplateArg(fs afero.Fs, templateFlagValue string) (templateArg, error) {
	templateValue := strings.Split(templateFlagValue, ":")
	if len(templateValue) != 2 {
		return templateArg{}, fmt.Errorf("template flag format is wrong")
	}
	sourceFilePath := templateValue[0]
	targetFilePath := templateValue[1]

	sourceTemplateContents, err := getSourceContents(fs, sourceFilePath)
	if err != nil {
		return templateArg{}, err
	}

	target, err := getTargetFile(fs, targetFilePath)
	if err != nil {
		return templateArg{}, err
	}

	return templateArg{
		source: sourceTemplateContents,
		target: target,
	}, nil
}

func getTargetFile(fs afero.Fs, targetFilePath string) (afero.File, error) {
	if targetFilePath == "-" {
		return os.Stdout, nil
	}

	if exists, err := afero.Exists(fs, targetFilePath); err != nil {
		return nil, err
	} else if exists {
		return nil, fmt.Errorf("target file  \"%s\" already exists", targetFilePath)
	}

	targetFile, err := fs.Create(targetFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file handle to target: %w", err)
	}

	return targetFile, nil
}

func getSourceContents(fs afero.Fs, sourceFilePath string) (string, error) {
	if exists, err := afero.Exists(fs, sourceFilePath); err != nil {
		return "", err
	} else if !exists {
		return "", fmt.Errorf("source template \"%s\" does not exist", sourceFilePath)
	}

	sourceTemplateContentsBytes, err := afero.ReadFile(fs, sourceFilePath)
	if err != nil {
		return "", fmt.Errorf("error reading source template %s: %w", sourceFilePath, err)
	}

	return string(sourceTemplateContentsBytes), nil
}

type execCommand struct {
	command string
	args    []string
}

func newExecCommand(execFlag string) (execCommand, error) {
	if execFlag == "" {
		return execCommand{}, fmt.Errorf("execCommand flag cannot be empty")
	}
	split := strings.Split(execFlag, " ")
	return execCommand{
		command: split[0],
		args:    split[1:],
	}, nil
}
