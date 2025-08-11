import React from "react"
import { useForm } from "react-hook-form"
import {
  Modal,
  Box,
  Typography,
  TextField,
  Button,
  Input,
} from "@mui/material"

interface AddFunctionAppModalProps {
  open: boolean
  onClose: () => void
}

interface FormData {
  name: string
  files: FileList
}

const AddFunctionAppModal: React.FC<AddFunctionAppModalProps> = ({
  open,
  onClose,
}) => {
  const { register, handleSubmit } = useForm<FormData>()

  const onSubmit = (data: FormData) => {
    console.log(data)
    onClose()
  }

  return (
    <Modal open={open} onClose={onClose}>
      <Box
        sx={{
          position: "absolute",
          top: "50%",
          left: "50%",
          transform: "translate(-50%, -50%)",
          width: 400,
          bgcolor: "background.paper",
          boxShadow: 24,
          p: 4,
          borderRadius: 1,
        }}
      >
        <Typography variant="h6" component="h2" mb={2} color="#000">
          Add New Function App
        </Typography>
        <form onSubmit={handleSubmit(onSubmit)}>
          <TextField
            label="Function App Name"
            fullWidth
            margin="normal"
            {...register("name", { required: true })}
          />
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
      </Box>
    </Modal>
  )
}

export default AddFunctionAppModal
