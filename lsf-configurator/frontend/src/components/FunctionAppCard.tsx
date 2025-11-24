import React from "react";
import type { FunctionApp } from "../models/models";
import {
  Paper,
  Typography,
  List,
  ListItem,
  ListItemText,
  Button,
  Box,
  Divider,
} from "@mui/material";
import { useNavigate } from "react-router";
import { useDeleteFunctionApp } from "../hooks/functionAppsHooks";

interface FunctionAppCardProps {
  app: FunctionApp;
}

const FunctionAppCard: React.FC<FunctionAppCardProps> = ({ app }) => {
  const navigate = useNavigate();
  const { mutate: deleteFunctionApp } = useDeleteFunctionApp();

  return (
    <Paper
      elevation={2}
      sx={{
        borderRadius: 2,
        p: 2,
        display: "flex",
        flexDirection: "column",
        background: "#f5faff",
        transition: "box-shadow 0.2s ease-in-out, transform 0.1s ease-in-out",
        "&:hover": {
          boxShadow: 4,
          transform: "translateY(-2px)",
        },
        height: "100%",
      }}
    >
      <Typography variant="h6" fontWeight={600} gutterBottom>
        {app.name}({app.id})
      </Typography>
      <Divider sx={{ my: 1 }} />
      {/* Runtime */}
      <Typography variant="subtitle2" color="textSecondary" gutterBottom>
        Runtime: {app.runtime}
      </Typography>
      <Typography
        variant="h6"
        color="text.secondary"
        sx={{ mb: 1, fontWeight: 500 }}
      >
        Components
      </Typography>
      <List dense disablePadding sx={{ flexGrow: 1 }}>
        {app.components?.map((component) => (
          <ListItem
            key={component.name}
            sx={{
              px: 0,
              py: 0.5,
              borderRadius: 1,
              backgroundColor: "#f9f9f9",
              border: "1px solid #e0e0e0",
              mb: 1,
            }}
          >
            <ListItemText
              primaryTypographyProps={{
                variant: "body2",
                color: "text.primary",
                textAlign: "center",
              }}
              primary={component.name}
            />
          </ListItem>
        ))}
      </List>
      <Box mt={2} sx={{ display: "flex", gap: 1 }}>
        <Button
          variant="contained"
          size="small"
          onClick={() => navigate(`/function-apps/${app.id}`)}
          sx={{ textTransform: "none", fontWeight: 500, flex: 1 }}
        >
          View Details
        </Button>
        <Button
          variant="outlined"
          color="error"
          size="small"
          onClick={() => deleteFunctionApp(app.id!)}
          sx={{ textTransform: "none", fontWeight: 500, flex: 1 }}
        >
          Delete
        </Button>
      </Box>
    </Paper>
  );
};

export default FunctionAppCard;
