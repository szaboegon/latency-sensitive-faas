import React, { useMemo } from "react";
import ReactFlow, {
  MiniMap,
  Controls,
  Background,
  type Node,
  type Edge,
  Position,
  MarkerType,
} from "reactflow";
import "reactflow/dist/style.css";
import dagre from "dagre";
import type { Deployment } from "../models/models";
import { generateComponentColor } from "../helpers/utilities";

interface CallGraphViewProps {
  deployments: Deployment[];
}

/**
 * Layout constants
 */
const nodeWidth = 240;
const nodeHeight = 50;
const COMPONENT_GAP_Y = 50;
const COMPOSITION_PADDING = 20;
const TITLE_HEIGHT = 36;
const TITLE_TO_COMPONENT_GAP = 100;
const DEPLOYMENT_LABEL_WIDTH = nodeWidth * 1.5;
const DEPLOYMENT_LABEL_HEIGHT = TITLE_HEIGHT*2;

const layoutGraph = (deployments: Deployment[]) => {
  const g = new dagre.graphlib.Graph();
  g.setDefaultEdgeLabel(() => ({}));
  g.setGraph({ rankdir: "LR", nodesep: 150, ranksep: 200 }); // deployments laid out left-to-right

  // Add deployment "container" nodes
  deployments.forEach((deployment) => {
    g.setNode(deployment.id ?? "", {
      width: nodeWidth * 2,
      height: nodeHeight * 2,
    });
  });

  // Add edges between deployments (based on cross-deployment routes)
  deployments.forEach((deployment) => {
    const depId = deployment.id ?? "";
    Object.entries(deployment.routingTable ?? {}).forEach(([_, routes]) => {
      (routes || []).forEach((route) => {
        const targetDepId = route.function === "local" ? depId : route.function;
        if (targetDepId && targetDepId !== depId) {
          g.setEdge(depId, targetDepId);
        }
      });
    });
  });

  dagre.layout(g);

  // Map deployment positions
  const deploymentPositions: Record<string, { x: number; y: number }> = {};
  deployments.forEach((deployment) => {
    const depId = deployment.id ?? "";
    const pos = g.node(depId);
    deploymentPositions[depId] = { x: pos.x, y: pos.y };
  });

  return deploymentPositions;
};

const CallGraphView: React.FC<CallGraphViewProps> = ({ deployments }) => {
  const { nodes, edges } = useMemo(() => {
    const rfNodes: Node[] = [];
    const rfEdges: Edge[] = [];

    // 1. Layout deployments
    const deploymentPositions = layoutGraph(deployments);

    // 2. Build nodes
    deployments.forEach((deployment) => {
      const depId = deployment.id ?? "";
      const components = Object.keys(deployment.routingTable ?? {});

      // Calculate deployment height based on number of components
      const depHeight =
        TITLE_HEIGHT +
        TITLE_TO_COMPONENT_GAP +
        components.length * (nodeHeight + COMPONENT_GAP_Y) +
        COMPOSITION_PADDING * 2;

      // Deployment group node
      const depPos = deploymentPositions[depId];
      rfNodes.push({
        id: depId,
        type: "group",
        position: { x: depPos.x - nodeWidth, y: depPos.y - depHeight / 2 },
        data: {},
        style: {
          border: "2px solid rgba(0,0,0,0.6)",
          borderRadius: 8,
          background: "transparent",
          boxSizing: "border-box",
          width: nodeWidth * 2,
          height: depHeight,
          padding: COMPOSITION_PADDING,
        },
      });

      // Title bar
      rfNodes.push({
        id: `${depId}-title`,
        data: { label: `${deployment.id} (${deployment.node})` },
        parentNode: depId,
        extent: "parent",
        position: { x: COMPOSITION_PADDING, y: 6 },
        style: {
          width: DEPLOYMENT_LABEL_WIDTH,
          height: DEPLOYMENT_LABEL_HEIGHT,
          background: "#e9e9e9",
          borderRadius: 6,
          padding: "6px",
          fontWeight: 700,
          display: "flex",
          alignItems: "center",
        },
      });

      // Components inside deployment
      components.forEach((componentName, idx) => {
        const localX = COMPOSITION_PADDING;
        const localY =
          TITLE_HEIGHT +
          TITLE_TO_COMPONENT_GAP +
          idx * (nodeHeight + COMPONENT_GAP_Y);

        rfNodes.push({
          id: `${depId}-${componentName}`,
          data: { label: componentName },
          parentNode: depId,
          extent: "parent",
          position: { x: localX, y: localY },
          sourcePosition: Position.Right,
          targetPosition: Position.Left,
          style: {
            width: nodeWidth,
            height: nodeHeight,
            backgroundColor: generateComponentColor(componentName),
            borderRadius: 6,
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            color: "#fff",
            fontWeight: 600,
          },
        });
      });
    });

    // 3. Build edges
    deployments.forEach((deployment) => {
      const depId = deployment.id ?? "";
      Object.entries(deployment.routingTable ?? {}).forEach(([sourceComponent, routes]) => {
        (routes || []).forEach((route, idx) => {
          const targetDepId = route.function === "local" ? depId : route.function;
          const targetComponent = route.to;
          if (!targetDepId || !targetComponent) return;

          const sourceId = `${depId}-${sourceComponent}`;
          const targetId = `${targetDepId}-${targetComponent}`;
          const isLocal = targetDepId === depId;

          rfEdges.push({
            id: `e-${sourceId}-to-${targetId}-${idx}`,
            source: sourceId,
            target: targetId,
            type: isLocal ? "step" : "smoothstep",
            animated: true,
            style: { stroke: "#4a5568", strokeWidth: 2 },
            markerEnd: {
              type: MarkerType.ArrowClosed,
              color: "#4a5568",
            },
          });
        });
      });
    });

    return { nodes: rfNodes, edges: rfEdges };
  }, [deployments]);

  return (
    <div style={{ width: "100%", height: "80vh" }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        fitView
        nodesDraggable
        nodesConnectable={false}
      >
        <MiniMap />
        <Controls />
        <Background gap={16} />
      </ReactFlow>
    </div>
  );
};

export default CallGraphView;
