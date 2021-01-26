package main

import (
	"reflect"
	"strconv"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/protowire"
	"github.com/pkg/errors"
)

func parseCommand(args []string, commandDescs []*commandDescription) (*protowire.KaspadMessage, error) {
	commandName, parameterStrings := args[0], args[1:]

	var commandDesc *commandDescription
	for _, cd := range commandDescs {
		if cd.name == commandName {
			commandDesc = cd
			break
		}
	}
	if commandDesc == nil {
		return nil, errors.Errorf("unknown command: %s. Use --list-commands to list all commands", commandName)
	}
	if len(parameterStrings) != len(commandDesc.parameters) {
		return nil, errors.Errorf("command '%s' expects %d parameters but got %d",
			commandName, len(commandDesc.parameters), len(parameterStrings))
	}

	commandValue := reflect.New(unwrapCommandType(commandDesc.typeof))
	for i, parameterDesc := range commandDesc.parameters {
		parameterValue, err := stringToValue(parameterDesc, parameterStrings[i])
		if err != nil {
			return nil, err
		}
		setField(commandValue, parameterValue, parameterDesc)
	}

	return generateKaspadMessage(commandValue, commandDesc)
}

func setField(commandValue reflect.Value, parameterValue reflect.Value, parameterDesc *parameterDescription) {
	parameterField := commandValue.Elem().FieldByName(parameterDesc.name)

	parameterField.Set(parameterValue)
}

func stringToValue(parameterDesc *parameterDescription, valueStr string) (reflect.Value, error) {
	var value interface{}
	var err error
	switch parameterDesc.typeof.Kind() {
	case reflect.Bool:
		value, err = strconv.ParseBool(valueStr)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
	case reflect.Int8:
		var valueInt64 int64
		valueInt64, err = strconv.ParseInt(valueStr, 10, 8)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
		value = int8(valueInt64)
	case reflect.Int16:
		var valueInt64 int64
		valueInt64, err = strconv.ParseInt(valueStr, 10, 16)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
		value = int16(valueInt64)
	case reflect.Int32:
		var valueInt64 int64
		valueInt64, err = strconv.ParseInt(valueStr, 10, 32)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
		value = int32(valueInt64)
	case reflect.Int64:
		value, err = strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
	case reflect.Uint8:
		var valueUInt64 uint64
		valueUInt64, err = strconv.ParseUint(valueStr, 10, 8)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
		value = uint8(valueUInt64)
	case reflect.Uint16:
		var valueUInt64 uint64
		valueUInt64, err = strconv.ParseUint(valueStr, 10, 16)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
		value = uint16(valueUInt64)
	case reflect.Uint32:
		var valueUInt64 uint64
		valueUInt64, err = strconv.ParseUint(valueStr, 10, 32)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
		value = uint32(valueUInt64)
	case reflect.Uint64:
		value, err = strconv.ParseUint(valueStr, 10, 64)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
	case reflect.Float32:
		var valueFloat64 float64
		valueFloat64, err = strconv.ParseFloat(valueStr, 32)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
		value = float32(valueFloat64)
	case reflect.Float64:
		value, err = strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
	case reflect.String:
		value = valueStr
	case reflect.Struct:
		pointer := reflect.New(parameterDesc.typeof) // create pointer to this type
		fieldInterface := pointer.Interface().(proto.Message)
		err := protojson.Unmarshal([]byte(valueStr), fieldInterface)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
		// Unpointer the value once it's ready
		fieldInterfaceValue := reflect.ValueOf(fieldInterface)
		value = fieldInterfaceValue.Elem().Interface()
	case reflect.Ptr:
		dummyParameterDesc := &parameterDescription{
			name:   "valuePointedTo",
			typeof: parameterDesc.typeof.Elem(),
		}
		valuePointedTo, err := stringToValue(dummyParameterDesc, valueStr)
		if err != nil {
			return reflect.Value{}, errors.WithStack(err)
		}
		pointer := pointerToValue(valuePointedTo)

		value = pointer.Interface()

	// Int and uint are not supported because their size is platform-dependant
	case reflect.Int:
	case reflect.Uint:
	// Other types are not supported simply because they are not used in any command right now
	// but support can be added if and when needed
	case reflect.Slice:
	case reflect.Func:
	case reflect.Interface:
	case reflect.Map:
	case reflect.UnsafePointer:
	case reflect.Invalid:
	case reflect.Uintptr:
	case reflect.Complex64:
	case reflect.Complex128:
	case reflect.Array:
	case reflect.Chan:
	default:
		return reflect.Value{},
			errors.Errorf("Unsupported type '%s' for parameter '%s'", parameterDesc.typeof.Kind(), parameterDesc.name)
	}

	return reflect.ValueOf(value), nil
}
