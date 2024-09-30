package cmd

import (
	"fmt"
	"os"

	"github.com/jlewi/monogo/helpers"

	"github.com/jlewi/foyle/app/api"
	"github.com/jlewi/foyle/app/pkg/analyze"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewLLMsCmd returns a command to work with llms
func NewLLMsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "llms",
	}

	cmd.AddCommand(NewLLMsRenderCmd())

	return cmd
}

func NewLLMsRenderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "render",
	}

	cmd.AddCommand(NewLLMsRenderRequestCmd())
	cmd.AddCommand(NewLLMsRenderResponseCmd())

	return cmd
}

func NewLLMsRenderRequestCmd() *cobra.Command {
	var provider string
	var inputFile string
	var outputFile string
	cmd := &cobra.Command{
		Use: "request",
		Run: func(cmd *cobra.Command, args []string) {
			err := func() error {
				p := api.ModelProvider(provider)

				data, err := os.ReadFile(inputFile)
				if err != nil {
					return errors.Wrapf(err, "Failed to read file %s", inputFile)
				}

				htmlData, err := analyze.RenderRequestHTML(string(data), p)
				if err != nil {
					return err
				}

				if outputFile != "" {
					if err := os.WriteFile(outputFile, []byte(htmlData), 0644); err != nil {
						return errors.Wrapf(err, "Failed to write file %s", outputFile)
					}
				} else {
					fmt.Fprint(os.Stdout, htmlData)
				}

				return nil
			}()

			if err != nil {
				fmt.Printf("Error rendering request;\n %+v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVarP(&provider, "provider", "p", string(api.ModelProviderOpenAI), "The model provider for the request.")
	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "The file containing the JSON representation of the request.")
	cmd.Flags().StringVarP(&inputFile, "output", "o", "", "The file to write the output to. If blank output to stdout.")
	helpers.IgnoreError(cmd.MarkFlagRequired("input"))

	return cmd
}

func NewLLMsRenderResponseCmd() *cobra.Command {
	var provider string
	var inputFile string
	var outputFile string
	cmd := &cobra.Command{
		Use: "response",
		Run: func(cmd *cobra.Command, args []string) {
			err := func() error {
				p := api.ModelProvider(provider)

				data, err := os.ReadFile(inputFile)
				if err != nil {
					return errors.Wrapf(err, "Failed to read file %s", inputFile)
				}

				htmlData, err := analyze.RenderResponseHTML(string(data), p)
				if err != nil {
					return err
				}

				if outputFile != "" {
					if err := os.WriteFile(outputFile, []byte(htmlData), 0644); err != nil {
						return errors.Wrapf(err, "Failed to write file %s", outputFile)
					}
				} else {
					fmt.Fprint(os.Stdout, htmlData)
				}

				return nil
			}()

			if err != nil {
				fmt.Printf("Error rendering response;\n %+v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVarP(&provider, "provider", "p", string(api.ModelProviderOpenAI), "The model provider for the request.")
	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "The file containing the JSON representation of the request.")
	cmd.Flags().StringVarP(&inputFile, "output", "o", "", "The file to write the output to. If blank output to stdout.")
	helpers.IgnoreError(cmd.MarkFlagRequired("input"))

	return cmd
}
