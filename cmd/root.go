package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/thecasualcoder/kube-template/pkg/kubernetes"
	"github.com/thecasualcoder/kube-template/pkg/manager"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const (
	kubeConfigFlag = "kubeconfig"
	templateFlag   = "template"
	execFlag       = "exec"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "kube-template",
	Short:         "Watch kubernetes resources, render templates and run applications",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("kube-tempalte does not accept args")
		}

		templateFlag, _ := cmd.Flags().GetString(templateFlag)
		kubeconfig, _ := cmd.Flags().GetString(kubeConfigFlag)
		execFlagValue, _ := cmd.Flags().GetString(execFlag)

		fs := afero.NewOsFs()

		templateArg, err := newTemplateArg(fs, templateFlag)
		if err != nil {
			_ = cmd.Help()
			return err
		}
		defer func() {
			_ = templateArg.target.Close()
			_ = fs.Remove(templateArg.target.Name())
		}()

		execCommand, err := newExecCommand(execFlagValue)
		if err != nil {
			_ = cmd.Help()
			return err
		}

		return run(templateArg, kubeconfig, execCommand)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.Flags().StringP(templateFlag, "t", "", "template to render. Should be of the format \"/path/to/template.tmpl:/path/to/rendered.conf\". \"-\" in target means STDOUT")
	rootCmd.Flags().StringP(execFlag, "e", "", "process to run after rendering")

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

func run(
	templateArg templateArg,
	kubeconfig string,
	execCommand execCommand,
) error {
	client, err := kubernetes.NewClient(kubeconfig)
	if err != nil {
		return fmt.Errorf("error creating kube-client: %w", err)
	}

	m := manager.New(client)

	eventChan := m.EventChan()
	errChan := make(chan error)

	go func() {
		var previousCmd *exec.Cmd
		for range eventChan {
			err := resetFileContent(templateArg.target)
			if err != nil {
				errChan <- err
				return
			}
			if err = renderTemplate(m, templateArg.source, templateArg.target); err != nil {
				errChan <- err
				return
			}

			command := exec.Command(execCommand.command, execCommand.args...)
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr

			if previousCmd != nil {
				if err := killCommandIfRunning(previousCmd); err != nil {
					errChan <- err
					return
				}
			}

			if err := execute(command); err != nil {
				errChan <- err
				return
			}

			previousCmd = command
		}
	}()

	return <-errChan
}

func resetFileContent(file afero.File) error {
	err := file.Truncate(0)
	if err != nil {
		return fmt.Errorf("error when truncating file %s: %w", file.Name(), err)
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("error when seeking to start of file %s: %w", file.Name(), err)
	}
	return nil
}

func killCommandIfRunning(command *exec.Cmd) error {
	if command.ProcessState != nil && !command.ProcessState.Exited() {
		err := command.Process.Signal(os.Interrupt)
		if err != nil {
			return err
		}

		const defaultInterruptWaitTimeout = 5
		timeOutCh := time.After(defaultInterruptWaitTimeout * time.Second)
		checkTimeOutCh := time.After(500 * time.Millisecond)
		for {
			select {
			case <-timeOutCh:
				return command.Process.Kill()
			case <-checkTimeOutCh:
				if command.ProcessState == nil || command.ProcessState.Exited() {
					return nil
				}
				checkTimeOutCh = time.After(500 * time.Millisecond)
			}
		}
	}

	return nil
}

func execute(command *exec.Cmd) error {
	if err := command.Start(); err != nil {
		return fmt.Errorf("error starting process %s: %v", command.Path, err)
	}
	go func() {
		if err := command.Wait(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}()
	return nil
}

func renderTemplate(m manager.Manager, source string, target io.Writer) error {
	tmpl := template.New("").Funcs(template.FuncMap{
		"endpoints": m.Endpoints,
		"pods":      m.PodsWithLabels,
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
