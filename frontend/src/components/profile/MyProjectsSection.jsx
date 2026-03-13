import React from "react";
import { Box, Button, Grid, Stack, Typography } from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import { useNavigate } from "react-router-dom";
import OpportunityCard from "../../features/opportunities/components/OpportunityCard";

export default function MyProjectsSection({ items, onRefresh }) {
  const navigate = useNavigate();

  return (
    <Box sx={{ p: 0 }}>
      <Stack direction="row" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h6" fontWeight={600}>
          My Posted Projects
        </Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => navigate("/opportunities")}
          sx={{ textTransform: "none", fontWeight: 600 }}
        >
          Post a Project Brief
        </Button>
      </Stack>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
        Projects you&apos;ve posted for talent to apply. Manage them from the Opportunities page.
      </Typography>
      {items.length === 0 ? (
        <Box
          sx={{
            py: 6,
            textAlign: "center",
            border: "1px dashed #e0e0e0",
            borderRadius: 2,
            bgcolor: "#fafafa",
          }}
        >
          <Typography variant="body1" color="text.secondary" gutterBottom>
            No projects posted yet
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Post a project brief to find talented students for your work.
          </Typography>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => navigate("/opportunities")}
            sx={{ textTransform: "none", fontWeight: 600 }}
          >
            Post a Project Brief
          </Button>
        </Box>
      ) : (
        <Grid container spacing={2}>
          {items.map((opp) => (
            <Grid item xs={12} sm={6} md={4} key={opp.id}>
              <OpportunityCard opportunity={opp} />
            </Grid>
          ))}
        </Grid>
      )}
    </Box>
  );
}
