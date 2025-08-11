import React from "react";
import { Container, Typography, Box, Grid } from "@mui/material";
import { useParams } from "react-router";
import FunctionAppCard from "../components/FunctionAppCard";
import FunctionCompositionCard from "../components/FunctionCompositionCard";
import { useFunctionApps } from "../hooks/useFunctionApps";
import type { FunctionComposition } from "../models/models";

const FunctionAppDetails: React.FC = () => {
  const { id } = useParams();
  const { data: functionApps, isLoading, error } = useFunctionApps();
  const app = functionApps?.find((app) => app.id === id);

  if (isLoading) {
    return (
      <Container>
        <Typography variant="h6" align="center">
          Loading...
        </Typography>
      </Container>
    );
  }

  if (error) {
    return (
      <Container>
        <Typography variant="h6" align="center" color="error">
          Error: {error.message}
        </Typography>
      </Container>
    );
  }

  if (!app) {
    return (
      <Container>
        <Typography variant="body1" align="center">
          Function app not found.
        </Typography>
      </Container>
    );
  }

  return (
    <Container>
      <Box my={4}>
        <Typography variant="h4" gutterBottom>
          {app.name} Details
        </Typography>

        {/* Components section */}
        <Typography variant="h5" gutterBottom>
        Components
        </Typography>
        <Grid container spacing={2} mb={4} sx={{ maxHeight: 400, overflow: "auto" }}>
          {app.components?.slice(0, 6).map((component, index) => (
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
        <Grid container rowSpacing={4} columnSpacing={3} sx={{ maxHeight: 400, overflow: "auto" }}>
          {app.compositions?.slice(0, 6).map((composition: FunctionComposition) => (
            <Grid size={{ xs: 12, sm: 6, md: 4 }} key={composition.id}>
              <FunctionCompositionCard composition={composition} />
            </Grid>
          ))}
        </Grid>
      </Box>
    </Container>
  );
};

export default FunctionAppDetails;
  );
};

export default FunctionAppDetails;
