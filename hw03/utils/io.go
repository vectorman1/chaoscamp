package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

var _technologies *ScannerData

func ReadUrls(loc string) ([]string, error) {
	file, err := os.Open(loc)

	if err != nil {
		log.Fatalf("Failed to read urls %v", err)
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, nil
}

func ReadTechnologies() {
	wd, _ := os.Getwd()
	file, err := ioutil.ReadFile(fmt.Sprintf("%s\\hw03\\technologies.json", wd))

	if err != nil {
		log.Fatalln(err)
		return
	}

	var technologies ScannerData

	err = json.Unmarshal(file, &technologies)

	if err != nil {
		log.Fatalln(err)
		return
	}

	_technologies = &technologies
}

func GetScannerData() *ScannerData {
	return _technologies
}
