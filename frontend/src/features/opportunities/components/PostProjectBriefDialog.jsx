import React, { useState, useEffect } from "react";
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  IconButton,
  Box,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Slider,
  FormControlLabel,
  Switch,
  Stack,
  Typography,
  Button,
  InputAdornment,
  Paper,
  Divider,
} from "@mui/material";
import CloseIcon from "@mui/icons-material/Close";
import LinkIcon from "@mui/icons-material/Link";
import CloudUploadIcon from "@mui/icons-material/CloudUpload";
import BoltOutlinedIcon from "@mui/icons-material/BoltOutlined";
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
  { value: "2weeks", label: "~2 weeks", months: 1 },
  { value: "1month", label: "~1 month", months: 1 },
  { value: "3months", label: "~3 months", months: 3 },
];

const accent = "#048B7F";
const accentDark = "#036b62";
const surfaceMuted = "#e6f4f3";

function SectionHeader({ step, title, subtitle }) {
  return (
    <Box sx={{ mb: 2 }}>
      <Stack direction="row" alignItems="flex-start" spacing={1.5}>
        <Box
          sx={{
            minWidth: 32,
            height: 32,
            borderRadius: 1.5,
            bgcolor: accent,
            color: "#fff",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            fontSize: 15,
            fontWeight: 800,
            lineHeight: 1,
            boxShadow: `0 2px 8px ${accent}44`,
          }}
        >
          {step}
        </Box>
        <Box>
          <Typography variant="subtitle1" fontWeight={700} color="text.primary" sx={{ lineHeight: 1.3 }}>
            {title}
          </Typography>
          {subtitle ? (
            <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
              {subtitle}
            </Typography>
          ) : null}
        </Box>
      </Stack>
    </Box>
  );
}

