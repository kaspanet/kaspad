dnsseeder
=========

## Requirements

[Go](http://golang.org) 1.10 or newer.

## Getting Started

- dnsseeder will now be installed in either ```$GOROOT/bin``` or
  ```$GOPATH/bin``` depending on your configuration.  If you did not already
  add the bin directory to your system path during Go installation, we
  recommend you do so now.

### Build from source (all platforms)

Building or updating from source requires the following build dependencies:

- **Go 1.10 or 1.11**

  Installation instructions can be found here: https://golang.org/doc/install.
  It is recommended to add `$GOPATH/bin` to your `PATH` at this point.

- **Vgo (Go 1.10 only)**

To build and install from a checked-out repo, run `go install` in the repo's
root directory.  Some notes:

* Replace `go` with `vgo` when using Go 1.10.

* The `dnsseeder` executable will be installed to `$GOPATH/bin`.  `GOPATH`
  defaults to `$HOME/go` (or `%USERPROFILE%\go` on Windows) if unset.

For more information about Daglabs and how to set up your software please go to
our docs page at [docs.daglabs.org](https://docs.daglabs.org/getting-started/beginner-guide/).

To start dnsseeder listening on udp 127.0.0.1:5354 with an initial connection to working testnet node 192.168.0.1:

```
$ ./dnsseeder -n nameserver.example.com -H network-seed.example.com -s 192.168.0.1 --testnet
```

You will then need to redirect DNS traffic on your public IP port 53 to 127.0.0.1:5354

## Issue Tracker

The [integrated github issue tracker](https://github.com/daglabs/dnsseeder/issues)
is used for this project.

## License

dnsseeder is licensed under the [copyfree](http://copyfree.org) ISC License.
