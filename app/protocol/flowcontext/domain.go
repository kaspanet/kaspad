package flowcontext

import (
	"github.com/kaspanet/kaspad/domain"
)

// DAG returns the DAG associated to the flow context.
func (f *FlowContext) Domain() domain.Domain {
	return f.domain
}
