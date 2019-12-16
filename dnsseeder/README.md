dnsseeder
=========

## Requirements

Latest version of [Go](http://golang.org) (currently 1.13)

## Getting Started

- Install Go according to the installation instructions here:
  http://golang.org/doc/install

- Ensure Go was installed properly and is a supported version:

```bash
$ go version
$ go env GOROOT GOPATH
```

NOTE: The `GOROOT` and `GOPATH` above must not be the same path. It is
recommended that `GOPATH` is set to a directory in your home directory such as
`~/dev/go` to avoid write permission issues. It is also recommended to add
`$GOPATH/bin` to your `PATH` at this point.

- Run the following commands to obtain dnsseeder, all dependencies, and install it:

```bash
$ git clone https://github.com/kaspanet/dnsseeder $GOPATH/src/github.com/kaspanet/dnsseeder
$ cd $GOPATH/src/github.com/kaspanet/dnsseeder
$ go install . 
```

- dnsseeder will now be installed in either ```$GOROOT/bin``` or
  ```$GOPATH/bin``` depending on your configuration.  If you did not already
  add the bin directory to your system path during Go installation, we
  recommend you do so now.

To start dnsseeder listening on udp 127.0.0.1:5354 with an initial connection to working testnet node 192.168.0.1:

```
$ ./dnsseeder -n nameserver.example.com -H network-seed.example.com -s 192.168.0.1 --testnet
```

You will then need to redirect DNS traffic on your public IP port 53 to 127.0.0.1:5354

