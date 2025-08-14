import React, { useState } from "react";
import {
  Modal,
  Box,
  Typography,
  TextField,
  Button,
  Select,
  MenuItem,
  Checkbox,
  ListItemText,
  InputLabel,
  FormControl,
} from "@mui/material";
import RoutingTableForm from "./RoutingTableForm";
import type { Build, Component, FunctionComposition, RoutingTable } from "../models/models";

interface FunctionCompositionAddModalProps {
  open: boolean;
  onClose: () => void;
  appFiles: string[];
  appComponents: Component[];
  allCompositions: FunctionComposition[];
}

const FunctionCompositionAddModal: React.FC<FunctionCompositionAddModalProps> = ({
  open,
  onClose,
  appFiles,
  appComponents,
  allCompositions
}) => {
  const [compositionData, setCompositionData] = useState({
    node: "",
    namespace: "",
    runtime: "",
    files: [] as string[],
    components: [] as string[],
    routingTable: {} as RoutingTable,
  });

  const handleSave = () => {
    console.log("Saving composition:", compositionData);
    onClose();
  };

  const handleRoutingTableChange = (data: RoutingTable) => {
    setCompositionData((prev) => ({
      ...prev,
      routingTable: { ...prev.routingTable, ...data },
    }));
  };

  return (
    <Modal open={open} onClose={onClose}>
      <Box
        sx={{
          position: "absolute",
          top: "50%",
          left: "50%",
          transform: "translate(-50%, -50%)",
          width: 800,
          bgcolor: "background.paper",
          borderRadius: 2,
          boxShadow: 24,
          p: 4,
        }}
      >
        <Typography variant="h6" mb={2}>
          Add New Function Composition
        </Typography>
        <TextField
          label="Node"
          fullWidth
          margin="normal"
          value={compositionData.node}
          onChange={(e) =>
            setCompositionData((prev) => ({ ...prev, node: e.target.value }))
          }
        />
        <TextField
          label="Namespace"
          fullWidth
          margin="normal"
          value={compositionData.namespace}
          onChange={(e) =>
            setCompositionData((prev) => ({
              ...prev,
              namespace: e.target.value,
            }))
          }
        />
        <FormControl fullWidth margin="normal">
          <InputLabel id="runtime-label">Runtime</InputLabel>
          <Select
            labelId="runtime-label"
            label="Runtime"
            value={compositionData.runtime}
            onChange={(e) =>
              setCompositionData((prev) => ({
                ...prev,
                runtime: e.target.value as string,
              }))
            }
          >
            <MenuItem value="Node.js">Node.js</MenuItem>
            <MenuItem value="Python">Python</MenuItem>
            <MenuItem value="Go">Go</MenuItem>
          </Select>
        </FormControl>
        <FormControl fullWidth margin="normal">
          <InputLabel id="files-label">Files</InputLabel>
          <Select
            labelId="files-label"
            label="Files"
            multiple
            value={compositionData.files}
            onChange={(e) =>
              setCompositionData((prev) => ({
                ...prev,
                files: e.target.value as string[],
              }))
            }
            renderValue={(selected) => selected.join(", ")}
          >
            {appFiles.map((file) => (
              <MenuItem key={file} value={file}>
                <Checkbox checked={compositionData.files.includes(file)} />
                <ListItemText primary={file} />
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        <FormControl fullWidth margin="normal">
          <InputLabel id="components-label">Components</InputLabel>
          <Select
            labelId="components-label"
            label="Components"
            multiple
            value={compositionData.components}
            onChange={(e) =>
              setCompositionData((prev) => ({
                ...prev,
                components: e.target.value as string[],
              }))
            }
            renderValue={(selected) => selected.join(", ")}
          >
            {appComponents.map((component) => (
              <MenuItem key={component} value={component}>
                <Checkbox
                  checked={compositionData.components.includes(component)}
                />
                <ListItemText primary={component} />
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        <RoutingTableForm
          composition={{
            id: "",
            node: "",
            components: compositionData.components.reduce((acc, comp) => {
              acc[comp] = [];
              return acc;
            }, {} as RoutingTable),
            namespace: "",
            runtime: "",
            files: [],
            build: {} as Build,
          }}
          allCompositions={allCompositions}
          onSave={handleRoutingTableChange}
        />
        <Box mt={2} display="flex" justifyContent="space-between">
          <Button variant="outlined" onClick={onClose}>
            Cancel
          </Button>
          <Button variant="contained" onClick={handleSave}>
            Save
          </Button>
        </Box>
      </Box>
    </Modal>
  );
};

export default FunctionCompositionAddModal;
