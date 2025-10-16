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

function isBase64Image(str: string): boolean {
  return (
    /^data:image\/(png|jpeg|jpg|gif|webp);base64,/.test(str) ||
    /^[A-Za-z0-9+/=]{100,}$/.test(str)
  );
}

const ResultString: React.FC<{ value: string; isImage?: boolean }> = ({
  value,
  isImage,
}) => {
  const [expanded, setExpanded] = useState(false);
  const [copied, setCopied] = useState(false);
  const [showImage, setShowImage] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      setTimeout(() => setCopied(false), 1200);
    } catch {
      setCopied(false);
    }
  };

  if (isImage) {
    return (
      <Box sx={{ position: "relative" }}>
        <Box sx={{ display: "flex", gap: 1, mb: 1 }}>
          <Button
            size="small"
            variant="text"
            onClick={() => setShowImage((v) => !v)}
          >
            {showImage ? "Show as text" : "View as image"}
          </Button>
          <Tooltip title={copied ? "Copied!" : "Copy"}>
            <IconButton size="small" onClick={handleCopy}>
              <ContentCopyIcon fontSize="small" />
            </IconButton>
          </Tooltip>
        </Box>
        {showImage ? (
          <Box
            component="img"
            src={
              value.startsWith("data:image")
                ? value
                : `data:image/png;base64,${value}`
            }
            alt="Base64 result"
            sx={{
              maxWidth: "100%",
              maxHeight: 300,
              borderRadius: 1,
              bgcolor: "#fffde7",
              p: 1,
            }}
          />
        ) : (
          <Box
            component="pre"
            sx={{
              whiteSpace: "pre-wrap",
              wordBreak: "break-word",
              fontSize: 14,
              bgcolor: "#fffde7",
              p: 1,
              borderRadius: 1,
              maxHeight: 300,
              overflow: "auto",
            }}
          >
            {expanded ? value : value.slice(0, LONG_STRING_LIMIT) + "..."}{" "}
            <Button
              size="small"
              variant="text"
              onClick={() => setExpanded((e) => !e)}
            >
              {expanded ? "Show less" : "Show more"}
            </Button>
          </Box>
        )}
      </Box>
    );
  }

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
      {data.map((result, idx) => {
        // Try to detect if result.event.result is a base64 image string
        let resultString = "";
        let isImage = false;
        if (
          result.event &&
          typeof result.event === "object" &&
          "result" in result.event &&
          typeof result.event.result === "string"
        ) {
          resultString = result.event.result;
          isImage = isBase64Image(resultString);
        } else {
          resultString = JSON.stringify(result.event, null, 2);
        }

        return (
          <Paper key={idx} sx={{ p: 2, background: "#f5faff" }}>
            <Typography variant="subtitle2" color="text.secondary" gutterBottom>
              {result.timestamp}
            </Typography>
            <ResultString value={resultString} isImage={isImage} />
          </Paper>
        );
      })}
    </Stack>
  );
};

export default ResultsView;
