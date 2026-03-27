import React, { useState, useEffect, useMemo, useRef, useCallback } from "react";
import {
  Box,
  Button,
  Chip,
  CircularProgress,
  Dialog,
  DialogContent,
  Divider,
  IconButton,
  InputAdornment,
  MenuItem,
  Snackbar,
  Stack,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Typography,
  Alert,
} from "@mui/material";
import CloseIcon from "@mui/icons-material/Close";
import AddIcon from "@mui/icons-material/Add";
import OpenInNewIcon from "@mui/icons-material/OpenInNew";
import CloudUploadIcon from "@mui/icons-material/CloudUpload";
import CheckIcon from "@mui/icons-material/Check";
import { portfolioApi } from "../../features/profile/api/profile";

const INPUT_SX = {
  "& .MuiOutlinedInput-root": {
    bgcolor: "#F7F7F5",
    borderRadius: "10px",
    "& fieldset": { borderColor: "rgba(0,0,0,0.08)" },
  },
};

const RELATED_SERVICES = ["Logo Design", "Brand Identity", "Poster Design", "Pitch Deck"];

const PROJECT_TYPES = [
  "Course project",
  "Personal project",
  "Freelance / client work",
  "Competition / exhibition",
  "Other",
];

const FILE_TYPE_TAGS = ["JPG", "PNG", "PDF", "MP4", "GIF"];

const MAX_FILE_BYTES = 50 * 1024 * 1024;

function StepBadge({ n }) {
  return (
    <Box
      sx={{
        width: 36,
        height: 36,
        borderRadius: "50%",
        bgcolor: "#111",
        color: "#fff",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        fontWeight: 700,
        fontSize: "0.95rem",
        flexShrink: 0,
      }}
    >
      {n}
    </Box>
  );
}

function emptyForm() {
  return {
    title: "",
    year: "",
    projectType: "",
    projectUrl: "",
    mediaUrls: [],
    description: "",
    shortCaption: "",
    skillInput: "",
    skillTags: [],
    coEmail: "",
    coCreators: [],
    relatedSelected: [],
    visibility: "public",
  };
}

