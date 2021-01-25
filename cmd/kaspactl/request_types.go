package main

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/kaspanet/kaspad/app/appmessage"
)

var requestTypes = []reflect.Type{
	reflect.TypeOf(appmessage.AddPeerRequestMessage{}),
	reflect.TypeOf(appmessage.GetConnectedPeerInfoRequestMessage{}),
	reflect.TypeOf(appmessage.GetPeerAddressesRequestMessage{}),
	reflect.TypeOf(appmessage.GetCurrentNetworkRequestMessage{}),

	reflect.TypeOf(appmessage.GetBlockRequestMessage{}),
	reflect.TypeOf(appmessage.GetBlocksRequestMessage{}),
	reflect.TypeOf(appmessage.GetHeadersRequestMessage{}),
	reflect.TypeOf(appmessage.GetBlockCountRequestMessage{}),
	reflect.TypeOf(appmessage.GetBlockDAGInfoRequestMessage{}),
	reflect.TypeOf(appmessage.GetSelectedTipHashRequestMessage{}),
	reflect.TypeOf(appmessage.GetVirtualSelectedParentBlueScoreRequestMessage{}),
	reflect.TypeOf(appmessage.GetVirtualSelectedParentChainFromBlockRequestMessage{}),
	reflect.TypeOf(appmessage.ResolveFinalityConflictRequestMessage{}),

	reflect.TypeOf(appmessage.GetBlockTemplateRequestMessage{}),
	reflect.TypeOf(appmessage.SubmitBlockRequestMessage{}),

	reflect.TypeOf(appmessage.GetMempoolEntryRequestMessage{}),
	reflect.TypeOf(appmessage.GetMempoolEntriesRequestMessage{}),
	reflect.TypeOf(appmessage.SubmitTransactionRequestMessage{}),

	reflect.TypeOf(appmessage.GetUTXOsByAddressesRequestMessage{}),
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

func requestDescriptions() []*requestDescription {
	requestDescriptions := make([]*requestDescription, len(requestTypes))

	for i, requestType := range requestTypes {
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
			typeof:     requestType,
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
