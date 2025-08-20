import React, { useState } from "react"
import { useForm } from "react-hook-form"
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
} from "@mui/material"
import { useCreateFunctionApp } from "../hooks/functionAppsHooks"
import FunctionAppJsonForm from "./FunctionAppJsonForm"
import type { FunctionAppCreateDto } from "../models/dto"

interface AddFunctionAppModalProps {
  open: boolean
  onClose: () => void
}

interface FormData {
  name: string
  files: FileList
  runtime: string
}

const AddFunctionAppModal: React.FC<AddFunctionAppModalProps> = ({
  open,
  onClose,
}) => {
  const { register, handleSubmit } = useForm<FormData>()
  const { mutate: createFunctionApp } = useCreateFunctionApp()
  const [tabValue, setTabValue] = useState<"form" | "json">("form")

  const onSubmit = (data: FormData) => {
    const fcApp: FunctionAppCreateDto = {
      name: data.name,
      runtime: data.runtime,
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
                label="Runtime"
                defaultValue=""
              >
                <MenuItem value="Node.js">Node.js</MenuItem>
                <MenuItem value="Python">Python</MenuItem>
                <MenuItem value="Go">Go</MenuItem>
              </Select>
            </FormControl>
            <Input
              type="file"
              fullWidth
              {...register("files", { required: true })}
              inputProps={{ multiple: true }}
            />
            <Box mt={2} display="flex" justifyContent="flex-end">
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
