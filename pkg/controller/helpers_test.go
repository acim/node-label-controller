package controller_test

import (
	"testing"

	"github.com/acim/node-label-controller/pkg/controller"
	api "k8s.io/api/core/v1"
)

func TestOSPrefix(t *testing.T) {
	tests := []struct {
		in  string
		out bool
	}{
		{
			in:  "Ubuntu",
			out: true,
		},
		{
			in:  "Ubuntu 18.04",
			out: true,
		},
		{
			in:  "Ubuntu 16.04",
			out: false,
		},
		{
			in:  "Linux Mint",
			out: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			f := controller.OSPrefix(tt.in)
			out := f(nodeStub("Ubuntu 18.04"))
			if out != tt.out {
				t.Errorf("got %t, want %t", out, tt.out)
			}
		})
	}
}

func nodeStub(os string) *api.Node {
	return &api.Node{
		Status: api.NodeStatus{
			NodeInfo: api.NodeSystemInfo{
				OSImage: os,
			},
		},
	}
}
