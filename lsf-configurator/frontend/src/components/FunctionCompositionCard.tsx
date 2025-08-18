import React, { useState } from "react";
import { Card, CardContent, CardHeader, Typography, Divider, Box, Chip, Stack, Button } from "@mui/material";
import type { FunctionComposition, Deployment } from "../models/models"; 
import DeploymentDetailsDrawer from "../components/DeploymentDetailsDrawer"; 
import CreateDeploymentModal from "./CreateDeploymentModal"; 
import { generateComponentColor } from "../helpers/utilities";

const StatusColorMap: Record<"pending" | "built" | "deployed" | "error", string> = {
    pending: "#ff9800",
    built: "#4caf50",   
    deployed: "#2196f3", 
    error: "#f44336"    
};

interface Props {
  composition: FunctionComposition;
  allDeployments: Deployment[];
  onDelete: (id: string) => void;
}

const FunctionCompositionCard: React.FC<Props> = ({ composition, allDeployments, onDelete }) => {
  const [selectedDeployment, setSelectedDeployment] = useState<Deployment | null>(null);
  const [isAddDeploymentModalOpen, setAddDeploymentModalOpen] = useState(false);

  const handleDelete = () => {
    if (composition.id) {
      onDelete(composition.id);
    }
  };

  const handleOpenDrawer = (deployment: Deployment) => setSelectedDeployment(deployment);
  const handleCloseDrawer = () => setSelectedDeployment(null);

  const handleOpenAddDeploymentModal = () => setAddDeploymentModalOpen(true);
  const handleCloseAddDeploymentModal = () => setAddDeploymentModalOpen(false);

  return (
    <>
      <Card
        sx={{
          borderRadius: 3,
          boxShadow: 4,
          background: "linear-gradient(135deg, #e0f7fa 0%, #fff 100%)",
          border: "1px solid #b2ebf2",
        }}
      >
        <CardHeader
          title={<Typography variant="h6">{composition.id || "Unnamed Composition"}</Typography>}
        />

        <CardContent>
          {/* Status Indicator */}
          <Box mb={2}>
            <Typography variant="subtitle2" gutterBottom>
              Status:
            </Typography>
            <Box
              sx={{
                display: "inline-block",
                padding: "4px 8px",
                borderRadius: 2,
                backgroundColor: StatusColorMap[composition.status],
                color: "#fff",
                fontWeight: "bold",
                textAlign: "center",
              }}
            >
              {composition.status.charAt(0).toUpperCase() + composition.status.slice(1)}
            </Box>
          </Box>

          {/* Components Visualization */}
          {composition.components && (
            <Box mb={2}>
              <Typography variant="subtitle2" gutterBottom>
                Components:
              </Typography>
              <Stack direction="row" spacing={1} flexWrap="wrap">
                {composition.components.map((component) => (
                  <Box
                    key={component}
                    sx={{
                      backgroundColor: generateComponentColor(component),
                      borderRadius: 2,
                      padding: 1,
                      textAlign: "center",
                      boxShadow: 1,
                    }}
                  >
                    <Typography variant="body2">{component}</Typography>
                  </Box>
                ))}
              </Stack>
            </Box>
          )}

          {/* Build info */}
          {composition.build && (
            <Box mt={1} mb={2}>
              <Typography variant="subtitle2" color="textSecondary">
                Last Build:
              </Typography>
              <Typography variant="body2">
                Image: {composition.build.image || "N/A"}
              </Typography>
              <Typography variant="body2">
                Timestamp: {composition.build.timestamp || "N/A"}
              </Typography>
            </Box>
          )}

          <Divider sx={{ my: 2 }} />

          {/* Files */}
          {composition.files && composition.files.length > 0 && (
            <Box mb={2}>
              <Typography variant="subtitle2" gutterBottom>
                Files:
              </Typography>
              <Stack direction="row" flexWrap="wrap" gap={1}>
                {composition.files.map((file, idx) => (
                  <Chip key={idx} label={file} variant="outlined" />
                ))}
              </Stack>
            </Box>
          )}

          {/* Deployments Visualization */}
          {composition.deployments && composition.deployments.length > 0 && (
            <Box>
              <Typography variant="subtitle2" gutterBottom>
                Deployments:
              </Typography>
              <Stack direction="row" spacing={1} flexWrap="wrap">
                {composition.deployments.map((deployment, idx) => (
                  <Chip
                    key={idx}
                    label={`${deployment.node} / ${deployment.namespace}`}
                    onClick={() => handleOpenDrawer(deployment)}
                    clickable
                    sx={{ cursor: "pointer" }}
                  />
                ))}
              </Stack>
            </Box>
          )}

          <Divider sx={{ my: 2 }} />
          <Stack direction="row" spacing={2}>
            <Button variant="contained" color="primary" onClick={handleOpenAddDeploymentModal}>
              New Deployment
            </Button>
            <Button variant="outlined" color="error" onClick={handleDelete}>
              Delete
            </Button>
          </Stack>
        </CardContent>
      </Card>

      {/* Drawer for Deployment Details */}
      <DeploymentDetailsDrawer
        deployment={selectedDeployment}
        onClose={handleCloseDrawer}
        allDeployments={allDeployments} // Pass all deployments for routing table editing
      />

      {/* Add Deployment Modal */}
      <CreateDeploymentModal
        open={isAddDeploymentModalOpen}
        onClose={handleCloseAddDeploymentModal}
        compositionId={composition.id ?? ""}
        components={composition.components || []}
        allDeployments={allDeployments}
      />
    </>
  );
};

export default FunctionCompositionCard;

