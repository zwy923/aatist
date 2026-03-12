import React from "react";
import { Card, CardContent, Stack, Box, Typography } from "@mui/material";

export default function StatsCard({ icon, label, value, accent }) {
  return (
    <Card
      sx={{
        background: "#ffffff",
        border: "1px solid #e5e7eb",
        borderRadius: 3,
        height: "100%",
      }}
    >
      <CardContent>
        <Stack direction="row" spacing={2} alignItems="center">
          <Box sx={{ p: 1.5, borderRadius: 2, background: accent.bg }}>{icon}</Box>
          <Box>
            <Typography variant="body2" color="text.secondary">
              {label}
            </Typography>
            <Typography variant="h4" fontWeight={700} sx={{ color: "text.primary" }}>
              {value}
            </Typography>
          </Box>
        </Stack>
      </CardContent>
    </Card>
  );
}

