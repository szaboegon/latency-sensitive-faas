import React, { useState } from "react";
import { Card, CardContent, CardHeader, Typography, Divider, Box, Chip, Stack, Button, Grid } from "@mui/material";
import type { FunctionComposition } from "../models/models"; 
import ArrowForwardIcon from "@mui/icons-material/ArrowForward";
import RoutingTableModal from "../components/RoutingTableModal"; // Import the modal component

interface Props {
  composition: FunctionComposition;
}

const FunctionCompositionCard: React.FC<Props> = ({ composition }) => {
  const [isModalOpen, setModalOpen] = useState(false);

  const handleOpenModal = () => setModalOpen(true);
  const handleCloseModal = () => setModalOpen(false);

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
          subheader={`Node: ${composition.node || "N/A"} • Namespace: ${composition.namespace}`}
        />

        <CardContent>
            {/* Components Visualization */}
          {composition.components && (
            <Box mb={2}>
              <Typography variant="subtitle2" gutterBottom>
                Components:
              </Typography>
              <Grid container spacing={1}>
                {Object.keys(composition.components).map((component, index) => (
                  <Grid size={{ xs: 6, sm: 4 }} key={component}>
                    <Box
                      sx={{
                        backgroundColor: `hsl(${(index * 60) % 360}, 70%, 80%)`,
                        borderRadius: 2,
                        padding: 1,
                        textAlign: "center",
                        boxShadow: 1,
                      }}
                    >
                      <Typography variant="body2">{component}</Typography>
                    </Box>
                  </Grid>
                ))}
              </Grid>
            </Box>
          )}
          {/* Runtime */}
          <Typography variant="subtitle2" color="textSecondary">
            Runtime: {composition.runtime}
          </Typography>

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

          {/* Routing Table Visualization */}
          {composition.components && (
            <Box>
              <Typography variant="subtitle2" gutterBottom>
                Routing Table:
              </Typography>
              {Object.entries(composition.components).map(([component, routes], idx) => (
                <Box
                  key={idx}
                  sx={{
                    mb: 1.5,
                    p: 1,
                    border: "1px dashed #80deea",
                    borderRadius: 2,
                    background: "#f0fafa",
                  }}
                >
                  <Typography variant="body2" fontWeight="bold">
                    Component: {component}
                  </Typography>
                  {routes.map((route, rIdx) => (
                    <Stack
                      key={rIdx}
                      direction="row"
                      alignItems="center"
                      spacing={1}
                      sx={{ pl: 2, mt: 0.5 }}
                    >
                      <Typography variant="body2">{route.function}</Typography>
                      <ArrowForwardIcon fontSize="small" color="action" />
                      <Typography variant="body2">{route.to}</Typography>
                    </Stack>
                  ))}
                </Box>
              ))}
            </Box>
          )}


          <Divider sx={{ my: 2 }} />
          <Button variant="contained" color="primary" onClick={handleOpenModal}>
            Edit Routing Table
          </Button>
        </CardContent>
      </Card>
      <RoutingTableModal
        open={isModalOpen}
        onClose={handleCloseModal}
        composition={composition}
      />
    </>
  );
};

export default FunctionCompositionCard;
