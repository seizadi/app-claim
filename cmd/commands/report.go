package commands

import (
	"bufio"
	"fmt"
	"github.com/seizadi/app-claim/pkg/reporting"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// addSearch implements the search command
var addReport = &cobra.Command{
	Use:   "report",
	Short: "report based on query parameters",
	Long: `generate report based on query parameters supplied.
`,
	Run: func(cmd *cobra.Command, args []string) {

		// If stage is specified search only selected
		stage, err := cmd.Flags().GetString("stage")
		if err != nil {
			fmt.Println(err)
			return
		}

		// If environment is specified search only selected
		env, err := cmd.Flags().GetString("env")
		if err != nil {
			fmt.Println(err)
			return
		}
		// If application is specified search only selected
		app, err := cmd.Flags().GetString("app")
		if err != nil {
			fmt.Println(err)
			return
		}

		// The graphdb parameter for the database to use for queries
		graphOptions, err := cmd.Flags().GetString("graphdb")
		if err != nil {
			fmt.Println(err)
			return
		}

		// The apps file allow you to specify the list of applications from CSV file
		appsFile, err := cmd.Flags().GetString("appsfile")
		if err != nil {
			fmt.Println(err)
			return
		}

		// Indicate if CSV import file has header row for column types,
		// NOTE we assume first column is application name rather than read header
		ignoreHeader, err := cmd.Flags().GetBool("ignoreheader")
		if err != nil {
			fmt.Println(err)
			return
		}
		// Report
		// We get the list of applications either from command line args or the
		// CSV file and generate a query from the database and return the query
		// results as a report. The stage and env paramters narrow the list of
		// applications that are returned. You can report on a single application
		// or all the applications. The focus of this report is the resource usage
		// by the applications.
		apps := []string{}

		if len(appsFile) > 0 {
			// Read the file with list of applications
			readFile, err := os.Open(appsFile)

			if err != nil {
				fmt.Println(err)
				return
			}
			fileScanner := bufio.NewScanner(readFile)

			fileScanner.Split(bufio.ScanLines)

			for fileScanner.Scan() {
				// Skip first line if we have header and ignore it
				if ignoreHeader {
					ignoreHeader = false
				}
				values := strings.Split(fileScanner.Text(), ",")
				if len(values) == 1 {
					apps = append(apps, values[0])
				}
			}

			if len(apps) == 0 {
				fmt.Printf("No applications retrieved from file %s", appsFile)
				return
			}

			readFile.Close()
		}

		apps = append(apps, args...)

		if len(app) > 0 {
			apps = append(apps, app)
		}

		if len(apps) == 0 {
			fmt.Println("No applications specified either supply from command line or file")
			return
		}
		err = reporting.Report(graphOptions, stage, env, apps)
		if err != nil {
			fmt.Println(err)
		}
		return
	},
}

func init() {
	rootCmd.AddCommand(addReport)
	addReport.Flags().StringP("dir", "d", "", "todo remove directory")
	addReport.Flags().StringP("stage", "s", "", "search stage")
	addReport.Flags().StringP("env", "e", "", "search environment")
	addReport.Flags().StringP("app", "a", "", "search application")
	addReport.Flags().StringP("graphdb", "g", "", "use graph database")
	addReport.Flags().StringP("appsfile", "l", "", "list of files for report")
	addReport.Flags().BoolP("ignoreheader", "i", true, "ignore header line on import")
}
