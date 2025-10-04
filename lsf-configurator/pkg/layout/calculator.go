package layout

import (
	"bytes"
	"encoding/json"
	"fmt"
	"lsf-configurator/pkg/core"
	"os/exec"
)

type slambucCalculator struct {
	pythonCmd     string
	script        string
	platformNodes []string
	platformDelay int
}

func NewLayoutCalculator(pythonCmd, script string, platformNodes []string, platformDelay int) core.LayoutCalculator {
	return &slambucCalculator{
		pythonCmd:     pythonCmd,
		script:        script,
		platformNodes: platformNodes,
		platformDelay: platformDelay,
	}
}

func (c *slambucCalculator) CalculateLayout(profiles []core.ComponentProfile, links []core.ComponentLink, appLatencyReq, memoryAvailable int) (core.Layout, error) {
	idMap := make(map[string]int)
	profileMap := make(map[int]core.ComponentProfile)
	nodes := []map[string]interface{}{}

	for i, p := range profiles {
		id := i + 1
		idMap[p.Name] = id
		profileMap[id] = p
		nodes = append(nodes, map[string]interface{}{
			"id":      id,
			"mem":     p.TotalMemory(),
			"runtime": p.Runtime,
		})
	}

	edges := []map[string]interface{}{}
	for _, l := range links {
		fromId, ok := idMap[l.From]
		if !ok {
			continue
		}
		toId, ok := idMap[l.To]
		if !ok {
			continue
		}
		edges = append(edges, map[string]interface{}{
			"from": fromId,
			"to":   toId,
			"attr": map[string]interface{}{
				"rate": l.InvocationRate,
				"data": l.DataDelay,
			},
		})
	}
	input := map[string]interface{}{
		"params": map[string]interface{}{
			"root":   1,
			"M":      memoryAvailable,
			"L":      appLatencyReq,
			"cp_end": len(profiles),
			"delay":  c.platformDelay,
		},
		"nodes": nodes,
		"edges": edges,
	}

	jsonInput, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	// Run Python script
	cmd := exec.Command(c.pythonCmd, c.script)
	cmd.Stdin = bytes.NewReader(jsonInput)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("python script failed: %v, stderr: %s", err, stderr.String())
	}

	// Parse JSON output
	var pyOutput struct {
		Layout  [][]int `json:"layout"`
		OptCost float64 `json:"opt_cost"`
		Latency int     `json:"latency"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &pyOutput); err != nil {
		return nil, fmt.Errorf("failed to parse python output: %w, stdout: %s", err, stdout.String())
	}

	if len(pyOutput.Layout) > len(c.platformNodes) {
		return nil, fmt.Errorf("insufficient memory: layout has more groups (%d) than platform nodes (%d)", len(pyOutput.Layout), len(c.platformNodes))
	}

	layout := make(core.Layout)
	for i, group := range pyOutput.Layout {
		var groupProfiles []core.ComponentProfile
		for _, id := range group {
			if prof, ok := profileMap[id]; ok {
				groupProfiles = append(groupProfiles, prof)
			}
		}
		layout[c.platformNodes[i]] = groupProfiles
	}

	return layout, nil
}
