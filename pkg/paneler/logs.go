package paneler

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/charm/models/panels"
	"github.com/everettraven/buoy/pkg/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

var _ Paneler = &Log{}

type Log struct {
	typedClient     *kubernetes.Clientset
	dynamicClient   dynamic.Interface
	discoveryClient *discovery.DiscoveryClient
	restMapper      meta.RESTMapper
}

func NewLog(typedClient *kubernetes.Clientset, dynamicClient dynamic.Interface, discoveryClient *discovery.DiscoveryClient, restMapper meta.RESTMapper) *Log {
	return &Log{
		typedClient:     typedClient,
		dynamicClient:   dynamicClient,
		discoveryClient: discoveryClient,
		restMapper:      restMapper,
	}
}

func (t *Log) Model(panel types.Panel) (tea.Model, error) {
	log := types.Logs{}
	err := json.Unmarshal(panel.Blob, &log)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling panel to table type: %s", err)
	}
	logItem := modelWrapperForLogPanel(log)
	pod, err := t.getPodForObject(log)
	if err != nil {
		return nil, fmt.Errorf("error getting pod for object: %w", err)
	}
	go streamLogs(t.typedClient, pod, logItem, log.Container) //nolint: errcheck
	return logItem, nil
}

func modelWrapperForLogPanel(logsPanel types.Logs) *panels.Logs {
	vp := viewport.New(100, 20)
	vpw := panels.NewLogs(logsPanel.Name, vp)
	return vpw
}

func (t *Log) getPodForObject(logsPanel types.Logs) (*v1.Pod, error) {
	gvk := schema.GroupVersionKind{
		Group:   logsPanel.Group,
		Version: logsPanel.Version,
		Kind:    logsPanel.Kind,
	}

	if gvk == v1.SchemeGroupVersion.WithKind("Pod") {
		return t.typedClient.CoreV1().Pods(logsPanel.Key.Namespace).Get(context.Background(), logsPanel.Key.Name, metav1.GetOptions{})
	}

	mapping, err := t.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("error creating resource mapping: %w", err)
	}
	u, err := t.dynamicClient.Resource(mapping.Resource).Namespace(logsPanel.Key.Namespace).Get(context.Background(), logsPanel.Key.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting object: %w", err)
	}

	selector, err := getPodSelectorForUnstructured(u)
	if err != nil {
		return nil, fmt.Errorf("error getting pod selector for object: %w", err)
	}
	pods, err := t.typedClient.CoreV1().Pods(u.GetNamespace()).List(context.Background(), metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, fmt.Errorf("error getting pods for object: %w", err)
	}
	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods found for object")
	}
	return &pods.Items[0], nil
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

func streamLogs(kc *kubernetes.Clientset, pod *v1.Pod, logItem *panels.Logs, container string) {
	req := kc.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &v1.PodLogOptions{
		Container: container,
		Follow:    true,
	})

	rc, err := req.Stream(context.Background())
	if err != nil {
		logItem.AddContent(fmt.Errorf("fetching logs for %s/%s: %w", pod.Namespace, pod.Name, err).Error())
		return
	}
	defer rc.Close()

	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		logItem.AddContent(scanner.Text())
	}
}
