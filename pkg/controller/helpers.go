package controller

import (
	"strings"

	api "k8s.io/api/core/v1"
)

// NodeLabel contains node match criteria function and label key and value.
type NodeLabel struct {
	matcher func(*api.Node) bool
	key     string
	value   string
}

// NewNodeLabel creates new node label object containg node match critieria function and label key and value.
func NewNodeLabel(matcher func(*api.Node) bool, labelKey, labelValue string) *NodeLabel {
	return &NodeLabel{
		matcher: matcher,
		key:     labelKey,
		value:   labelValue,
	}
}

// OSPrefix returns node mathinc function to match node by opearating system name's prefix.
func OSPrefix(os string) func(n *api.Node) bool {
	return func(n *api.Node) bool {
		return strings.HasPrefix(n.Status.NodeInfo.OSImage, os)
	}
}
