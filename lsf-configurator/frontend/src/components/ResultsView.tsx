import React, { useState } from "react";
import {
  Box,
  Typography,
  Paper,
  CircularProgress,
  Alert,
  Stack,
  Button,
  IconButton,
  Tooltip,
} from "@mui/material";
import ContentCopyIcon from "@mui/icons-material/ContentCopy";
import { useGetAppResults } from "../hooks/resultsHooks";

interface ResultsViewProps {
  appId: string;
  count?: number;
}

const LONG_STRING_LIMIT = 100;

const ResultString: React.FC<{ value: string }> = ({ value }) => {
  const [expanded, setExpanded] = useState(false);
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      setTimeout(() => setCopied(false), 1200);
    } catch {
      setCopied(false);
    }
  };

  if (value.length > LONG_STRING_LIMIT) {
    return (
      <Box sx={{ position: "relative" }}>
        <Box
          component="pre"
          sx={{
            whiteSpace: "pre-wrap",
            wordBreak: "break-word",
            fontSize: 14,
            bgcolor: "#fffde7",
            p: 1,
            borderRadius: 1,
          }}
        >
          {expanded ? value : value.slice(0, LONG_STRING_LIMIT) + "..."}{" "}
        </Box>
        <Box sx={{ display: "flex", gap: 1, mt: 0.5 }}>
          <Button
            size="small"
            variant="text"
            onClick={() => setExpanded((e) => !e)}
          >
            {expanded ? "Show less" : "Show more"}
          </Button>
          <Tooltip title={copied ? "Copied!" : "Copy"}>
            <IconButton size="small" onClick={handleCopy}>
              <ContentCopyIcon fontSize="small" />
            </IconButton>
          </Tooltip>
        </Box>
      </Box>
    );
  }
  return (
    <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
      <Box
        component="pre"
        sx={{
          whiteSpace: "pre-wrap",
          wordBreak: "break-word",
          fontSize: 14,
        }}
      >
        {value}
      </Box>
      <Tooltip title={copied ? "Copied!" : "Copy"}>
        <IconButton size="small" onClick={handleCopy}>
          <ContentCopyIcon fontSize="small" />
        </IconButton>
      </Tooltip>
    </Box>
  );
};

const ResultsView: React.FC<ResultsViewProps> = ({ appId, count = 5 }) => {
  const { data, isLoading, error } = useGetAppResults(appId, count);

  if (isLoading) {
    return (
      <Box
        display="flex"
        justifyContent="center"
        alignItems="center"
        minHeight={200}
      >
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Alert severity="error">Failed to load results: {error.message}</Alert>
    );
  }

  if (!data || !Array.isArray(data) || data.length === 0) {
    return <Typography>No results available.</Typography>;
  }

  return (
    <Stack spacing={2}>
      {data.map((result, idx) => (
        <Paper key={idx} sx={{ p: 2, background: "#f5faff" }}>
          <Typography variant="subtitle2" color="text.secondary" gutterBottom>
            {result.timestamp}
          </Typography>
          <ResultString value={result.event} />
        </Paper>
      ))}
    </Stack>
  );
};

export default ResultsView;
