import React, { useState } from "react";
import {
  Modal,
  Box,
  Typography,
  Button,
  Select,
  MenuItem,
  Checkbox,
  ListItemText,
  InputLabel,
  FormControl,
  TextField,
} from "@mui/material";
import { useForm, Controller } from "react-hook-form";
import RoutingTableEditor from "./RoutingTableEditor";
import type { Build, Component, FunctionComposition, RoutingTable } from "../models/models";
import { useCreateFunctionComposition } from "../hooks/functionCompositionHooks";

interface FunctionCompositionAddModalProps {
  open: boolean;
  onClose: () => void;
  appId: string;
  appFiles: string[];
  appComponents: Component[];
  allCompositions: FunctionComposition[];
}

const FunctionCompositionAddModal: React.FC<FunctionCompositionAddModalProps> = ({
  open,
  onClose,
  appId,
  appFiles,
  appComponents,
  allCompositions,
}) => {
  const { control, handleSubmit, setValue } = useForm<FunctionComposition>({
    defaultValues: {
      node: "",
      namespace: "",
      files: [],
      components: {},
      build: {} as Build,
    },
  });

  const { mutate: createComposition } = useCreateFunctionComposition();

  const [selectedComponents, setSelectedComponents] = useState<string[]>([]);

  const handleComponentSelection = (selected: string[]) => {
    setSelectedComponents(selected);
    const newRoutingTable = selected.reduce((acc, component) => {
      acc[component] = [];
      return acc;
    }, {} as RoutingTable);
    setValue("components", newRoutingTable);
  };

  const onSubmit = (data: FunctionComposition) => {
    console.log("Saving composition:", data);
    createComposition({ appId, functionComposition: data });
    onClose();
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
        <form onSubmit={handleSubmit(onSubmit)}>
          <Controller
            name="node"
            control={control}
            render={({ field }) => (
              <TextField {...field} label="Node" fullWidth margin="normal" />
            )}
          />
          <Controller
            name="namespace"
            control={control}
            render={({ field }) => (
              <TextField {...field} label="Namespace" fullWidth margin="normal" />
            )}
          />
          <FormControl fullWidth margin="normal">
            <InputLabel id="files-label">Files</InputLabel>
            <Controller
              name="files"
              control={control}
              render={({ field }) => (
                <Select
                  {...field}
                  labelId="files-label"
                  label="Files"
                  multiple
                  renderValue={(selected) => selected.join(", ")}
                >
                  {appFiles.map((file) => (
                    <MenuItem key={file} value={file}>
                      <Checkbox checked={field.value.includes(file)} />
                      <ListItemText primary={file} />
                    </MenuItem>
                  ))}
                </Select>
              )}
            />
          </FormControl>
          <FormControl fullWidth margin="normal">
            <InputLabel id="components-label">Components</InputLabel>
            <Select
              labelId="components-label"
              label="Components"
              multiple
              value={selectedComponents}
              onChange={(e) =>
                handleComponentSelection(e.target.value as string[])
              }
              renderValue={(selected) => selected.join(", ")}
            >
              {appComponents.map((component) => (
                <MenuItem key={component} value={component}>
                  <Checkbox checked={selectedComponents.includes(component)} />
                  <ListItemText primary={component} />
                </MenuItem>
              ))}
            </Select>
          </FormControl>
          <Box
            sx={{
              borderTop: "1px solid #ccc",
              my: 3,
            }}
          />
          <Typography variant="body1" mb={2}>
            Routing Table
          </Typography>
          <RoutingTableEditor
            composition={{
              id: "",
              node: "",
              components: selectedComponents.reduce((acc, component) => {
                acc[component] = [];
                return acc;
              }, {} as RoutingTable),
              namespace: "",
              files: [],
              build: {} as Build,
            }}
            allCompositions={allCompositions}
            onChange={(data) => setValue("components", data)}
          />
          <Box mt={2} display="flex" justifyContent="space-between">
            <Button variant="outlined" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" variant="contained">
              Save
            </Button>
          </Box>
        </form>
      </Box>
    </Modal>
  );
};

export default FunctionCompositionAddModal;
