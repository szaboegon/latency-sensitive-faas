import React from "react";
import { useFunctionApps } from "../hooks/useFunctionApps";
import { Container, Typography, Box } from "@mui/material";
import FunctionAppCard from "../components/FunctionAppCard";
import Grid  from "@mui/material/Grid";


const Home: React.FC = () => {
  const { data: functionApps, isLoading, error } = useFunctionApps();

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

  return (
    <Container>
      <Box my={4}>
        <Typography variant="h4" gutterBottom>
          Function Apps Overview
        </Typography>
        {functionApps?.length === 0 ? (
          <Typography variant="body1" color="textSecondary">
            No function apps found.
          </Typography>
        ) : (
          <Grid container rowSpacing={9} columnSpacing={3}>
            {functionApps?.map((app) => (
              <Grid size={{ xs: 12, sm: 6, md: 4 }} key={app.id}>
                <FunctionAppCard app={app} />
              </Grid>
            ))}
          </Grid>
        )}
      </Box>
    </Container>
  );
};

export default Home;
