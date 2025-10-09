import React from "react";
import {
  Drawer,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  Toolbar,
  Typography,
  Box,
  ListItemButton,
} from "@mui/material";
import { Home } from "@mui/icons-material";
import { useNavigate } from "react-router";
import { colors } from "../helpers/constants";

const Sidebar: React.FC = () => {
  const navigate = useNavigate();

  const menuItems = [{ text: "Home", icon: <Home />, path: "/" }];

  return (
    <Drawer
      variant="permanent"
      sx={{
        width: 240,
        flexShrink: 0,
        "& .MuiDrawer-paper": {
          width: 240,
          boxSizing: "border-box",
          backgroundColor: colors.dark,
          color: "#fff",
        },
      }}
    >
      <Toolbar>
        <Typography
          variant="h6"
          sx={{ color: "#fff", textAlign: "center", width: "100%" }}
        >
          Dashboard
        </Typography>
      </Toolbar>
      <Box sx={{ overflow: "auto" }}>
        <List>
          {menuItems.map((item) => (
            <ListItem component="li" key={item.text} disablePadding>
              <ListItemButton onClick={() => navigate(item.path)}>
                <ListItemIcon sx={{ color: "#fff" }}>{item.icon}</ListItemIcon>
                <ListItemText primary={item.text} />
              </ListItemButton>
            </ListItem>
          ))}
        </List>
      </Box>
    </Drawer>
  );
};

export default Sidebar;
