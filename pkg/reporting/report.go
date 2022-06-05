package reporting

import (
	"fmt"
	"strings"

	"github.com/seizadi/app-claim/pkg/graphdb"
)

func Report(graphOptions string, stage string, env string, apps []string) error {
	// Initialize Driver and Services
	driver, err := graphdb.NewDriver(graphOptions)
	if err != nil {
		return err
	}

	appSvc := graphdb.NewAppService(driver)

	// Print the report in comma comma-separated values for CSV format
	for _, app := range apps {
		// For app in the list query database and return result
		records, err := appSvc.QueryAppResources(stage, env, app)

		if err != nil {
			return err
		}

		for _, r := range records {
			fmt.Println(strings.Join(r, ","))
		}
	}
	return nil
}
