import React, { useState } from "react";
import {
  Modal,
  Box,
  Typography,
  Button,
} from "@mui/material";
import type { FunctionComposition, RoutingTable } from "../models/models";
import { useModifyRoutingTable } from "../hooks/functionCompositionHooks";
import RoutingTableEditor from "./RoutingTableEditor";

interface Props {
  open: boolean;
  onClose: () => void;
  composition: FunctionComposition;
  allCompositions: FunctionComposition[];
}

const RoutingTableModal: React.FC<Props> = ({
  open,
  onClose,
  composition,
  allCompositions,
}) => {
  const { mutate: modifyRoutingTable } = useModifyRoutingTable();
  const [routingTable, setRoutingTable] = useState<RoutingTable>(composition.components);

  const handleSaveFormInput = () => {
    modifyRoutingTable({
      functionCompositionId: composition.id!,
      routingTable,
    });
    onClose();
  };

  const onRoutingTableChange = (newRoutingTable: RoutingTable) => {
    setRoutingTable(newRoutingTable);
  }

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
          <RoutingTableEditor
            composition={composition}
            allCompositions={allCompositions}
            onChange={onRoutingTableChange}
          />
          <Button
            variant="contained"
            sx={{ mt: 2, width: "100%" }}
            onClick={() => handleSaveFormInput()}
          >
            Save
          </Button>
        </Box>
      </Box>
    </Modal>
  );
};

export default RoutingTableModal;