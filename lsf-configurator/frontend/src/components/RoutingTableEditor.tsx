import React from "react";
import { Stack, Select, MenuItem, Typography, Button, IconButton } from "@mui/material";
import { useForm, useFieldArray, Controller } from "react-hook-form";
import DeleteIcon from "@mui/icons-material/Delete";
import type { FunctionComposition, RoutingTable } from "../models/models";

interface RoutingTableFormProps {
  composition: FunctionComposition;
  allCompositions: FunctionComposition[];
  onChange: (routingTable: RoutingTable) => void;
}

interface Rule {
  component: string;
  targetComposition: string;
  targetComponent: string;
}

const RoutingTableEditor: React.FC<RoutingTableFormProps> = ({
  composition,
  allCompositions,
  onChange,
}) => {
  const { control, watch } = useForm<{ rules: Rule[] }>({
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

  const watchedRules = watch("rules");

  const notifyChange = () => {
    const routingTable: RoutingTable = watchedRules.reduce((acc, rule) => {
      if (!acc[rule.component]) {
        acc[rule.component] = [];
      }
      acc[rule.component].push({
        to: rule.targetComponent === "None" ? "" : rule.targetComponent,
        function: rule.targetComposition === "None" ? "" : rule.targetComposition,
      });
      return acc;
    }, {} as RoutingTable);

    onChange(routingTable);
  };

  const handleAddRule = () => {
    append({ component: "", targetComposition: "", targetComponent: "" });
    notifyChange();
  };

  const handleRemoveRule = (index: number) => {
    remove(index);
    notifyChange();
  };

  return (
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
                <Select
                  {...field}
                  fullWidth
                  onChange={(e) => {
                    field.onChange(e);
                    notifyChange();
                  }}
                >
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
                <Select
                  {...field}
                  fullWidth
                  value={field.value || "None"}
                  onChange={(e) => {
                    field.onChange(e);
                    notifyChange();
                  }}
                >
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
                <Select
                  {...field}
                  fullWidth
                  value={field.value || "None"}
                  onChange={(e) => {
                    field.onChange(e);
                    notifyChange();
                  }}
                >
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
            <IconButton onClick={() => handleRemoveRule(index)} color="error">
              <DeleteIcon />
            </IconButton>
          </Stack>
        );
      })}

      <Button variant="outlined" onClick={handleAddRule}>
        + Add Rule
      </Button>
    </Stack>
  );
};

export default RoutingTableEditor;