package paneler

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/charm/models/panels"
	"github.com/everettraven/buoy/pkg/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

var _ Paneler = &Log{}

type Log struct {
	KubeClient *kubernetes.Clientset
}

func (t *Log) Model(panel types.Panel) (tea.Model, error) {
	log := types.Logs{}
	err := json.Unmarshal(panel.Blob, &log)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling panel to table type: %s", err)
	}
	logItem := modelWrapperForLogPanel(t.KubeClient, log)
	go streamLogs(t.KubeClient, log, logItem) //nolint: errcheck
	return logItem, nil
}

func modelWrapperForLogPanel(kc *kubernetes.Clientset, logsPanel types.Logs) *panels.Logs {
	vp := viewport.New(100, 20)
	vpw := panels.NewLogs(logsPanel.Name, vp)
	return vpw
}

func streamLogs(kc *kubernetes.Clientset, logsPanel types.Logs, logItem *panels.Logs) error {
	//TODO: expand this beyond just a pod
	req := kc.CoreV1().Pods(logsPanel.Key.Namespace).GetLogs(logsPanel.Key.Name, &v1.PodLogOptions{
		Container: logsPanel.Container,
		Follow:    true,
	})

	rc, err := req.Stream(context.Background())
	if err != nil {
		return fmt.Errorf("fetching logs for %s/%s: %w", logsPanel.Key.Namespace, logsPanel.Key.Name, err)
	}
	defer rc.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(rc)
		for scanner.Scan() {
			logs := wrapLogs(scanner.Bytes())
			logItem.AddContent(logs)
		}
	}()

	wg.Wait()
	return nil
}

func wrapLogs(logs []byte) string {
	logStr := string(logs)
	splitLogs := strings.Split(logStr, "\n")
	var logsBuilder strings.Builder
	for _, log := range splitLogs {
		if len(log) > 100 {
			segs := (len(log) / 100)
			for seg := 0; seg < segs; seg++ {
				logsBuilder.WriteString(log[:100])
				logsBuilder.WriteString("\n")
				log = log[100:]
			}
			//write any leftovers
			logsBuilder.WriteString(log)
		} else {
			logsBuilder.WriteString(log)
		}
		logsBuilder.WriteString("\n")
	}
	return logsBuilder.String()
}
