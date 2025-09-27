import React, { useState } from "react";
import { useForm, useFieldArray, Controller } from "react-hook-form";
import {
  Modal,
  Box,
  Typography,
  TextField,
  Button,
  Input,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Tabs,
  Tab,
  IconButton,
  Checkbox,
  FormControlLabel,
  OutlinedInput,
  Chip,
} from "@mui/material";
import { useCreateFunctionApp } from "../hooks/functionAppsHooks";
import FunctionAppJsonForm from "./FunctionAppJsonForm";
import type { FunctionAppCreateDto } from "../models/dto";
import type { Component, ComponentLink } from "../models/models";
import AddIcon from "@mui/icons-material/Add";
import DeleteIcon from "@mui/icons-material/Delete";
import { fileExtensions } from "../helpers/constants";

interface AddFunctionAppModalProps {
  open: boolean;
  onClose: () => void;
}

interface FormData {
  name: string;
  files: FileList;
  runtime: string;
  latencyLimit: number;
  components: (Component & { files?: string[] })[]; // add files field
  links: ComponentLink[];
  platformManaged?: boolean;
}

const AddFunctionAppModal: React.FC<AddFunctionAppModalProps> = ({
  open,
  onClose,
}) => {
  const { register, handleSubmit, control } = useForm<FormData>({
    defaultValues: {
      components: [],
      links: [],
      platformManaged: false,
    },
  });

  const {
    fields: componentFields,
    append: appendComponent,
    remove: removeComponent,
  } = useFieldArray({ control, name: "components" });

  const {
    fields: linkFields,
    append: appendLink,
    remove: removeLink,
  } = useFieldArray({ control, name: "links" });

  const { mutate: createFunctionApp } = useCreateFunctionApp();
  const [tabValue, setTabValue] = useState<"form" | "json">("form");
  const [uploadedFileNames, setUploadedFileNames] = useState<string[]>([]);

  const onSubmit = (data: FormData) => {
    const fcApp: FunctionAppCreateDto = {
      name: data.name,
      runtime: data.runtime,
      latencyLimit: data.latencyLimit,
      components: data.components,
      links: data.links,
      platformManaged: data.platformManaged || false,
    };
    createFunctionApp({ functionApp: fcApp, files: data.files });
    onClose();
  };

  // Handle file input change to update uploadedFileNames
  const handleFilesChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (files) {
      setUploadedFileNames(Array.from(files).map((f) => f.name));
    } else {
      setUploadedFileNames([]);
    }
  };

  const handleTabChange = (
    _: React.SyntheticEvent,
    newValue: "form" | "json"
  ) => {
    setTabValue(newValue);
  };

  const getExtensionForRuntime = (runtime: string) => {
    switch (runtime) {
      case "Python":
        return fileExtensions.python;
      case "Node.js":
        return fileExtensions.nodeJs;
      case "Go":
        return fileExtensions.go;
      default:
        return "";
    }
  };

  return (
    <Modal open={open} onClose={onClose}>
      <Box
        sx={{
          position: "absolute",
          top: "50%",
          left: "50%",
          transform: "translate(-50%, -50%)",
          width: 700,
          maxHeight: "90vh",
          overflowY: "auto",
          bgcolor: "background.paper",
          boxShadow: 24,
          p: 4,
          borderRadius: 1,
        }}
      >
        <Typography variant="h6" component="h2" mb={2} color="#000">
          Add New Function App
        </Typography>
        <Tabs value={tabValue} onChange={handleTabChange} sx={{ mb: 2 }}>
          <Tab label="Form" value="form" />
          <Tab label="JSON" value="json" />
        </Tabs>
        {tabValue === "form" && (
          <form onSubmit={handleSubmit(onSubmit)}>
            <TextField
              label="Function App Name"
              fullWidth
              margin="normal"
              {...register("name", { required: true })}
            />
            <FormControl fullWidth margin="normal">
              <InputLabel id="runtime-label">Runtime</InputLabel>
              <Select
                {...register("runtime", { required: true })}
                labelId="runtime-label"
                defaultValue=""
              >
                <MenuItem value="Node.js">Node.js</MenuItem>
                <MenuItem value="Python">Python</MenuItem>
                <MenuItem value="Go">Go</MenuItem>
              </Select>
            </FormControl>
            <TextField
              label="Latency Limit (ms)"
              type="number"
              fullWidth
              margin="normal"
              {...register("latencyLimit", {
                required: true,
                valueAsNumber: true,
              })}
            />
            <Input
              type="file"
              fullWidth
              {...register("files", { required: true })}
              inputProps={{ multiple: true }}
              onChange={handleFilesChange}
            />

            {/* Components Section */}
            <Box mt={3}>
              <Typography variant="subtitle1">Components</Typography>
              {componentFields.map((field, index) => {
                // Get runtime from react-hook-form state
                const runtime = control._formValues?.runtime || "";
                const ext = getExtensionForRuntime(runtime);
                // Only show files that do NOT end with the extension for files select
                const filteredFileNames = uploadedFileNames.filter(
                  (fname) => !fname.endsWith(ext)
                );
                // Only show files that DO end with the extension for name select
                const nameFileNames = uploadedFileNames.filter((fname) =>
                  fname.endsWith(ext)
                );
                return (
                  <Box
                    key={field.id}
                    display="flex"
                    gap={2}
                    alignItems="center"
                    mt={1}
                  >
                    {/* Name select: only files with matching extension, show name without extension */}
                    <Controller
                      control={control}
                      name={`components.${index}.name`}
                      rules={{ required: true }}
                      render={({ field }) => (
                        <FormControl sx={{ minWidth: "30%", maxWidth: "40%" }}>
                          <InputLabel id={`component-name-label-${index}`}>
                            Name
                          </InputLabel>
                          <Select
                            labelId={`component-name-label-${index}`}
                            {...field}
                            value={field.value || ""}
                            onChange={(e) => field.onChange(e.target.value)}
                          >
                            {nameFileNames.map((fname) => {
                              const name = fname.replace(ext, "");
                              return (
                                <MenuItem key={fname} value={name}>
                                  {name}
                                </MenuItem>
                              );
                            })}
                          </Select>
                        </FormControl>
                      )}
                    />
                    <TextField
                      label="Memory Limit"
                      type="number"
                      sx={{ minWidth: "10%", maxWidth: "15%" }}
                      {...register(`components.${index}.memory`, {
                        required: true,
                        valueAsNumber: true,
                      })}
                    />
                    <TextField
                      label="Runtime (ms)"
                      type="number"
                      sx={{ minWidth: "10%", maxWidth: "15%" }}
                      {...register(`components.${index}.runtime`, {
                        required: true,
                        valueAsNumber: true,
                      })}
                    />
                    {/* MultiSelect for files */}
                    <Controller
                      control={control}
                      name={`components.${index}.files`}
                      render={({ field }) => (
                        <FormControl sx={{ minWidth: "30%", maxWidth: "40%" }}>
                          <InputLabel id={`component-files-label-${index}`}>
                            Files
                          </InputLabel>
                          <Select
                            labelId={`component-files-label-${index}`}
                            multiple
                            {...field}
                            input={<OutlinedInput label="Files" />}
                            renderValue={(selected) => (
                              <Box
                                sx={{
                                  display: "flex",
                                  flexWrap: "wrap",
                                  gap: 0.5,
                                }}
                              >
                                {(selected as string[]).map((value) => (
                                  <Chip key={value} label={value} />
                                ))}
                              </Box>
                            )}
                          >
                            {filteredFileNames.map((fname) => (
                              <MenuItem key={fname} value={fname}>
                                <Checkbox
                                  checked={
                                    field.value?.includes(fname) || false
                                  }
                                />
                                {fname}
                              </MenuItem>
                            ))}
                          </Select>
                        </FormControl>
                      )}
                    />
                    <IconButton onClick={() => removeComponent(index)}>
                      <DeleteIcon />
                    </IconButton>
                  </Box>
                );
              })}
              <Button
                startIcon={<AddIcon />}
                sx={{ mt: 1 }}
                onClick={() =>
                  appendComponent({
                    name: "",
                    memory: 0,
                    runtime: 0,
                    files: [],
                  })
                }
              >
                Add Component
              </Button>
            </Box>

            {/* Links Section */}
            <Box mt={3}>
              <Typography variant="subtitle1">Links</Typography>
              {linkFields.map((field, index) => {
                // Get runtime from react-hook-form state
                const runtime = control._formValues?.runtime || "";
                const ext = getExtensionForRuntime(runtime);
                // Only show files that DO end with the extension for from/to selects
                const nameFileNames = uploadedFileNames.filter((fname) =>
                  fname.endsWith(ext)
                );
                return (
                  <Box
                    key={field.id}
                    display="flex"
                    gap={2}
                    alignItems="center"
                    mt={1}
                  >
                    <Controller
                      control={control}
                      name={`links.${index}.from`}
                      rules={{ required: true }}
                      render={({ field }) => (
                        <FormControl sx={{ minWidth: "30%", maxWidth: "40%" }}>
                          <InputLabel id={`link-from-label-${index}`}>
                            From
                          </InputLabel>
                          <Select
                            labelId={`link-from-label-${index}`}
                            {...field}
                            value={field.value || ""}
                            onChange={(e) => field.onChange(e.target.value)}
                          >
                            {nameFileNames.map((fname) => {
                              const name = fname.replace(ext, "");
                              return (
                                <MenuItem key={fname} value={name}>
                                  {name}
                                </MenuItem>
                              );
                            })}
                          </Select>
                        </FormControl>
                      )}
                    />
                    <Controller
                      control={control}
                      name={`links.${index}.to`}
                      rules={{ required: true }}
                      render={({ field }) => (
                        <FormControl sx={{ minWidth: "30%", maxWidth: "40%" }}>
                          <InputLabel id={`link-to-label-${index}`}>
                            To
                          </InputLabel>
                          <Select
                            labelId={`link-to-label-${index}`}
                            {...field}
                            value={field.value || ""}
                            onChange={(e) => field.onChange(e.target.value)}
                          >
                            {nameFileNames.map((fname) => {
                              const name = fname.replace(ext, "");
                              return (
                                <MenuItem key={fname} value={name}>
                                  {name}
                                </MenuItem>
                              );
                            })}
                          </Select>
                        </FormControl>
                      )}
                    />
                    <TextField
                      label="Invocation Rate"
                      type="number"
                      sx={{ minWidth: "10%", maxWidth: "15%" }}
                      {...register(`links.${index}.invocationRate`, {
                        required: true,
                        valueAsNumber: true,
                      })}
                    />
                    <IconButton onClick={() => removeLink(index)}>
                      <DeleteIcon />
                    </IconButton>
                  </Box>
                );
              })}
              <Button
                startIcon={<AddIcon />}
                sx={{ mt: 1 }}
                onClick={() =>
                  appendLink({ from: "", to: "", invocationRate: 0 })
                }
              >
                Add Link
              </Button>
            </Box>
            <FormControlLabel
              control={<Checkbox {...register("platformManaged")} />}
              label="Platform managed"
              sx={{ mt: 1 }}
            />

            <Box mt={3} display="flex" justifyContent="flex-end">
              <Button onClick={onClose} sx={{ mr: 1 }}>
                Cancel
              </Button>
              <Button type="submit" variant="contained" color="primary">
                Add
              </Button>
            </Box>
          </form>
        )}
        {tabValue === "json" && <FunctionAppJsonForm onClose={onClose} />}
      </Box>
    </Modal>
  );
};

export default AddFunctionAppModal;
