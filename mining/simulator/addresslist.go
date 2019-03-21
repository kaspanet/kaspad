package main

import (
	"bufio"
	"os"
)

const addressListPath = "addresses"

func getAddressList() ([]string, error) {
	file, err := os.Open(addressListPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	addressList := []string{}
	for scanner.Scan() {
		addressList = append(addressList, scanner.Text())
	}

	return addressList, nil
}
