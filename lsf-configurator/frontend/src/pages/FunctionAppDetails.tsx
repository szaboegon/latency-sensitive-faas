import React from "react";
import {  Typography, Box, Grid, Container } from "@mui/material";
import { useParams } from "react-router";
import FunctionCompositionCard from "../components/FunctionCompositionCard";
import { useFunctionApps } from "../hooks/functionAppsHooks";
import type { FunctionComposition } from "../models/models";

const FunctionAppDetails: React.FC = () => {
  const { id } = useParams();
  const { data: functionApps, isLoading, error } = useFunctionApps();
  const app = functionApps?.find((app) => app.id === id);

  if (isLoading) {
    return (
      <Box
        sx={{
          height: "100vh",
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
        }}
      >
        <Typography variant="h6" align="center">
          Loading...
        </Typography>
      </Box>
    );
  }

  if (error) {
    return (
      <Box
        sx={{
          height: "100vh",
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
        }}
      >
        <Typography variant="h6" align="center" color="error">
          Error: {error.message}
        </Typography>
      </Box>
    );
  }

  if (!app) {
    return (
      <Box
        sx={{
          height: "100vh",
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
        }}
      >
        <Typography variant="body1" align="center">
          Function app not found.
        </Typography>
      </Box>
    );
  }

  return (
      <Box my={4} >
        <Typography variant="h4" gutterBottom>
          {app.name} Details
        </Typography>

        {/* Components section */}
        <Typography variant="h5" gutterBottom>
          Components
        </Typography>
        <Grid container spacing={2} mb={4}>
          {app.components?.map((component, index) => (
            <Grid size={{ xs: 12, sm: 6, md: 4 }} key={component}>
              <Box
                sx={{
                  backgroundColor: `hsl(${(index * 60) % 360}, 70%, 80%)`,
                  borderRadius: 2,
                  padding: 2,
                  textAlign: "center",
                  boxShadow: 2,
                }}
              >
                <Typography variant="h6">{component}</Typography>
              </Box>
            </Grid>
          ))}
        </Grid>

        {/* Function compositions */}
        <Typography variant="h5" gutterBottom>
          Function Compositions
        </Typography>
        <Grid container spacing={4}>
          {app.compositions?.map((composition: FunctionComposition) => (
            <Grid size={{ xs: 12, sm: 6, md: 4 }} key={composition.id}>
              <FunctionCompositionCard composition={composition} />
            </Grid>
          ))}
        </Grid>
      </Box>
  );
};

export default FunctionAppDetails;

