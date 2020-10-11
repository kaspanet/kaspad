package model

// TransactionSelector is a function for selecting transaction from
// some ReadOnlyUTXOSet
type TransactionSelector func(readOnlyUTXOSet ReadOnlyUTXOSet) []*DomainTransaction
