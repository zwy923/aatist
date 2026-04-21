import React, { useCallback, useEffect, useState } from "react";
import { useLocation, useNavigate, useParams } from "react-router-dom";
import {
  Avatar,
  Box,
  Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  Grid,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  TextField,
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
  const location = useLocation();
  const navigate = useNavigate();
  const { isAuthenticated, user } = useAuthStore();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [opportunity, setOpportunity] = useState(null);
  const [applyModalOpen, setApplyModalOpen] = useState(false);
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [editSaving, setEditSaving] = useState(false);
  const [sidebarError, setSidebarError] = useState("");
  const [editForm, setEditForm] = useState({
    title: "",
    organization: "",
    position: "",
    category: "",
    budgetType: "fixed",
    budgetValue: "",
    location: "",
    durationMonths: "",
    startDate: "",
    description: "",
    urgent: false,
  });

  const reloadOpportunity = useCallback(async () => {
    if (!id) return;
    try {
      setLoading(true);
      setError(null);
      const res = await opportunitiesApi.getOpportunity(id);
      const data = res.data?.data;
      setOpportunity(data || null);
    } catch (e) {
      if (e.code === "ERR_CANCELED" || e.name === "CanceledError") return;
      console.error("Opportunity load error:", e);
      setError(e.message || "Failed to load opportunity");
      setOpportunity(null);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    if (!id) {
      setLoading(false);
      setError("Invalid opportunity link.");
      setOpportunity(null);
      return;
    }
    const ac = new AbortController();
    (async () => {
      try {
        setLoading(true);
        setError(null);
        const res = await opportunitiesApi.getOpportunity(id, { signal: ac.signal });
        if (ac.signal.aborted) return;
        const data = res.data?.data;
        setOpportunity(data || null);
      } catch (e) {
        if (e.code === "ERR_CANCELED" || e.name === "CanceledError") return;
        if (ac.signal.aborted) return;
        console.error("Opportunity load error:", e);
        setError(e.message || "Failed to load opportunity");
        setOpportunity(null);
      } finally {
        if (!ac.signal.aborted) {
          setLoading(false);
        }
      }
    })();
    return () => ac.abort();
  }, [id]);

  const formatDate = (dateStr) => {
    if (!dateStr) return "N/A";
    return new Date(dateStr).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  };

  const toDateInputValue = (dateStr) => {
    if (!dateStr) return "";
    const d = new Date(dateStr);
    if (Number.isNaN(d.getTime())) return "";
    return d.toISOString().slice(0, 10);
  };

  const myUserId = user?.id ?? user?.user_id;
  const isOwner =
    opportunity?.created_by != null &&
    myUserId != null &&
    String(opportunity.created_by) === String(myUserId);

  const buildEditFormFromOpportunity = useCallback((opp) => ({
    title: opp?.title || "",
    organization: opp?.organization || "",
    position: opp?.position || "",
    category: opp?.category || "",
    budgetType: opp?.budget_type || "fixed",
    budgetValue: opp?.budget_value != null ? String(opp.budget_value) : "",
    location: opp?.location || "",
    durationMonths: opp?.duration_months != null ? String(opp.duration_months) : "",
    startDate: toDateInputValue(opp?.start_date),
    description: opp?.description || "",
    urgent: !!opp?.urgent,
  }), []);

  useEffect(() => {
    if (!opportunity) return;
    setEditForm(buildEditFormFromOpportunity(opportunity));
  }, [opportunity, buildEditFormFromOpportunity]);

  useEffect(() => {
    const shouldOpenEdit = !!location.state?.openEdit;
    if (!shouldOpenEdit || !opportunity || !isOwner) return;
    setSidebarError("");
    setEditForm(buildEditFormFromOpportunity(opportunity));
    setEditModalOpen(true);
    navigate(location.pathname, { replace: true, state: {} });
  }, [location.pathname, location.state, navigate, opportunity, isOwner, buildEditFormFromOpportunity]);

  const handleApplySuccess = () => {
    reloadOpportunity();
  };

  const handleDeleteOpportunity = async () => {
    if (!opportunity?.id) return;
    setSidebarError("");
    if (!window.confirm("Delete this opportunity? This action cannot be undone.")) return;
    try {
      await opportunitiesApi.deleteOpportunity(opportunity.id);
      navigate("/opportunities");
    } catch (e) {
      console.error("Opportunity delete error:", e);
      setSidebarError(e?.response?.data?.message || e?.message || "Failed to delete opportunity");
    }
  };

  const handleSaveEdit = async () => {
    if (!opportunity?.id) return;
    setSidebarError("");
    const title = editForm.title.trim();
    const organization = editForm.organization.trim();
    const position = editForm.position.trim();
    const category = editForm.category.trim();
    const location = editForm.location.trim();
    if (!title || !organization || !position || !category || !location) {
      setSidebarError("Please complete title, organization, position, category and location.");
      return;
    }
    const payload = {
      title,
      organization,
      position,
      category,
      budgetType: editForm.budgetType,
      budgetValue: editForm.budgetValue === "" ? null : Math.max(0, Number(editForm.budgetValue) || 0),
      location,
      durationMonths: editForm.durationMonths === "" ? null : Math.max(0, Number(editForm.durationMonths) || 0),
      startDate: editForm.startDate ? `${editForm.startDate}T12:00:00Z` : null,
      description: editForm.description.trim() || "",
      urgent: !!editForm.urgent,
    };
    setEditSaving(true);
    try {
      await opportunitiesApi.updateOpportunity(opportunity.id, payload);
      setEditModalOpen(false);
      await reloadOpportunity();
    } catch (e) {
      console.error("Opportunity update error:", e);
      setSidebarError(e?.response?.data?.message || e?.message || "Failed to update opportunity");
    } finally {
      setEditSaving(false);
    }
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
                  <SavedButton targetId={opportunity.id} initialSaved={opportunity.is_favorite} iconSet="star" />
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

              <Stack direction="row" spacing={2} sx={{ mb: 2, color: "text.secondary", flexWrap: "wrap" }}>
                {opportunity?.published_at && (
                  <Typography variant="body2" sx={{ display: "flex", alignItems: "center", gap: 0.5 }}>
                    <CalendarTodayIcon fontSize="small" />
                    Published: {formatDate(opportunity.published_at)}
                  </Typography>
                )}
                {opportunity?.start_date && (
                  <Typography variant="body2" sx={{ display: "flex", alignItems: "center", gap: 0.5 }}>
                    <CalendarTodayIcon fontSize="small" />
                    Deadline: {formatDate(opportunity.start_date)}
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
              <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5 }}>
                Posted by
              </Typography>
              <Typography fontWeight={700} sx={{ mb: 0.5 }}>
                {opportunity?.creator_name || "—"}
              </Typography>
              {opportunity?.position ? (
                <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                  {opportunity.position}
                </Typography>
              ) : null}
              <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5 }}>
                Organization
              </Typography>
              <Typography fontWeight={600} sx={{ mb: 2 }}>
                {opportunity?.organization || "—"}
              </Typography>

              {isOwner ? (
                <Stack direction="row" spacing={1} sx={{ mb: 2 }}>
                  <Button
                    fullWidth
                    variant="outlined"
                    onClick={() => {
                      setSidebarError("");
                      setEditForm(buildEditFormFromOpportunity(opportunity));
                      setEditModalOpen(true);
                    }}
                    sx={{ textTransform: "none" }}
                  >
                    Edit
                  </Button>
                  <Button
                    fullWidth
                    color="error"
                    variant="contained"
                    onClick={handleDeleteOpportunity}
                    sx={{ textTransform: "none" }}
                  >
                    Delete
                  </Button>
                </Stack>
              ) : (
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
              )}

              {sidebarError ? (
                <Typography color="error" variant="body2" sx={{ mb: 1.5 }}>
                  {sidebarError}
                </Typography>
              ) : null}

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

        <Dialog open={editModalOpen} onClose={() => !editSaving && setEditModalOpen(false)} fullWidth maxWidth="sm">
          <DialogTitle>Edit opportunity</DialogTitle>
          <DialogContent>
            <Stack spacing={2} sx={{ mt: 1 }}>
              <TextField
                label="Title"
                fullWidth
                value={editForm.title}
                onChange={(e) => setEditForm((prev) => ({ ...prev, title: e.target.value }))}
              />
              <Stack direction={{ xs: "column", sm: "row" }} spacing={1.5}>
                <TextField
                  label="Organization"
                  fullWidth
                  value={editForm.organization}
                  onChange={(e) => setEditForm((prev) => ({ ...prev, organization: e.target.value }))}
                />
                <TextField
                  label="Position"
                  fullWidth
                  value={editForm.position}
                  onChange={(e) => setEditForm((prev) => ({ ...prev, position: e.target.value }))}
                />
              </Stack>
              <Stack direction={{ xs: "column", sm: "row" }} spacing={1.5}>
                <TextField
                  label="Category"
                  fullWidth
                  value={editForm.category}
                  onChange={(e) => setEditForm((prev) => ({ ...prev, category: e.target.value }))}
                />
                <TextField
                  label="Location"
                  fullWidth
                  value={editForm.location}
                  onChange={(e) => setEditForm((prev) => ({ ...prev, location: e.target.value }))}
                />
              </Stack>
              <Stack direction={{ xs: "column", sm: "row" }} spacing={1.5}>
                <FormControl fullWidth>
                  <InputLabel id="edit-budget-type-label">Budget type</InputLabel>
                  <Select
                    labelId="edit-budget-type-label"
                    label="Budget type"
                    value={editForm.budgetType}
                    onChange={(e) => setEditForm((prev) => ({ ...prev, budgetType: e.target.value }))}
                  >
                    <MenuItem value="fixed">Fixed</MenuItem>
                    <MenuItem value="hourly">Hourly</MenuItem>
                  </Select>
                </FormControl>
                <TextField
                  label="Budget value"
                  type="number"
                  fullWidth
                  value={editForm.budgetValue}
                  onChange={(e) => setEditForm((prev) => ({ ...prev, budgetValue: e.target.value }))}
                />
              </Stack>
              <Stack direction={{ xs: "column", sm: "row" }} spacing={1.5}>
                <TextField
                  label="Duration (months)"
                  type="number"
                  fullWidth
                  value={editForm.durationMonths}
                  onChange={(e) => setEditForm((prev) => ({ ...prev, durationMonths: e.target.value }))}
                />
                <TextField
                  label="Deadline"
                  type="date"
                  fullWidth
                  value={editForm.startDate}
                  onChange={(e) => setEditForm((prev) => ({ ...prev, startDate: e.target.value }))}
                  InputLabelProps={{ shrink: true }}
                />
              </Stack>
              <TextField
                label="Description"
                fullWidth
                multiline
                rows={4}
                value={editForm.description}
                onChange={(e) => setEditForm((prev) => ({ ...prev, description: e.target.value }))}
              />
            </Stack>
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setEditModalOpen(false)} disabled={editSaving} sx={{ textTransform: "none" }}>
              Cancel
            </Button>
            <Button variant="contained" onClick={handleSaveEdit} disabled={editSaving} sx={{ textTransform: "none" }}>
              {editSaving ? "Saving..." : "Save"}
            </Button>
          </DialogActions>
        </Dialog>
      </StateContainer>
    </PageLayout>
  );
}
