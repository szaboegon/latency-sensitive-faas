import React from "react";
import { Stack, Select, MenuItem, Typography, Button, IconButton } from "@mui/material";
import { useForm, useFieldArray, Controller } from "react-hook-form";
import DeleteIcon from "@mui/icons-material/Delete";
import type { Deployment, RoutingTable } from "../models/models";

interface RoutingTableFormProps {
  deployment: Deployment;
  allDeployments: Deployment[];
  onChange: (routingTable: RoutingTable) => void;
}

interface Rule {
  component: string;
  targetDeployment: string;
  targetComponent: string;
}

const RoutingTableEditor: React.FC<RoutingTableFormProps> = ({
  deployment,
  allDeployments,
  onChange,
}) => {
  const { control, watch } = useForm<{ rules: Rule[] }>({
    defaultValues: {
      rules: Object.entries(deployment.routingTable).flatMap(
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
        function: rule.targetDeployment === "None" ? "" : rule.targetDeployment,
      });
      return acc;
    }, {} as RoutingTable);

    onChange(routingTable);
  };

  const handleAddRule = () => {
    append({ component: "", targetDeployment: "", targetComponent: "" });
    notifyChange();
  };

  const handleRemoveRule = (index: number) => {
    remove(index);
    notifyChange();
  };

  return (
    <Stack spacing={2}>
      {fields.map((field, index) => {
        const selectedTargetDepId = watchedRules?.[index]?.targetDeployment;
        const availableTargetComponents = Object.keys(
          allDeployments.find((dep) => dep.id === selectedTargetDepId)?.routingTable ?? {}
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
                  {Object.keys(deployment.routingTable).map((component) => (
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

            {/* Target deployment */}
            <Controller
              name={`rules.${index}.targetDeployment`}
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
                  {allDeployments.map((dep) => (
                    <MenuItem key={dep.id} value={dep.id}>
                      {dep.id} ({dep.node})
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