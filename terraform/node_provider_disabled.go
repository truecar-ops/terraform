package terraform

import (
	"fmt"

	"github.com/truecar-ops/terraform/dag"
)

// NodeDisabledProvider represents a provider that is disabled. A disabled
// provider does nothing. It exists to properly set inheritance information
// for child providers.
type NodeDisabledProvider struct {
	*NodeAbstractProvider
}

var (
	_ GraphNodeModulePath     = (*NodeDisabledProvider)(nil)
	_ GraphNodeReferencer     = (*NodeDisabledProvider)(nil)
	_ GraphNodeProvider       = (*NodeDisabledProvider)(nil)
	_ GraphNodeAttachProvider = (*NodeDisabledProvider)(nil)
	_ dag.GraphNodeDotter     = (*NodeDisabledProvider)(nil)
)

func (n *NodeDisabledProvider) Name() string {
	return fmt.Sprintf("%s (disabled)", n.NodeAbstractProvider.Name())
}
