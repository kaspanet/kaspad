# c4exctl

c4exctl is an RPC client for c4exd

## Requirements

Go 1.19 or later.

## Installation

#### Build from Source

- Install Go according to the installation instructions here:
  http://golang.org/doc/install

- Ensure Go was installed properly and is a supported version:

```bash
$ go version
```

- Run the following commands to obtain and install c4exd including all dependencies:

```bash
$ git clone https://github.com/c4ei/c4exd
$ cd c4exd/cmd/c4exctl
$ go install .
```

- Kaspactl should now be installed in `$(go env GOPATH)/bin`. If you did not already add the bin directory to your
  system path during Go installation, you are encouraged to do so now.
- 이제 Kaspad(및 유틸리티)가 $(go env GOPATH)/bin에 설치됩니다. Go 설치 중에 시스템 경로에 bin 디렉터리를 아직 추가하지 않았다면 지금 추가하는 것이 좋습니다.

## Usage

The full kaspctl configuration options can be seen with:

```bash
$ kaspctl --help
```

But the minimum configuration needed to run it is:

```bash
$ c4exctl <REQUEST_JSON>
```

For example:

```
$ c4exctl '{"getBlockDagInfoRequest":{}}'
```

For a list of all available requests check out the [RPC documentation](infrastructure/network/netadapter/server/grpcserver/protowire/rpc.md)