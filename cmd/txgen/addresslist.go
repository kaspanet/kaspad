package main

import (
	"bufio"
	"os"
)

func getAddressList(cfg *config) ([]string, error) {
	file, err := os.Open(cfg.AddressListPath)
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
