package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/pkg/errors"
)

const instanceStateCodeActive = "16"

func getAddressList(cfg *config) ([]string, error) {
	if cfg.AddressListPath != "" {
		return getAddressListFromPath(cfg)
	}
	return getAddressListFromAWS(cfg)
}

func getAddressListFromAWS(cfg *config) ([]string, error) {
	log.Infof("Getting hosts list for autoscaling group %s", cfg.AutoScalingGroup)
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(cfg.Region)}))
	ec2Client := ec2.New(sess)
	instances, err := ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{Name: aws.String("tag:aws:autoscaling:groupName"), Values: []*string{&cfg.AutoScalingGroup}},
			&ec2.Filter{Name: aws.String("instance-state-code"), Values: []*string{aws.String(instanceStateCodeActive)}},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error describing instances")
	}

	addressList := []string{}
	for _, reservation := range instances.Reservations {
		for _, instance := range reservation.Instances {
			if instance.PrivateDnsName == nil {
				continue
			}
			addressList = append(addressList, fmt.Sprintf("%s:%s", *instance.PrivateDnsName, dagconfig.DevNetParams.RPCPort))
		}
	}

	return addressList, nil
}

func getAddressListFromPath(cfg *config) ([]string, error) {
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
