package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/thecasualcoder/kube-template/pkg/kubernetes"
	"github.com/thecasualcoder/kube-template/pkg/manager"
	"io"
	"os"
	"path/filepath"
	"strings"

	"text/template"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const (
	kubeConfigFlag = "kubeconfig"
	templateFlag   = "template"
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

		templateFlag, err := cmd.Flags().GetString(templateFlag)
		if err != nil {
			return fmt.Errorf("error parsing template flag: %w", err)
		}

		kubeconfig, _ := cmd.Flags().GetString(kubeConfigFlag)

		fs := afero.NewOsFs()
		return run(fs, templateFlag, kubeconfig)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.Flags().StringP(templateFlag, "t", "", "template to render. Should be of the format \"/path/to/template.tmpl:/path/to/rendered.conf\". \"-\" in target means STDOUT")

	kubeconfig := os.Getenv("KUBECONFIG")

	if kubeconfig == "" {
		home, err := homedir.Dir()
		if err == nil {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	rootCmd.Flags().String(kubeConfigFlag, kubeconfig, "(optional) absolute path to the kubeconfig file")

	err := rootCmd.MarkFlagRequired(templateFlag)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseTemplateFlag(flagValue string) (string, string, error) {
	templateValue := strings.Split(flagValue, ":")
	if len(templateValue) != 2 {
		return "", "", fmt.Errorf("template flag format is wrong")
	}
	return templateValue[0], templateValue[1], nil
}

func run(fs afero.Fs, templateFlag string, kubeconfig string) error {
	sourceTemplate, targetFile, err := parseTemplateFlag(templateFlag)
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
	clientset, err := kubernetes.NewClient(kubeconfig)
	if err != nil {
		return fmt.Errorf("error creating kube-client: %w", err)
	}

	m := manager.New(clientset)

	err = renderTemplate(m, sourceTemplateContents, target)
	if err != nil {
		return err
	}

	return nil
}

func renderTemplate(m manager.Manager, source string, target io.Writer) error {
	tmpl := template.New("").Funcs(template.FuncMap{
		"endpoints": m.Endpoints,
	})

	tmpl, err := tmpl.Parse(source)
	if err != nil {
		return fmt.Errorf("source template is not a valid template file: %w", err)
	}

	err = tmpl.Execute(target, nil)
	if err != nil {
		return fmt.Errorf("error rendering template: %w", err)
	}
	return nil
}
