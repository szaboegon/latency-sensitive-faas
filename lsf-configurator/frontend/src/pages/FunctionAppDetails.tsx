import React, { useState, useMemo } from "react";
import { Typography, Box, Grid, Button, Tabs, Tab } from "@mui/material";
import { useParams } from "react-router";
import FunctionCompositionCard from "../components/FunctionCompositionCard";
import { useFunctionAppById } from "../hooks/functionAppsHooks";
import type { FunctionComposition } from "../models/models";
import { generateComponentColor } from "../helpers/utilities";
import { useDeleteFunctionComposition } from "../hooks/functionCompositionHooks";
import AddIcon from "@mui/icons-material/Add";
import CallGraphView from "../components/CallGraphView";
import FunctionCompositionAddModal from "../components/FunctionCompositionAddModal";

const FunctionAppDetails: React.FC = () => {
  const { id } = useParams();
  const { data: app, isLoading, error } = useFunctionAppById(id ?? "");
  const { mutate: deleteComposition } = useDeleteFunctionComposition();
  console.log("FunctionAppDetails app:", app);

  const [tabValue, setTabValue] = useState<"list" | "graph">("list");
  const [isAddModalOpen, setAddModalOpen] = useState(false);

  const allDeployments = useMemo(
    () => app?.compositions?.flatMap((composition) => composition.deployments) ?? [],
    [app]
  );

  const handleAddComposition = () => {
    setAddModalOpen(true);
  };

  const handleCloseAddModal = () => {
    setAddModalOpen(false);
  };

  const handleTabChange = (
    _: React.SyntheticEvent,
    newValue: "list" | "graph"
  ) => {
    setTabValue(newValue);
  };

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
    <>
      <Box my={4}>
        <Typography variant="h4" gutterBottom>
          {app.name} Details
        </Typography>

        {/* Components section */}
        <Typography variant="h5" gutterBottom>
          Components
        </Typography>
        <Grid container spacing={2} mb={4}>
          {app.components?.map((component) => (
            <Grid key={component.name} size={{ xs: 12, sm: 6, md: 4 }}>
              <Box
                sx={{
                  backgroundColor: generateComponentColor(component.name),
                  borderRadius: 2,
                  padding: 2,
                  textAlign: "center",
                  boxShadow: 2,
                }}
              >
                <Typography variant="h6">{component.name}</Typography>
              </Box>
            </Grid>
          ))}
        </Grid>

        {/* Tabs for compositions */}
        <Tabs value={tabValue} onChange={handleTabChange} sx={{ mb: 2 }}>
          <Tab label="List View" value="list" />
          <Tab label="Call Graph View" value="graph" />
        </Tabs>

        {tabValue === "list" && (
          <Grid container spacing={4}>
            {app.compositions?.map((composition: FunctionComposition) => (
              <Grid key={composition.id} size={{ xs: 12, sm: 6, md: 4 }}>
                <FunctionCompositionCard
                  composition={composition}
                  onDelete={deleteComposition}
                  allDeployments={allDeployments} 
                />
              </Grid>
            ))}
            <Grid size={{ xs: 12, sm: 6, md: 4 }}>
              <Box
                sx={{
                  display: "flex",
                  justifyContent: "center",
                  alignItems: "center",
                  height: "100%",
                  border: "1px dashed #b2ebf2",
                  borderRadius: 2,
                  padding: 2,
                  boxShadow: 2,
                  cursor: "pointer",
                  minHeight: "600px",
                }}
              >
                <Button
                  variant="outlined"
                  startIcon={<AddIcon />}
                  onClick={handleAddComposition}
                >
                  Add Composition
                </Button>
              </Box>
            </Grid>
          </Grid>
        )}

        {tabValue === "graph" && (
          <CallGraphView deployments={allDeployments} /> 
        )}
      </Box>
      <FunctionCompositionAddModal
        open={isAddModalOpen}
        onClose={handleCloseAddModal}
        appId={app.id ?? ""}
        appFiles={app.files ?? []}
        appComponents={app.components?.map(comp => comp.name) ?? []}
      />
    </>
  );
};

export default FunctionAppDetails;
