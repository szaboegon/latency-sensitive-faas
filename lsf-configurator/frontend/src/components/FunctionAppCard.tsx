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

interface FunctionAppCardProps {
  app: FunctionApp;
}

const FunctionAppCard: React.FC<FunctionAppCardProps> = ({ app }) => {
  const navigate = useNavigate();

  return (
    <Paper
      elevation={0}
      sx={{
        p: 3,
        display: "flex",
        flexDirection: "column",
        borderRadius: 2,
        border: "1px solid",
        borderColor: "divider",
        backgroundColor: "background.paper",
        transition: "box-shadow 0.2s ease-in-out, transform 0.1s ease-in-out",
        "&:hover": {
          boxShadow: 4,
          transform: "translateY(-2px)",
        },
        height: "100%",
      }}
    >
      <Typography variant="h6" fontWeight={600} gutterBottom>
        {app.name}
      </Typography>
      <Divider sx={{ my: 1 }} />
      <Typography
        variant="subtitle2"
        color="text.secondary"
        sx={{ mb: 1, fontWeight: 500 }}
      >
        Components
      </Typography>
      <List dense disablePadding sx={{ flexGrow: 1 }}>
        {app.components?.map((component) => (
          <ListItem
            key={component}
            sx={{
              px: 0,
              py: 0.5,
              borderBottom: "1px solid",
              borderColor: "divider",
            }}
          >
            <ListItemText
              primaryTypographyProps={{ variant: "body2", color: "text.primary" }}
              primary={component}
            />
          </ListItem>
        ))}
      </List>
      <Box mt={2}>
        <Button
          variant="contained"
          size="small"
          onClick={() => navigate(`/function-apps/${app.id}`)}
          sx={{ textTransform: "none", fontWeight: 500 }}
          fullWidth
        >
          View Details
        </Button>
      </Box>
    </Paper>
  );
};

export default FunctionAppCard;