export default function AddPortfolioProjectDialog({
  open,
  onClose,
  editingItem,
  onSaved,
}) {
  const [step, setStep] = useState(0);
  const [form, setForm] = useState(emptyForm);
  const [saving, setSaving] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [snackbar, setSnackbar] = useState({ open: false, message: "", severity: "info" });
  const fileInputRef = useRef(null);

  const yearOptions = useMemo(() => {
    const end = new Date().getFullYear() + 1;
    const out = [];
    for (let y = end; y >= 1990; y--) out.push(y);
    return out;
  }, []);

  const resetForOpen = useCallback(() => {
    if (editingItem) {
      setForm({
        ...emptyForm(),
        title: editingItem.title || "",
        year: editingItem.year ?? "",
        projectType: editingItem.client_name || "",
        projectUrl: editingItem.project_link || "",
        mediaUrls: [
          editingItem.cover_image_url,
          ...(editingItem.media_urls || []),
        ].filter(Boolean),
        description: editingItem.description || "",
        shortCaption: editingItem.short_caption || "",
        skillTags: [...(editingItem.tags || [])],
        coCreators: [...(editingItem.co_creators || [])],
        relatedSelected: [...(editingItem.related_services || [])],
        visibility: editingItem.is_public !== false ? "public" : "private",
      });
    } else {
      setForm(emptyForm());
    }
    setStep(0);
  }, [editingItem]);

  useEffect(() => {
    if (open) resetForOpen();
  }, [open, resetForOpen]);

  const setField = (key) => (e) => {
    const v = e?.target?.value ?? e;
    setForm((p) => ({ ...p, [key]: v }));
  };

  const coverUrl = form.mediaUrls[0] || null;
  const extraPreviews = form.mediaUrls.slice(1, 4);

  const uploadFiles = async (files) => {
    const list = Array.from(files || []);
    if (!list.length) return;
    setUploading(true);
    try {
      const next = [...form.mediaUrls];
      for (const file of list) {
        if (file.size > MAX_FILE_BYTES) {
          setSnackbar({ open: true, message: "Each file must be 50MB or less.", severity: "error" });
          continue;
        }
        if (!file.type.startsWith("image/")) {
          setSnackbar({
            open: true,
            message: "Please upload images for the gallery (JPG, PNG, WebP). Other types coming soon.",
            severity: "warning",
          });
          continue;
        }
        const res = await portfolioApi.uploadProjectCover(file);
        const url = res?.data?.data?.url;
        if (url) next.push(url);
      }
      setForm((p) => ({ ...p, mediaUrls: next }));
    } catch (err) {
      setSnackbar({ open: true, message: err?.message || "Upload failed", severity: "error" });
    } finally {
      setUploading(false);
    }
  };

  const onDrop = (e) => {
    e.preventDefault();
    e.stopPropagation();
    uploadFiles(e.dataTransfer?.files);
  };

  const removeMediaAt = (idx) => {
    setForm((p) => ({
      ...p,
      mediaUrls: p.mediaUrls.filter((_, i) => i !== idx),
    }));
  };

  const addSkillTag = () => {
    const t = form.skillInput.trim().replace(/^,+|,+$/g, "");
    if (!t) return;
    if (form.skillTags.includes(t)) {
      setForm((p) => ({ ...p, skillInput: "" }));
      return;
    }
    if (form.skillTags.length >= 20) return;
    setForm((p) => ({ ...p, skillTags: [...p.skillTags, t], skillInput: "" }));
  };

  const onSkillKeyDown = (e) => {
    if (e.key === "Enter" || e.key === ",") {
      e.preventDefault();
      addSkillTag();
    }
  };

  const addCoCreator = () => {
    const email = form.coEmail.trim().toLowerCase();
    if (!email.endsWith("@aalto.fi")) {
      setSnackbar({ open: true, message: "Use an @aalto.fi email address.", severity: "warning" });
      return;
    }
    if (form.coCreators.some((c) => c.email === email)) return;
    const local = email.split("@")[0].replace(/\./g, " ");
    const name = local.replace(/\b\w/g, (c) => c.toUpperCase());
    setForm((p) => ({
      ...p,
      coCreators: [...p.coCreators, { email, name }],
      coEmail: "",
    }));
  };

  const toggleRelated = (label) => {
    setForm((p) => {
      const has = p.relatedSelected.includes(label);
      return {
        ...p,
        relatedSelected: has ? p.relatedSelected.filter((x) => x !== label) : [...p.relatedSelected, label],
      };
    });
  };

  const validateStep = (s) => {
    if (s === 0) {
      if (!form.title.trim()) return "Project title is required.";
      if (form.year === "" || form.year == null) return "Year is required.";
      if (!form.projectType) return "Project type is required.";
    }
    if (s === 1) {
      if (!form.mediaUrls.length) return "Upload at least one image (used as cover).";
    }
    if (s === 2) {
      if (!form.description.trim()) return "Project description is required.";
      if (!form.shortCaption.trim()) return "Short caption is required.";
      if (form.shortCaption.length > 120) return "Short caption must be 120 characters or less.";
    }
    if (s === 3) {
      if (!form.skillTags.length) return "Add at least one skill or tool.";
    }
    return null;
  };

  const goNext = () => {
    const err = validateStep(step);
    if (err) {
      setSnackbar({ open: true, message: err, severity: "error" });
      return;
    }
    setStep((x) => Math.min(3, x + 1));
  };

  const goBack = () => setStep((x) => Math.max(0, x - 1));

  const buildPayload = (published) => {
    const [cover, ...rest] = form.mediaUrls;
    const primaryService = form.relatedSelected[0] || null;
    const y = typeof form.year === "number" ? form.year : parseInt(String(form.year), 10);
    return {
      title: form.title.trim(),
      short_caption: form.shortCaption.trim(),
      description: form.description.trim() || null,
      year: Number.isFinite(y) ? y : null,
      service_category: primaryService,
      client_name: form.projectType || null,
      project_link: form.projectUrl.trim() || null,
      cover_image_url: cover || null,
      media_urls: rest,
      tags: form.skillTags,
      related_services: form.relatedSelected,
      co_creators: form.coCreators.map((c) => ({ email: c.email, name: c.name })),
      is_published: published,
      is_public: form.visibility === "public",
    };
  };

  const submit = async (published) => {
    const err = validateStep(0) || validateStep(1) || validateStep(2) || validateStep(3);
    if (err) {
      setSnackbar({ open: true, message: err, severity: "error" });
      return;
    }
    setSaving(true);
    try {
      const payload = buildPayload(published);
      let res;
      if (editingItem?.id) {
        res = await portfolioApi.updatePortfolioItem(editingItem.id, payload);
      } else {
        res = await portfolioApi.createPortfolioItem(payload);
      }
      if (res?.data && res.data.success === false) throw new Error("Save failed");
      setSnackbar({
        open: true,
        message: published ? "Project published." : "Draft saved.",
        severity: "success",
      });
      onSaved?.(res?.data?.data);
      onClose();
    } catch (e) {
      setSnackbar({ open: true, message: e?.message || "Could not save project", severity: "error" });
    } finally {
      setSaving(false);
    }
  };

  return (
    <>
      <Dialog
        open={open}
        onClose={() => !saving && onClose()}
        maxWidth="md"
        fullWidth
        PaperProps={{ sx: { borderRadius: "12px", maxHeight: "calc(100vh - 48px)" } }}
      >
        <Box sx={{ px: { xs: 2, sm: 3 }, pt: 2.5, pb: 1 }}>
          <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
            <Box>
              <Typography sx={{ color: "#1976d2", fontWeight: 700, fontSize: "1.35rem" }}>
                {editingItem ? "Edit Portfolio Project" : "Add Portfolio Project"}
              </Typography>
              <Typography variant="body2" sx={{ color: "#666", mt: 0.75, maxWidth: 520 }}>
                Select projects that represent your strengths or work you&apos;re most proud of. Course projects,
                personal work, and freelance work are all welcome.
              </Typography>
            </Box>
            <Button
              variant="outlined"
              onClick={() => !saving && onClose()}
              sx={{ textTransform: "none", borderRadius: "10px", color: "#555", borderColor: "#ccc", flexShrink: 0 }}
            >
              Cancel
            </Button>
          </Stack>
        </Box>
        <Divider />
        <DialogContent sx={{ px: { xs: 2, sm: 3 }, py: 3 }}>
          {step === 0 && (
            <Stack spacing={3}>
              <Stack direction="row" alignItems="center" spacing={1.5}>
                <StepBadge n={1} />
                <Typography fontWeight={700} color="#111">
                  What is this project?
                </Typography>
              </Stack>
              <Stack spacing={2.5}>
                <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
                  <TextField
                    fullWidth
                    required
                    label="Project Title"
                    placeholder="e.g., Brand Identity for Startup X"
                    value={form.title}
                    onChange={setField("title")}
                    sx={INPUT_SX}
                    InputLabelProps={{ required: false }}
                  />
                  <TextField
                    select
                    fullWidth
                    required
                    label="Year Completed"
                    value={form.year === "" ? "" : String(form.year)}
                    onChange={(e) => {
                      const v = e.target.value;
                      setForm((p) => ({ ...p, year: v === "" ? "" : parseInt(v, 10) }));
                    }}
                    SelectProps={{ displayEmpty: true }}
                    sx={INPUT_SX}
                    InputLabelProps={{ required: false }}
                  >
                    <MenuItem value="" disabled>
                      Select year
                    </MenuItem>
                    {yearOptions.map((y) => (
                      <MenuItem key={y} value={String(y)}>
                        {y}
                      </MenuItem>
                    ))}
                  </TextField>
                </Stack>
                <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
                  <TextField
                    select
                    fullWidth
                    required
                    label="Project Type"
                    value={form.projectType}
                    onChange={(e) => setForm((p) => ({ ...p, projectType: e.target.value }))}
                    SelectProps={{ displayEmpty: true }}
                    sx={INPUT_SX}
                    InputLabelProps={{ required: false }}
                  >
                    <MenuItem value="" disabled>
                      Select type
                    </MenuItem>
                    {PROJECT_TYPES.map((t) => (
                      <MenuItem key={t} value={t}>
                        {t}
                      </MenuItem>
                    ))}
                  </TextField>
                  <TextField
                    fullWidth
                    label={
                      <span>
                        Project URL{" "}
                        <Typography component="span" variant="body2" sx={{ color: "#757575", fontWeight: 400 }}>
                          (optional)
                        </Typography>
                      </span>
                    }
                    placeholder="yourportfolio.com/project"
                    value={form.projectUrl}
                    onChange={setField("projectUrl")}
                    sx={INPUT_SX}
                    InputProps={{
                      endAdornment: (
                        <InputAdornment position="end">
                          <OpenInNewIcon sx={{ color: "#999", fontSize: 20 }} />
                        </InputAdornment>
                      ),
                    }}
                    InputLabelProps={{ shrink: true }}
                  />
                </Stack>
                <Typography variant="caption" sx={{ color: "#757575", mt: -1 }}>
                  Behance, Dribbble, personal site, etc.
                </Typography>
              </Stack>
            </Stack>
          )}

          {step === 1 && (
            <Stack spacing={3}>
              <Stack direction="row" alignItems="center" spacing={1.5}>
                <StepBadge n={2} />
                <Typography fontWeight={700} color="#111">
                  Upload your work
                </Typography>
              </Stack>
              <Stack direction="row" spacing={1.5} flexWrap="wrap" useFlexGap>
                {coverUrl && (
                  <Box
                    sx={{
                      width: 100,
                      height: 100,
                      borderRadius: "10px",
                      overflow: "hidden",
                      position: "relative",
                      background: "linear-gradient(135deg, #7c3aed 0%, #a78bfa 100%)",
                    }}
                  >
                    <Box component="img" src={coverUrl} alt="" sx={{ width: "100%", height: "100%", objectFit: "cover" }} />
                    <Chip
                      label="COVER"
                      size="small"
                      sx={{
                        position: "absolute",
                        top: 6,
                        left: 6,
                        height: 22,
                        fontSize: "0.65rem",
                        fontWeight: 700,
                        bgcolor: "#111",
                        color: "#fff",
                      }}
                    />
                    <IconButton
                      size="small"
                      onClick={() => removeMediaAt(0)}
                      sx={{ position: "absolute", bottom: 4, right: 4, bgcolor: "rgba(255,255,255,0.9)" }}
                    >
                      <CloseIcon fontSize="small" />
                    </IconButton>
                  </Box>
                )}
                {extraPreviews.map((url, i) => (
                  <Box
                    key={url + i}
                    sx={{
                      width: 100,
                      height: 100,
                      borderRadius: "10px",
                      overflow: "hidden",
                      position: "relative",
                      background: i % 2 ? "linear-gradient(135deg, #ec4899 0%, #fda4af 100%)" : "linear-gradient(135deg, #06b6d4 0%, #38bdf8 100%)",
                    }}
                  >
                    <Box component="img" src={url} alt="" sx={{ width: "100%", height: "100%", objectFit: "cover" }} />
                    <IconButton
                      size="small"
                      onClick={() => removeMediaAt(i + 1)}
                      sx={{ position: "absolute", bottom: 4, right: 4, bgcolor: "rgba(255,255,255,0.9)" }}
                    >
                      <CloseIcon fontSize="small" />
                    </IconButton>
                  </Box>
                ))}
                <Box
                  onClick={() => fileInputRef.current?.click()}
                  sx={{
                    width: 100,
                    height: 100,
                    borderRadius: "10px",
                    border: "2px dashed #ccc",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    cursor: "pointer",
                    bgcolor: "#faf9f7",
                  }}
                >
                  <AddIcon sx={{ fontSize: 36, color: "#bbb" }} />
                </Box>
              </Stack>
              <input
                ref={fileInputRef}
                type="file"
                multiple
                accept="image/jpeg,image/png,image/webp,image/gif"
                hidden
                onChange={(e) => {
                  uploadFiles(e.target.files);
                  e.target.value = "";
                }}
              />
              <Box
                onDragOver={(e) => e.preventDefault()}
                onDrop={onDrop}
                sx={{
                  border: "2px dashed #ccc",
                  borderRadius: "12px",
                  bgcolor: "#faf9f7",
                  py: 4,
                  px: 2,
                  textAlign: "center",
                }}
              >
                <Box
                  sx={{
                    width: 48,
                    height: 48,
                    borderRadius: "10px",
                    bgcolor: "#ede8e0",
                    display: "inline-flex",
                    alignItems: "center",
                    justifyContent: "center",
                    mb: 1.5,
                  }}
                >
                  {uploading ? <CircularProgress size={24} /> : <CloudUploadIcon sx={{ color: "#8d7a68" }} />}
                </Box>
                <Typography variant="body1" sx={{ color: "#333" }}>
                  Drag & drop files here, or{" "}
                  <Box
                    component="button"
                    type="button"
                    onClick={() => fileInputRef.current?.click()}
                    sx={{
                      border: "none",
                      background: "none",
                      p: 0,
                      color: "#1976d2",
                      cursor: "pointer",
                      font: "inherit",
                      fontWeight: 600,
                    }}
                  >
                    browse
                  </Box>
                </Typography>
                <Typography variant="body2" sx={{ color: "#757575", mt: 0.5 }}>
                  The first image will be used as the cover
                </Typography>
                <Stack direction="row" spacing={0.75} justifyContent="center" flexWrap="wrap" useFlexGap sx={{ mt: 2 }}>
                  {FILE_TYPE_TAGS.map((t) => (
                    <Box
                      key={t}
                      sx={{
                        px: 1.25,
                        py: 0.35,
                        borderRadius: "6px",
                        border: "1px solid #ddd",
                        bgcolor: "#fff",
                        fontSize: "0.75rem",
                        color: "#555",
                      }}
                    >
                      {t}
                    </Box>
                  ))}
                  <Typography variant="caption" sx={{ color: "#888", alignSelf: "center", ml: 1 }}>
                    · Max 50MB per file
                  </Typography>
                </Stack>
              </Box>
            </Stack>
          )}

          {step === 2 && (
            <Stack spacing={3}>
              <Stack direction="row" alignItems="center" spacing={1.5}>
                <StepBadge n={3} />
                <Typography fontWeight={700} color="#111">
                  Describe this project
                </Typography>
              </Stack>
              <Box>
                <Typography variant="body2" fontWeight={700} sx={{ mb: 0.75 }}>
                  Project Description <Box component="span" sx={{ color: "error.main" }}>*</Box>
                </Typography>
                <TextField
                  fullWidth
                  multiline
                  minRows={6}
                  placeholder="Describe the project background, your role, the process, and what you're most proud of..."
                  value={form.description}
                  onChange={setField("description")}
                  sx={INPUT_SX}
                />
                <Typography variant="caption" sx={{ color: "#757575", mt: 0.75, display: "block" }}>
                  This is shown on your project detail page. Be specific — clients want to know your thought process.
                </Typography>
              </Box>
              <Box>
                <Typography variant="body2" fontWeight={700} sx={{ mb: 0.75 }}>
                  Short Caption <Box component="span" sx={{ color: "error.main" }}>*</Box>
                </Typography>
                <TextField
                  fullWidth
                  placeholder="One sentence that summarises this project (shown on preview cards)"
                  value={form.shortCaption}
                  onChange={setField("shortCaption")}
                  inputProps={{ maxLength: 120 }}
                  sx={INPUT_SX}
                />
                <Typography variant="caption" sx={{ color: "#757575", mt: 0.75, display: "block" }}>
                  Max 120 characters · shown on Hire Talent listing cards
                </Typography>
              </Box>
            </Stack>
          )}

          {step === 3 && (
            <Stack spacing={3}>
              <Stack direction="row" alignItems="center" spacing={1.5}>
                <StepBadge n={4} />
                <Typography fontWeight={700} color="#111">
                  Tags, skills & connections
                </Typography>
              </Stack>
              <Box>
                <Typography variant="body2" fontWeight={700} sx={{ mb: 0.75 }}>
                  Skills & Tools Used <Box component="span" sx={{ color: "error.main" }}>*</Box>
                </Typography>
                <Box
                  sx={{
                    border: "1px solid rgba(0,0,0,0.12)",
                    borderRadius: "10px",
                    p: 1,
                    minHeight: 52,
                    display: "flex",
                    flexWrap: "wrap",
                    gap: 0.75,
                    alignItems: "center",
                    bgcolor: "#F7F7F5",
                  }}
                >
                  {form.skillTags.map((tag) => (
                    <Chip
                      key={tag}
                      label={tag}
                      onDelete={() => setForm((p) => ({ ...p, skillTags: p.skillTags.filter((t) => t !== tag) }))}
                      sx={{ bgcolor: "#111", color: "#fff", "& .MuiChip-deleteIcon": { color: "rgba(255,255,255,0.7)" } }}
                    />
                  ))}
                  <TextField
                    variant="standard"
                    placeholder="Add a skill or tool..."
                    value={form.skillInput}
                    onChange={setField("skillInput")}
                    onKeyDown={onSkillKeyDown}
                    InputProps={{ disableUnderline: true }}
                    sx={{ flex: 1, minWidth: 160 }}
                  />
                </Box>
                <Typography variant="caption" sx={{ color: "#757575", mt: 0.75, display: "block" }}>
                  Press Enter or comma to add · these appear on the Hire Talent filter
                </Typography>
              </Box>
              <Box>
                <Typography variant="body2" sx={{ color: "#757575", mb: 0.75 }}>
                  Co-creators (optional)
                </Typography>
                {form.coCreators.map((c) => (
                  <Stack
                    key={c.email}
                    direction="row"
                    alignItems="center"
                    spacing={1.5}
                    sx={{ p: 1.5, borderRadius: "10px", bgcolor: "#F7F7F5", mb: 1 }}
                  >
                    <Box
                      sx={{
                        width: 40,
                        height: 40,
                        borderRadius: "50%",
                        bgcolor: "#7c3aed",
                        color: "#fff",
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                        fontWeight: 700,
                      }}
                    >
                      {(c.name || c.email)[0].toUpperCase()}
                    </Box>
                    <Box sx={{ flex: 1 }}>
                      <Typography fontWeight={700}>{c.name}</Typography>
                      <Typography variant="caption" sx={{ color: "#666" }}>
                        {c.email}
                      </Typography>
                    </Box>
                    <IconButton
                      size="small"
                      onClick={() =>
                        setForm((p) => ({ ...p, coCreators: p.coCreators.filter((x) => x.email !== c.email) }))
                      }
                    >
                      <CloseIcon fontSize="small" />
                    </IconButton>
                  </Stack>
                ))}
                <Stack direction="row" spacing={1} sx={{ mt: 1 }}>
                  <TextField
                    fullWidth
                    placeholder="Tag by @aalto.fi email address"
                    value={form.coEmail}
                    onChange={setField("coEmail")}
                    sx={INPUT_SX}
                  />
                  <Button
                    variant="contained"
                    onClick={addCoCreator}
                    sx={{ bgcolor: "#111", textTransform: "none", borderRadius: "10px", px: 2, flexShrink: 0 }}
                  >
                    + Add
                  </Button>
                </Stack>
                <Typography variant="caption" sx={{ color: "#757575", mt: 0.75, display: "block" }}>
                  Tagged co-creators will see this project on their own profile
                </Typography>
              </Box>
              <Box>
                <Typography variant="body2" sx={{ color: "#757575", mb: 0.75 }}>
                  Related Services (optional)
                </Typography>
                <Alert
                  severity="info"
                  icon={false}
                  sx={{
                    mb: 2,
                    bgcolor: "#e3f2fd",
                    borderLeft: "4px solid #1976d2",
                    color: "#1565c0",
                    borderRadius: "10px",
                  }}
                >
                  Linking this project to a service helps clients discover your work when browsing that service.
                </Alert>
                <GridTwoCol>
                  {RELATED_SERVICES.map((name) => {
                    const sel = form.relatedSelected.includes(name);
                    return (
                      <Box
                        key={name}
                        onClick={() => toggleRelated(name)}
                        sx={{
                          p: 1.5,
                          borderRadius: "10px",
                          border: sel ? "2px solid #1976d2" : "1px solid #e0e0e0",
                          bgcolor: sel ? "rgba(25,118,210,0.06)" : "#F7F7F5",
                          cursor: "pointer",
                          display: "flex",
                          alignItems: "center",
                          gap: 1,
                        }}
                      >
                        <Box
                          sx={{
                            width: 22,
                            height: 22,
                            borderRadius: "4px",
                            border: sel ? "none" : "2px solid #bbb",
                            bgcolor: sel ? "#1976d2" : "transparent",
                            display: "flex",
                            alignItems: "center",
                            justifyContent: "center",
                          }}
                        >
                          {sel && <CheckIcon sx={{ fontSize: 16, color: "#fff" }} />}
                        </Box>
                        <Typography fontWeight={600} fontSize="0.9rem">
                          {name}
                        </Typography>
                      </Box>
                    );
                  })}
                </GridTwoCol>
              </Box>
              <Box>
                <Typography fontWeight={700} sx={{ mb: 1 }}>
                  Visibility
                </Typography>
                <Box
                  sx={{
                    p: 2,
                    borderRadius: "10px",
                    bgcolor: "#F7F7F5",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "space-between",
                    flexWrap: "wrap",
                    gap: 2,
                  }}
                >
                  <Box>
                    <Typography fontWeight={700}>Who can see this project?</Typography>
                    <Typography variant="body2" sx={{ color: "#666" }}>
                      Public projects appear on your profile and in Hire Talent search results
                    </Typography>
                  </Box>
                  <ToggleButtonGroup
                    exclusive
                    value={form.visibility}
                    onChange={(_, v) => v && setForm((p) => ({ ...p, visibility: v }))}
                    sx={{
                      "& .MuiToggleButton-root": {
                        px: 2,
                        py: 0.75,
                        textTransform: "none",
                        borderRadius: "8px !important",
                        border: "1px solid #ddd !important",
                      },
                      "& .Mui-selected": {
                        bgcolor: "#111 !important",
                        color: "#fff !important",
                      },
                    }}
                  >
                    <ToggleButton value="public">Public</ToggleButton>
                    <ToggleButton value="private">Private</ToggleButton>
                  </ToggleButtonGroup>
                </Box>
              </Box>
            </Stack>
          )}

          <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mt: 4, pt: 2 }}>
            {step < 3 ? (
              <>
                <Button disabled={step === 0} onClick={goBack} sx={{ textTransform: "none", color: "#555" }}>
                  Back
                </Button>
                <Button variant="contained" onClick={goNext} sx={{ textTransform: "none", borderRadius: "10px", px: 3 }}>
                  Continue
                </Button>
              </>
            ) : (
              <>
                <Typography variant="caption" sx={{ color: "#757575" }}>
                  <Box component="span" sx={{ color: "error.main" }}>
                    *
                  </Box>{" "}
                  Required fields
                </Typography>
                <Stack direction="row" spacing={1.5}>
                  <Button
                    variant="outlined"
                    disabled={saving}
                    onClick={() => submit(false)}
                    sx={{ textTransform: "none", borderRadius: "10px", borderColor: "#ccc", color: "#111" }}
                  >
                    Save as Draft
                  </Button>
                  <Button
                    variant="contained"
                    disabled={saving}
                    onClick={() => submit(true)}
                    endIcon={saving ? <CircularProgress size={18} color="inherit" /> : <CheckIcon />}
                    sx={{ textTransform: "none", borderRadius: "10px", px: 2.5, bgcolor: "#1976d2" }}
                  >
                    Publish Project
                  </Button>
                </Stack>
              </>
            )}
          </Stack>
        </DialogContent>
      </Dialog>
      <Snackbar
        open={snackbar.open}
        autoHideDuration={4000}
        onClose={() => setSnackbar((p) => ({ ...p, open: false }))}
        anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
      >
        <Alert severity={snackbar.severity} onClose={() => setSnackbar((p) => ({ ...p, open: false }))}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </>
  );
}

function GridTwoCol({ children }) {
  return (
    <Box
      sx={{
        display: "grid",
        gridTemplateColumns: { xs: "1fr", sm: "1fr 1fr" },
        gap: 1.5,
      }}
    >
      {children}
    </Box>
  );
}
