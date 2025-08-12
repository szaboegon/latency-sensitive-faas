import React, { useState } from "react";
import { Modal, Box, Tabs, Tab, Typography, Button, Stack, Select, MenuItem, TextField } from "@mui/material";
import { useForm, useFieldArray } from "react-hook-form";
import type { FunctionComposition } from "../models/models";

interface Props {
  open: boolean;
  onClose: () => void;
  composition: FunctionComposition;
}

const RoutingTableModal: React.FC<Props> = ({ open, onClose, composition }) => {
  const [tab, setTab] = useState(0);
  const { control, handleSubmit } = useForm({
    defaultValues: {
      rules: [],
    },
  });
  const { fields, append } = useFieldArray({
    control,
    name: "rules",
  });

  const handleTabChange = (_: React.SyntheticEvent, newValue: number) => setTab(newValue);

  const handleAddRule = () => {
    append({ component: "", to: "" });
  };

  const handleFormSubmit = (data: any) => {
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
          width: 600,
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
                {fields.map((field, index) => (
                  <Stack key={field.id} direction="row" spacing={2}>
                    <Select
                      fullWidth
                      defaultValue=""
                      {...control.register(`rules.${index}.component`)}
                    >
                      {Object.keys(composition.components).map((component) => (
                        <MenuItem key={component} value={component}>
                          {component}
                        </MenuItem>
                      ))}
                    </Select>
                    <Select
                      fullWidth
                      defaultValue=""
                      {...control.register(`rules.${index}.to`)}
                    >
                      {/* Replace with actual function compositions */}
                      <MenuItem value="FunctionA">FunctionA</MenuItem>
                      <MenuItem value="FunctionB">FunctionB</MenuItem>
                    </Select>
                  </Stack>
                ))}
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
