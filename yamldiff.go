package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// loadYAML loads a YAML file and returns its content as a map
func loadYAML(filePath string) (map[interface{}]interface{}, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var content map[interface{}]interface{}
	err = yaml.Unmarshal(data, &content)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// compareMaps recursively compares two maps and calls printDifference when a difference is found.
// compareMaps recursively compares two maps and calls printDifference when a difference is found.
// It skips printing differences where a key is missing in one of the maps.
func compareMaps(map1, map2 map[interface{}]interface{}, path string, diffMap map[interface{}]interface{}, print bool) {
	for key := range map1 {
		val1 := map1[key]
		val2, ok := map2[key]
		if !ok {
			// Skip cases where the key is missing in the second map
			continue
		}

		switch val1Typed := val1.(type) {
		case map[interface{}]interface{}:
			if nestedMap2, ok := val2.(map[interface{}]interface{}); ok {
				newPath := path + "." + fmt.Sprint(key)
				subDiffMap := make(map[interface{}]interface{})
				compareMaps(val1Typed, nestedMap2, newPath, subDiffMap, print)
				if len(subDiffMap) > 0 {
					diffMap[key] = subDiffMap
				}
			} else {
				if print && !reflect.DeepEqual(val1, val2) {
					printDifference(path, key, val1, val2)
				}
				diffMap[key] = val1
			}
		default:
			if !reflect.DeepEqual(val1, val2) {
				if print {
					printDifference(path, key, val1, val2)
				}
				diffMap[key] = val1
			}
		}
	}

	// Also check if there are keys in map2 that are missing in map1
	for key := range map2 {
		if _, ok := map1[key]; !ok {
			// Skip cases where the key is missing in the first map
			continue
		}
	}
}

// printDifference prints differing values along with their key paths
func printDifference(path string, key interface{}, val1, val2 interface{}) {
	fullPath := path + "." + fmt.Sprint(key)

	// Format the output for better readability
	fmt.Printf("\nDifference at: %s\n", fullPath)
	fmt.Printf("  First file:  %v\n", val1)
	fmt.Printf("  Second file: %v\n", val2)
}

// printYAML prints the content as YAML to the console with an optional header
func printYAML(content map[interface{}]interface{}, diff bool) error {
	data, err := yaml.Marshal(content)
	if err != nil {
		return err
	}

	if diff {
		// ASCII header and line break
		fmt.Println("\n==============================")
		fmt.Println("Differing Values from First File")
		fmt.Println("==============================\n")
	}

	fmt.Println(string(data))
	return nil
}

func main() {
	var outputFormat string

	// Root command
	var rootCmd = &cobra.Command{
		Use:   "yamldiff [file1.yaml] [file2.yaml]",
		Short: "Compare two YAML files and output the differences.",
		Long: `yamldiff compares two YAML files and shows the differences.
By default, it outputs the differences as YAML with additional formatting for clarity.
You can choose other output format using the -o flag:

- yaml: Outputs the differences as plain YAML without additional formatting.
- yamldiff: Outputs the differences with an ASCII header and extra formatting for clarity.`,
		Args: cobra.ExactArgs(2), // Expect exactly two arguments
		Run: func(cmd *cobra.Command, args []string) {
			file1 := args[0]
			file2 := args[1]

			data1, err := loadYAML(file1)
			if err != nil {
				log.Fatalf("Error loading first file: %v\n", err)
			}

			data2, err := loadYAML(file2)
			if err != nil {
				log.Fatalf("Error loading second file: %v\n", err)
			}

			diffMap := make(map[interface{}]interface{})

			if outputFormat == "yaml" {
				compareMaps(data1, data2, "", diffMap, false)
				err := printYAML(diffMap, false)
				if err != nil {
					log.Fatalf("Error printing YAML: %v\n", err)
				}
			} else {
				compareMaps(data1, data2, "", diffMap, true)

				if outputFormat == "yamldiff" {
					err := printYAML(diffMap, true)
					if err != nil {
						log.Fatalf("Error printing YAML: %v\n", err)
					}
				}
			}
		},
	}

	// Adding the output format flag
	rootCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Set the output format (yaml, yamldiff).")

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
