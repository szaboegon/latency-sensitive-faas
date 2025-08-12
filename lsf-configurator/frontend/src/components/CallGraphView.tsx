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
import type { FunctionComposition } from "../models/models";
import { generateComponentColor } from "../helpers/utilities";

interface CallGraphViewProps {
  compositions: FunctionComposition[];
}

/**
 * Layout constants - tweak to taste
 */
const COMPOSITION_WIDTH = 320;
const TITLE_HEIGHT = 36;
const COMPONENT_HEIGHT = 40;
const COMPONENT_GAP = 12;
const COMPOSITION_PADDING = 12;

const CallGraphView: React.FC<CallGraphViewProps> = ({ compositions }) => {
  const { nodes, edges } = useMemo(() => {
    const rfNodes: Node[] = [];
    const rfEdges: Edge[] = [];

    // Directly render compositions without dagre layout
    compositions.forEach((c, compIndex) => {
      const nComponents = Object.keys(c.components ?? {}).length;
      const height =
        TITLE_HEIGHT +
        COMPOSITION_PADDING * 2 +
        nComponents * (COMPONENT_HEIGHT + COMPONENT_GAP);

      const compId = c.id ?? "";
      const topLeftX = compIndex * (COMPOSITION_WIDTH + 120); // simple horizontal spacing
      const topLeftY = 0;

      // Composition group node
      rfNodes.push({
        id: compId,
        type: "group",
        position: { x: topLeftX, y: topLeftY },
        data: {},
        style: {
          width: COMPOSITION_WIDTH,
          height,
          border: "2px solid rgba(0,0,0,0.6)",
          borderRadius: 8,
          background: "transparent",
          boxSizing: "border-box",
          overflow: "visible",
        },
      });

      // Title bar node inside the group
      rfNodes.push({
        id: `${compId}-title`,
        data: { label: `${c.id} (${c.node})` },
        parentNode: compId,
        extent: "parent",
        position: { x: COMPOSITION_PADDING, y: 6 },
        style: {
          width: COMPOSITION_WIDTH - COMPOSITION_PADDING * 2,
          height: TITLE_HEIGHT - 6,
          background: "#e9e9e9",
          borderRadius: 6,
          padding: "6px",
          fontWeight: 700,
          display: "flex",
          alignItems: "center",
        },
      });

      // Component nodes inside the composition
      const componentNames = Object.keys(c.components ?? {});
      componentNames.forEach((componentName, idx) => {
        const localX = COMPOSITION_PADDING;
        const localY =
          TITLE_HEIGHT + COMPOSITION_PADDING + idx * (COMPONENT_HEIGHT + COMPONENT_GAP);

        rfNodes.push({
          id: `${compId}-${componentName}`,
          data: { label: componentName },
          parentNode: compId,
          extent: "parent",
          position: { x: localX, y: localY },
          sourcePosition: Position.Right,
          targetPosition: Position.Left,
          style: {
            width: COMPOSITION_WIDTH - COMPOSITION_PADDING * 2,
            height: COMPONENT_HEIGHT,
            backgroundColor: generateComponentColor(componentName),
            borderRadius: 6,
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            color: "#fff",
            fontWeight: 600,
            boxSizing: "border-box",
            padding: "0 8px",
          },
        });
      });
    });

    // Build edges
    compositions.forEach((c) => {
      const compId = c.id ?? "";
      Object.entries(c.components ?? {}).forEach(([sourceComponent, routes]) => {
        (routes || []).forEach((r: any, idx: number) => {
          const targetComp = r.function;
          const targetComponent = r.to;
          if (!targetComp || !targetComponent) return;

          const sourceId = `${compId}-${sourceComponent}`;
          const targetId = `${targetComp}-${targetComponent}`;

          rfEdges.push({
            id: `e-${sourceId}-to-${targetId}-${idx}`,
            source: sourceId,
            target: targetId,
            type: "smoothstep",
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
  }, [compositions]);

  return (
    <div style={{ width: "100%", height: "80vh" }}>
      <ReactFlow nodes={nodes} edges={edges} fitView>
        <MiniMap />
        <Controls />
        <Background gap={16} />
      </ReactFlow>
    </div>
  );
};

export default CallGraphView;
