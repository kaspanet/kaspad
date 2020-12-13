# -- multistage docker build: stage #1: build stage
FROM golang:1.14-alpine AS build

RUN mkdir -p /go/src/github.com/kaspanet/kaspad

WORKDIR /go/src/github.com/kaspanet/kaspad

RUN apk add --no-cache curl git openssh binutils gcc musl-dev
RUN go get -u golang.org/x/lint/golint \
      github.com/kisielk/errcheck \
      github.com/opennota/check/cmd/aligncheck \
      github.com/opennota/check/cmd/structcheck \
      github.com/opennota/check/cmd/varcheck \
      honnef.co/go/tools/cmd/staticcheck 

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN GOFMT_RESULT=`go fmt ./...`; echo $GOFMT_RESULT; test -z "$GOFMT_RESULT"
RUN go vet -composites=false ./...
RUN golint -set_exit_status ./...
# RUN aligncheck ./...
# RUN structcheck -e ./...
# RUN varcheck -e ./...
RUN staticcheck -checks "SA4006,SA4008,SA4009,SA4010,SA5003,SA1004,SA1014,SA1021,SA1023,SA1024,SA1025,SA1026,SA1027,SA1028,SA2000,SA2001,SA2003,SA4000,SA4001,SA4003,SA4004,SA4011,SA4012,SA4013,SA4014,SA4015,SA4016,SA4017,SA4018,SA4019,SA4020,SA4021,SA4022,SA4023,SA5000,SA5002,SA5004,SA5005,SA5007,SA5008,SA5009,SA5010,SA5011,SA5012,SA6001,SA6002,SA9001,SA9002,SA9003,SA9004,SA9005,SA9006,ST1019" ./...
RUN GOOS=linux go build -a -installsuffix cgo -o kaspad .

# Remove the line below and uncomment the line after it for testing with coverage
RUN go test ./...

# --- multistage docker build: stage #2: runtime image
FROM alpine
WORKDIR /app

RUN apk add --no-cache ca-certificates tini

COPY --from=build /go/src/github.com/kaspanet/kaspad/kaspad /app/
COPY --from=build /go/src/github.com/kaspanet/kaspad/sample-kaspad.conf /app/

USER nobody
ENTRYPOINT [ "/sbin/tini", "--" ]
