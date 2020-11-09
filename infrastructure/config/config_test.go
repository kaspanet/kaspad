package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func TestCreateDefaultConfigFile(t *testing.T) {
	// find out where the sample config lives
	_, path, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("Failed finding config file path")
	}
	sampleConfigFile := filepath.Join(filepath.Dir(path), "..", "..", "sample-kaspad.conf")

	// Setup a temporary directory
	tmpDir, err := ioutil.TempDir("", "kaspad")
	if err != nil {
		t.Fatalf("Failed creating a temporary directory: %v", err)
	}
	testpath := filepath.Join(tmpDir, "test.conf")

	// copy config file to location of kaspad binary
	data, err := ioutil.ReadFile(sampleConfigFile)
	if err != nil {
		t.Fatalf("Failed reading sample config file: %v", err)
	}
	appPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		t.Fatalf("Failed obtaining app path: %v", err)
	}
	tmpConfigFile := filepath.Join(appPath, "sample-kaspad.conf")
	err = ioutil.WriteFile(tmpConfigFile, data, 0644)
	if err != nil {
		t.Fatalf("Failed copying sample config file: %v", err)
	}

	// Clean-up
	defer func() {
		os.Remove(testpath)
		os.Remove(tmpConfigFile)
		os.Remove(tmpDir)
	}()

	err = createDefaultConfigFile(testpath)
	if err != nil {
		t.Fatalf("Failed to create a default config file: %v", err)
	}
}

// TestConstants makes sure that all constants hard-coded into the help text were not modified.
func TestConstants(t *testing.T) {
	zero := externalapi.DomainSubnetworkID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	if subnetworks.SubnetworkIDNative != zero {
		t.Errorf("subnetworks.SubnetworkIDNative value was changed from 0, therefore you probably need to update the help text for SubnetworkID")
	}
	one := externalapi.DomainSubnetworkID{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	if subnetworks.SubnetworkIDCoinbase != one {
		t.Errorf("subnetworks.SubnetworkIDCoinbase value was changed from 1, therefore you probably need to update the help text for SubnetworkID")
	}
	two := externalapi.DomainSubnetworkID{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	if subnetworks.SubnetworkIDRegistry != two {
		t.Errorf("subnetworks.SubnetworkIDRegistry value was changed from 2, therefore you probably need to update the help text for SubnetworkID")
	}
}
