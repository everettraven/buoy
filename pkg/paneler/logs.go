package paneler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/everettraven/buoy/pkg/charm/models"
	"github.com/everettraven/buoy/pkg/charm/styles"
	"github.com/everettraven/buoy/pkg/types"
	"github.com/treilik/bubbleboxer"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

var _ Paneler = &Log{}

type Log struct {
	KubeClient *kubernetes.Clientset
}

func (t *Log) Node(panel types.Panel, bxr *bubbleboxer.Boxer) (bubbleboxer.Node, error) {
	log := types.Logs{}
	err := json.Unmarshal(panel.Blob, &log)
	if err != nil {
		return bubbleboxer.Node{}, fmt.Errorf("unmarshalling panel to table type: %s", err)
	}
	mw, err := modelWrapperForLogPanel(t.KubeClient, log)
	if err != nil {
		return bubbleboxer.Node{}, fmt.Errorf("getting table widget: %s", err)
	}
	return nodeForModelWrapper(log.Name, mw, bxr)
}

func modelWrapperForLogPanel(kc *kubernetes.Clientset, logsPanel types.Logs) (*models.ModelWrapper, error) {
	//TODO: expand this beyond just a pod
	req := kc.CoreV1().Pods(logsPanel.Key.Namespace).GetLogs(logsPanel.Key.Name, &v1.PodLogOptions{})
	rc, err := req.Stream(context.Background())
	if err != nil {
		return nil, fmt.Errorf("fetching logs for %s/%s: %w", logsPanel.Key.Namespace, logsPanel.Key.Name, err)
	}
	defer rc.Close()

	logs, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("reading logs from stream: %w", err)
	}

	// TODO: Sort out word wrapping
	vp := viewport.New(100, 8)

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

	vp.SetContent(logsBuilder.String())

	vpw := &models.ModelWrapper{
		Model:   vp,
		UpdateF: models.ViewportUpdateFunc,
		HeightF: models.ViewportHeightFunc,
	}
	vpw.SetStyle(styles.FocusedModelStyle)

	return vpw, nil
}
