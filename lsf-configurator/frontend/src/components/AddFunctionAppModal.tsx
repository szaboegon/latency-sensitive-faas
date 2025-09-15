import React, { useState } from "react"
import { useForm, useFieldArray } from "react-hook-form"
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
} from "@mui/material"
import { useCreateFunctionApp } from "../hooks/functionAppsHooks"
import FunctionAppJsonForm from "./FunctionAppJsonForm"
import type { FunctionAppCreateDto } from "../models/dto"
import type { Component, ComponentLink } from "../models/models"
import AddIcon from "@mui/icons-material/Add"
import DeleteIcon from "@mui/icons-material/Delete"

interface AddFunctionAppModalProps {
  open: boolean
  onClose: () => void
}

interface FormData {
  name: string
  files: FileList
  runtime: string
  latencyLimit: number
  components: Component[]
  links: ComponentLink[]
}

const AddFunctionAppModal: React.FC<AddFunctionAppModalProps> = ({
  open,
  onClose,
}) => {
  const { register, handleSubmit, control } = useForm<FormData>({
    defaultValues: {
      components: [],
      links: [],
    },
  })

  const {
    fields: componentFields,
    append: appendComponent,
    remove: removeComponent,
  } = useFieldArray({ control, name: "components" })

  const {
    fields: linkFields,
    append: appendLink,
    remove: removeLink,
  } = useFieldArray({ control, name: "links" })

  const { mutate: createFunctionApp } = useCreateFunctionApp()
  const [tabValue, setTabValue] = useState<"form" | "json">("form")

  const onSubmit = (data: FormData) => {
    const fcApp: FunctionAppCreateDto = {
      name: data.name,
      runtime: data.runtime,
      latencyLimit: data.latencyLimit,
      components: data.components,
      links: data.links,
    }
    createFunctionApp({ functionApp: fcApp, files: data.files })
    onClose()
  }

  const handleTabChange = (
    _: React.SyntheticEvent,
    newValue: "form" | "json"
  ) => {
    setTabValue(newValue)
  }

  return (
    <Modal open={open} onClose={onClose}>
      <Box
        sx={{
          position: "absolute",
          top: "50%",
          left: "50%",
          transform: "translate(-50%, -50%)",
          width: 600,
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
              {...register("latencyLimit", { required: true, valueAsNumber: true })}
            />
            <Input
              type="file"
              fullWidth
              {...register("files", { required: true })}
              inputProps={{ multiple: true }}
            />

            {/* Components Section */}
            <Box mt={3}>
              <Typography variant="subtitle1">Components</Typography>
              {componentFields.map((field, index) => (
                <Box key={field.id} display="flex" gap={2} alignItems="center" mt={1}>
                  <TextField
                    label="Name"
                    {...register(`components.${index}.name`, { required: true })}
                  />
                  <TextField
                    label="Memory Limit"
                    type="number"
                    {...register(`components.${index}.memory`, { required: true, valueAsNumber: true })}
                  />
                  <TextField
                    label="Runtime (ms)"
                    type="number"
                    {...register(`components.${index}.runtime`, { required: true, valueAsNumber: true })}
                  />
                  <IconButton onClick={() => removeComponent(index)}>
                    <DeleteIcon />
                  </IconButton>
                </Box>
              ))}
              <Button
                startIcon={<AddIcon />}
                sx={{ mt: 1 }}
                onClick={() =>
                  appendComponent({ name: "", memory: 0, runtime: 0 })
                }
              >
                Add Component
              </Button>
            </Box>

            {/* Links Section */}
            <Box mt={3}>
              <Typography variant="subtitle1">Links</Typography>
              {linkFields.map((field, index) => (
                <Box key={field.id} display="flex" gap={2} alignItems="center" mt={1}>
                  <TextField
                    label="From"
                    {...register(`links.${index}.from`, { required: true })}
                  />
                  <TextField
                    label="To"
                    {...register(`links.${index}.to`, { required: true })}
                  />
                  <TextField
                    label="Invocation Rate"
                    type="number"
                    {...register(`links.${index}.invocationRate`, {
                      required: true,
                      valueAsNumber: true,
                    })}
                  />
                  <IconButton onClick={() => removeLink(index)}>
                    <DeleteIcon />
                  </IconButton>
                </Box>
              ))}
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
  )
}

export default AddFunctionAppModal