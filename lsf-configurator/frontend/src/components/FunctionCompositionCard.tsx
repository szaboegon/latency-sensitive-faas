import React from "react";
import { Card, CardContent, CardHeader, Typography, Divider, Box, Chip, Stack } from "@mui/material";
import type { FunctionComposition } from "../models/models"; 
import ArrowForwardIcon from "@mui/icons-material/ArrowForward";

interface Props {
  composition: FunctionComposition;
}

const FunctionCompositionCard: React.FC<Props> = ({ composition }) => {
  return (
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
      </CardContent>
    </Card>
  );
};

export default FunctionCompositionCard;
