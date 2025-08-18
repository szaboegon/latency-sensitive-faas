import React, { useState } from "react";
import {
  Modal,
  Box,
  Typography,
  Button,
} from "@mui/material";
import type { Deployment, RoutingTable } from "../models/models";
import RoutingTableEditor from "./RoutingTableEditor";
import { useModifyRoutingTable } from "../hooks/deploymentHooks";

interface Props {
  open: boolean;
  onClose: () => void;
  deployment: Deployment;
  allDeployments: Deployment[];
}

const RoutingTableModal: React.FC<Props> = ({
  open,
  onClose,
  deployment,
  allDeployments,
}) => {
  const { mutate: modifyRoutingTable } = useModifyRoutingTable();
  const [routingTable, setRoutingTable] = useState<RoutingTable>(deployment.routingTable);

  const handleSaveFormInput = () => {
    modifyRoutingTable({
      deploymentId: deployment.id!,
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
            deployment={deployment}
            allDeployments={allDeployments}
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