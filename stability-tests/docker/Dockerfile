ARG KASPAD_VERSION

FROM 578712463641.dkr.ecr.eu-central-1.amazonaws.com/kaspad-release-candidate:$KASPAD_VERSION as kaspad
FROM 578712463641.dkr.ecr.eu-central-1.amazonaws.com/kaspaminer-release-candidate:$KASPAD_VERSION as kaspaminer

FROM golang:1.16-alpine

RUN apk add bash build-base git

ARG KASPAD_VERSION

COPY --from=kaspad /app/ /app/
COPY --from=kaspaminer /app/ /app/
ENV PATH="/app:${PATH}"

COPY . /tests

WORKDIR /tests

RUN git ls-remote https://github.com/kaspanet/kaspad.git $KASPAD_VERSION | awk '{print $1;}' > /tmp/kaspad_git_commit
RUN go mod edit -dropreplace github.com/kaspanet/kaspad
RUN go mod edit -replace github.com/kaspanet/kaspad=github.com/kaspanet/kaspad@`cat /tmp/kaspad_git_commit` ;
RUN go mod download

RUN go install ./...

ENTRYPOINT ["./run/run.sh"]