package client

import (
	"errors"
	"strings"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func isK8sPodReady(pod *core.Pod) bool {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == core.PodReady && cond.Status == core.ConditionTrue {
			return true
		}
	}

	return false
}

func makeInClusterK8sClient() (kubernetes.Interface, error) {
	conf, err := defK8sConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(conf)
}

var (
	errK8sNameTooShort = errors.New("name of kuberenets pod is too short")
	errK8sNameInvalid  = errors.New("name of kuberenets pod isn't valid")
)

type k8sName struct {
	namespace string
	selector  labels.Selector
}

func makeK8sName(s string) (k8sName, error) {
	out := k8sName{}

	ss := strings.Split(s, ".")
	if len(ss) < 3 {
		return out, errK8sNameTooShort
	}
	if len(ss)%2 == 0 {
		return out, errK8sNameInvalid
	}

	out.namespace = ss[len(ss)-1]
	ss = ss[:len(ss)-1]

	set := make(labels.Set)
	for i := 0; i < len(ss); i += 2 {
		set[ss[i+1]] = ss[i]
	}

	out.selector = labels.SelectorFromValidatedSet(set)

	return out, nil
}

func matchK8sPod(pod *core.Pod, s labels.Selector) bool {
	return len(pod.Status.PodIP) > 0 && s.Matches(labels.Set(pod.Labels))
}
