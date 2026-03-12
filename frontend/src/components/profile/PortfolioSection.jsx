import React, { useState } from "react";
import {
  Box,
  Button,
  CircularProgress,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Grid,
  IconButton,
  Stack,
  TextField,
  Typography,
  Alert,
  Snackbar,
  MenuItem,
  InputAdornment,
} from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import EditIcon from "@mui/icons-material/Edit";
import DeleteIcon from "@mui/icons-material/Delete";
import CloseIcon from "@mui/icons-material/Close";
import ImageIcon from "@mui/icons-material/Image";
import UploadFileIcon from "@mui/icons-material/UploadFile";
import LinkIcon from "@mui/icons-material/Link";
import { portfolioApi } from "../../features/profile/api/profile";

const PORTFOLIO_TIPS = [
  "Use high-quality images (at least 1200px wide)",
  "Write clear descriptions explaining the context",
  "Tag your work with relevant skills (max 10 tags)",
  "Link each work to a service category",
];

const SERVICE_CATEGORIES = [
  "Logo & Brand Identity",
  "Illustration & Drawing",
  "Print Design",
  "Presentation Design",
  "Web Design",
  "UI/UX Design",
  "Video Editing",
  "Animation",
  "Photography",
];

export default function PortfolioSection({ items, onCreate, onUpdate, onDelete }) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingItem, setEditingItem] = useState(null);
  const [formData, setFormData] = useState({
    title: "",
    description: "",
    service_category: "",
    client_name: "",
    project_link: "",
    cover_image_url: "",
    tags: [],
    year: new Date().getFullYear(),
  });
  const [newTag, setNewTag] = useState("");
  const [saving, setSaving] = useState(false);
  const [uploadingImage, setUploadingImage] = useState(false);
  const [snackbar, setSnackbar] = useState({ open: false, message: "", severity: "success" });

  const handleOpenCreate = () => {
    setEditingItem(null);
    setFormData({
      title: "",
      description: "",
      service_category: "",
      client_name: "",
      project_link: "",
      cover_image_url: "",
      tags: [],
      year: new Date().getFullYear(),
    });
    setDialogOpen(true);
  };

  const handleOpenEdit = (item) => {
    setEditingItem(item);
    setFormData({
      title: item.title || "",
      description: item.description || "",
      service_category: item.service_category || "",
      client_name: item.client_name || "",
      project_link: item.project_link || "",
      cover_image_url: item.cover_image_url || "",
      tags: item.tags || [],
      year: item.year || new Date().getFullYear(),
    });
    setDialogOpen(true);
  };

  const handleClose = () => {
    setDialogOpen(false);
    setEditingItem(null);
    setNewTag("");
  };

  const handleChange = (field) => (event) => {
    setFormData((prev) => ({ ...prev, [field]: event.target.value }));
  };

  const handleAddTag = () => {
    if (!newTag.trim()) return;
    if (formData.tags.length >= 10) {
      setSnackbar({ open: true, message: "Maximum 10 tags allowed", severity: "warning" });
      return;
    }
    if (formData.tags.includes(newTag.trim())) {
      setSnackbar({ open: true, message: "Tag already exists", severity: "warning" });
      return;
    }
    setFormData((prev) => ({ ...prev, tags: [...prev.tags, newTag.trim()] }));
    setNewTag("");
  };

  const handleRemoveTag = (tag) => {
    setFormData((prev) => ({
      ...prev,
      tags: prev.tags.filter((t) => t !== tag),
    }));
  };

  const handleCoverUpload = async (event) => {
    const file = event.target.files?.[0];
    if (!file) return;
    setUploadingImage(true);
    try {
      const response = await portfolioApi.uploadProjectCover(file);
      const uploaded = response?.data?.data;
      if (!uploaded?.url) {
        throw new Error("Missing file URL");
      }
      setFormData((prev) => ({ ...prev, cover_image_url: uploaded.url }));
      setSnackbar({ open: true, message: "Cover image uploaded", severity: "success" });
    } catch (err) {
      setSnackbar({ open: true, message: err?.message || "Image upload failed", severity: "error" });
    } finally {
      setUploadingImage(false);
      event.target.value = "";
    }
  };

  const handleSave = async () => {
    if (!formData.title.trim()) {
      setSnackbar({ open: true, message: "Title is required", severity: "error" });
      return;
    }

    setSaving(true);
    try {
      const data = {
        ...formData,
        service_category: formData.service_category?.trim() || null,
        project_link: formData.project_link?.trim() || null,
        cover_image_url: formData.cover_image_url?.trim() || null,
        year: formData.year ? parseInt(formData.year, 10) : null,
      };
      const result = editingItem
        ? await onUpdate(editingItem.id, data)
        : await onCreate(data);

      if (result.success) {
        setSnackbar({
          open: true,
          message: editingItem ? "Portfolio item updated" : "Portfolio item created",
          severity: "success",
        });
        handleClose();
      } else {
        setSnackbar({ open: true, message: result.error || "Save failed", severity: "error" });
      }
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id) => {
    if (!window.confirm("Are you sure you want to delete this portfolio item?")) return;
    const result = await onDelete(id);
    if (result.success) {
      setSnackbar({ open: true, message: "Portfolio item deleted", severity: "success" });
    } else {
      setSnackbar({ open: true, message: result.error || "Delete failed", severity: "error" });
    }
  };

  return (
    <>
      <Typography variant="h6" fontWeight={600} color="#1a1a1a" gutterBottom>
        Portfolio Works
      </Typography>
      <Typography variant="body2" color="#666" sx={{ mb: 3 }}>
        Upload your best work. Each piece should have a title, description, and tags.
      </Typography>

      <Grid container spacing={2}>
        {items.map((item) => (
          <Grid item xs={6} md={3} key={item.id}>
            <Box
              sx={{
                position: "relative",
                borderRadius: 2,
                overflow: "hidden",
                border: "1px solid #e0e0e0",
                "&:hover .actions": { opacity: 1 },
              }}
            >
              <Box
                sx={{
                  aspectRatio: "4/3",
                  bgcolor: "#1a1a1a",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                }}
              >
                {item.cover_image_url ? (
                  <Box
                    component="img"
                    src={item.cover_image_url}
                    alt={item.title}
                    sx={{ width: "100%", height: "100%", objectFit: "cover" }}
                  />
                ) : (
                  <ImageIcon sx={{ color: "rgba(255,255,255,0.3)", fontSize: 48 }} />
                )}
              </Box>
              <Box
                className="actions"
                sx={{
                  position: "absolute",
                  top: 8,
                  right: 8,
                  opacity: 0,
                  transition: "opacity 0.2s",
                }}
              >
                <IconButton
                  size="small"
                  onClick={() => handleOpenEdit(item)}
                  sx={{ bgcolor: "rgba(255,255,255,0.9)", "&:hover": { bgcolor: "#fff" } }}
                >
                  <EditIcon fontSize="small" />
                </IconButton>
                <IconButton
                  size="small"
                  onClick={() => handleDelete(item.id)}
                  sx={{ bgcolor: "rgba(255,255,255,0.9)", "&:hover": { bgcolor: "#fff" } }}
                >
                  <DeleteIcon fontSize="small" color="error" />
                </IconButton>
              </Box>
              {item.title && (
                <Box sx={{ p: 1.5, bgcolor: "#fff" }}>
                  <Typography variant="body2" fontWeight={600} noWrap>
                    {item.title}
                  </Typography>
                </Box>
              )}
            </Box>
          </Grid>
        ))}

        <Grid item xs={6} md={3}>
          <Box
            onClick={handleOpenCreate}
            sx={{
              aspectRatio: "4/3",
              border: "2px dashed #ccc",
              borderRadius: 2,
              display: "flex",
              flexDirection: "column",
              alignItems: "center",
              justifyContent: "center",
              cursor: "pointer",
              "&:hover": { borderColor: "#1976d2", bgcolor: "rgba(25, 118, 210, 0.04)" },
            }}
          >
            <AddIcon sx={{ color: "#666", fontSize: 40, mb: 1 }} />
            <Typography variant="body2" color="#666">
              + Upload New Work
            </Typography>
          </Box>
        </Grid>
      </Grid>

      <Box sx={{ mt: 4 }}>
        <Typography variant="subtitle1" fontWeight={600} color="#1a1a1a" gutterBottom>
          Tips for great portfolio pieces:
        </Typography>
        <Stack component="ul" sx={{ m: 0, pl: 2.5, color: "#666" }}>
          {PORTFOLIO_TIPS.map((tip, i) => (
            <Typography key={i} component="li" variant="body2" sx={{ mb: 0.5 }}>
              {tip}
            </Typography>
          ))}
        </Stack>
      </Box>

      <Dialog open={dialogOpen} onClose={handleClose} maxWidth="sm" fullWidth>
        <DialogTitle>
          <Stack direction="row" justifyContent="space-between" alignItems="center">
            <Typography variant="h6">{editingItem ? "Edit Work" : "Add Work"}</Typography>
            <IconButton onClick={handleClose} size="small">
              <CloseIcon />
            </IconButton>
          </Stack>
        </DialogTitle>
        <DialogContent>
          <Stack spacing={3} sx={{ mt: 1 }}>
            <TextField
              fullWidth
              label="Title"
              value={formData.title}
              onChange={handleChange("title")}
              required
            />
            <TextField
              select
              fullWidth
              label="Service Category"
              value={formData.service_category}
              onChange={handleChange("service_category")}
              helperText="Choose the main service this work represents"
            >
              {SERVICE_CATEGORIES.map((category) => (
                <MenuItem key={category} value={category}>
                  {category}
                </MenuItem>
              ))}
            </TextField>
            <TextField
              fullWidth
              label="Client Name"
              value={formData.client_name}
              onChange={handleChange("client_name")}
              placeholder="Optional"
            />
            <TextField
              fullWidth
              label="Year"
              type="number"
              value={formData.year}
              onChange={handleChange("year")}
              inputProps={{ min: 2000, max: 2100 }}
            />
            <TextField
              fullWidth
              label="Description"
              value={formData.description}
              onChange={handleChange("description")}
              multiline
              rows={4}
            />
            <TextField
              fullWidth
              label="Project Link"
              value={formData.project_link}
              onChange={handleChange("project_link")}
              placeholder="https://..."
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <LinkIcon sx={{ color: "text.secondary", fontSize: 20 }} />
                  </InputAdornment>
                ),
              }}
            />

            <Box>
              <Typography variant="subtitle2" gutterBottom>
                Cover Image
              </Typography>
              <Stack direction="row" spacing={1.2} alignItems="center">
                <Button
                  variant="outlined"
                  component="label"
                  startIcon={uploadingImage ? <CircularProgress size={16} /> : <UploadFileIcon />}
                  disabled={uploadingImage}
                >
                  {uploadingImage ? "Uploading..." : "Upload Image"}
                  <input hidden type="file" accept="image/*" onChange={handleCoverUpload} />
                </Button>
                <Typography variant="body2" color="text.secondary">
                  JPG, PNG, WEBP, GIF. Max 50MB.
                </Typography>
              </Stack>
              {formData.cover_image_url && (
                <Box
                  component="img"
                  src={formData.cover_image_url}
                  alt="Cover preview"
                  sx={{
                    mt: 1.5,
                    width: "100%",
                    maxHeight: 220,
                    objectFit: "cover",
                    borderRadius: 1.5,
                    border: "1px solid #e0e0e0",
                  }}
                />
              )}
            </Box>
            <Box>
              <Typography variant="subtitle2" gutterBottom>
                Tags (max 10)
              </Typography>
              <Stack direction="row" flexWrap="wrap" gap={1} mb={1}>
                {formData.tags.map((tag) => (
                  <Chip key={tag} label={tag} size="small" onDelete={() => handleRemoveTag(tag)} />
                ))}
              </Stack>
              <Stack direction="row" spacing={1}>
                <TextField
                  size="small"
                  placeholder="Add tag"
                  value={newTag}
                  onChange={(e) => setNewTag(e.target.value)}
                  onKeyPress={(e) => e.key === "Enter" && handleAddTag()}
                  sx={{ flex: 1 }}
                />
                <Button variant="outlined" size="small" onClick={handleAddTag} disabled={formData.tags.length >= 10}>
                  Add
                </Button>
              </Stack>
            </Box>
          </Stack>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          <Button onClick={handleClose} color="inherit">
            Cancel
          </Button>
          <Button onClick={handleSave} variant="contained" disabled={saving}>
            {saving ? "Saving..." : "Save"}
          </Button>
        </DialogActions>
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
