import React, { useCallback, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  Avatar,
  Box,
  Button,
  Chip,
  Grid,
  Paper,
  Stack,
  Typography,
} from "@mui/material";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import LocationOnIcon from "@mui/icons-material/LocationOn";
import AttachMoneyIcon from "@mui/icons-material/AttachMoney";
import CalendarTodayIcon from "@mui/icons-material/CalendarToday";
import PageLayout from "../shared/components/PageLayout";
import { StateContainer } from "../shared/components/ui/StateContainer";
import { opportunitiesApi } from "../features/opportunities/api/opportunities";
import ApplyModal from "../features/opportunities/components/ApplyModal";
import SavedButton from "../features/opportunities/components/SavedButton";
import useAuthStore from "../shared/stores/authStore";

export default function OpportunityDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { isAuthenticated } = useAuthStore();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [opportunity, setOpportunity] = useState(null);
  const [applyModalOpen, setApplyModalOpen] = useState(false);

  const loadData = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await opportunitiesApi.getOpportunity(id);
      const data = res.data?.data;
      setOpportunity(data || null);
    } catch (e) {
      console.error("Opportunity load error:", e);
      setError("Failed to load opportunity");
      setOpportunity(null);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const formatDate = (dateStr) => {
    if (!dateStr) return "N/A";
    return new Date(dateStr).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  };

  const handleApplySuccess = () => {
    loadData();
  };

  return (
    <PageLayout maxWidth="xl" variant="light">
      <StateContainer loading={loading} error={error} empty={!opportunity}>
        <Button
          startIcon={<ArrowBackIcon />}
          onClick={() => navigate("/opportunities")}
          sx={{ mb: 2, textTransform: "none" }}
        >
          Back to opportunities
        </Button>

        <Grid container spacing={3}>
          <Grid item xs={12} md={8}>
            <Paper sx={{ p: 3, border: "1px solid #e5e7eb", borderRadius: 3 }}>
              <Stack direction="row" justifyContent="space-between" alignItems="flex-start" sx={{ mb: 2 }}>
                <Typography variant="h4" fontWeight={700}>
                  {opportunity?.title}
                </Typography>
                {opportunity?.id && (
                  <SavedButton targetId={opportunity.id} initialSaved={opportunity.is_favorite} />
                )}
              </Stack>

              <Stack direction="row" spacing={2} flexWrap="wrap" sx={{ mb: 2 }}>
                {opportunity?.budget_value != null ? (
                  <Chip
                    icon={<AttachMoneyIcon />}
                    label={`€${opportunity.budget_value} (${opportunity.budget_type || "fixed"})`}
                    variant="outlined"
                    color="primary"
                  />
                ) : (
                  <Chip label="Budget negotiable" variant="outlined" />
                )}
                {opportunity?.location && (
                  <Chip
                    icon={<LocationOnIcon />}
                    label={opportunity.location}
                    variant="outlined"
                  />
                )}
                {opportunity?.urgent && (
                  <Chip label="Urgent" color="error" size="small" />
                )}
              </Stack>

              <Stack direction="row" spacing={2} sx={{ mb: 2, color: "text.secondary" }}>
                {opportunity?.start_date && (
                  <Typography variant="body2" sx={{ display: "flex", alignItems: "center", gap: 0.5 }}>
                    <CalendarTodayIcon fontSize="small" />
                    Starts: {formatDate(opportunity.start_date)}
                  </Typography>
                )}
                {opportunity?.duration_months != null && (
                  <Typography variant="body2">
                    Duration: {opportunity.duration_months} month(s)
                  </Typography>
                )}
              </Stack>

              <Typography color="text.secondary" sx={{ whiteSpace: "pre-wrap", lineHeight: 1.8 }}>
                {opportunity?.description || "No description provided."}
              </Typography>

              {(opportunity?.tags || []).length > 0 && (
                <Stack direction="row" spacing={0.5} flexWrap="wrap" sx={{ mt: 2 }}>
                  {opportunity.tags.map((tag) => (
                    <Chip key={tag} label={tag} size="small" />
                  ))}
                </Stack>
              )}
            </Paper>
          </Grid>

          <Grid item xs={12} md={4}>
            <Paper sx={{ p: 2.5, border: "1px solid #e5e7eb", borderRadius: 3, position: "sticky", top: 20 }}>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                Posted by organization
              </Typography>
              <Typography fontWeight={600} sx={{ mb: 2 }}>
                {opportunity?.organization || "Client"}
              </Typography>

              <Button
                fullWidth
                variant="contained"
                onClick={() => {
                  if (!isAuthenticated) {
                    navigate("/auth/login", { state: { from: `/opportunities/${id}` } });
                    return;
                  }
                  setApplyModalOpen(true);
                }}
                sx={{ mb: 2, textTransform: "none" }}
              >
                Apply for this opportunity
              </Button>

              <Box sx={{ borderTop: "1px solid #e5e7eb", pt: 1.5 }}>
                <Typography color="text.secondary" variant="body2">✓ Fast response time</Typography>
                <Typography color="text.secondary" variant="body2">✓ Professional quality</Typography>
              </Box>
            </Paper>
          </Grid>
        </Grid>

        <ApplyModal
          open={applyModalOpen}
          onClose={() => setApplyModalOpen(false)}
          opportunityId={id}
          opportunityTitle={opportunity?.title}
          onSuccess={handleApplySuccess}
        />
      </StateContainer>
    </PageLayout>
  );
}
