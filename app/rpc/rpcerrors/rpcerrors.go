package rpcerrors

type RPCError struct {
	Message string
}

func (e RPCError) Error() string {
	return e.Message
}
