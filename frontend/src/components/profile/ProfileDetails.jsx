import React, { useState, useEffect } from "react";
import {
    Box,
    Button,
    Chip,
    Grid,
    IconButton,
    Paper,
    Stack,
    TextField,
    Typography,
    Autocomplete,
    Alert,
    Snackbar,
} from "@mui/material";
import EditIcon from "@mui/icons-material/Edit";
import SaveIcon from "@mui/icons-material/Save";
import CancelIcon from "@mui/icons-material/Cancel";
import AddIcon from "@mui/icons-material/Add";

// Common skill levels
const SKILL_LEVELS = ["Beginner", "Intermediate", "Advanced", "Expert"];

export default function ProfileDetails({ profile, isStudentRole, onUpdate }) {
    const [editing, setEditing] = useState(false);
    const [saving, setSaving] = useState(false);
    const [formData, setFormData] = useState({});
    const [newSkill, setNewSkill] = useState({ name: "", level: "Intermediate" });
    const [snackbar, setSnackbar] = useState({ open: false, message: "", severity: "success" });

    // Initialize form data when profile changes
    useEffect(() => {
        if (profile) {
            setFormData({
                name: profile.name || "",
                bio: profile.bio || "",
                // Student fields
                student_id: profile.student_id || "",
                school: profile.school || "",
                faculty: profile.faculty || "",
                major: profile.major || "",
                skills: profile.skills || [],
                // Organization fields
                organization_name: profile.organization_name || "",
                organization_bio: profile.organization_bio || "",
                contact_title: profile.contact_title || "",
                org_size: profile.org_size || "",
            });
        }
    }, [profile]);

    const handleChange = (field) => (event) => {
        setFormData((prev) => ({ ...prev, [field]: event.target.value }));
    };

    const handleAddSkill = () => {
        if (!newSkill.name.trim()) return;

        // Check if skill already exists
        const exists = formData.skills?.some(
            (s) => s.name.toLowerCase() === newSkill.name.toLowerCase()
        );
        if (exists) {
            setSnackbar({ open: true, message: "Skill already added", severity: "warning" });
            return;
        }

        setFormData((prev) => ({
            ...prev,
            skills: [...(prev.skills || []), { name: newSkill.name.trim(), level: newSkill.level }],
        }));
        setNewSkill({ name: "", level: "Intermediate" });
    };

    const handleRemoveSkill = (skillName) => {
        setFormData((prev) => ({
            ...prev,
            skills: prev.skills.filter((s) => s.name !== skillName),
        }));
    };

    const handleSave = async () => {
        setSaving(true);
        try {
            // Prepare update data based on role
            const updateData = {
                name: formData.name,
                bio: formData.bio,
            };

            if (isStudentRole) {
                updateData.student_id = formData.student_id;
                updateData.school = formData.school;
                updateData.faculty = formData.faculty;
                updateData.major = formData.major;
                updateData.skills = formData.skills;
            } else {
                updateData.organization_name = formData.organization_name;
                updateData.organization_bio = formData.organization_bio;
                updateData.contact_title = formData.contact_title;
                updateData.org_size = formData.org_size ? parseInt(formData.org_size, 10) : null;
            }

            const result = await onUpdate(updateData);

            if (result.success) {
                setSnackbar({ open: true, message: "Profile updated successfully", severity: "success" });
                setEditing(false);
            } else {
                setSnackbar({ open: true, message: result.error || "Update failed", severity: "error" });
            }
        } finally {
            setSaving(false);
        }
    };

    const handleCancel = () => {
        // Reset form data to original profile
        if (profile) {
            setFormData({
                name: profile.name || "",
                bio: profile.bio || "",
                student_id: profile.student_id || "",
                school: profile.school || "",
                faculty: profile.faculty || "",
                major: profile.major || "",
                skills: profile.skills || [],
                organization_name: profile.organization_name || "",
                organization_bio: profile.organization_bio || "",
                contact_title: profile.contact_title || "",
                org_size: profile.org_size || "",
            });
        }
        setEditing(false);
    };

    return (
        <>
            <Paper
                sx={{
                    background: "#ffffff",
                    border: "1px solid #e5e7eb",
                    borderRadius: 3,
                    p: 4,
                }}
            >
                <Stack direction="row" justifyContent="space-between" alignItems="center" mb={3}>
                    <Typography variant="h5" fontWeight={600}>
                        Profile Details
                    </Typography>
                    {!editing ? (
                        <Button
                            startIcon={<EditIcon />}
                            onClick={() => setEditing(true)}
                            variant="outlined"
                            sx={{ borderColor: "primary.main" }}
                        >
                            Edit
                        </Button>
                    ) : (
                        <Stack direction="row" spacing={1}>
                            <Button
                                startIcon={<CancelIcon />}
                                onClick={handleCancel}
                                variant="outlined"
                                color="inherit"
                            >
                                Cancel
                            </Button>
                            <Button
                                startIcon={<SaveIcon />}
                                onClick={handleSave}
                                variant="contained"
                                disabled={saving}
                            >
                                {saving ? "Saving..." : "Save"}
                            </Button>
                        </Stack>
                    )}
                </Stack>

                <Grid container spacing={3}>
                    {/* Common fields */}
                    <Grid item xs={12} md={6}>
                        <TextField
                            fullWidth
                            label="Name"
                            value={formData.name || ""}
                            onChange={handleChange("name")}
                            disabled={!editing}
                            variant="outlined"
                        />
                    </Grid>

                    <Grid item xs={12}>
                        <TextField
                            fullWidth
                            label="Bio"
                            value={formData.bio || ""}
                            onChange={handleChange("bio")}
                            disabled={!editing}
                            multiline
                            rows={3}
                            variant="outlined"
                            placeholder="Tell us about yourself..."
                        />
                    </Grid>

                    {/* Student/Alumni specific fields */}
                    {isStudentRole && (
                        <>
                            <Grid item xs={12}>
                                <Typography variant="subtitle1" fontWeight={600} gutterBottom sx={{ mt: 2 }}>
                                    Academic Information
                                </Typography>
                            </Grid>

                            <Grid item xs={12} md={6}>
                                <TextField
                                    fullWidth
                                    label="Student ID"
                                    value={formData.student_id || ""}
                                    onChange={handleChange("student_id")}
                                    disabled={!editing}
                                    variant="outlined"
                                />
                            </Grid>

                            <Grid item xs={12} md={6}>
                                <TextField
                                    fullWidth
                                    label="School"
                                    value={formData.school || ""}
                                    onChange={handleChange("school")}
                                    disabled={!editing}
                                    variant="outlined"
                                />
                            </Grid>

                            <Grid item xs={12} md={6}>
                                <TextField
                                    fullWidth
                                    label="Faculty"
                                    value={formData.faculty || ""}
                                    onChange={handleChange("faculty")}
                                    disabled={!editing}
                                    variant="outlined"
                                />
                            </Grid>

                            <Grid item xs={12} md={6}>
                                <TextField
                                    fullWidth
                                    label="Major"
                                    value={formData.major || ""}
                                    onChange={handleChange("major")}
                                    disabled={!editing}
                                    variant="outlined"
                                />
                            </Grid>

                            {/* Skills section */}
                            <Grid item xs={12}>
                                <Typography variant="subtitle1" fontWeight={600} gutterBottom sx={{ mt: 2 }}>
                                    Skills
                                </Typography>

                                <Stack direction="row" flexWrap="wrap" gap={1} mb={2}>
                                    {formData.skills?.map((skill, index) => (
                                        <Chip
                                            key={index}
                                            label={`${skill.name} · ${skill.level}`}
                                            onDelete={editing ? () => handleRemoveSkill(skill.name) : undefined}
                                            sx={{
                                                backgroundColor: "rgba(93, 224, 255, 0.15)",
                                                borderColor: "primary.main",
                                                "& .MuiChip-deleteIcon": {
                                                    color: "text.secondary",
                                                    "&:hover": { color: "error.main" },
                                                },
                                            }}
                                            variant="outlined"
                                        />
                                    ))}
                                    {formData.skills?.length === 0 && !editing && (
                                        <Typography variant="body2" color="text.secondary">
                                            No skills added yet
                                        </Typography>
                                    )}
                                </Stack>

                                {editing && (
                                    <Stack direction="row" spacing={2} alignItems="flex-end">
                                        <TextField
                                            label="Skill Name"
                                            value={newSkill.name}
                                            onChange={(e) => setNewSkill((prev) => ({ ...prev, name: e.target.value }))}
                                            size="small"
                                            sx={{ flex: 1 }}
                                        />
                                        <Autocomplete
                                            value={newSkill.level}
                                            onChange={(e, value) => setNewSkill((prev) => ({ ...prev, level: value || "Intermediate" }))}
                                            options={SKILL_LEVELS}
                                            renderInput={(params) => <TextField {...params} label="Level" size="small" />}
                                            sx={{ width: 150 }}
                                            disableClearable
                                        />
                                        <IconButton
                                            onClick={handleAddSkill}
                                            color="primary"
                                            sx={{
                                                backgroundColor: "rgba(93, 224, 255, 0.15)",
                                                "&:hover": { backgroundColor: "rgba(93, 224, 255, 0.25)" },
                                            }}
                                        >
                                            <AddIcon />
                                        </IconButton>
                                    </Stack>
                                )}
                            </Grid>
                        </>
                    )}

                    {/* Organization specific fields */}
                    {!isStudentRole && (
                        <>
                            <Grid item xs={12}>
                                <Typography variant="subtitle1" fontWeight={600} gutterBottom sx={{ mt: 2 }}>
                                    Organization Information
                                </Typography>
                            </Grid>

                            <Grid item xs={12} md={6}>
                                <TextField
                                    fullWidth
                                    label="Organization Name"
                                    value={formData.organization_name || ""}
                                    onChange={handleChange("organization_name")}
                                    disabled={!editing}
                                    variant="outlined"
                                />
                            </Grid>

                            <Grid item xs={12} md={6}>
                                <TextField
                                    fullWidth
                                    label="Contact Title"
                                    value={formData.contact_title || ""}
                                    onChange={handleChange("contact_title")}
                                    disabled={!editing}
                                    variant="outlined"
                                    placeholder="e.g., HR Manager"
                                />
                            </Grid>

                            <Grid item xs={12} md={6}>
                                <TextField
                                    fullWidth
                                    label="Organization Size"
                                    type="number"
                                    value={formData.org_size || ""}
                                    onChange={handleChange("org_size")}
                                    disabled={!editing}
                                    variant="outlined"
                                    placeholder="Number of employees"
                                />
                            </Grid>

                            <Grid item xs={12}>
                                <TextField
                                    fullWidth
                                    label="Organization Bio"
                                    value={formData.organization_bio || ""}
                                    onChange={handleChange("organization_bio")}
                                    disabled={!editing}
                                    multiline
                                    rows={3}
                                    variant="outlined"
                                    placeholder="Tell us about your organization..."
                                />
                            </Grid>
                        </>
                    )}
                </Grid>
            </Paper>

            <Snackbar
                open={snackbar.open}
                autoHideDuration={4000}
                onClose={() => setSnackbar((prev) => ({ ...prev, open: false }))}
                anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
            >
                <Alert
                    severity={snackbar.severity}
                    onClose={() => setSnackbar((prev) => ({ ...prev, open: false }))}
                >
                    {snackbar.message}
                </Alert>
            </Snackbar>
        </>
    );
}
