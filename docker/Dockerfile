# -- multistage docker build: stage #1: build stage
FROM golang:1.15-alpine AS build

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
RUN go vet ./...
RUN golint -set_exit_status ./...
# RUN aligncheck ./...
# RUN structcheck -e ./...
# RUN varcheck -e ./...
RUN staticcheck -checks SA4006 ./...
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
