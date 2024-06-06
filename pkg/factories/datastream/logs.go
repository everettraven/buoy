package datastream

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

var _ Datastream = &logDatastream{}

type logDatastream struct {
	logReadCloser io.ReadCloser
	contentAdder  ContentAdder
}

func (l *logDatastream) Run(stopCh <-chan struct{}) {
	go streamLogs(l.logReadCloser, l.contentAdder)
}

type Log interface {
	Key() types.NamespacedName
	GVK() schema.GroupVersionKind
	Container() string
	ContentAdder
}

func LogsDatastreamFunc(typedClient *kubernetes.Clientset, dynamicClient *dynamic.DynamicClient, restMapper meta.RESTMapper) DatastreamFactoryFunc {
	return func(obj interface{}) (Datastream, error) {
		log, ok := obj.(Log)
		if !ok {
			return nil, &InvalidPanelType{fmt.Errorf("object does not implement Log interface. Unable to determine how to fetch logs")}
		}

		if log.GVK() == v1.SchemeGroupVersion.WithKind("Pod") {
			pod, err := typedClient.CoreV1().Pods(log.Key().Namespace).Get(context.Background(), log.Key().Name, metav1.GetOptions{})
			if err != nil {
				return nil, fmt.Errorf("error getting pod: %w", err)
			}
			rc, err := logsForPod(typedClient, pod, log.Container())
			if err != nil {
				return nil, fmt.Errorf("error getting logs for pod: %w", err)
			}
			return &logDatastream{
				logReadCloser: rc,
				contentAdder:  log,
			}, nil
		}

		mapping, err := restMapper.RESTMapping(log.GVK().GroupKind(), log.GVK().Version)
		if err != nil {
			return nil, fmt.Errorf("error creating resource mapping: %w", err)
		}
		u, err := dynamicClient.Resource(mapping.Resource).Namespace(log.Key().Namespace).Get(context.Background(), log.Key().Name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("error getting object: %w", err)
		}

		selector, err := getPodSelectorForUnstructured(u)
		if err != nil {
			return nil, fmt.Errorf("error getting pod selector for object: %w", err)
		}
		pods, err := typedClient.CoreV1().Pods(u.GetNamespace()).List(context.Background(), metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			return nil, fmt.Errorf("error getting pods for object: %w", err)
		}
		if len(pods.Items) == 0 {
			return nil, fmt.Errorf("no pods found for object")
		}
		pod := &pods.Items[0]
		rc, err := logsForPod(typedClient, pod, log.Container())
		if err != nil {
			return nil, fmt.Errorf("error getting logs for pod: %w", err)
		}
		return &logDatastream{
			logReadCloser: rc,
			contentAdder:  log,
		}, nil
	}
}

func getPodSelectorForUnstructured(u *unstructured.Unstructured) (labels.Selector, error) {
	selector, found, err := unstructured.NestedFieldCopy(u.Object, "spec", "selector")
	if !found {
		return nil, fmt.Errorf("no pod label selector found in object spec: %s", u.Object)
	}
	if err != nil {
		return nil, fmt.Errorf("error getting pod label selector from object spec: %w", err)
	}
	sel := &metav1.LabelSelector{}
	bytes, err := json.Marshal(selector)
	if err != nil {
		return nil, fmt.Errorf("error marshalling selector: %w", err)
	}
	err = json.Unmarshal(bytes, sel)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling selector: %w", err)
	}
	return metav1.LabelSelectorAsSelector(sel)
}

type ContentAdder interface {
	AddContent(string)
}

func streamLogs(rc io.ReadCloser, ca ContentAdder) {
	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		ca.AddContent(scanner.Text())
	}
}

func logsForPod(kc *kubernetes.Clientset, pod *v1.Pod, container string) (io.ReadCloser, error) {
	req := kc.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &v1.PodLogOptions{
		Container: container,
		Follow:    true,
	})

	rc, err := req.Stream(context.Background())
	if err != nil {
		return nil, fmt.Errorf("fetching logs for %s/%s: %w", pod.Namespace, pod.Name, err)
	}
	return rc, nil
}
