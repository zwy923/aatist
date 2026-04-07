import React, { useState, useEffect, useCallback } from "react";
import {
  Box,
  Button,
  Dialog,
  DialogContent,
  DialogTitle,
  Checkbox,
  FormControl,
  FormControlLabel,
  FormGroup,
  FormLabel,
  IconButton,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  TextField,
  Typography,
  Paper,
  Alert,
  Snackbar,
} from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import SaveIcon from "@mui/icons-material/Save";
import EditIcon from "@mui/icons-material/Edit";
import DeleteIcon from "@mui/icons-material/Delete";
import CloudUploadIcon from "@mui/icons-material/CloudUpload";
import PhotoLibraryIcon from "@mui/icons-material/PhotoLibrary";
import { profileApi, portfolioApi } from "../../features/profile/api/profile";
import { encodePriceTypePayload, formatServicePriceLine, parsePriceTypeTokens } from "../../shared/utils/priceType";

const BROAD_CATEGORIES = {
  Design: ["Logo Design", "Brand Identity", "Illustration", "Print Design", "Presentation Design", "Infographics"],
  "Graphics & Design": ["Logo & Brand Identity", "Illustration & Drawing", "Print Design", "Presentation Design"],
  "Website & Digital": ["Web Design", "UI/UX Design", "Digital Product"],
  "Video & Animation": ["Video Editing", "Animation", "Motion Graphics", "Explainer Videos"],
  Photography: ["Event Photography", "Portrait Photography", "Product Photography", "Food Photography"],
};

const FLAT_SERVICES = [];
Object.entries(BROAD_CATEGORIES).forEach(([broad, specifics]) => {
  specifics.forEach((s) => FLAT_SERVICES.push({ broad, specific: s }));
});

/** Map stored category string back to broad + specific selects (handles legacy "Specific only" values). */
function resolveCategoryFields(stored) {
  const raw = (stored || "").trim();
  if (!raw) return { broadCategory: "", specificService: "" };
  if (raw.includes(" > ")) {
    const idx = raw.indexOf(" > ");
    const broad = raw.slice(0, idx).trim();
    const specific = raw.slice(idx + 3).trim();
    return { broadCategory: broad, specificService: specific || broad };
  }
  const hit = FLAT_SERVICES.find((x) => x.specific === raw || x.broad === raw);
  if (hit) return { broadCategory: hit.broad, specificService: hit.specific };
  return { broadCategory: "", specificService: raw };
}

