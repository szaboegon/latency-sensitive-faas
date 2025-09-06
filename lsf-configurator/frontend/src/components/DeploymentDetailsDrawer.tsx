import React, { useState } from "react";
import { Drawer, Box, Typography, Divider, Stack, Button, Card, IconButton, Tooltip } from "@mui/material";
import ContentCopyIcon from "@mui/icons-material/ContentCopy";
import ArrowForwardIcon from "@mui/icons-material/ArrowForward";
import RoutingTableModal from "./RoutingTableModal";
import type { Deployment } from "../models/models";

interface Props {
  deployment: Deployment | null;
  onClose: () => void;
  allDeployments: Deployment[]; 
}

const DeploymentDetailsDrawer: React.FC<Props> = ({ deployment, onClose, allDeployments }) => {
  const [isModalOpen, setModalOpen] = useState(false);
  const [copied, setCopied] = useState(false);

  const handleOpenModal = () => setModalOpen(true);
  const handleCloseModal = () => setModalOpen(false);

  const url = deployment
    ? `${deployment.id}.${deployment.namespace}.127.0.0.1.sslip.io`
    : "";

  const handleCopyUrl = () => {
    navigator.clipboard.writeText(url);
    setCopied(true);
    setTimeout(() => setCopied(false), 1200);
  };

  return (
    <Drawer
      anchor="right"
      open={!!deployment}
      onClose={onClose}
      slotProps={{
        paper: {
          sx: {
            width: 400,
            borderTopLeftRadius: 16,
            borderBottomLeftRadius: 16,
            boxShadow: 6,
            background: "linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%)"
          }
        }
      }}
    >
      {deployment && (
        <Box p={2} width="400px">
          <Typography variant="h6" gutterBottom sx={{ fontWeight: 700 }}>
            Deployment <span style={{ color: "#333" }}>{deployment.id}</span>
          </Typography>
          <Box mb={2}>
            <Stack spacing={1}>
              <Stack direction="row" spacing={1} alignItems="center">
                <Typography variant="body2" sx={{ fontWeight: 600, minWidth: 90 }}>
                  Node:
                </Typography>
                <Typography variant="body2">{deployment.node}</Typography>
              </Stack>
              <Stack direction="row" spacing={1} alignItems="center">
                <Typography variant="body2" sx={{ fontWeight: 600, minWidth: 90 }}>
                  Namespace:
                </Typography>
                <Typography variant="body2">{deployment.namespace}</Typography>
              </Stack>
            </Stack>
          </Box>
          <Divider sx={{ my: 2 }} />
          <Stack direction="row" alignItems="center" spacing={1} sx={{ overflow: "auto" }}>
            <Typography variant="subtitle2" sx={{ fontWeight: 700, color: "#388e3c" }}>
              URL:
            </Typography>
            <Typography
              variant="body2"
              sx={{
                fontFamily: "monospace",
                color: "#333",
                whiteSpace: "nowrap",
                overflow: "auto",
                flex: 1
              }}
            >
              {url}
            </Typography>
            <Tooltip title={copied ? "Copied!" : "Copy"}>
              <IconButton size="small" onClick={handleCopyUrl}>
                <ContentCopyIcon fontSize="small" />
              </IconButton>
            </Tooltip>
          </Stack>
          <Divider sx={{ my: 2 }} />
          <Typography variant="subtitle2" gutterBottom sx={{ fontWeight: 700 }}>
            Routing Table
          </Typography>
          <Box>
            <Stack direction="column" spacing={1}>
              {Object.entries(deployment.routingTable).map(([component, routes]) => (
                <React.Fragment key={component}>
                  <Card elevation={1} sx={{ p: 1, borderRadius: 2, background: "#e8f5e9" }}>
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
                        <Typography variant="body2" sx={{ color: "#333" }}>
                          {route.to} <span style={{ color: "#888" }}>({route.function})</span>
                        </Typography>
                      </Stack>
                    ))}
                    {(!routes || routes.length === 0) && (
                      <Stack direction="row" alignItems="center" spacing={1} sx={{ mt: 0.5 }}>
                        <ArrowForwardIcon fontSize="small" color="action" />
                        <Typography variant="body2" sx={{ fontStyle: "italic", color: "#aaa" }}>
                          &lt;none&gt;
                        </Typography>
                      </Stack>
                    )}
                  </Card>
                </React.Fragment>
              ))}
            </Stack>
          </Box>
          <Divider sx={{ my: 2 }} />
          <Stack direction="row" spacing={2} alignItems="center" justifyContent="flex-end">
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
            deployment={deployment}
            allDeployments={allDeployments}
          />
        </Box>
      )}
    </Drawer>
  );
};

export default DeploymentDetailsDrawer;
