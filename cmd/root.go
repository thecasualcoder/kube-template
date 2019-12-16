package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"text/template"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "kube-template",
	Short:         "Watch kubernetes resources, render templates and run applications",
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("kube-tempalte does not accept args")
		}

		templateFlag, err := cmd.Flags().GetString("template")
		if err != nil {
			return fmt.Errorf("error parsing template flag: %w", err)
		}

		fs := afero.NewOsFs()
		return run(fs, templateFlag)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.
		Flags().
		StringP("template", "t", "", "template to render. Should be of the format \"/path/to/template.tmpl:/path/to/rendered.conf\". \"-\" in target means STDOUT")

	err := rootCmd.MarkFlagRequired("template")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseTempalateFlag(flag string) (string, string, error) {
	templateValue := strings.Split(flag, ":")
	if len(templateValue) != 2 {
		return "", "", fmt.Errorf("template flag format is wrong")
	}
	return templateValue[0], templateValue[1], nil
}

func run(fs afero.Fs, templateFlag string) error {
	sourceTemplate, targetFile, err := parseTempalateFlag(templateFlag)
	if err != nil {
		return err
	}

	var (
		sourceTemplateContents string
		target                 io.Writer
	)
	{
		if exists, err := afero.Exists(fs, sourceTemplate); err != nil {
			return err
		} else if !exists {
			return fmt.Errorf("source template \"%s\" does not exist", sourceTemplate)
		}

		sourceTemplateContentsBytes, err := afero.ReadFile(fs, sourceTemplate)
		if err != nil {
			return fmt.Errorf("error reading source template %s: %w", sourceTemplate, err)
		}

		sourceTemplateContents = string(sourceTemplateContentsBytes)

		if exists, err := afero.Exists(fs, targetFile); err != nil {
			return err
		} else if exists {
			return fmt.Errorf("target file  \"%s\" already exists", targetFile)
		}

		if targetFile == "-" {
			target = os.Stdout
		} else {
			target, err = fs.Open(targetFile)
			if err != nil {
				return fmt.Errorf("error opening file handle to target: %w", err)
			}
		}
	}

	tmpl := template.
		New("").
		Funcs(template.FuncMap{
			"itemsWith": itemsWith,
		})

	tmpl, err = tmpl.Parse(sourceTemplateContents)
	if err != nil {
		return fmt.Errorf("source template is not a valid template file: %w", err)
	}

	err = tmpl.Execute(target, nil)
	if err != nil {
		return fmt.Errorf("error rendering template: %w", err)
	}

	return nil
}

func itemsWith(suffix string) []string {
	output := make([]string, 0, 5)
	for i := 0; i < 5; i++ {
		output = append(output, fmt.Sprintf("item-%d-%s", i, suffix))
	}
	return output
}
