import React from "react";
import {
  Modal,
  Box,
  Typography,
  TextField,
  Button,
  Stack,
} from "@mui/material";
import RoutingTableEditor from "./RoutingTableEditor";
import type { Component, Deployment, RoutingTable } from "../models/models";
import { useForm, Controller } from "react-hook-form";
import { useCreateDeployment } from "../hooks/deploymentHooks";

interface Props {
  open: boolean;
  onClose: () => void;
  compositionId: string;
  components: Component[];
  allDeployments: Deployment[];
}

const CreateDeploymentModal: React.FC<Props> = ({
  open,
  onClose,
  compositionId,
  components,
  allDeployments,
}) => {
  const { control, handleSubmit, setValue } = useForm<{
    node: string;
    namespace: string;
    routingTable: RoutingTable;
  }>({
    defaultValues: {
      node: "",
      namespace: "",
      routingTable: components.reduce((acc, comp) => {
        acc[comp] = [];
        return acc;
      }, {} as RoutingTable),
    },
  });

  const { mutate: createDeployment } = useCreateDeployment();

  const handleAddDeployment = (data: {
    node: string;
    namespace: string;
    routingTable: RoutingTable;
  }) => {
    const newDeployment: Deployment = {
      functionCompositionId: compositionId,
      node: data.node,
      namespace: data.namespace,
      routingTable: data.routingTable,
    };
    createDeployment(newDeployment);
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
        <Typography variant="h6" gutterBottom>
          Add New Deployment
        </Typography>
        <form onSubmit={handleSubmit(handleAddDeployment)}>
          <Stack spacing={2}>
            <Controller
              name="node"
              control={control}
              render={({ field }) => (
                <TextField {...field} label="Node" fullWidth />
              )}
            />
            <Controller
              name="namespace"
              control={control}
              render={({ field }) => (
                <TextField {...field} label="Namespace" fullWidth />
              )}
            />
            <Typography variant="body1" mt={2}>
              Routing Table
            </Typography>
            <RoutingTableEditor
              deployment={{
                id: "",
                functionCompositionId: compositionId,
                node: "",
                namespace: "",
                routingTable: components.reduce((acc, comp) => {
                  acc[comp] = [];
                  return acc;
                }, {} as RoutingTable),
              }}
              allDeployments={allDeployments}
              onChange={(data) => setValue("routingTable", data)}
            />
            <Stack direction="row" spacing={2}>
              <Button variant="contained" color="primary" type="submit">
                Add
              </Button>
              <Button variant="outlined" color="error" onClick={onClose}>
                Cancel
              </Button>
            </Stack>
          </Stack>
        </form>
      </Box>
    </Modal>
  );
};

export default CreateDeploymentModal;
