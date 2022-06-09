package commands

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/seizadi/app-claim/pkg/reporting"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addSearch)
	addSearch.Flags().StringP("dir", "d", "", "search directory")
	addSearch.Flags().StringP("stage", "s", "", "search stage")
	addSearch.Flags().StringP("env", "e", "", "search environment")
	addSearch.Flags().StringP("app", "a", "", "search application")
	addSearch.Flags().StringP("graphdb", "g", "", "use graph database")
	addSearch.Flags().StringP("aws", "", "", "search for aws resources")
	addSearch.Flags().StringP("appsfile", "", "", "FIXME -- REMOVE")
	addSearch.Flags().BoolP("claims", "c", false, "search for claims")
}

// addSearch implements the search command
var addSearch = &cobra.Command{
	Use:   "search",
	Short: "search kubernetes manifest yaml",
	Long: `search kubernetes manifest yaml
It assumes that the directory supplied has the manifests
in it in YAML format.`,
	Run: func(cmd *cobra.Command, args []string) {

		dir, err := cmd.Flags().GetString("dir")
		if err != nil {
			fmt.Println(err)
			return
		} else if len(dir) == 0 {
			fmt.Println("--dir argument must be specified with path to manifests")
			return
		}

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

		// Check if we should save results to graphdb
		graphOptions, err := cmd.Flags().GetString("graphdb")
		if err != nil {
			fmt.Println(err)
			return
		}

		awsOptions, err := cmd.Flags().GetString("aws")
		if err != nil {
			fmt.Println(err)
			return
		}

		claims, err := cmd.Flags().GetBool("claims")
		if err != nil {
			fmt.Println(err)
			return
		}

		// Search
		// There are two types of search free form from list of args passed or AWS
		// For AWS search we expect a file to passed as argument that contains the
		// AWS S3 bucket names as search terms
		r := NewSearchRunner("", "")

		if claims { // Claim Search
			searchResults, err := r.SearchForClaims(dir, stage, env, app)
			if err != nil {
				fmt.Println(err)
				return
			}

			for k, v := range searchResults {
				fmt.Printf("%s %v\n", k, v)
			}

			if len(graphOptions) > 0 {
				err = reporting.Discover(graphOptions, searchResults, "CLAIM")
				if err != nil {
					log.Print(err)
					return
				}
			}
		} else if len(awsOptions) > 0 { // AWS Search
			// Read the file with AWS Buckets
			readFile, err := os.Open(awsOptions)

			if err != nil {
				fmt.Println(err)
				return
			}
			fileScanner := bufio.NewScanner(readFile)

			fileScanner.Split(bufio.ScanLines)

			tokens := []string{}
			for fileScanner.Scan() {
				values := strings.Split(fileScanner.Text(), " ")
				if len(values) == 3 {
					tokens = append(tokens, values[2])
				}
			}

			readFile.Close()

			searchResults, err := r.SearchForTokens(tokens, dir, stage, env, app)
			if err != nil {
				fmt.Println(err)
				return
			}

			for k, v := range searchResults {
				fmt.Printf("%s %v\n", k, v)
			}

			if len(graphOptions) > 0 {
				err = reporting.Discover(graphOptions, searchResults, "s3")
				if err != nil {
					log.Print(err)
					return
				}
			}

			// Now search by amazonaws.com and identify other resources like RDS, DynamoDB and ElasticSearch
			searchResults, err = r.SearchForTokens([]string{"amazonaws.com"}, dir, stage, env, app)
			if err != nil {
				fmt.Println(err)
				return
			}

			// Patterns:
			// RDS: *.rds.amazonaws.com:<port>
			// REDIS: redis://*.cache.amazonaws.com:<port>
			// KAFKA: *.kafka.<region>.amazonaws.com:<port>
			// ES (ElasticSearch): *.es.amazonaws.com:<port>
			runners := []*SearchRunner{
				NewSearchRunner(".*\\.rds\\.amazonaws.com", "RDS"),
				NewSearchRunner(".*\\.cache\\.amazonaws.com", "REDIS"),
				NewSearchRunner(".*\\.kafka\\..*\\.amazonaws.com", "KAFKA"),
				NewSearchRunner(".*\\.es\\.amazonaws.com", "ES"),
				NewSearchRunner(".*s3\\.amazonaws.com", "S3"),
				NewSearchRunner(".*s3-fips\\..*\\.amazonaws.com", "S3"),
			}

			for k, v := range searchResults {
				found := false
				for _, r := range runners {
					found, err = r.SearchForResource(k, v)
					if err != nil {
						log.Print(err)
						return
					}
					if found {
						fmt.Printf("[%s] %s %v\n", r.ResourceKind, k, v)
						break
					}
				}
			}

			if len(graphOptions) > 0 {
				for _, r := range runners {
					err = reporting.Discover(graphOptions, r.SearchResults, r.ResourceKind)
					if err != nil {
						log.Print(err)
						return
					}
				}
			}

		} else { // General Token Search
			searchResults, err := r.SearchForTokens(args, dir, stage, env, app)
			if err != nil {
				fmt.Println(err)
				return
			}

			for k, v := range searchResults {
				fmt.Printf("%s %v\n", k, v)
			}

			if len(graphOptions) > 0 {
				err = reporting.Discover(graphOptions, searchResults, "")
				if err != nil {
					log.Print(err)
					return
				}
			}
		}
	},
}

// DatabaseClaimReconciler reconciles a DatabaseClaim object
type SearchRunner struct {
	SearchResults map[string][]string
	SearchPattern string // regex
	ResourceKind  string
	SearchClaims  bool
}

