import React, { useState } from "react";
import { TextField, Button, Box, Typography, Input } from "@mui/material";
import { useBulkCreateFunctionApp } from "../hooks/functionAppsHooks";
import { useForm } from "react-hook-form";
import type { BulkCreateRequest } from "../models/dto";

interface FunctionAppJsonFormProps {
  onClose: () => void;
}

interface FormData {
  json: string
  files: FileList
}


const FunctionAppJsonForm: React.FC<FunctionAppJsonFormProps> = ({ onClose }) => {
    const { register, handleSubmit } = useForm<FormData>()
  const { mutate: bulkCreateFunctionApp } = useBulkCreateFunctionApp();

  const onSubmit = (data: FormData) => {
    const request: BulkCreateRequest = JSON.parse(data.json);
    bulkCreateFunctionApp({ req: request, files: data.files });
    onClose();
  }


  return (
    <Box>
      <Typography variant="body1" mb={2}>
        Enter Function App JSON:
      </Typography>
      <form onSubmit={handleSubmit(onSubmit) }>
      <TextField
        multiline
        rows={10}
        fullWidth
        {...register("json", { required: true })}
        placeholder='TODO'
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
        <Button variant="contained" color="primary" type="submit">
          Add
        </Button>
      </Box>
      </form>
    </Box>
  );
};

export default FunctionAppJsonForm;