export default function PostProjectBriefDialog({ open, onClose, onSuccess, defaultOrganization = "" }) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [formData, setFormData] = useState({
    title: "",
    organization: "",
    position: "",
    category: "",
    urgent: false,
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

  useEffect(() => {
    if (!open) return;
    setFormData((prev) => ({
      ...prev,
      organization: (defaultOrganization || "").trim(),
    }));
  }, [open, defaultOrganization]);

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

  const resetForm = () => {
    setFormData({
      title: "",
      organization: (defaultOrganization || "").trim(),
      position: "",
      category: "",
      urgent: false,
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
  };

  const handleSubmit = async () => {
    setError(null);
    if (!formData.title?.trim()) {
      setError("Please describe what you need help with.");
      return;
    }
    if (!formData.category) {
      setError("Please select a category.");
      return;
    }
    const org = formData.organization?.trim() || "";
    const pos = formData.position?.trim() || "";
    if (!org) {
      setError("Please enter your organization or company name.");
      return;
    }
    if (!pos) {
      setError("Please enter your role or job title (shown under your name on the listing).");
      return;
    }

    setLoading(true);
    try {
      const timelineOpt = TIMELINE_OPTIONS.find((o) => o.value === formData.timeline);
      let desc = formData.description?.trim() || "";
      if (formData.referenceLink?.trim()) {
        desc += (desc ? "\n\n" : "") + `Reference: ${formData.referenceLink.trim()}`;
      }

      const budgetVal = formData.budgetNegotiable ? null : Math.max(0, Number(formData.budgetValue) || 0);
      const payload = {
        title: formData.title.trim(),
        organization: org,
        position: pos,
        category: formData.category,
        budgetType: formData.budgetType,
        budgetValue: budgetVal,
        location: formData.location,
        description: desc || undefined,
        tags: [],
        urgent: formData.urgent,
        durationMonths: formData.timelineFlexible ? null : (timelineOpt?.months ?? 1),
        startDate: formData.timelineFlexible || !formData.startDate ? null : `${formData.startDate}T12:00:00Z`,
      };

      await opportunitiesApi.createOpportunity(payload);
      onSuccess?.();
      onClose();
      resetForm();
    } catch (err) {
      const msg = err?.response?.data?.error?.message || err?.response?.data?.message || err?.message || "Failed to publish project";
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="md"
      fullWidth
      lang="en-US"
      PaperProps={{
        elevation: 8,
        sx: {
          borderRadius: 3,
          overflow: "hidden",
          maxHeight: "min(92vh, 900px)",
        },
      }}
    >
      <DialogTitle
        component="div"
        sx={{
          background: `linear-gradient(125deg, ${accentDark} 0%, ${accent} 42%, #2ea89d 100%)`,
          color: "#fff",
          py: 2.5,
          px: 3,
          pr: 5,
        }}
      >
        <Stack direction="row" alignItems="flex-start" justifyContent="space-between" spacing={2}>
          <Box>
            <Typography variant="h6" fontWeight={800} sx={{ letterSpacing: "-0.02em", mb: 0.5 }}>
              Post a Project Brief
            </Typography>
            <Typography variant="body2" sx={{ opacity: 0.92, maxWidth: 480, lineHeight: 1.5 }}>
              Reach Aalto talent with a clear brief. Add your company and role so talent knows who is posting.
            </Typography>
          </Box>
          <IconButton
            onClick={onClose}
            size="small"
            aria-label="close"
            sx={{
              color: "#fff",
              bgcolor: "rgba(255,255,255,0.12)",
              "&:hover": { bgcolor: "rgba(255,255,255,0.22)" },
            }}
          >
            <CloseIcon />
          </IconButton>
        </Stack>
      </DialogTitle>

      <Divider />

      <DialogContent sx={{ bgcolor: surfaceMuted, pt: 3, pb: 1 }}>
        <Stack spacing={2.75}>
          <Paper
            elevation={0}
            sx={{
              p: 2.5,
              borderRadius: 2.5,
              border: "1px solid",
              borderColor: "rgba(15, 118, 110, 0.18)",
              bgcolor: "#fff",
              boxShadow: "0 4px 24px rgba(15, 118, 110, 0.06)",
            }}
          >
            <SectionHeader
              step={1}
              title="What do you need help with?"
              subtitle="One clear sentence works best — talent will see this as the headline."
            />
            <TextField
              fullWidth
              placeholder="e.g. Album cover design for Spotify release, team photos for our website…"
              value={formData.title}
              onChange={handleChange("title")}
              size="small"
              sx={{ "& .MuiOutlinedInput-root": { bgcolor: "#fafafa", borderRadius: 2 } }}
            />
            <Stack direction={{ xs: "column", sm: "row" }} spacing={2} sx={{ mt: 2 }}>
              <TextField
                fullWidth
                required
                label="Organization / company"
                placeholder="Company or team name"
                value={formData.organization}
                onChange={handleChange("organization")}
                size="small"
                sx={{ "& .MuiOutlinedInput-root": { bgcolor: "#fafafa", borderRadius: 2 } }}
              />
              <TextField
                fullWidth
                required
                label="Your role / position"
                placeholder="e.g. Marketing Lead, Founder"
                value={formData.position}
                onChange={handleChange("position")}
                size="small"
                sx={{ "& .MuiOutlinedInput-root": { bgcolor: "#fafafa", borderRadius: 2 } }}
              />
            </Stack>
          </Paper>

          <Paper
            elevation={0}
            sx={{
              p: 2.5,
              borderRadius: 2.5,
              border: "1px solid",
              borderColor: "divider",
              bgcolor: "#fff",
            }}
          >
            <SectionHeader step={2} title="Category" />
            <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1 }}>
              {CATEGORIES.map((cat) => (
                <Button
                  key={cat}
                  variant={formData.category === cat ? "contained" : "outlined"}
                  size="small"
                  onClick={() => setFormData((p) => ({ ...p, category: cat }))}
                  sx={{
                    textTransform: "none",
                    fontWeight: 600,
                    borderRadius: 2,
                    ...(formData.category === cat
                      ? { bgcolor: accent, "&:hover": { bgcolor: accentDark } }
                      : { borderColor: "rgba(0,0,0,0.12)" }),
                  }}
                >
                  {cat}
                </Button>
              ))}
            </Box>
          </Paper>

          <Paper
            elevation={0}
            sx={{
              p: 2.5,
              borderRadius: 2.5,
              border: "2px solid",
              borderColor: formData.urgent ? `${accent}55` : "divider",
              bgcolor: formData.urgent ? "rgba(20, 184, 166, 0.06)" : "#fff",
              transition: "border-color 0.2s, background 0.2s",
            }}
          >
            <Stack direction={{ xs: "column", sm: "row" }} alignItems={{ sm: "center" }} justifyContent="space-between" spacing={2}>
              <Stack direction="row" spacing={1.5} alignItems="flex-start">
                <BoltOutlinedIcon sx={{ color: formData.urgent ? accent : "text.disabled", mt: 0.25 }} />
                <Box>
                  <Typography fontWeight={700}>Urgent</Typography>
                  <Typography variant="body2" color="text.secondary">
                    Highlights this brief to talent and pairs with filters on the opportunities page.
                  </Typography>
                </Box>
              </Stack>
              <FormControlLabel
                control={
                  <Switch
                    checked={formData.urgent}
                    onChange={(e) => setFormData((p) => ({ ...p, urgent: e.target.checked }))}
                    color="primary"
                    sx={{
                      "& .MuiSwitch-switchBase.Mui-checked": { color: accent },
                      "& .MuiSwitch-switchBase.Mui-checked + .MuiSwitch-track": { bgcolor: `${accent}99` },
                    }}
                  />
                }
                label={formData.urgent ? "Yes" : "No"}
                labelPlacement="start"
                sx={{ m: 0, ml: { sm: "auto" } }}
              />
            </Stack>
          </Paper>

          <Paper elevation={0} sx={{ p: 2.5, borderRadius: 2.5, border: "1px solid", borderColor: "divider", bgcolor: "#fff" }}>
            <SectionHeader step={3} title="Budget" />
            <ToggleButtonGroup value={formData.budgetType} exclusive onChange={handleBudgetTypeChange} sx={{ mb: 2 }}>
              <ToggleButton value="fixed" sx={{ textTransform: "none", px: 2 }}>
                Fixed price
              </ToggleButton>
              <ToggleButton value="hourly" sx={{ textTransform: "none", px: 2 }}>
                Hourly rate
              </ToggleButton>
            </ToggleButtonGroup>
            <Box sx={{ display: "flex", alignItems: "center", gap: 2, flexWrap: "wrap" }}>
              <Slider
                value={formData.budgetValue}
                onChange={(_, v) =>
                  setFormData((p) => ({ ...p, budgetValue: Math.max(0, typeof v === "number" ? v : 0) }))
                }
                min={0}
                max={3100}
                step={100}
                valueLabelDisplay="auto"
                valueLabelFormat={(v) => (v >= 3000 ? "€3000+" : `€${v}`)}
                sx={{
                  maxWidth: 300,
                  color: accent,
                  "& .MuiSlider-thumb": { bgcolor: accent },
                  "& .MuiSlider-track": { bgcolor: accent },
                }}
                disabled={formData.budgetNegotiable}
              />
              <TextField
                type="number"
                size="small"
                label="Up to (€)"
                value={formData.budgetNegotiable ? "" : formData.budgetValue}
                onChange={(e) =>
                  setFormData((p) => ({
                    ...p,
                    budgetValue: Math.max(0, Number(e.target.value) || 0),
                  }))
                }
                sx={{ width: 120 }}
                disabled={formData.budgetNegotiable}
                InputProps={{ inputProps: { min: 0 } }}
              />
            </Box>
            <FormControlLabel
              control={
                <Switch
                  checked={formData.budgetNegotiable}
                  onChange={(e) => setFormData((p) => ({ ...p, budgetNegotiable: e.target.checked }))}
                  size="small"
                />
              }
              label="Budget negotiable"
              sx={{ mt: 1.5 }}
            />
          </Paper>

          <Paper elevation={0} sx={{ p: 2.5, borderRadius: 2.5, border: "1px solid", borderColor: "divider", bgcolor: "#fff" }}>
            <SectionHeader step={4} title="Timeline" subtitle="Rough duration; you can add a specific deadline below." />
            <ToggleButtonGroup value={formData.timeline} exclusive onChange={handleTimelineChange} sx={{ mb: 2, flexWrap: "wrap" }}>
              {TIMELINE_OPTIONS.map((o) => (
                <ToggleButton key={o.value} value={o.value} sx={{ textTransform: "none" }}>
                  {o.label}
                </ToggleButton>
              ))}
            </ToggleButtonGroup>
            <Stack direction={{ xs: "column", sm: "row" }} spacing={2} alignItems={{ sm: "center" }}>
              <TextField
                type="date"
                size="small"
                label="Deadline (optional)"
                value={formData.startDate}
                onChange={handleChange("startDate")}
                InputLabelProps={{ shrink: true }}
                slotProps={{ htmlInput: { lang: "en-US" } }}
                sx={{ minWidth: 180 }}
                disabled={formData.timelineFlexible}
              />
              <FormControlLabel
                control={
                  <Switch
                    checked={formData.timelineFlexible}
                    onChange={(e) => setFormData((p) => ({ ...p, timelineFlexible: e.target.checked }))}
                    size="small"
                  />
                }
                label="Flexible — no fixed end date"
              />
            </Stack>
          </Paper>

          <Paper elevation={0} sx={{ p: 2.5, borderRadius: 2.5, border: "1px solid", borderColor: "divider", bgcolor: "#fff" }}>
            <SectionHeader step={5} title="Description" />
            <TextField
              fullWidth
              multiline
              rows={4}
              placeholder="Context, deliverables, constraints, who you’re looking for…"
              value={formData.description}
              onChange={handleChange("description")}
              size="small"
              sx={{ "& .MuiOutlinedInput-root": { bgcolor: "#fafafa", borderRadius: 2 } }}
            />
          </Paper>

          <Paper
            elevation={0}
            sx={{
              p: 2.5,
              borderRadius: 2.5,
              border: "1px dashed",
              borderColor: "rgba(0,0,0,0.14)",
              bgcolor: "rgba(255,255,255,0.85)",
            }}
          >
            <Typography fontWeight={700} sx={{ mb: 2 }}>
              Optional details
            </Typography>
            <Stack spacing={2.5}>
              <Box>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 0.75 }}>
                  Reference link
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
                        <LinkIcon fontSize="small" color="action" />
                      </InputAdornment>
                    ),
                  }}
                />
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                  Work location
                </Typography>
                <Box sx={{ display: "flex", gap: 1, flexWrap: "wrap" }}>
                  {LOCATIONS.map((loc) => (
                    <Button
                      key={loc}
                      variant={formData.location === loc ? "contained" : "outlined"}
                      size="small"
                      onClick={() => setFormData((p) => ({ ...p, location: loc }))}
                      sx={{
                        textTransform: "none",
                        fontWeight: 600,
                        borderRadius: 2,
                        ...(formData.location === loc
                          ? { bgcolor: accent, "&:hover": { bgcolor: accentDark } }
                          : {}),
                      }}
                    >
                      {loc}
                    </Button>
                  ))}
                </Box>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                  Attachment
                </Typography>
                <Box
                  sx={{
                    border: "2px dashed rgba(15, 118, 110, 0.25)",
                    borderRadius: 2,
                    p: 2.5,
                    textAlign: "center",
                    cursor: "pointer",
                    bgcolor: "rgba(240, 253, 250, 0.5)",
                    transition: "border-color 0.2s, background 0.2s",
                    "&:hover": { borderColor: accent, bgcolor: "rgba(240, 253, 250, 0.9)" },
                  }}
                  onClick={() => document.getElementById("opp-file-input")?.click()}
                  onKeyDown={(e) => e.key === "Enter" && document.getElementById("opp-file-input")?.click()}
                  role="button"
                  tabIndex={0}
                >
                  <input
                    id="opp-file-input"
                    type="file"
                    hidden
                    accept=".png,.jpg,.jpeg,.pdf"
                    onChange={(e) => setFormData((p) => ({ ...p, file: e.target.files?.[0] }))}
                  />
                  <CloudUploadIcon sx={{ fontSize: 36, color: accent, opacity: 0.75, mb: 0.5 }} />
                  <Typography variant="body2" color="text.secondary">
                    {formData.file ? formData.file.name : "Click to upload — PNG, JPG, or PDF (UI only for now)"}
                  </Typography>
                </Box>
              </Box>
            </Stack>
          </Paper>

          {error ? (
            <Typography color="error" variant="body2" sx={{ px: 0.5 }}>
              {error}
            </Typography>
          ) : null}
        </Stack>
      </DialogContent>

      <Divider />

      <DialogActions sx={{ px: 3, py: 2, bgcolor: surfaceMuted, gap: 1 }}>
        <Button onClick={onClose} disabled={loading} sx={{ textTransform: "none", fontWeight: 600, color: "text.secondary" }}>
          Cancel
        </Button>
        <Button
          variant="contained"
          onClick={handleSubmit}
          disabled={loading}
          sx={{
            textTransform: "none",
            fontWeight: 700,
            px: 3,
            py: 1,
            borderRadius: 2,
            bgcolor: accent,
            boxShadow: `0 4px 14px ${accent}55`,
            "&:hover": { bgcolor: accentDark, boxShadow: `0 6px 18px ${accent}66` },
          }}
        >
          {loading ? "Publishing…" : "Publish brief"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
