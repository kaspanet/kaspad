package addressexchange

import (
	"errors"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/flowcontext"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func setupTestDomain(testName string) (domainInstance domain.Domain, teardown func(), err error) {
	dataDir, err := ioutil.TempDir("", testName)
	if err != nil {
		return nil, nil, err
	}
	db, err := ldb.NewLevelDB(dataDir)
	if err != nil {
		return nil, nil, err
	}
	teardown = func() {
		db.Close()
		os.RemoveAll(dataDir)
	}

	params := dagconfig.SimnetParams
	domainInstance, err = domain.New(&params, db)
	if err != nil {
		teardown()
		return nil, nil, err
	}

	return domainInstance, teardown, nil
}

func TestSendAddresses(t *testing.T) {
	testDomain, teardown, err := setupTestDomain(t.Name())
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}
	defer teardown()

	defaultCfg.ActiveNetParams.AcceptUnroutable = true
	addressManager, err := addressmanager.New(addressmanager.NewConfig(defaultCfg))
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	addresses := generateAddressesForTest(5)
	addressManager.AddAddresses(addresses...)

	netAdapter, err := netadapter.NewNetAdapter(defaultCfg)
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	connManager, err := connmanager.New(defaultCfg, netAdapter, addressManager)
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	ctx := flowcontext.New(defaultCfg, testDomain, addressManager, netAdapter, connManager)
	incomingRoute := router.NewRoute()
	outgoingRoute := router.NewRoute()
	err = incomingRoute.Enqueue(appmessage.NewMsgRequestAddresses(true, &subnetworks.SubnetworkIDNative))
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	go func() {
		message, err := outgoingRoute.Dequeue()
		if err != nil {
			t.Fatalf("Unexpected error in incomingRoute.Dequeue()")
		}

		msgAddresses, ok := message.(*appmessage.MsgAddresses)
		if !ok {
			t.Fatalf("Unexpected message, expected: %s, got: %s", appmessage.CmdAddresses, message.Command())
		}

		matchCount := 0
		for _, address := range addresses {
			for _, msgAddress := range msgAddresses.AddressList {
				if reflect.DeepEqual(address, msgAddress) {
					matchCount++
					break
				}
			}
		}

		if matchCount != len(addresses) {
			t.Fatalf("Not all add")
		}
		incomingRoute.Close()
	}()

	err = SendAddresses(ctx, incomingRoute, outgoingRoute)
	if err != nil && !errors.Is(err, router.ErrRouteClosed) {
		t.Fatalf("SendAddresses: %s", err)
	}
}
