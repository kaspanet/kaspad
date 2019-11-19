package main

import (
	"bufio"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

func getAddressList(cfg *config) ([]string, error) {
	if cfg.AddressListPath != "" {
		return getAddressListFromPath(cfg.AddressListPath)
	}
	return getAddressListFromAWS(cfg.AutoScalingGroup, cfg.Region)
}

func getAddressListFromAWS(autoScalingGroup string, region string) ([]string, error) {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	ec2Client := ec2.New(sess)
	instances, err := ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{Name: aws.String("tag:aws:autoscaling:groupName"), Values: []*string{&autoScalingGroup}},
			&ec2.Filter{Name: aws.String("instance-state-code"), Values: []*string{aws.String("16")}},
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
			addressList = append(addressList, *instance.PrivateDnsName)
		}
	}

	return addressList, nil
}

func getAddressListFromPath(addressListPath string) ([]string, error) {
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
