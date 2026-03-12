import React, { useState, useEffect, useRef } from "react";
import {
  Avatar,
  Box,
  Button,
  Grid,
  Paper,
  Stack,
  TextField,
  Typography,
  Alert,
  Snackbar,
  Collapse,
  IconButton,
} from "@mui/material";
import SaveIcon from "@mui/icons-material/Save";
import PersonIcon from "@mui/icons-material/Person";
import PhotoCameraIcon from "@mui/icons-material/PhotoCamera";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import ExpandLessIcon from "@mui/icons-material/ExpandLess";
import LanguageIcon from "@mui/icons-material/Language";
import LinkedInIcon from "@mui/icons-material/LinkedIn";
import PaletteIcon from "@mui/icons-material/Palette";
import InputAdornment from "@mui/material/InputAdornment";

export default function BasicInfoSection({ profile, isStudentRole, onUpdate, onAvatarUpload }) {
  const [formData, setFormData] = useState({});
  const [saving, setSaving] = useState(false);
  const [avatarUploading, setAvatarUploading] = useState(false);
  const [snackbar, setSnackbar] = useState({ open: false, message: "", severity: "success" });
  const [guidedOpen, setGuidedOpen] = useState(false);
  const avatarInputRef = useRef(null);

  useEffect(() => {
    if (profile) {
      const skillsStr = Array.isArray(profile.skills)
        ? profile.skills.map((s) => (typeof s === "string" ? s : s.name)).join(", ")
        : "";
      setFormData({
        bio: profile.bio || "",
        website: profile.website || "",
        linkedin: profile.linkedin || "",
        behance: profile.behance || "",
        languages: profile.languages || "",
        hard_skills: skillsStr,
        professional_interests: profile.professional_interests || "",
      });
    }
  }, [profile]);

  const handleChange = (field) => (e) => {
    setFormData((prev) => ({ ...prev, [field]: e.target.value }));
  };

  const parseCommaList = (str) =>
    (str || "")
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);

  const handleSave = async () => {
    setSaving(true);
    try {
      const skills = parseCommaList(formData.hard_skills).map((name) => ({
        name,
        level: "intermediate",
      }));
      const updateData = {
        bio: formData.bio || null,
        website: (formData.website || "").trim() || null,
        linkedin: (formData.linkedin || "").trim() || null,
        behance: (formData.behance || "").trim() || null,
        languages: (formData.languages || "").trim() || null,
        professional_interests: (formData.professional_interests || "").trim() || null,
        skills,
      };
      const result = await onUpdate(updateData);
      if (result.success) {
        setSnackbar({ open: true, message: "Profile saved successfully", severity: "success" });
      } else {
        setSnackbar({ open: true, message: result.error || "Save failed", severity: "error" });
      }
    } finally {
      setSaving(false);
    }
  };

  const handleAvatarClick = () => avatarInputRef.current?.click();
  const handleAvatarChange = async (e) => {
    const file = e.target.files?.[0];
    if (!file || !onAvatarUpload) return;
    setAvatarUploading(true);
    try {
      const result = await onAvatarUpload(file);
      if (result.success) {
        setSnackbar({ open: true, message: "Avatar uploaded successfully", severity: "success" });
      } else {
        setSnackbar({ open: true, message: result.error || "Upload failed", severity: "error" });
      }
    } finally {
      setAvatarUploading(false);
      e.target.value = "";
    }
  };

  return (
    <>
      <Box
        sx={{
          p: 2,
          mb: 3,
          borderRadius: 2,
          background: "linear-gradient(135deg, rgba(25, 118, 210, 0.15) 0%, rgba(156, 39, 176, 0.15) 100%)",
          border: "1px solid rgba(25, 118, 210, 0.3)",
        }}
      >
        <Stack direction="row" alignItems="center" spacing={1}>
          <PersonIcon sx={{ color: "#1976d2", fontSize: 28 }} />
          <Box>
            <Typography variant="subtitle1" fontWeight={600} color="#1a1a1a">
              Edit Your Public Profile
            </Typography>
            <Typography variant="body2" color="#666">
              All information below will be visible on your public profile to potential clients.
              Make sure to showcase your best self!
            </Typography>
          </Box>
        </Stack>
      </Box>

      {onAvatarUpload && (
        <Box sx={{ mb: 3 }}>
          <Typography variant="subtitle2" fontWeight={600} color="#1a1a1a" gutterBottom>
            Profile Photo
          </Typography>
          <Stack direction="row" alignItems="center" spacing={2}>
            <Box
              onClick={handleAvatarClick}
              sx={{
                position: "relative",
                cursor: avatarUploading ? "wait" : "pointer",
                "&:hover .avatar-overlay": { opacity: 1 },
              }}
            >
              <Avatar
                src={profile?.avatar_url}
                sx={{
                  width: 80,
                  height: 80,
                  bgcolor: "primary.main",
                  fontSize: "2rem",
                }}
              >
                {(profile?.name || "?").charAt(0)}
              </Avatar>
              <Box
                className="avatar-overlay"
                sx={{
                  position: "absolute",
                  inset: 0,
                  borderRadius: "50%",
                  bgcolor: "rgba(0,0,0,0.5)",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  opacity: 0,
                  transition: "opacity 0.2s",
                }}
              >
                <PhotoCameraIcon sx={{ color: "#fff", fontSize: 28 }} />
              </Box>
            </Box>
            <Box>
              <Button
                variant="outlined"
                size="small"
                startIcon={<PhotoCameraIcon />}
                onClick={handleAvatarClick}
                disabled={avatarUploading}
              >
                {avatarUploading ? "Uploading..." : "Upload avatar"}
              </Button>
              <Typography variant="caption" color="text.secondary" sx={{ display: "block", mt: 0.5 }}>
                JPG, PNG or WebP. Max 5MB.
              </Typography>
            </Box>
            <input
              ref={avatarInputRef}
              type="file"
              accept="image/jpeg,image/png,image/webp"
              onChange={handleAvatarChange}
              hidden
            />
          </Stack>
        </Box>
      )}

      <Grid container spacing={3}>
        <Grid item xs={12}>
          <TextField
            fullWidth
            label="Bio"
            value={formData.bio || ""}
            onChange={handleChange("bio")}
            multiline
            rows={4}
            placeholder="Tell us about yourself..."
            sx={{ "& .MuiOutlinedInput-root": { bgcolor: "#fafafa" } }}
          />
        </Grid>

        <Grid item xs={12} md={4}>
          <TextField
            fullWidth
            label="Website"
            value={formData.website || ""}
            onChange={handleChange("website")}
            placeholder="https://yoursite.com"
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <LanguageIcon sx={{ color: "#666", fontSize: 20 }} />
                </InputAdornment>
              ),
            }}
            sx={{ "& .MuiOutlinedInput-root": { bgcolor: "#fafafa" } }}
          />
        </Grid>
        <Grid item xs={12} md={4}>
          <TextField
            fullWidth
            label="LinkedIn"
            value={formData.linkedin || ""}
            onChange={handleChange("linkedin")}
            placeholder="https://linkedin.com/in/username"
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <LinkedInIcon sx={{ color: "#0a66c2", fontSize: 20 }} />
                </InputAdornment>
              ),
            }}
            sx={{ "& .MuiOutlinedInput-root": { bgcolor: "#fafafa" } }}
          />
        </Grid>
        <Grid item xs={12} md={4}>
          <TextField
            fullWidth
            label="Behance"
            value={formData.behance || ""}
            onChange={handleChange("behance")}
            placeholder="https://behance.net/username"
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <PaletteIcon sx={{ color: "#1769ff", fontSize: 20 }} />
                </InputAdornment>
              ),
            }}
            sx={{ "& .MuiOutlinedInput-root": { bgcolor: "#fafafa" } }}
          />
        </Grid>

        <Grid item xs={12}>
          <TextField
            fullWidth
            label="Languages"
            value={formData.languages || ""}
            onChange={handleChange("languages")}
            placeholder="English, Finnish, Swedish"
            helperText="Separate with commas"
            sx={{ "& .MuiOutlinedInput-root": { bgcolor: "#fafafa" } }}
          />
        </Grid>
        <Grid item xs={12}>
          <TextField
            fullWidth
            label="Hard Skills"
            value={formData.hard_skills || ""}
            onChange={handleChange("hard_skills")}
            placeholder="Adobe Lightroom, Adobe Photoshop, Event Photography"
            helperText="Separate with commas"
            sx={{ "& .MuiOutlinedInput-root": { bgcolor: "#fafafa" } }}
          />
        </Grid>
        <Grid item xs={12}>
          <TextField
            fullWidth
            label="Professional Interests"
            value={formData.professional_interests || ""}
            onChange={handleChange("professional_interests")}
            placeholder="Documentary Photography, Nordic Aesthetics, Sustainable Events"
            helperText="Separate with commas"
            sx={{ "& .MuiOutlinedInput-root": { bgcolor: "#fafafa" } }}
          />
        </Grid>

        <Grid item xs={12}>
          <Paper
            variant="outlined"
            sx={{ borderRadius: 2, overflow: "hidden", borderColor: "#e0e0e0" }}
          >
            <Box
              onClick={() => setGuidedOpen(!guidedOpen)}
              sx={{
                p: 2,
                cursor: "pointer",
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                "&:hover": { bgcolor: "#f5f5f5" },
              }}
            >
              <Box>
                <Typography fontWeight={600} color="#1a1a1a">
                  Guided Profile Questions (Optional)
                </Typography>
                <Typography variant="body2" color="#666">
                  Help us understand you better - answers are private
                </Typography>
              </Box>
              <IconButton size="small">
                {guidedOpen ? <ExpandLessIcon /> : <ExpandMoreIcon />}
              </IconButton>
            </Box>
            <Collapse in={guidedOpen}>
              <Box sx={{ p: 2, pt: 0, color: "#666" }}>
                <Typography variant="body2">
                  Coming soon - optional questions to help match you with relevant opportunities.
                </Typography>
              </Box>
            </Collapse>
          </Paper>
        </Grid>

        <Grid item xs={12}>
          <Button
            variant="contained"
            startIcon={<SaveIcon />}
            onClick={handleSave}
            disabled={saving}
            sx={{ bgcolor: "#1976d2", "&:hover": { bgcolor: "#1565c0" } }}
          >
            {saving ? "Saving..." : "Save Profile"}
          </Button>
        </Grid>
      </Grid>

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
