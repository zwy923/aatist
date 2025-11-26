import React from "react";
import { Box, Chip, Paper, Stack, Typography } from "@mui/material";

export default function OpportunitiesOverview({ opportunities, onSelect }) {
  return (
    <Paper
      sx={{
        p: 3,
        background: "rgba(7, 12, 30, 0.96)",
        border: "1px solid rgba(93, 224, 255, 0.25)",
        borderRadius: 3,
        height: "100%",
      }}
    >
      <Typography variant="h6" fontWeight={600} mb={2}>
        Opportunities Overview
      </Typography>
      <Stack spacing={2}>
        {opportunities.slice(0, 5).map((opp) => (
          <Box
            key={opp.id}
            sx={{
              p: 2,
              borderRadius: 2,
              background: "rgba(93, 224, 255, 0.05)",
              border: "1px solid rgba(93, 224, 255, 0.1)",
              cursor: "pointer",
              "&:hover": {
                background: "rgba(93, 224, 255, 0.1)",
                borderColor: "rgba(93, 224, 255, 0.3)",
              },
            }}
            onClick={() => onSelect(opp.id)}
          >
            <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
              <Box sx={{ flex: 1, minWidth: 0 }}>
                <Typography
                  variant="subtitle1"
                  fontWeight={600}
                  sx={{ color: "text.primary", mb: 0.5 }}
                >
                  {opp.title}
                </Typography>
                <Typography variant="body2" color="text.secondary" noWrap>
                  {opp.description || opp.company || ""}
                </Typography>
                {(opp.tags || []).length > 0 && (
                  <Stack direction="row" spacing={0.5} mt={1} flexWrap="wrap">
                    {opp.tags.slice(0, 3).map((tag) => (
                      <Chip
                        key={`${opp.id}-${tag}`}
                        label={tag}
                        size="small"
                        sx={{
                          background: "rgba(93,224,255,0.1)",
                          color: "#5de0ff",
                          fontSize: "0.65rem",
                        }}
                      />
                    ))}
                  </Stack>
                )}
              </Box>
              <Chip
                label={opp.type || "open"}
                size="small"
                sx={{
                  ml: 2,
                  background: "rgba(93, 224, 255, 0.2)",
                  color: "#5de0ff",
                }}
              />
            </Stack>
          </Box>
        ))}
        {opportunities.length === 0 && (
          <Typography variant="body2" color="text.secondary" textAlign="center" py={2}>
            No opportunities available
          </Typography>
        )}
      </Stack>
    </Paper>
  );
}

