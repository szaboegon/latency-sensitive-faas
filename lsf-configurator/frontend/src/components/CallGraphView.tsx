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

const COMPOSITION_HEIGHT = 300;
const TITLE_HEIGHT = 36;
const COMPONENT_WIDTH = 160;
const COMPONENT_HEIGHT = 40;
const COMPONENT_GAP_Y = 40; // vertical spacing between components
const COMPOSITION_PADDING = 40;

const getLayoutedElements = (deployments: Deployment[], edges: Edge[]) => {
  const g = new dagre.graphlib.Graph();
  g.setDefaultEdgeLabel(() => ({}));
  g.setGraph({ rankdir: "LR", nodesep: 200, ranksep: 200 });

  deployments.forEach((deployment) => {
    g.setNode(deployment.id ?? "", { width: 300, height: COMPOSITION_HEIGHT });
  });

  edges.forEach((edge) => {
    const sourceDep = edge.source.split("-")[0];
    const targetDep = edge.target.split("-")[0];
    if (sourceDep && targetDep && sourceDep !== targetDep) {
      g.setEdge(sourceDep, targetDep);
    }
  });

  dagre.layout(g);

  const rfNodes: Node[] = [];

  deployments.forEach((deployment) => {
    const depId = deployment.id ?? "";
    const depBox = g.node(depId);

    const componentNames = Object.keys(deployment.routingTable ?? {});
    const width = COMPONENT_WIDTH + COMPOSITION_PADDING * 2;
    const height =
      TITLE_HEIGHT +
      COMPOSITION_PADDING +
      componentNames.length * (COMPONENT_HEIGHT + COMPONENT_GAP_Y) +
      COMPOSITION_PADDING;

    // Group node (deployment box)
    rfNodes.push({
      id: depId,
      type: "group",
      position: { x: depBox.x - width / 2, y: depBox.y - height / 2 },
      data: {},
      style: {
        width,
        height,
        border: "2px solid rgba(0,0,0,0.6)",
        borderRadius: 8,
        background: "transparent",
      },
    });

    // Title bar
    rfNodes.push({
      id: `${depId}-title`,
      data: { label: `${deployment.id} (${deployment.node})` },
      parentNode: depId,
      extent: "parent",
      position: { x: COMPOSITION_PADDING, y: 10 },
      style: {
        width: width - COMPOSITION_PADDING * 2,
        height: TITLE_HEIGHT - 6,
        background: "#e9e9e9",
        borderRadius: 6,
        padding: "6px",
        fontWeight: 700,
        display: "flex",
        alignItems: "center",
      },
    });

    componentNames.forEach((componentName, idx) => {
      const localX = COMPOSITION_PADDING;
      const localY =
        TITLE_HEIGHT + COMPOSITION_PADDING + idx * (COMPONENT_HEIGHT + COMPONENT_GAP_Y);

      rfNodes.push({
        id: `${depId}-${componentName}`,
        data: { label: componentName },
        parentNode: depId,
        extent: "parent",
        position: { x: localX, y: localY },
        sourcePosition: Position.Bottom,
        targetPosition: Position.Top,
        style: {
          width: COMPONENT_WIDTH,
          height: COMPONENT_HEIGHT,
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

  return { nodes: rfNodes, edges };
};

const CallGraphView: React.FC<CallGraphViewProps> = ({ deployments }) => {
  const { nodes, edges } = useMemo(() => {
    const rfEdges: Edge[] = [];

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
            type: "smoothstep",
            animated: true,
            sourceHandle: undefined,
            targetHandle: undefined,
            style: { stroke: isLocal ? "#2b6cb0" : "#4a5568", strokeWidth: 2 },
            markerEnd: {
              type: MarkerType.ArrowClosed,
              color: isLocal ? "#2b6cb0" : "#4a5568",
            },
          });
        });
      });
    });

    return getLayoutedElements(deployments, rfEdges);
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
