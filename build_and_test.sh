#!/bin/sh -ex

FLAGS=$@

go version

go get $FLAGS -t -d ./...
# This is to bypass a go bug: https://github.com/golang/go/issues/27643
GO111MODULE=off go get $FLAGS golang.org/x/lint/golint \
                          honnef.co/go/tools/cmd/staticcheck

test -z "$(go fmt ./...)"

golint -set_exit_status ./...

staticcheck -checks SA4006,SA4008,SA4009,SA4010,SA5003,SA1004,SA1014,SA1021,SA1023,SA1024,SA1025,SA1026,SA1027,SA1028,SA2000,SA2001,SA2003,SA4000,SA4001,SA4003,SA4004,SA4011,SA4012,SA4013,SA4014,SA4015,SA4016,SA4017,SA4018,SA4019,SA4020,SA4021,SA4022,SA4023,SA5000,SA5002,SA5004,SA5005,SA5007,SA5008,SA5009,SA5010,SA5011,SA5012,SA6001,SA6002,SA9001,SA9002,SA9003,SA9004,SA9005,SA9006,ST1019 ./...

go vet -composites=false $FLAGS ./...

go build $FLAGS -o kaspad .

if [ -n "${NO_PARALLEL}" ]
then
  go test -parallel=1 -timeout 3600s $FLAGS ./...
else
  go test $FLAGS -timeout 3600s ./...
fi