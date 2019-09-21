package controller

import (
	"testing"

	api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

const (
	labelKey   = "foo"
	labelValue = "bar"
	nodeName   = "baz-node"
)

func TestSyncHandler(t *testing.T) {
	c := fakeController()
	n, err := c.kubeClient.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
	if err != nil {
		t.Error(err)
	}

	key, err := cache.MetaNamespaceKeyFunc(n)
	if err != nil {
		t.Error(err)
	}

	err = c.syncHandler(key)
	if err != nil {
		t.Error(err)
	}

	n, err = c.kubeClient.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
	if err != nil {
		t.Error(err)
	}

	if _, ok := n.Labels[labelKey]; !ok {
		t.Errorf("expected %s label", labelKey)
	}

	if n.Labels[labelKey] != labelValue {
		t.Errorf("want label %s equal %s, got %s", labelKey, labelValue, n.Labels[labelKey])
	}
}

func TestLabelNode(t *testing.T) {
	c := fakeController()
	n, err := c.kubeClient.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
	if err != nil {
		t.Error(err)
	}

	err = c.labelNode(n)
	if err != nil {
		t.Error(err)
	}

	n, err = c.kubeClient.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
	if err != nil {
		t.Error(err)
	}

	if _, ok := n.Labels[labelKey]; !ok {
		t.Errorf("expected %s label", labelKey)
	}

	if n.Labels[labelKey] != labelValue {
		t.Errorf("want label %s equal %s, got %s", labelKey, labelValue, n.Labels[labelKey])
	}
}

func fakeController() *Controller {
	n := &api.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   nodeName,
			Labels: make(map[string]string, 1),
		},
	}
	kubeClient := fake.NewSimpleClientset(n)
	kubeInformerFactory := informers.NewSharedInformerFactory(kubeClient, 0)
	kubeInformerFactory.Core().V1().Nodes().Informer().GetIndexer().Add(n)

	return NewController(kubeClient, kubeInformerFactory.Core().V1().Nodes(), NewNodeLabel(OSPrefix(""), labelKey, labelValue))
}
