import React from "react";
import { Box, Paper, Stack, Typography } from "@mui/material";
import TrendingUp from "@mui/icons-material/TrendingUp";

export default function PopularSkillsCard({ skills }) {
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
      <Stack direction="row" spacing={1} alignItems="center" mb={2}>
        <TrendingUp sx={{ color: "#5de0ff" }} />
        <Typography variant="h6" fontWeight={600}>
          Popular Skills
        </Typography>
      </Stack>
      {skills.length === 0 ? (
        <Typography variant="body2" color="text.secondary">
          No skills detected yet. Publish opportunities with tags to populate this section.
        </Typography>
      ) : (
        <Stack spacing={1.5}>
          {skills.map((item) => (
            <Box key={item.tag}>
              <Stack direction="row" justifyContent="space-between" alignItems="center" mb={0.5}>
                <Typography variant="body1" sx={{ color: "text.primary" }}>
                  {item.tag}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {item.count} listings
                </Typography>
              </Stack>
              <Box
                sx={{
                  height: 6,
                  borderRadius: 3,
                  background: "rgba(93, 224, 255, 0.1)",
                  overflow: "hidden",
                }}
              >
                <Box
                  sx={{
                    height: "100%",
                    width: `${Math.min((item.count / (skills[0]?.count || 1)) * 100, 100)}%`,
                    background: "linear-gradient(90deg, #5de0ff, #7f5dff)",
                  }}
                />
              </Box>
            </Box>
          ))}
        </Stack>
      )}
    </Paper>
  );
}

