import React from "react";
import { Box, Typography, Grid, Paper, Chip, Stack } from "@mui/material";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import LayersIcon from "@mui/icons-material/Layers";
import SpeedIcon from "@mui/icons-material/Speed";
import MemoryIcon from "@mui/icons-material/Memory";
import type { Layout } from "../models/models";
import { colors } from "../helpers/constants";
import { toSnakeCase } from "../helpers/utilities";

interface Props {
  layoutCandidates: Record<string, Layout>;
  activeLayoutId?: string;
}

const LayoutCandidatesView: React.FC<Props> = ({
  layoutCandidates,
  activeLayoutId,
}) => {
  const keys = Object.keys(layoutCandidates);

  if (!layoutCandidates || keys.length === 0) {
    return (
      <Typography variant="body2" color="text.secondary">
        No layout candidates available.
      </Typography>
    );
  }

  return (
    <Box>
      <Typography variant="h5" gutterBottom>
        Layout Candidates
      </Typography>
      <Grid container spacing={2}>
        {keys.map((key) => {
          const candidateLayout = layoutCandidates[key];
          const isActive =
            toSnakeCase(key) === toSnakeCase(activeLayoutId ?? "");
          return (
            <Grid size={{ xs: 12, sm: 6, md: 4 }} key={key}>
              <Paper
                elevation={6}
                sx={{
                  borderRadius: 2,
                  padding: 2,
                  boxShadow: isActive ? 4 : 1,
                  position: "relative",
                }}
              >
                <Stack direction="row" alignItems="center" spacing={1} mb={1}>
                  <LayersIcon color="primary" />
                  <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
                    Layout Key: {toSnakeCase(key)}
                  </Typography>
                  <Chip
                    label={key}
                    color={isActive ? "success" : "default"}
                    size="small"
                    sx={{ ml: 1, fontWeight: 500 }}
                  />
                  {isActive && (
                    <Chip
                      icon={<CheckCircleIcon color="success" />}
                      label="Active"
                      color="success"
                      size="small"
                      sx={{ ml: 1, fontWeight: 500 }}
                    />
                  )}
                </Stack>
                {/* Show layout structure */}
                <Box mt={1}>
                  {Object.entries(candidateLayout).map(([group, profiles]) => (
                    <Box key={group} mb={2}>
                      <Typography variant="subtitle2" sx={{ fontWeight: 600 }}>
                        Group: {group}
                      </Typography>
                      <Stack
                        spacing={1}
                        sx={{
                          border: `2px solid ${colors.dark}`,
                          borderRadius: 1,
                        }}
                      >
                        {profiles.map((profile, idx) => (
                          <Paper
                            key={profile.name + idx}
                            variant="outlined"
                            sx={{
                              outlineWidth: 0,
                              p: 1,
                              display: "flex",
                              alignItems: "center",
                              gap: 1,
                              background: "#f9f9f9",
                            }}
                          >
                            <Typography
                              variant="body2"
                              sx={{ fontWeight: 500 }}
                            >
                              {profile.name}
                            </Typography>
                            <Chip
                              icon={<MemoryIcon />}
                              label={`${profile.memory} MB`}
                              size="small"
                              color="secondary"
                              sx={{ fontWeight: 500 }}
                            />
                            <Chip
                              icon={<SpeedIcon />}
                              label={`${profile.runtime} ms`}
                              size="small"
                              color="info"
                              sx={{ fontWeight: 500 }}
                            />
                            <Chip
                              label={`Maximum Replicas: ${profile.requiredReplicas}`}
                              size="small"
                              color="default"
                              sx={{ fontWeight: 500 }}
                            />
                          </Paper>
                        ))}
                      </Stack>
                    </Box>
                  ))}
                </Box>
              </Paper>
            </Grid>
          );
        })}
      </Grid>
    </Box>
  );
};

export default LayoutCandidatesView;
