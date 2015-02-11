package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

func main() {
	// 1. Parse command line arguments
	curDir, err := os.Getwd()
	check(err)
	metadataDirVal := flag.String("metadata", curDir,
		"Path to directory containig metadata files (api_data.json and attributes.json)")
	destDirVal := flag.String("output", curDir,
		"Path to output file")
	flag.Parse()

	metadataDir := *metadataDirVal
	if stat, _ := os.Stat(metadataDir); !stat.IsDir() {
		check(fmt.Errorf("%s is not a valid directory", metadataDir))
	}

	destDir := *destDirVal
	if stat, _ := os.Stat(destDir); stat != nil && !stat.IsDir() {
		check(fmt.Errorf("%s is not a valid directory", destDir))
	}

	apiDataFile := path.Join(metadataDir, "api_data.json")
	var apiData map[string]interface{}
	apiDataText, err := loadFile(apiDataFile)
	check(err)
	err = json.Unmarshal(apiDataText, &apiData)
	if err != nil {
		check(fmt.Errorf("Cannot unmarshal JSON read from '%s': %s", apiDataFile, err.Error()))
	}

	attributesFile := path.Join(metadataDir, "attributes.json")
	var attributes map[string]string
	attributesText, err := loadFile(attributesFile)
	check(err)
	err = json.Unmarshal(attributesText, &attributes)
	if err != nil {
		check(fmt.Errorf("Cannot unmarshal JSON read from '%s': %s", attributesFile, err.Error()))
	}

	// 2. Analyze
	analyzer := NewApiAnalyzer(apiData, attributes)
	descriptor := analyzer.Analyze()

	// 3. Write codegen.go
	check(NewClientGenerator("codegen.go").Generate(descriptor, destDir))

	// 4. Write codegen_cmds.go
	check(NewCmdGenerator("codegen_cmds.go").Generate(descriptor, destDir))
}

// Helper function that reads content from given file
func loadFile(file string) ([]byte, error) {
	if _, err := os.Stat(file); err != nil {
		return nil, fmt.Errorf("Cannot find '%s'", file)
	}
	js, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Cannot read '%s': %s", file, err.Error())
	}
	return js, nil
}

// Panic if error is not nil
func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "api15gen: %s\n", err.Error())
		os.Exit(1)
	}
}
