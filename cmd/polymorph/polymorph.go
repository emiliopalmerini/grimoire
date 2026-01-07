package polymorph

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/emiliopalmerini/grimorio/internal/cantrip/polymorph"
	"github.com/emiliopalmerini/grimorio/internal/metrics"
	"github.com/spf13/cobra"
)

var (
	fromFormat string
	toFormat   string
	outputFile string
)

var Cmd = &cobra.Command{
	Use:   "polymorph [input-file]",
	Short: "[Cantrip] Transform data between formats",
	Long: `Polymorph transforms data between different formats.

Supported formats: json, yaml, toml, xml, csv, markdown (md), html

The input format is auto-detected from file extension, or specify with --from.
Arrays of objects render as tables in markdown/html, single objects as key-value pairs.

Examples:
  grimorio polymorph data.json --to yaml
  grimorio polymorph config.yaml --to toml
  grimorio polymorph config.xml --to json
  grimorio polymorph users.json --to xml
  grimorio polymorph data.json --to csv -o output.csv
  cat data.json | grimorio polymorph --from json --to yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPolymorph,
}

func init() {
	Cmd.Flags().StringVarP(&fromFormat, "from", "f", "", "Input format (auto-detected from extension if not specified)")
	Cmd.Flags().StringVarP(&toFormat, "to", "t", "", "Output format (required)")
	Cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	Cmd.MarkFlagRequired("to")
}

func runPolymorph(cmd *cobra.Command, args []string) error {
	flags, _ := json.Marshal(map[string]any{"from": fromFormat, "to": toFormat})
	return metrics.Track("polymorph", metrics.Cantrip, string(flags), func() error {
		var input []byte
		var err error
		var inputPath string

		if len(args) == 1 {
			inputPath = args[0]
			input, err = os.ReadFile(inputPath)
			if err != nil {
				return fmt.Errorf("failed to read input file: %w", err)
			}
		} else {
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return fmt.Errorf("no input file provided and stdin is empty")
			}
			input, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
		}

		srcFormat := fromFormat
		if srcFormat == "" {
			if inputPath != "" {
				srcFormat = polymorph.DetectFormat(inputPath)
			}
			if srcFormat == "" {
				return fmt.Errorf("cannot detect input format, use --from to specify")
			}
		}

		result, err := polymorph.Convert(input, srcFormat, toFormat)
		if err != nil {
			return fmt.Errorf("polymorph failed: %w", err)
		}

		if outputFile != "" {
			if err := os.WriteFile(outputFile, result, 0644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}
			fmt.Printf("Polymorphed %s → %s: %s\n", srcFormat, toFormat, outputFile)
		} else {
			fmt.Print(string(result))
		}

		return nil
	})
}
