import React, { useState } from "react";
import {
  Dialog,
  DialogTitle,
  DialogContent,
  IconButton,
  Box,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Slider,
  FormControlLabel,
  Radio,
  RadioGroup,
  Stack,
  Typography,
  Button,
  InputAdornment,
  Paper,
} from "@mui/material";
import CloseIcon from "@mui/icons-material/Close";
import LinkIcon from "@mui/icons-material/Link";
import CloudUploadIcon from "@mui/icons-material/CloudUpload";
import { opportunitiesApi } from "../api/opportunities";

const CATEGORIES = [
  "Software Development",
  "Design",
  "Marketing",
  "Photography",
  "Video & Animation",
  "Logo & Brand Identity",
  "Web Design",
  "UI/UX Design",
  "Other",
];

const LOCATIONS = ["Remote", "On-site", "Hybrid"];

const TIMELINE_OPTIONS = [
  { value: "urgent", label: "Urgent (5 days)", urgent: true, months: 1 },
  { value: "2weeks", label: "2 weeks", urgent: false, months: 1 },
  { value: "1month", label: "1 month", urgent: false, months: 1 },
];

export default function PostProjectBriefDialog({ open, onClose, onSuccess }) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [formData, setFormData] = useState({
    title: "",
    organization: "",
    category: "",
    budgetType: "fixed",
    budgetValue: 1500,
    budgetNegotiable: false,
    timeline: "2weeks",
    timelineFlexible: false,
    startDate: "",
    description: "",
    referenceLink: "",
    location: "Remote",
    file: null,
  });

  const handleChange = (field) => (e) => {
    const value = e.target.type === "checkbox" ? e.target.checked : e.target.value;
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  const handleBudgetTypeChange = (_, value) => {
    if (value) setFormData((prev) => ({ ...prev, budgetType: value }));
  };

  const handleTimelineChange = (_, value) => {
    if (value) setFormData((prev) => ({ ...prev, timeline: value }));
  };

  const handleSubmit = async () => {
    setError(null);
    if (!formData.title?.trim()) {
      setError("Please enter what you need help with.");
      return;
    }
    if (!formData.organization?.trim()) {
      setError("Please enter your organization name.");
      return;
    }
    if (!formData.category) {
      setError("Please select a category.");
      return;
    }

    setLoading(true);
    try {
      const timelineOpt = TIMELINE_OPTIONS.find((o) => o.value === formData.timeline);
      let desc = formData.description?.trim() || "";
      if (formData.referenceLink?.trim()) {
        desc += (desc ? "\n\n" : "") + `Reference: ${formData.referenceLink.trim()}`;
      }

      const payload = {
        title: formData.title.trim(),
        organization: formData.organization.trim(),
        category: formData.category,
        budgetType: formData.budgetType,
        budgetValue: formData.budgetNegotiable ? null : formData.budgetValue,
        location: formData.location,
        description: desc || undefined,
        tags: [],
        urgent: formData.timelineFlexible ? false : (timelineOpt?.urgent ?? false),
        durationMonths: formData.timelineFlexible ? null : (timelineOpt?.months ?? 1),
        startDate: formData.timelineFlexible || !formData.startDate ? null : `${formData.startDate}T12:00:00Z`,
      };

      await opportunitiesApi.createOpportunity(payload);
      onSuccess?.();
      onClose();
      setFormData({
        title: "",
        organization: "",
        category: "",
        budgetType: "fixed",
        budgetValue: 1500,
        budgetNegotiable: false,
        timeline: "2weeks",
        timelineFlexible: false,
        startDate: "",
        description: "",
        referenceLink: "",
        location: "Remote",
        file: null,
      });
    } catch (err) {
      const msg = err?.response?.data?.error?.message || err?.response?.data?.message || err?.message || "Failed to publish project";
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth PaperProps={{ sx: { borderRadius: 3 } }}>
      <DialogTitle component="div" sx={{ display: "flex", alignItems: "center", justifyContent: "space-between", pb: 1 }}>
        <Typography variant="h6" component="span" fontWeight={700}>
          Post a Project Brief
        </Typography>
        <IconButton onClick={onClose} size="small" aria-label="close">
          <CloseIcon />
        </IconButton>
      </DialogTitle>
      <DialogContent sx={{ bgcolor: "rgba(147, 197, 253, 0.06)" }}>
        <Stack spacing={3} sx={{ pt: 1 }}>
          {/* 1. What do you need help with? */}
          <Paper variant="outlined" sx={{ p: 2, borderRadius: 2, bgcolor: "#fff" }}>
            <Typography fontWeight={600} sx={{ mb: 1 }}>
              1. What do you need help with?
            </Typography>
            <TextField
              fullWidth
              placeholder="Example: Need photographer for startup team photos."
              value={formData.title}
              onChange={handleChange("title")}
              size="small"
            />
          </Paper>

          {/* Organization (required by backend) */}
          <Paper variant="outlined" sx={{ p: 2, borderRadius: 2, bgcolor: "#fff" }}>
            <Typography fontWeight={600} sx={{ mb: 1 }}>
              Your organization
            </Typography>
            <TextField
              fullWidth
              placeholder="Company or organization name"
              value={formData.organization}
              onChange={handleChange("organization")}
              size="small"
            />
          </Paper>

          {/* Category */}
          <Paper variant="outlined" sx={{ p: 2, borderRadius: 2, bgcolor: "#fff" }}>
            <Typography fontWeight={600} sx={{ mb: 1 }}>
              Category
            </Typography>
            <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1 }}>
              {CATEGORIES.map((cat) => (
                <Button
                  key={cat}
                  variant={formData.category === cat ? "contained" : "outlined"}
                  size="small"
                  onClick={() => setFormData((p) => ({ ...p, category: cat }))}
                  sx={{ textTransform: "none" }}
                >
                  {cat}
                </Button>
              ))}
            </Box>
          </Paper>

          {/* 2. Budget */}
          <Paper variant="outlined" sx={{ p: 2, borderRadius: 2, bgcolor: "#fff" }}>
            <Typography fontWeight={600} sx={{ mb: 1 }}>
              2. Budget
            </Typography>
            <ToggleButtonGroup
              value={formData.budgetType}
              exclusive
              onChange={handleBudgetTypeChange}
              sx={{ mb: 2 }}
            >
              <ToggleButton value="fixed" sx={{ textTransform: "none" }}>
                Fixed price
              </ToggleButton>
              <ToggleButton value="hourly" sx={{ textTransform: "none" }}>
                Hourly rate
              </ToggleButton>
            </ToggleButtonGroup>
            <Box sx={{ display: "flex", alignItems: "center", gap: 2, flexWrap: "wrap" }}>
              <Slider
                value={formData.budgetValue}
                onChange={(_, v) => setFormData((p) => ({ ...p, budgetValue: v }))}
                min={0}
                max={3100}
                step={100}
                valueLabelDisplay="auto"
                valueLabelFormat={(v) => (v >= 3000 ? "3000+" : v)}
                sx={{ maxWidth: 280 }}
                disabled={formData.budgetNegotiable}
              />
              <TextField
                type="number"
                size="small"
                label="Up to €"
                value={formData.budgetNegotiable ? "" : formData.budgetValue}
                onChange={(e) => setFormData((p) => ({ ...p, budgetValue: Number(e.target.value) || 0 }))}
                sx={{ width: 120 }}
                disabled={formData.budgetNegotiable}
                InputProps={{ inputProps: { min: 0 } }}
              />
            </Box>
            <FormControlLabel
              control={
                <Radio
                  checked={formData.budgetNegotiable}
                  onChange={(e) => setFormData((p) => ({ ...p, budgetNegotiable: e.target.checked }))}
                />
              }
              label="Budget negotiable"
              sx={{ mt: 1 }}
            />
          </Paper>

          {/* 3. Timeline */}
          <Paper variant="outlined" sx={{ p: 2, borderRadius: 2, bgcolor: "#fff" }}>
            <Typography fontWeight={600} sx={{ mb: 1 }}>
              3. Timeline
            </Typography>
            <ToggleButtonGroup
              value={formData.timeline}
              exclusive
              onChange={handleTimelineChange}
              sx={{ mb: 2 }}
            >
              {TIMELINE_OPTIONS.map((o) => (
                <ToggleButton key={o.value} value={o.value} sx={{ textTransform: "none" }}>
                  {o.label}
                </ToggleButton>
              ))}
            </ToggleButtonGroup>
            <Box sx={{ display: "flex", alignItems: "center", gap: 2, flexWrap: "wrap" }}>
              <TextField
                type="date"
                size="small"
                label="Deadline"
                value={formData.startDate}
                onChange={handleChange("startDate")}
                InputLabelProps={{ shrink: true }}
                sx={{ minWidth: 160 }}
                disabled={formData.timelineFlexible}
              />
              <FormControlLabel
                control={
                  <Radio
                    checked={formData.timelineFlexible}
                    onChange={(e) => setFormData((p) => ({ ...p, timelineFlexible: e.target.checked }))}
                  />
                }
                label="Flexible"
              />
            </Box>
          </Paper>

          {/* 4. Description */}
          <Paper variant="outlined" sx={{ p: 2, borderRadius: 2, bgcolor: "#fff" }}>
            <Typography fontWeight={600} sx={{ mb: 1 }}>
              4. Description
            </Typography>
            <TextField
              fullWidth
              multiline
              rows={4}
              placeholder="Provide project context and scope..."
              value={formData.description}
              onChange={handleChange("description")}
              size="small"
            />
          </Paper>

          {/* Advanced */}
          <Paper variant="outlined" sx={{ p: 2, borderRadius: 2, bgcolor: "#fff" }}>
            <Typography fontWeight={600} sx={{ mb: 2 }}>
              Advanced (Optional)
            </Typography>
            <Stack spacing={2}>
              <Box>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5 }}>
                  Reference Link
                </Typography>
                <TextField
                  fullWidth
                  size="small"
                  placeholder="https://"
                  value={formData.referenceLink}
                  onChange={handleChange("referenceLink")}
                  InputProps={{
                    startAdornment: (
                      <InputAdornment position="start">
                        <LinkIcon fontSize="small" />
                      </InputAdornment>
                    ),
                  }}
                />
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5 }}>
                  Location
                </Typography>
                <Box sx={{ display: "flex", gap: 1, flexWrap: "wrap" }}>
                  {LOCATIONS.map((loc) => (
                    <Button
                      key={loc}
                      variant={formData.location === loc ? "contained" : "outlined"}
                      size="small"
                      onClick={() => setFormData((p) => ({ ...p, location: loc }))}
                      sx={{ textTransform: "none" }}
                    >
                      {loc}
                    </Button>
                  ))}
                </Box>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5 }}>
                  File Upload
                </Typography>
                <Box
                  sx={{
                    border: "2px dashed #ccc",
                    borderRadius: 2,
                    p: 3,
                    textAlign: "center",
                    cursor: "pointer",
                    "&:hover": { borderColor: "primary.main", bgcolor: "action.hover" },
                  }}
                  onClick={() => document.getElementById("opp-file-input").click()}
                >
                  <input
                    id="opp-file-input"
                    type="file"
                    hidden
                    accept=".png,.jpg,.jpeg,.pdf"
                    onChange={(e) => setFormData((p) => ({ ...p, file: e.target.files?.[0] }))}
                  />
                  <CloudUploadIcon sx={{ fontSize: 40, color: "text.disabled", mb: 1 }} />
                  <Typography variant="body2" color="text.secondary">
                    Upload a file or drag and drop
                  </Typography>
                  <Typography variant="caption" color="text.disabled">
                    PNG, JPG, PDF up to 10MB
                  </Typography>
                </Box>
              </Box>
            </Stack>
          </Paper>

          {error && (
            <Typography color="error" variant="body2">
              {error}
            </Typography>
          )}

          <Box sx={{ display: "flex", justifyContent: "flex-end" }}>
            <Button
              variant="contained"
              onClick={handleSubmit}
              disabled={loading}
              sx={{
                bgcolor: "#22c55e",
                "&:hover": { bgcolor: "#16a34a" },
                textTransform: "none",
                fontWeight: 600,
                px: 3,
              }}
            >
              {loading ? "Publishing..." : "Publish Project"}
            </Button>
          </Box>
        </Stack>
      </DialogContent>
    </Dialog>
  );
}
