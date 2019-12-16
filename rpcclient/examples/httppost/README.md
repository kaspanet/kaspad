HTTP POST Example
==============================

This example shows how to use the rpcclient package to connect to a Kaspa
RPC server using HTTP POST mode with TLS disabled and gets the current
block count.

## Running the Example

Modify the `main.go` source to specify the correct RPC username and
password for the RPC server:

```Go
	User: "yourrpcuser",
	Pass: "yourrpcpass",
```

Finally, navigate to the example's directory and run it with:

```bash
$ cd $GOPATH/src/github.com/kaspanet/kaspad/rpcclient/examples/httppost
$ go run *.go
```

