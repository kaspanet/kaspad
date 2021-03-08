module github.com/kaspanet/automation/stability-tests

go 1.16

require (
	github.com/jessevdk/go-flags v1.4.0
	github.com/kaspanet/go-secp256k1 v0.0.3
	github.com/kaspanet/kaspad v0.7.2
	github.com/pkg/errors v0.9.1
	golang.org/x/tools v0.0.0-20200228224639-71482053b885 // indirect
)

replace github.com/kaspanet/kaspad => ../../kaspad
