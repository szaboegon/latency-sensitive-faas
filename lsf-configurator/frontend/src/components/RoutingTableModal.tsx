import React, { useState } from "react";
import {
  Modal,
  Box,
  Tabs,
  Tab,
  Typography,
  Button,
  Stack,
  Select,
  MenuItem,
  TextField,
  IconButton,
} from "@mui/material";
import { useForm, useFieldArray, useWatch, Controller } from "react-hook-form";
import type { FunctionComposition } from "../models/models";
import DeleteIcon from "@mui/icons-material/Delete";

interface Props {
  open: boolean;
  onClose: () => void;
  composition: FunctionComposition;
  allCompositions: FunctionComposition[];
}

interface Rule {
  component: string;
  targetComposition: string | null;
  targetComponent: string | null;
}

const RoutingTableModal: React.FC<Props> = ({
  open,
  onClose,
  composition,
  allCompositions,
}) => {
  const [tab, setTab] = useState(0);

  const { control, handleSubmit } = useForm<{ rules: Rule[] }>({
    defaultValues: {
      rules: Object.entries(composition.components).flatMap(
        ([component, routes]) =>
          routes.map((route) => ({
            component,
            targetComposition: route.function ?? null,
            targetComponent: route.to ?? null,
          }))
      ),
    },
  });

  const { fields, append, remove } = useFieldArray({
    control,
    name: "rules",
  });

  const watchedRules = useWatch({
    control,
    name: "rules",
  });

  const handleTabChange = (_: React.SyntheticEvent, newValue: number) =>
    setTab(newValue);

  const handleAddRule = () => {
    append({ component: "", targetComposition: "", targetComponent: "" });
  };

  const handleFormSubmit = (data: { rules: Rule[] }) => {
    console.log("Form Data:", data);
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
          width: 900,
          bgcolor: "background.paper",
          boxShadow: 24,
          p: 4,
          borderRadius: 2,
        }}
      >
        <Typography variant="h6" mb={2}>
          Edit Routing Table
        </Typography>

        <Tabs value={tab} onChange={handleTabChange}>
          <Tab label="JSON Input" />
          <Tab label="Form Input" />
        </Tabs>

        {tab === 0 && (
          <Box mt={2}>
            <TextField
              fullWidth
              multiline
              rows={10}
              placeholder="Enter routing table in JSON format"
              variant="outlined"
            />
          </Box>
        )}

        {tab === 1 && (
          <Box mt={2}>
            <form onSubmit={handleSubmit(handleFormSubmit)}>
              <Stack spacing={2}>
               {fields.map((field, index) => {
  const selectedTargetCompId = watchedRules?.[index]?.targetComposition;
  const availableTargetComponents = Object.keys(
    allCompositions.find((comp) => comp.id === selectedTargetCompId)?.components ?? {}
  );

  return (
    <Stack key={field.id} direction="row" spacing={2} alignItems="center">
      {/* Current composition component */}
      <Controller
        name={`rules.${index}.component`}
        control={control}
        render={({ field }) => (
          <Select {...field} fullWidth>
            {Object.keys(composition.components).map((component) => (
              <MenuItem key={component} value={component}>
                {component}
              </MenuItem>
            ))}
          </Select>
        )}
      />

      {/* Arrow */}
      <Typography variant="body2" sx={{ mx: 1 }}>
        →
      </Typography>

      {/* Target composition */}
      <Controller
        name={`rules.${index}.targetComposition`}
        control={control}
        render={({ field }) => (
          <Select {...field} fullWidth>
            {allCompositions.map((comp) => (
              <MenuItem key={comp.id} value={comp.id}>
                {comp.id} ({comp.node})
              </MenuItem>
            ))}
          </Select>
        )}
      />

      {/* Target component */}
      <Controller
        name={`rules.${index}.targetComponent`}
        control={control}
        render={({ field }) => (
          <Select {...field} fullWidth>
            {availableTargetComponents.map((component) => (
              <MenuItem key={component} value={component}>
                {component}
              </MenuItem>
            ))}
          </Select>
        )}
      />

      {/* Delete */}
      <IconButton onClick={() => remove(index)} color="error">
        <DeleteIcon />
      </IconButton>
    </Stack>
  );
})}

                <Button variant="outlined" onClick={handleAddRule}>
                  + Add Rule
                </Button>
                <Button type="submit" variant="contained" color="primary">
                  Save
                </Button>
              </Stack>
            </form>
          </Box>
        )}
      </Box>
    </Modal>
  );
};

export default RoutingTableModal;
