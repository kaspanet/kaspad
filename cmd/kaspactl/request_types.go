package main

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/protowire"
)

var requestTypes = []reflect.Type{
	reflect.TypeOf(protowire.KaspadMessage_AddPeerRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetConnectedPeerInfoRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetPeerAddressesRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetCurrentNetworkRequest{}),

	reflect.TypeOf(protowire.KaspadMessage_GetBlockRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetBlocksRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetHeadersRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetBlockCountRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetBlockDagInfoRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetSelectedTipHashRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetVirtualSelectedParentBlueScoreRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetVirtualSelectedParentChainFromBlockRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_ResolveFinalityConflictRequest{}),

	reflect.TypeOf(protowire.KaspadMessage_GetBlockTemplateRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_SubmitBlockRequest{}),

	reflect.TypeOf(protowire.KaspadMessage_GetMempoolEntryRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_GetMempoolEntriesRequest{}),
	reflect.TypeOf(protowire.KaspadMessage_SubmitTransactionRequest{}),

	reflect.TypeOf(protowire.KaspadMessage_GetUtxosByAddressesRequest{}),
}

type requestDescription struct {
	name       string
	parameters []*parameterDescription
	typeof     reflect.Type
}

type parameterDescription struct {
	name   string
	typeof reflect.Type
}

func unwrapRequestType(requestTypeWrapped reflect.Type) reflect.Type {
	return requestTypeWrapped.Field(0).Type.Elem()
}

func requestDescriptions() []*requestDescription {
	requestDescriptions := make([]*requestDescription, len(requestTypes))

	for i, requestTypeWrapped := range requestTypes {
		requestType := unwrapRequestType(requestTypeWrapped)

		name := strings.TrimSuffix(requestType.Name(), "RequestMessage")
		numFields := requestType.NumField()

		var parameters []*parameterDescription
		for i := 0; i < numFields; i++ {
			field := requestType.Field(i)

			if !unicode.IsUpper(rune(field.Name[0])) { // Only exported fields are of interest
				continue
			}
			parameters = append(parameters, &parameterDescription{
				name:   field.Name,
				typeof: field.Type,
			})
		}
		requestDescriptions[i] = &requestDescription{
			name:       name,
			parameters: parameters,
			typeof:     requestTypeWrapped,
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
