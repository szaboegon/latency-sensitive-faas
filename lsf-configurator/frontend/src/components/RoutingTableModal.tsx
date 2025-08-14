import React from "react";
import {
  Modal,
  Box,
  Typography,
} from "@mui/material";
import type { FunctionComposition, RoutingTable } from "../models/models";
import { useModifyRoutingTable } from "../hooks/functionCompositionHooks";
import RoutingTableForm from "./RoutingTableForm";

interface Props {
  open: boolean;
  onClose: () => void;
  composition: FunctionComposition;
  allCompositions: FunctionComposition[];
}

interface Rule {
  component: string;
  targetComposition: string;
  targetComponent: string;
}

const RoutingTableModal: React.FC<Props> = ({
  open,
  onClose,
  composition,
  allCompositions,
}) => {
  const { mutate: modifyRoutingTable } = useModifyRoutingTable();

  const handleSaveFormInput = (rules: Rule[]) => {
    const routingTable: RoutingTable = rules.reduce((acc, rule) => {
      if (!acc[rule.component]) {
        acc[rule.component] = [];
      }
      acc[rule.component].push({
        to: rule.targetComponent === "None" ? "" : rule.targetComponent,
        function: rule.targetComposition === "None" ? "" : rule.targetComposition,
      });
      return acc;
    }, {} as RoutingTable);

    modifyRoutingTable({
      functionCompositionId: composition.id!,
      routingTable,
    });
    onClose();
  };

  return (
    <Modal open={open} onClose={onClose}>
      <Box
        sx={{
          position: "absolute",
          top: "50%",
          left: "50%",
          transform: "translate(-50%, -50%)",
          width: 900,
          bgcolor: "background.paper",
          boxShadow: 24,
          p: 4,
          borderRadius: 2,
        }}
      >
        <Typography variant="h6" mb={2}>
          Edit Routing Table
        </Typography>
          <Box mt={2}>
            <RoutingTableForm
              composition={composition}
              allCompositions={allCompositions}
              onSave={handleSaveFormInput}
            />
          </Box>
      </Box>
    </Modal>
  );
};

export default RoutingTableModal;