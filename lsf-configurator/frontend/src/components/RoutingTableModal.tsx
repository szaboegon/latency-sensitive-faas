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
import type { FunctionComposition, RoutingTable } from "../models/models";
import DeleteIcon from "@mui/icons-material/Delete";
import { useModifyRoutingTable } from "../hooks/functionCompositionHooks";

interface Props {
  open: boolean;
  onClose: () => void;
  composition: FunctionComposition;
  allCompositions: FunctionComposition[];
}

interface Rule {
  component: string;
  targetComposition: string;
  targetComponent: string;
}

const RoutingTableModal: React.FC<Props> = ({
  open,
  onClose,
  composition,
  allCompositions,
}) => {
  const [tab, setTab] = useState(0);
  const [jsonInput, setJsonInput] = useState("");
  const { control, handleSubmit } = useForm<{ rules: Rule[] }>({
    defaultValues: {
      rules: Object.entries(composition.components).flatMap(
        ([component, routes]) =>
          routes.length > 0
            ? routes.map((route) => ({
                component,
                targetComposition: route.function ?? "",
                targetComponent: route.to ?? "",
              }))
            : [{
                component,
                targetComposition: "",
                targetComponent: ""
              }]
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

  const {mutate: modifyRoutingTable} = useModifyRoutingTable()

  const handleTabChange = (_: React.SyntheticEvent, newValue: number) =>
    setTab(newValue);

  const handleAddRule = () => {
    append({ component: "", targetComposition: "", targetComponent: "" });
  };

  const handleSaveFormInput = (rules: Rule[]) => {
    const routingTable: RoutingTable = rules.reduce((acc, rule) => {
      if (!acc[rule.component]) {
        acc[rule.component] = [];
      }
      acc[rule.component].push({
        to: rule.targetComponent === "None" ? "" : rule.targetComponent,
        function: rule.targetComposition === "None" ? "" : rule.targetComposition,
      });
      return acc;
    }, {} as RoutingTable);

    modifyRoutingTable({
      functionCompositionId: composition.id!,
      routingTable,
    });
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
          <Tab label="Form Input" />
          <Tab label="JSON Input" />
        </Tabs>

        {tab === 0 && (
          <Box mt={2}>
            <form>
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
                          <Select {...field} fullWidth value={field.value || "None"}>
                            <MenuItem value="None">None</MenuItem>
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
                          <Select {...field} fullWidth value={field.value || "None"}>
                            <MenuItem value="None">None</MenuItem>
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
                <Button type="submit" variant="contained" color="primary" onClick={handleSubmit((data) => handleSaveFormInput(data.rules))}>
                  Save
                </Button>
              </Stack>
            </form>
          </Box>
        )}

        {tab === 1 && (
          <Box mt={2}>
            <TextField
              fullWidth
              multiline
              rows={10}
              placeholder="Enter routing table in JSON format"
              variant="outlined"
              onChange={(e) => setJsonInput(e.target.value)}
            />
            <Button
              variant="contained"
              color="primary"
              sx={{ mt: 2 }}
              onClick={() => {
                try {
                  const parsedTable: RoutingTable = JSON.parse(jsonInput);
                  modifyRoutingTable({
                    functionCompositionId: composition.id!,
                    routingTable: parsedTable,
                  });
                  onClose();
                } catch (error) {
                  console.error("Invalid JSON format:", error);
                }
              }}
            >
              Save
            </Button>
          </Box>
        )}
      </Box>
    </Modal>
  );
};

export default RoutingTableModal;
