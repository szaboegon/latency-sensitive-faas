import React from "react";
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
} from "@mui/material";
import { useForm, Controller } from "react-hook-form";
import type { Build, FunctionComposition } from "../models/models";
import { useCreateFunctionComposition } from "../hooks/functionCompositionHooks";

interface FunctionCompositionAddModalProps {
  open: boolean;
  onClose: () => void;
  appId: string;
  appFiles: string[];
  appComponents: string[];
}

const FunctionCompositionAddModal: React.FC<FunctionCompositionAddModalProps> = ({
  open,
  onClose,
  appId,
  appFiles,
  appComponents,
}) => {
  const { control, handleSubmit } = useForm<FunctionComposition>({
    defaultValues: {
      files: [],
      components: [],
      build: {} as Build,
    },
  });

  const { mutate: createComposition } = useCreateFunctionComposition();

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
            <Controller
              name="components"
              control={control}
              render={({ field }) => (
                <Select
                  {...field}
                  labelId="components-label"
                  label="Components"
                  multiple
                  renderValue={(selected) => selected.join(", ")}
                >
                  {appComponents.map((component) => (
                    <MenuItem key={component} value={component}>
                      <Checkbox checked={field.value.includes(component)} />
                      <ListItemText primary={component} />
                    </MenuItem>
                  ))}
                </Select>
              )}
            />
          </FormControl>
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