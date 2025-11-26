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
      { "name": "resize", "memory": 97, "runtime": 12, "files": [] },
      { "name": "grayscale", "memory": 86, "runtime": 2, "files": [] },
      {
        "name": "objectdetect",
        "memory": 153,
        "runtime": 256,
        "files": [
          "MobileNetSSD_deploy.caffemodel",
          "MobileNetSSD_deploy.prototxt.txt"
        ]
      },
      { "name": "cut", "memory": 80, "runtime": 12, "files": [] },
      {
        "name": "objectdetect2",
        "memory": 325,
        "runtime": 1016,
        "files": [
          "MobileNetSSD_deploy.caffemodel",
          "MobileNetSSD_deploy.prototxt.txt"
        ]
      },
      { "name": "tag", "memory": 120, "runtime": 24, "files": [] }
    ],
    "links": [
      {
        "from": "resize",
        "to": "grayscale",
        "invocationRate": {
          "min": 1.0,
          "max": 4.0
        },
        "dataDelay": 10
      },
      {
        "from": "grayscale",
        "to": "objectdetect",
        "invocationRate": {
          "min": 1.0,
          "max": 4.0
        },
        "dataDelay": 10
      },
      {
        "from": "objectdetect",
        "to": "cut",
        "invocationRate": {
          "min": 1.0,
          "max": 4.0
        },
        "dataDelay": 10
      },
      {
        "from": "cut",
        "to": "objectdetect2",
        "invocationRate": {
          "min": 2.0,
          "max": 8.0
        },
        "dataDelay": 10
      },
      {
        "from": "objectdetect2",
        "to": "tag",
        "invocationRate": {
          "min": 2.0,
          "max": 8.0
        },
        "dataDelay": 10
      }
    ],
    "latencyLimit": 1600,
    "platformManaged": true
  },
  "functionCompositions": [],
  "deployments": []
}
`,
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
