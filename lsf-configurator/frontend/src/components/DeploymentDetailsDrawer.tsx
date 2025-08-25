import React, { useState } from "react";
import { Drawer, Box, Typography, Divider, Stack, Button, Paper } from "@mui/material";
import type { Deployment } from "../models/models";
import ArrowForwardIcon from "@mui/icons-material/ArrowForward";
import RoutingTableModal from "./RoutingTableModal";

interface Props {
  deployment: Deployment | null;
  onClose: () => void;
  allDeployments: Deployment[]; 
}

const DeploymentDetailsDrawer: React.FC<Props> = ({ deployment, onClose, allDeployments }) => {
  const [isModalOpen, setModalOpen] = useState(false);

  const handleOpenModal = () => setModalOpen(true);
  const handleCloseModal = () => setModalOpen(false);

  return (
    <Drawer
      anchor="right"
      open={!!deployment}
      onClose={onClose}
      sx={{ width: 400 }}
    >
      {deployment && (
        <Box p={2} width="400px">
          <Typography variant="h6" gutterBottom sx={{ fontWeight: "bold" }}>
            Deployment Details
          </Typography>
          <Paper elevation={2} sx={{ p: 2, mb: 2, backgroundColor: "#e3f2fd" }}>
            <Typography variant="body1" sx={{ fontWeight: "bold" }}>
              Node:
            </Typography>
            <Typography variant="body2">{deployment.node}</Typography>
            <Typography variant="body1" sx={{ fontWeight: "bold", mt: 1 }}>
              Namespace:
            </Typography>
            <Typography variant="body2">{deployment.namespace}</Typography>
          </Paper>
          <Divider sx={{ my: 2 }} />
          <Typography variant="subtitle2" gutterBottom sx={{ fontWeight: "bold" }}>
            Routing Table:
          </Typography>
          <Box>
            <Stack direction="column" spacing={1}>
              {Object.entries(deployment.routingTable).map(([component, routes]) => (
                <React.Fragment key={component}>
                  <Paper elevation={1} sx={{ p: 1, backgroundColor: "#e8f5e9" }}>
                    <Typography variant="body2" fontWeight="bold" sx={{ color: "#388e3c" }}>
                      {component}
                    </Typography>
                    {routes && routes.map((route, rIdx) => (
                      <Stack
                        key={`${component}-${rIdx}`}
                        direction="row"
                        alignItems="center"
                        spacing={1}
                        sx={{ mt: 0.5 }}
                      >
                        <ArrowForwardIcon fontSize="small" color="action" />
                        <Typography variant="body2">
                          {route.to} ({route.function})
                        </Typography>
                      </Stack>
                    ))}
                    {(!routes || routes.length === 0) && (
                      <Stack direction="row" alignItems="center" spacing={1} sx={{ mt: 0.5 }}>
                        <ArrowForwardIcon fontSize="small" color="action" />
                        <Typography variant="body2" sx={{ fontStyle: "italic" }}>
                          &lt;none&gt;
                        </Typography>
                      </Stack>
                    )}
                  </Paper>
                </React.Fragment>
              ))}
            </Stack>
          </Box>
          <Divider sx={{ my: 2 }} />
          <Stack direction="row" spacing={2} alignItems="center">
            <Button
              variant="contained"
              color="primary"
              onClick={handleOpenModal}
            >
              Edit Routing Table
            </Button>
            <Button
              variant="outlined"
              color="primary"
              onClick={onClose}
            >
              Close
            </Button>
          </Stack>

          {/* Routing Table Modal */}
          <RoutingTableModal
            open={isModalOpen}
            onClose={handleCloseModal}
            deployment={deployment} // Pass the deployment for routing table editing
            allDeployments={allDeployments}
          />
        </Box>
      )}
    </Drawer>
  );
};

export default DeploymentDetailsDrawer;
