package controller

import (
	"strings"

	api "k8s.io/api/core/v1"
)

// OSPrefix returns node mathinc function to match node by opearating system name's prefix.
func OSPrefix(os string) func(n *api.Node) bool {
	return func(n *api.Node) bool {
		return strings.HasPrefix(n.Status.NodeInfo.OSImage, os)
	}
}
