package main

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/kaspanet/kaspad/app/appmessage"
)

var requestTypes = []reflect.Type{
	reflect.TypeOf(appmessage.GetCurrentNetworkRequestMessage{}),
	reflect.TypeOf(appmessage.SubmitBlockRequestMessage{}),
	reflect.TypeOf(appmessage.GetBlockTemplateRequestMessage{}),
	reflect.TypeOf(appmessage.GetPeerAddressesRequestMessage{}),
	reflect.TypeOf(appmessage.GetPeerAddressesKnownAddressMessage{}),
	reflect.TypeOf(appmessage.GetSelectedTipHashRequestMessage{}),
	reflect.TypeOf(appmessage.GetMempoolEntryRequestMessage{}),
	reflect.TypeOf(appmessage.GetMempoolEntriesRequestMessage{}),
	reflect.TypeOf(appmessage.GetConnectedPeerInfoRequestMessage{}),
	reflect.TypeOf(appmessage.AddPeerRequestMessage{}),
	reflect.TypeOf(appmessage.SubmitTransactionRequestMessage{}),
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

func requestDescriptions() map[string]*requestDescription {
	requestDescriptions := make(map[string]*requestDescription, len(requestTypes))

	for _, requestType := range requestTypes {
		name := strings.TrimSuffix(requestType.Name(), "RequestMessage")
		numFields := requestType.NumField()

		parameters := make([]*parameterDescription, numFields)
		for i := 0; i < numFields; i++ {
			field := requestType.Field(i)

			if unicode.IsUpper(rune(field.Name[0])) { // Only exported fields are of interest
				continue
			}
			parameters = append(parameters, &parameterDescription{
				name:   field.Name,
				typeof: field.Type,
			})
			fmt.Printf("\t%s: %s\n", field.Name, field.Type.Name())
		}
		requestDescriptions[name] = &requestDescription{
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
