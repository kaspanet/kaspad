package serverutils

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/connmgr"
	"github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/util"
)

// Peer extends the peer to maintain state shared by the server and
// the blockmanager.
type Peer struct {
	*peer.Peer

	// The following variables must only be used atomically
	FeeFilter int64

	relayMtx        sync.Mutex
	DynamicBanScore connmgr.DynamicBanScore
	quit            chan struct{}
	DisableRelayTx  bool

	// The following chans are used to sync blockmanager and server.
	txProcessed    chan struct{}
	blockProcessed chan struct{}
}

// BTCDLookup resolves the IP of the given host using the correct DNS lookup
// function depending on the configuration options.  For example, addresses will
// be resolved using tor when the --proxy flag was specified unless --noonion
// was also specified in which case the normal system DNS resolver will be used.
//
// Any attempt to resolve a tor address (.onion) will return an error since they
// are not intended to be resolved outside of the tor proxy.
func BTCDLookup(host string) ([]net.IP, error) {
	if strings.HasSuffix(host, ".onion") {
		return nil, errors.Errorf("attempt to resolve tor address %s", host)
	}

	return config.ActiveConfig().Lookup(host)
}

// GenCertPair generates a key/cert pair to the paths provided.
func GenCertPair(certFile, keyFile string) error {
	log.Infof("Generating TLS certificates...")

	org := "btcd autogenerated cert"
	validUntil := time.Now().Add(10 * 365 * 24 * time.Hour)
	cert, key, err := util.NewTLSCertPair(org, validUntil, nil)
	if err != nil {
		return err
	}

	// Write cert and key files.
	if err = ioutil.WriteFile(certFile, cert, 0666); err != nil {
		return err
	}
	if err = ioutil.WriteFile(keyFile, key, 0600); err != nil {
		os.Remove(certFile)
		return err
	}

	log.Infof("Done generating TLS certificates")
	return nil
}

// BTCDDial connects to the address on the named network using the appropriate
// dial function depending on the address and configuration options.  For
// example, .onion addresses will be dialed using the onion specific proxy if
// one was specified, but will otherwise use the normal dial function (which
// could itself use a proxy or not).
func BTCDDial(addr net.Addr) (net.Conn, error) {
	if strings.Contains(addr.String(), ".onion:") {
		return config.ActiveConfig().OnionDial(addr.Network(), addr.String(),
			config.DefaultConnectTimeout)
	}
	return config.ActiveConfig().Dial(addr.Network(), addr.String(), config.DefaultConnectTimeout)
}
