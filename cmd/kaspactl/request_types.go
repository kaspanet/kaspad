package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/protowire"
)

var requestTypes = []reflect.Type{
	reflect.TypeOf(protowire.GetCurrentNetworkRequestMessage{}),
	reflect.TypeOf(protowire.SubmitBlockRequestMessage{}),
	reflect.TypeOf(protowire.GetBlockTemplateRequestMessage{}),
	reflect.TypeOf(protowire.GetPeerAddressesRequestMessage{}),
	reflect.TypeOf(protowire.GetPeerAddressesKnownAddressMessage{}),
	reflect.TypeOf(protowire.GetSelectedTipHashRequestMessage{}),
	reflect.TypeOf(protowire.GetMempoolEntryRequestMessage{}),
	reflect.TypeOf(protowire.GetMempoolEntriesRequestMessage{}),
	reflect.TypeOf(protowire.GetConnectedPeerInfoRequestMessage{}),
	reflect.TypeOf(protowire.AddPeerRequestMessage{}),
	reflect.TypeOf(protowire.SubmitTransactionRequestMessage{}),
}

type requestDescription struct {
	name       string
	parameters []*parameter
}

type parameter struct {
	name   string
	typeof reflect.Type
}

func requestDescriptions() []*requestDescription {
	requestDescriptions := make([]*requestDescription, len(requestTypes))

	for i, requestType := range requestTypes {
		name := strings.TrimSuffix(requestType.Name(), "RequestMessage")
		numFields := requestType.NumField()

		parameters := make([]*parameter, numFields)
		for i := 0; i < numFields; i++ {
			field := requestType.Field(i)

			if field.Tag.Get("protobuf") == "" {
				// fields that do not have the protobuf tag are not part of the message
				continue
			}
			parameters = append(parameters, &parameter{
				name:   field.Name,
				typeof: field.Type,
			})
			fmt.Printf("\t%s: %s\n", field.Name, field.Type.Name())
		}
		requestDescriptions[i] = &requestDescription{
			name:       name,
			parameters: parameters,
		}
	}

	return requestDescriptions
}

func (rd *requestDescription) help() string {
	sb := &strings.Builder{}
	sb.WriteString(rd.name)
	for _, parameter := range rd.parameters {
		_, _ = fmt.Fprintf(sb, " [%s]", parameter.name)
	}
	return sb.String()
}
