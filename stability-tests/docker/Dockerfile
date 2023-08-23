ARG C4exD_IMAGE
ARG C4exMINER_IMAGE

FROM ${C4exD_IMAGE} as c4exd
FROM ${C4exMINER_IMAGE} as c4exminer

FROM golang:1.19-alpine

RUN mkdir -p /go/src/github.com/c4ei/c4exd

WORKDIR /go/src/github.com/c4ei/c4exd

RUN apk add bash build-base git

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

COPY --from=c4exd /app/ /app/
COPY --from=c4exminer /app/ /app/
ENV PATH="/app:${PATH}"

WORKDIR /go/src/github.com/c4ei/yunseokyeol/stability-tests

RUN go install ./...

ENTRYPOINT ["./run/run.sh"]