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
      { "name": "resize", "memory": 97, "runtime": 31 },
      { "name": "grayscale", "memory": 92, "runtime": 52 },
      { "name": "objectdetect", "memory": 592, "runtime": 1066 },
      { "name": "cut", "memory": 79, "runtime": 29 },
      { "name": "objectdetect2", "memory": 592, "runtime": 1041 },
      { "name": "tag", "memory": 83, "runtime": 33 }
    ],
    "links": [
      { "from": "resize", "to": "grayscale", "invocationRate": 1.0 },
      { "from": "grayscale", "to": "objectdetect", "invocationRate": 1.0 },
      { "from": "objectdetect", "to": "cut", "invocationRate": 1.0 },
      { "from": "cut", "to": "objectdetect2", "invocationRate": 3.0 },
      { "from": "objectdetect2", "to": "tag", "invocationRate": 1.0 }
    ],
    "latencyLimit": 2252,
    "platformManaged": false
  },
  "functionCompositions": [
    {
      "id": "composition-1",
      "components": ["resize"],
      "files": []
    },
    {
      "id": "composition-2",
      "components": ["grayscale", "cut"],
      "files": []
    },
    {
      "id": "composition-3",
      "components": ["objectdetect", "objectdetect2"],
      "files": [
        "MobileNetSSD_deploy.caffemodel",
        "MobileNetSSD_deploy.prototxt.txt"
      ]
    },
    {
      "id": "composition-4",
      "components": ["objectdetect2", "tag"],
      "files": [
        "MobileNetSSD_deploy.caffemodel",
        "MobileNetSSD_deploy.prototxt.txt"
      ]
    }
  ],
  "deployments": [
    {
      "id": "deployment-1",
      "functionCompositionId": "composition-1",
      "node": "knative",
      "namespace": "application",
      "routingTable": {
        "resize": [{ "to": "grayscale", "function": "deployment-2" }]
      }
    },
    {
      "id": "deployment-2",
      "functionCompositionId": "composition-2",
      "node": "knative-m02",
      "namespace": "application",
      "routingTable": {
        "grayscale": [{ "to": "objectdetect", "function": "deployment-3" }],
        "cut": [{ "to": "objectdetect2", "function": "deployment-3" }]
      }
    },
    {
      "id": "deployment-3",
      "functionCompositionId": "composition-3",
      "node": "knative-m02",
      "namespace": "application",
      "routingTable": {
        "objectdetect": [{ "to": "cut", "function": "deployment-2" }],
        "objectdetect2": [{ "to": "tag", "function": "deployment-4" }]
      }
    },
    {
      "id": "deployment-4",
      "functionCompositionId": "composition-4",
      "node": "knative-m03",
      "namespace": "application",
      "routingTable": {
        "objectdetect2": [{ "to": "tag", "function": "local" }],
        "tag": []
      }
    }
  ]
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
