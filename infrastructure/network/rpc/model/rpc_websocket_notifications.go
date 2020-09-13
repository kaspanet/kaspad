package model

//const (
//	// FinalityConflictNtfnMethod is the new method used for notifications
//	// from the kaspa rpc server that inform a client that a finality conflict
//	// has occured.
//	FinalityConflictNtfnMethod = "finalityConflict"
//
//	// FinalityConflictResolvedNtfnMethod is the new method used for notifications
//	// from the kaspa rpc server that inform a client that a finality conflict
//	// has been resolved.
//	FinalityConflictResolvedNtfnMethod = "finalityConflictResolved"
//)
//
//// FinalityConflictNtfn  defines the parameters to the finalityConflict
//// JSON-RPC notification.
//type FinalityConflictNtfn struct {
//	ViolatingBlockHash string `json:"violatingBlockHash"`
//}
//
//// NewFinalityConflictNtfn returns a new instance which can be used to issue a
//// finalityConflict JSON-RPC notification.
//func NewFinalityConflictNtfn(violatingBlockHash string) *FinalityConflictNtfn {
//	return &FinalityConflictNtfn{
//		ViolatingBlockHash: violatingBlockHash,
//	}
//}
//
//// FinalityConflictResolvedNtfn defines the parameters to the
//// finalityConflictResolved JSON-RPC notification.
//type FinalityConflictResolvedNtfn struct {
//	FinalityBlockHash string
//}
//
//// NewFinalityConflictResolvedNtfn returns a new instance which can be used to issue a
//// finalityConflictResolved JSON-RPC notification.
//func NewFinalityConflictResolvedNtfn(finalityBlockHash string) *FinalityConflictResolvedNtfn {
//	return &FinalityConflictResolvedNtfn{
//		FinalityBlockHash: finalityBlockHash,
//	}
//}