export default function ServicesSection({
  onSave,
  hideIntro = false,
  triggerEditForId = null,
  onTriggerEditConsumed,
}) {
  const [services, setServices] = useState([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingId, setEditingId] = useState(null);
  const [snackbar, setSnackbar] = useState({ open: false, message: "", severity: "success" });
  const [formData, setFormData] = useState({
    broadCategory: "",
    specificService: "",
    title: "",
    description: "",
    shortDescription: "",
    priceHourly: false,
    priceProject: false,
    priceNegotiable: true,
    priceMin: "",
    priceMax: "",
    mediaUrls: [],
  });
  const [uploadingMedia, setUploadingMedia] = useState(false);

  const loadServices = async () => {
    try {
      const res = await profileApi.getServices();
      const list = res.data?.data?.services || [];
      setServices(list.length ? list : []);
    } catch (err) {
      setServices([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadServices();
  }, []);

  const resetForm = () => {
    setFormData({
      broadCategory: "",
      specificService: "",
      title: "",
      description: "",
      shortDescription: "",
      priceHourly: false,
      priceProject: false,
      priceNegotiable: true,
      priceMin: "",
      priceMax: "",
      mediaUrls: [],
    });
    setEditingId(null);
  };

  const handleOpenAdd = () => {
    resetForm();
    setDialogOpen(true);
  };

  const handleOpenEdit = useCallback((s) => {
    const { broadCategory, specificService } = resolveCategoryFields(s.category);
    const tokens = parsePriceTypeTokens(s.price_type);
    const hasAny = tokens.length > 0;
    setFormData({
      broadCategory,
      specificService,
      title: s.title || "",
      description: s.description || s.experience_summary || "",
      shortDescription: s.short_description || "",
      priceHourly: tokens.includes("hourly"),
      priceProject: tokens.includes("project"),
      priceNegotiable: hasAny ? tokens.includes("negotiable") : true,
      priceMin: s.price_min != null ? String(s.price_min) : "",
      priceMax: s.price_max != null ? String(s.price_max) : "",
      mediaUrls: s.media_urls || [],
    });
    setEditingId(s.id);
    setDialogOpen(true);
  }, []);

  useEffect(() => {
    if (triggerEditForId == null || loading) return;
    const s = services.find((x) => String(x.id) === String(triggerEditForId));
    if (!s) {
      onTriggerEditConsumed?.();
      return;
    }
    handleOpenEdit(s);
    onTriggerEditConsumed?.();
  }, [triggerEditForId, loading, services, onTriggerEditConsumed, handleOpenEdit]);

  const handleChange = (field) => (e) => {
    const value = e.target.value;
    if (field === "broadCategory") {
      setFormData((prev) => ({ ...prev, broadCategory: value, specificService: "" }));
      return;
    }
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  const handleMediaUpload = async (e) => {
    const files = e.target.files;
    if (!files?.length) return;
    if (formData.mediaUrls.length + files.length > 9) {
      setSnackbar({ open: true, message: "Max 9 files", severity: "warning" });
      return;
    }
    setUploadingMedia(true);
    try {
      const urls = [];
      for (let i = 0; i < files.length; i++) {
        const res = await portfolioApi.uploadProjectCover(files[i]);
        const url = res?.data?.data?.url;
        if (url) urls.push(url);
      }
      setFormData((prev) => ({ ...prev, mediaUrls: [...prev.mediaUrls, ...urls] }));
    } catch (err) {
      setSnackbar({ open: true, message: err?.message || "Upload failed", severity: "error" });
    } finally {
      setUploadingMedia(false);
      e.target.value = "";
    }
  };

  const handleRemoveMedia = (idx) => {
    setFormData((prev) => ({
      ...prev,
      mediaUrls: prev.mediaUrls.filter((_, i) => i !== idx),
    }));
  };

  const handleSaveService = async () => {
    const broad = (formData.broadCategory || "").trim();
    const spec = (formData.specificService || "").trim();
    const cat = broad && spec && broad !== spec ? `${broad} > ${spec}` : spec || broad;
    if (!cat.trim()) {
      setSnackbar({ open: true, message: "Please select a category", severity: "error" });
      return;
    }
    const desc = formData.description || formData.shortDescription;
    if (!desc.trim()) {
      setSnackbar({ open: true, message: "Please add a description", severity: "error" });
      return;
    }
    if (!formData.priceHourly && !formData.priceProject && !formData.priceNegotiable) {
      setSnackbar({ open: true, message: "Select at least one pricing option", severity: "error" });
      return;
    }
    const rawMin = formData.priceMin != null && String(formData.priceMin).trim() !== "" ? parseInt(String(formData.priceMin).trim(), 10) : null;
    const rawMax = formData.priceMax != null && String(formData.priceMax).trim() !== "" ? parseInt(String(formData.priceMax).trim(), 10) : null;
    if (rawMin != null && (Number.isNaN(rawMin) || rawMin < 0)) {
      setSnackbar({ open: true, message: "Price minimum cannot be negative", severity: "error" });
      return;
    }
    if (rawMax != null && (Number.isNaN(rawMax) || rawMax < 0)) {
      setSnackbar({ open: true, message: "Price maximum cannot be negative", severity: "error" });
      return;
    }
    setSaving(true);
    try {
      const payload = {
        category: cat,
        experience_summary: desc,
        title: formData.title.trim() || null,
        description: formData.description.trim() || null,
        short_description: formData.shortDescription.trim() || null,
        price_type: encodePriceTypePayload({
          hourly: formData.priceHourly,
          project: formData.priceProject,
          negotiable: formData.priceNegotiable,
        }),
        price_min: rawMin,
        price_max: rawMax,
        media_urls: formData.mediaUrls,
      };
      if (editingId != null && editingId !== "") {
        await profileApi.updateService(Number(editingId), payload);
        setSnackbar({ open: true, message: "Service updated", severity: "success" });
      } else {
        await profileApi.createService(payload);
        setSnackbar({ open: true, message: "Service added", severity: "success" });
      }
      await loadServices();
      setDialogOpen(false);
      resetForm();
      onSave?.();
    } catch (err) {
      setSnackbar({ open: true, message: err?.message || "Save failed", severity: "error" });
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id) => {
    if (!window.confirm("Delete this service?")) return;
    try {
      await profileApi.deleteService(id);
      await loadServices();
      setSnackbar({ open: true, message: "Service deleted", severity: "success" });
      onSave?.();
    } catch (err) {
      setSnackbar({ open: true, message: err?.message || "Delete failed", severity: "error" });
    }
  };

  if (loading) return null;

  return (
    <>
      {!hideIntro && (
        <>
          <Typography variant="h6" fontWeight={600} color="#1a1a1a" gutterBottom>
            Your Service Offerings
          </Typography>
          <Typography variant="body2" color="#666" sx={{ mb: 3 }}>
            Add services you can offer to clients. Include descriptions, pricing, and examples of your work.
          </Typography>
        </>
      )}

      <Stack spacing={2}>
        {services.map((s) => (
          <Paper
            key={s.id}
            variant="outlined"
            sx={{
              p: 2,
              borderRadius: 2,
              borderColor: "#e0e0e0",
              bgcolor: "rgba(149, 227, 186, 0.08)",
              border: "1px solid rgba(149, 227, 186, 0.4)",
            }}
          >
            <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
              <Box>
                <Typography fontWeight={600}>{s.title || s.category}</Typography>
                <Typography variant="body2" color="text.secondary">
                  {s.short_description || s.experience_summary || s.description}
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  {formatServicePriceLine(s)}
                </Typography>
              </Box>
              <Stack direction="row" spacing={0.5}>
                <IconButton size="small" onClick={() => handleOpenEdit(s)}>
                  <EditIcon fontSize="small" />
                </IconButton>
                <IconButton size="small" color="error" onClick={() => handleDelete(s.id)}>
                  <DeleteIcon fontSize="small" />
                </IconButton>
              </Stack>
            </Stack>
          </Paper>
        ))}

        <Box
          onClick={handleOpenAdd}
          sx={{
            p: 3,
            border: "2px dashed rgba(149, 227, 186, 0.6)",
            borderRadius: 2,
            cursor: "pointer",
            textAlign: "center",
            bgcolor: "rgba(149, 227, 186, 0.06)",
            "&:hover": { borderColor: "#95e3ba", bgcolor: "rgba(149, 227, 186, 0.12)" },
          }}
        >
          <AddIcon sx={{ color: "#2e7d32", fontSize: 36, mb: 1 }} />
          <Typography fontWeight={600} color="#2e7d32">
            Add New Service
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Describe what you offer and set your pricing
          </Typography>
        </Box>
      </Stack>

      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="sm" fullWidth PaperProps={{ sx: { borderRadius: 3 } }}>
        <DialogTitle sx={{ display: "flex", alignItems: "center", gap: 1 }}>
          <AddIcon /> {editingId != null ? "Edit service" : "Add New Service"}
        </DialogTitle>
        <DialogContent>
          <Stack spacing={3} sx={{ mt: 1 }}>
            {/* Section 1: What are you offering? */}
            <Paper variant="outlined" sx={{ p: 2, borderRadius: 2, bgcolor: "rgba(149, 227, 186, 0.08)", borderColor: "rgba(149, 227, 186, 0.4)" }}>
              <Typography fontWeight={600} gutterBottom>What are you offering?</Typography>
              <Stack spacing={2}>
                <FormControl fullWidth size="small">
                  <InputLabel>Broad Category</InputLabel>
                  <Select
                    value={formData.broadCategory}
                    onChange={handleChange("broadCategory")}
                    label="Broad Category"
                  >
                    {Object.keys(BROAD_CATEGORIES).map((c) => (
                      <MenuItem key={c} value={c}>{c}</MenuItem>
                    ))}
                  </Select>
                </FormControl>
                <FormControl fullWidth size="small">
                  <InputLabel>Specific Service</InputLabel>
                  <Select
                    value={formData.specificService}
                    onChange={handleChange("specificService")}
                    label="Specific Service"
                    disabled={!formData.broadCategory}
                  >
                    {(BROAD_CATEGORIES[formData.broadCategory] || []).map((c) => (
                      <MenuItem key={c} value={c}>{c}</MenuItem>
                    ))}
                  </Select>
                </FormControl>
                <TextField
                  fullWidth
                  size="small"
                  label="Service Title"
                  placeholder="e.g., Professional Minimalist Logo Design"
                  value={formData.title}
                  onChange={handleChange("title")}
                />
              </Stack>
            </Paper>

            {/* Section 2: Tell clients about this service */}
            <Paper variant="outlined" sx={{ p: 2, borderRadius: 2, bgcolor: "rgba(149, 227, 186, 0.08)", borderColor: "rgba(149, 227, 186, 0.4)" }}>
              <Typography fontWeight={600} gutterBottom>Tell clients about this service</Typography>
              <Typography variant="caption" fontStyle="italic" color="text.secondary" sx={{ display: "block", mb: 1 }}>
                Tip: You don&apos;t need to be an expert. Just be honest about your skill level and what you can deliver!
              </Typography>
              <Stack spacing={2}>
                <TextField
                  fullWidth
                  multiline
                  rows={4}
                  size="small"
                  placeholder="Describe what you will do, what the client gets, and your process..."
                  value={formData.description}
                  onChange={handleChange("description")}
                />
                <TextField
                  fullWidth
                  size="small"
                  label="Short Description (for preview cards)"
                  placeholder="A brief one-sentence summary of the service"
                  value={formData.shortDescription}
                  onChange={handleChange("shortDescription")}
                />
              </Stack>
            </Paper>

            {/* Section 3: How much do you charge? */}
            <Paper variant="outlined" sx={{ p: 2, borderRadius: 2, bgcolor: "rgba(149, 227, 186, 0.08)", borderColor: "rgba(149, 227, 186, 0.4)" }}>
              <FormLabel component="legend" sx={{ fontWeight: 600, mb: 1 }}>How much do you charge?</FormLabel>
              <FormGroup>
                <FormControlLabel
                  control={
                    <Checkbox
                      checked={formData.priceHourly}
                      onChange={(e) => setFormData((p) => ({ ...p, priceHourly: e.target.checked }))}
                    />
                  }
                  label="Hourly range"
                />
                <FormControlLabel
                  control={
                    <Checkbox
                      checked={formData.priceProject}
                      onChange={(e) => setFormData((p) => ({ ...p, priceProject: e.target.checked }))}
                    />
                  }
                  label="Project-based"
                />
                <FormControlLabel
                  control={
                    <Checkbox
                      checked={formData.priceNegotiable}
                      onChange={(e) => setFormData((p) => ({ ...p, priceNegotiable: e.target.checked }))}
                    />
                  }
                  label="Negotiable"
                />
              </FormGroup>
              {(formData.priceHourly || formData.priceProject) && (
                <Stack direction="row" alignItems="center" spacing={1} sx={{ mt: 2 }}>
                  <span>€</span>
                  <TextField
                    type="number"
                    size="small"
                    label="Min"
                    value={formData.priceMin}
                    onChange={handleChange("priceMin")}
                    sx={{ width: 100 }}
                    inputProps={{ min: 0 }}
                  />
                  <span>~ Max</span>
                  <TextField
                    type="number"
                    size="small"
                    label="Max"
                    value={formData.priceMax}
                    onChange={handleChange("priceMax")}
                    sx={{ width: 100 }}
                    inputProps={{ min: 0 }}
                  />
                </Stack>
              )}
            </Paper>

            {/* Section 4: Show your work */}
            <Paper variant="outlined" sx={{ p: 2, borderRadius: 2, bgcolor: "rgba(149, 227, 186, 0.08)", borderColor: "rgba(149, 227, 186, 0.4)" }}>
              <Typography fontWeight={600} gutterBottom>Show your work</Typography>
              <Stack direction={{ xs: "column", sm: "row" }} spacing={2} alignItems="stretch">
                <Box
                  component="label"
                  sx={{
                    flex: 1,
                    p: 3,
                    border: "2px dashed #ccc",
                    borderRadius: 2,
                    cursor: "pointer",
                    display: "flex",
                    flexDirection: "column",
                    alignItems: "center",
                    justifyContent: "center",
                    "&:hover": { borderColor: "#95e3ba", bgcolor: "rgba(149, 227, 186, 0.06)" },
                  }}
                >
                  <input type="file" accept="image/*,video/mp4" multiple hidden onChange={handleMediaUpload} disabled={uploadingMedia} />
                  <CloudUploadIcon sx={{ fontSize: 40, color: "#666", mb: 1 }} />
                  <Typography variant="body2">Upload images or MP4</Typography>
                  <Typography variant="caption" color="text.secondary">Max 9 files</Typography>
                </Box>
                <Box
                  sx={{
                    flex: 1,
                    p: 3,
                    border: "1px solid #e0e0e0",
                    borderRadius: 2,
                    display: "flex",
                    flexDirection: "column",
                    alignItems: "center",
                    justifyContent: "center",
                    bgcolor: "#fafafa",
                  }}
                >
                  <PhotoLibraryIcon sx={{ fontSize: 40, color: "#999", mb: 1 }} />
                  <Typography variant="body2" color="text.secondary">Choose from my Gallery</Typography>
                  <Typography variant="caption" color="text.secondary">(Coming soon)</Typography>
                </Box>
              </Stack>
              {formData.mediaUrls.length > 0 && (
                <Stack direction="row" flexWrap="wrap" gap={1} sx={{ mt: 2 }}>
                  {formData.mediaUrls.map((url, i) => (
                    <Box key={i} sx={{ position: "relative" }}>
                      <Box
                        component="img"
                        src={url}
                        alt=""
                        sx={{
                          width: 60,
                          height: 60,
                          objectFit: "contain",
                          borderRadius: 1,
                          bgcolor: "#f1f3f4",
                        }}
                      />
                      <IconButton size="small" sx={{ position: "absolute", top: -8, right: -8, bgcolor: "#fff" }} onClick={() => handleRemoveMedia(i)}>
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </Box>
                  ))}
                </Stack>
              )}
            </Paper>

            <Button
              variant="contained"
              startIcon={<SaveIcon />}
              onClick={handleSaveService}
              disabled={saving}
              fullWidth
              sx={{
                bgcolor: "#9fe22e",
                color: "#17202f",
                fontWeight: 700,
                py: 1.5,
                borderRadius: 2,
                "&:hover": { bgcolor: "#8bc91e" },
              }}
            >
              {saving ? "Saving..." : "Save Service"}
            </Button>
          </Stack>
        </DialogContent>
      </Dialog>

      <Snackbar open={snackbar.open} autoHideDuration={4000} onClose={() => setSnackbar((p) => ({ ...p, open: false }))} anchorOrigin={{ vertical: "bottom", horizontal: "center" }}>
        <Alert severity={snackbar.severity} onClose={() => setSnackbar((p) => ({ ...p, open: false }))}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </>
  );
}
