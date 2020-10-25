package txscript

const (
	// StandardVerifyFlags are the script flags which are used when
	// executing transaction scripts to enforce additional checks which
	// are required for the script to be considered standard. These checks
	// help reduce issues related to transaction malleability as well as
	// allow pay-to-script hash transactions. Note these flags are
	// different than what is required for the consensus rules in that they
	// are more strict.
	StandardVerifyFlags = ScriptDiscourageUpgradableNops
)
