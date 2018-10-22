package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func TestIsK8sPodReady(t *testing.T) {
	pod := &core.Pod{
		Status: core.PodStatus{
			Conditions: []core.PodCondition{
				{
					Type:   core.PodReady,
					Status: core.ConditionTrue,
				},
			},
		},
	}
	assert.True(t, isK8sPodReady(pod))

	pod.Status.Conditions[0].Status = core.ConditionFalse
	assert.False(t, isK8sPodReady(pod))
}

func TestMakeInClusterK8sClient(t *testing.T) {
	c, err := makeInClusterK8sClient()
	assert.Zero(t, c)
	assert.Error(t, err)

	def := defK8sConfig
	defK8sConfig = func() (*rest.Config, error) {
		return new(rest.Config), nil
	}
	defer func() { defK8sConfig = def }()

	c, err = makeInClusterK8sClient()
	assert.IsType(t, &kubernetes.Clientset{}, c)
	assert.NoError(t, err)
}

func TestMakeK8sName(t *testing.T) {
	n, err := makeK8sName("value3.key3.value2.key2.value1.key1.namespace")
	if assert.NoError(t, err) {
		assert.Equal(t, "namespace", n.namespace)
		if assert.NotZero(t, n.selector) {
			assert.True(t, n.selector.Matches(labels.Set{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}))
		}
	}
}

func TestMakeK8sNameWithErrNameTooShort(t *testing.T) {
	n, err := makeK8sName("key.namespace")
	assert.Equal(t, errK8sNameTooShort, err, "name %#v", n)
}

func TestMakeK8sNameWithErrNameInvalid(t *testing.T) {
	n, err := makeK8sName("key2.value1.key1.namespace")
	assert.Equal(t, errK8sNameInvalid, err, "name %#v", n)
}

func TestMatchK8sPod(t *testing.T) {
	s := labels.SelectorFromValidatedSet(labels.Set{
		"app": "pip",
	})

	p := new(core.Pod)

	p.Status.PodIP = "127.0.0.1"
	p.Labels = map[string]string{
		"app": "pip",
	}
	assert.True(t, matchK8sPod(p, s))

	p.Status.PodIP = ""
	p.Labels = map[string]string{
		"app": "pip",
	}
	assert.False(t, matchK8sPod(p, s))

	p.Status.PodIP = "127.0.0.1"
	p.Labels = map[string]string{
		"app": "other",
	}
	assert.False(t, matchK8sPod(p, s))
}
