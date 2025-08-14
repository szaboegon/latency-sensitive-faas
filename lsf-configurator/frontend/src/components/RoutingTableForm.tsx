import React from "react";
import { Stack, Select, MenuItem, Typography, Button, IconButton } from "@mui/material";
import { useForm, useFieldArray, useWatch, Controller } from "react-hook-form";
import DeleteIcon from "@mui/icons-material/Delete";
import type { FunctionComposition } from "../models/models";

interface RoutingTableFormProps {
  composition: FunctionComposition;
  allCompositions: FunctionComposition[];
  onSave: (rules: Rule[]) => void;
}

interface Rule {
  component: string;
  targetComposition: string;
  targetComponent: string;
}

const RoutingTableForm: React.FC<RoutingTableFormProps> = ({
  composition,
  allCompositions,
  onSave,
}) => {
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
            : [{ component, targetComposition: "", targetComponent: "" }]
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

  const handleAddRule = () => {
    append({ component: "", targetComposition: "", targetComponent: "" });
  };

  return (
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
        <Button
          type="submit"
          variant="contained"
          color="primary"
          onClick={handleSubmit((data) => onSave(data.rules))}
        >
          Save
        </Button>
      </Stack>
    </form>
  );
};

export default RoutingTableForm;