func NewSearchRunner(pattern string, kind string) *SearchRunner {
	return &SearchRunner{
		SearchResults: map[string][]string{},
		SearchPattern: pattern,
		ResourceKind:  kind,
	}
}

func (r *SearchRunner) SearchForResource(k string, v []string) (bool, error) {
	found, err := regexp.MatchString(r.SearchPattern, k)
	if err != nil {
		log.Print(err)
		return false, err
	}

	if found {
		r.SearchResults[k] = v
	}

	return found, nil
}

func (r *SearchRunner) SearchForClaims(dir, stage, env, app string) (map[string][]string, error) {
	r.SearchClaims = true
	tokens := []string{"DatabaseClaim"}
	return r.SearchForTokens(tokens, dir, stage, env, app)
}

func (r *SearchRunner) SearchForTokens(tokens []string, dir, stage, env, app string) (map[string][]string, error) {
	stageFound := false

	stages, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	for _, s := range stages {
		if s.IsDir() && !IsHiddenFile(s.Name()) {
			if len(stage) > 0 {
				if stage == s.Name() {
					stageFound = true
					r.SearchEnv(dir, stage, env, app, tokens)
				}
			} else {
				r.SearchEnv(dir, s.Name(), env, app, tokens)
			}
		}
	}

	if len(stage) > 0 && !stageFound {
		return nil, errors.New(fmt.Sprintf("stage %s not found\n", stage))
	}

	return r.SearchResults, nil
}

func (r *SearchRunner) SearchEnv(dir string, stage string, env string, app string, args []string) {
	envFound := false
	path := dir + "/" + stage
	envs, err := ioutil.ReadDir(path)
	if err != nil {
		log.Print(err)
		return
	}

	for _, e := range envs {
		if e.IsDir() && !IsHiddenFile(e.Name()) {
			if len(env) > 0 {
				if env == e.Name() {
					envFound = true
					r.SearchManifest(dir, stage, env, app, args)
				}
			} else {
				r.SearchManifest(dir, stage, e.Name(), app, args)
			}
		}
	}

	if len(env) > 0 && !envFound {
		log.Printf("environment %s not found\n", env)
	}
}

func GetManifest(filename string) (*[]map[interface{}]interface{}, error) {
	var out []map[interface{}]interface{}
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Print(err)
		return &out, err
	}

	dec := yaml.NewDecoder(bytes.NewReader(source))

	for {
		m := make(map[interface{}]interface{})
		if dec.Decode(&m) != nil {
			break
		}
		out = append(out, m)
	}

	return &out, err
}

func (r *SearchRunner) SearchManifest(dir string, stage string, env string, app string, args []string) ([]string, error) {
	var searchOut []string
	apps, err := ioutil.ReadDir(dir + "/" + stage + "/" + env)
	if err != nil {
		log.Print(err)
		return searchOut, err
	}

	rootPath := dir + "/" + stage + "/" + env + "/"
	for _, a := range apps {
		if a.IsDir() && !IsHiddenFile(a.Name()) {
			out, err := GetManifest(rootPath + a.Name() + "/manifest.yaml")
			if err != nil {
				fmt.Printf("Got error %s\n", err.Error())
				return searchOut, err
			}

			for _, m := range *out {
				if r.SearchClaims {
					if k, ok := m["kind"].(string); ok {
						r.SearchMatch(k, stage, env, a.Name(), args)
					}
				} else {
					r.recurseSearch(m, stage, env, a.Name(), args)
				}
			}
		}
	}
	return searchOut, nil
}

func (r *SearchRunner) recurseSearch(m interface{}, stage string, env string, app string, args []string) {
	reflectM := reflect.ValueOf(m)

	switch reflectM.Kind() {
	case reflect.String:
		v := reflectM.String()
		r.SearchMatch(v, stage, env, app, args)

	case reflect.Slice:
		for i := 0; i < reflectM.Len(); i++ {
			r.recurseSearch(reflectM.Index(i).Interface(), stage, env, app, args)
		}

	case reflect.Map:
		for _, key := range reflectM.MapKeys() {
			strct := reflectM.MapIndex(key)
			value := reflect.ValueOf(strct.Interface())
			if value.Kind() == reflect.String {
				r.recurseSearch(fmt.Sprintf("%v:%s", key.Interface(), value.String()), stage, env, app, args)
			} else {
				// Search Key for match
				r.SearchMatch(fmt.Sprintf("%v", key.Interface()), stage, env, app, args)
				r.recurseSearch(strct.Interface(), stage, env, app, args)
			}
		}
	case reflect.Struct:
		fmt.Printf("****** We should not be getting structs with decoder %v\n", m)
		//for i := 0; i < reflectM.NumField(); i++ {
		//	if reflectM.Field(i).CanInterface() {
		//		r.recurseSearch(reflectM.Field(i).Interface(), stage, env, app, args)
		//	}
		//}
	}
}

func (r *SearchRunner) SearchMatch(s string, stage string, env string, app string, args []string) {
	for _, a := range args {
		if strings.Contains(strings.ToLower(s), strings.ToLower(a)) {
			if len(s) < 128 {
				// fmt.Printf("%s/%s/%s %s %s\n", stage, env, app, a, s)
				tag := fmt.Sprintf("%s/%s/%s", stage, env, app)
				r.SearchResults[s] = append(r.SearchResults[s], tag)
			}
		}
	}
}

func IsHiddenFile(filename string) bool {
	return filename[0:1] == "."
}
