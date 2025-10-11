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
        borderRadius: 3,
        boxShadow: 4,
        background: "linear-gradient(135deg, #e0f7fa 0%, #fff 100%)",
        border: "1px solid #b2ebf2",
        p: 3,
        display: "flex",
        flexDirection: "column",
        transition: "box-shadow 0.2s ease-in-out, transform 0.1s ease-in-out",
        "&:hover": {
          boxShadow: 6,
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
              borderRadius: 2,
              backgroundColor: "#f0fafa",
              boxShadow: 1,
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
