import React from "react";
import { TextField, Button, Box, Typography, Input } from "@mui/material";
import { useBulkCreateFunctionApp } from "../hooks/functionAppsHooks";
import { useForm } from "react-hook-form";
import type { BulkCreateRequest } from "../models/dto";

interface FunctionAppJsonFormProps {
  onClose: () => void;
}

interface FormData {
  json: string;
  files: FileList;
}

const FunctionAppJsonForm: React.FC<FunctionAppJsonFormProps> = ({
  onClose,
}) => {
  const { register, handleSubmit } = useForm<FormData>({
    defaultValues: {
      json: `{
  "functionApp": {
    "name": "object-detection",
    "runtime": "python",
    "components": [
      { "name": "resize", "memory": 97, "runtime": 31, "files": [] },
      { "name": "grayscale", "memory": 92, "runtime": 52, "files": [] },
      {
        "name": "objectdetect",
        "memory": 592,
        "runtime": 1066,
        "files": [
          "MobileNetSSD_deploy.caffemodel",
          "MobileNetSSD_deploy.prototxt.txt"
        ]
      },
      { "name": "cut", "memory": 79, "runtime": 29, "files": [] },
      {
        "name": "objectdetect2",
        "memory": 592,
        "runtime": 1041,
        "files": [
          "MobileNetSSD_deploy.caffemodel",
          "MobileNetSSD_deploy.prototxt.txt"
        ]
      },
      { "name": "tag", "memory": 83, "runtime": 33, "files": [] }
    ],
    "links": [
      {
        "from": "resize",
        "to": "grayscale",
        "invocationRate": 2.0,
        "dataDelay": 3
      },
      {
        "from": "grayscale",
        "to": "objectdetect",
        "invocationRate": 2.0,
        "dataDelay": 3
      },
      {
        "from": "objectdetect",
        "to": "cut",
        "invocationRate": 2.0,
        "dataDelay": 3
      },
      {
        "from": "cut",
        "to": "objectdetect2",
        "invocationRate": 4.0,
        "dataDelay": 3
      },
      {
        "from": "objectdetect2",
        "to": "tag",
        "invocationRate": 4.0,
        "dataDelay": 3
      }
    ],
    "latencyLimit": 2252,
    "platformManaged": true
  },
  "functionCompositions": [],
  "deployments": []
}`,
    },
  });
  const { mutate: bulkCreateFunctionApp } = useBulkCreateFunctionApp();

  const onSubmit = (data: FormData) => {
    const request: BulkCreateRequest = JSON.parse(data.json);
    bulkCreateFunctionApp({ req: request, files: data.files });
    onClose();
  };

  return (
    <Box>
      <Typography variant="body1" mb={2}>
        Enter Function App JSON:
      </Typography>
      <form onSubmit={handleSubmit(onSubmit)}>
        <TextField
          multiline
          rows={10}
          fullWidth
          {...register("json", { required: true })}
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
