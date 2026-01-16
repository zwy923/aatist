import React, { useState } from "react";
import {
    Box,
    Button,
    Card,
    CardContent,
    Chip,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    Grid,
    IconButton,
    Paper,
    Stack,
    TextField,
    Typography,
    Alert,
    Snackbar,
} from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import EditIcon from "@mui/icons-material/Edit";
import DeleteIcon from "@mui/icons-material/Delete";
import WorkIcon from "@mui/icons-material/Work";
import CloseIcon from "@mui/icons-material/Close";

export default function PortfolioSection({ items, onCreate, onUpdate, onDelete }) {
    const [dialogOpen, setDialogOpen] = useState(false);
    const [editingItem, setEditingItem] = useState(null);
    const [formData, setFormData] = useState({
        title: "",
        description: "",
        client_name: "",
        tags: [],
        year: new Date().getFullYear(),
    });
    const [newTag, setNewTag] = useState("");
    const [saving, setSaving] = useState(false);
    const [snackbar, setSnackbar] = useState({ open: false, message: "", severity: "success" });

    const handleOpenCreate = () => {
        setEditingItem(null);
        setFormData({
            title: "",
            description: "",
            client_name: "",
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
            client_name: item.client_name || "",
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

    const handleSave = async () => {
        if (!formData.title.trim()) {
            setSnackbar({ open: true, message: "Title is required", severity: "error" });
            return;
        }

        setSaving(true);
        try {
            const data = {
                ...formData,
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
        if (!window.confirm("Are you sure you want to delete this portfolio item?")) {
            return;
        }

        const result = await onDelete(id);
        if (result.success) {
            setSnackbar({ open: true, message: "Portfolio item deleted", severity: "success" });
        } else {
            setSnackbar({ open: true, message: result.error || "Delete failed", severity: "error" });
        }
    };

    return (
        <>
            <Paper
                sx={{
                    background: "rgba(7, 12, 30, 0.8)",
                    backdropFilter: "blur(20px)",
                    border: "1px solid rgba(93, 224, 255, 0.15)",
                    borderRadius: 3,
                    p: 4,
                }}
            >
                <Stack direction="row" justifyContent="space-between" alignItems="center" mb={3}>
                    <Stack direction="row" alignItems="center" spacing={1}>
                        <WorkIcon sx={{ color: "primary.main" }} />
                        <Typography variant="h5" fontWeight={600}>
                            Portfolio
                        </Typography>
                    </Stack>
                    <Button
                        startIcon={<AddIcon />}
                        onClick={handleOpenCreate}
                        variant="contained"
                        size="small"
                    >
                        Add Project
                    </Button>
                </Stack>

                {items.length === 0 ? (
                    <Box
                        sx={{
                            textAlign: "center",
                            py: 6,
                            color: "text.secondary",
                        }}
                    >
                        <WorkIcon sx={{ fontSize: 48, opacity: 0.5, mb: 2 }} />
                        <Typography variant="h6">No portfolio items yet</Typography>
                        <Typography variant="body2">
                            Showcase your work by adding projects to your portfolio
                        </Typography>
                    </Box>
                ) : (
                    <Grid container spacing={3}>
                        {items.map((item) => (
                            <Grid item xs={12} md={6} key={item.id}>
                                <Card
                                    sx={{
                                        backgroundColor: "rgba(7, 12, 30, 0.6)",
                                        border: "1px solid rgba(93, 224, 255, 0.1)",
                                        borderRadius: 2,
                                        height: "100%",
                                        transition: "all 0.2s ease",
                                        "&:hover": {
                                            borderColor: "primary.main",
                                            transform: "translateY(-2px)",
                                        },
                                    }}
                                >
                                    <CardContent>
                                        <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
                                            <Box sx={{ flex: 1 }}>
                                                <Typography variant="h6" fontWeight={600} gutterBottom>
                                                    {item.title}
                                                </Typography>
                                                {item.client_name && (
                                                    <Typography variant="body2" color="text.secondary" gutterBottom>
                                                        Client: {item.client_name}
                                                    </Typography>
                                                )}
                                                {item.year && (
                                                    <Typography variant="caption" color="text.secondary">
                                                        {item.year}
                                                    </Typography>
                                                )}
                                            </Box>
                                            <Stack direction="row" spacing={0.5}>
                                                <IconButton
                                                    size="small"
                                                    onClick={() => handleOpenEdit(item)}
                                                    sx={{ color: "text.secondary", "&:hover": { color: "primary.main" } }}
                                                >
                                                    <EditIcon fontSize="small" />
                                                </IconButton>
                                                <IconButton
                                                    size="small"
                                                    onClick={() => handleDelete(item.id)}
                                                    sx={{ color: "text.secondary", "&:hover": { color: "error.main" } }}
                                                >
                                                    <DeleteIcon fontSize="small" />
                                                </IconButton>
                                            </Stack>
                                        </Stack>

                                        {item.description && (
                                            <Typography
                                                variant="body2"
                                                color="text.secondary"
                                                sx={{
                                                    mt: 1,
                                                    display: "-webkit-box",
                                                    WebkitLineClamp: 3,
                                                    WebkitBoxOrient: "vertical",
                                                    overflow: "hidden",
                                                }}
                                            >
                                                {item.description}
                                            </Typography>
                                        )}

                                        {item.tags?.length > 0 && (
                                            <Stack direction="row" flexWrap="wrap" gap={0.5} mt={2}>
                                                {item.tags.map((tag, index) => (
                                                    <Chip
                                                        key={index}
                                                        label={tag}
                                                        size="small"
                                                        sx={{
                                                            backgroundColor: "rgba(93, 224, 255, 0.1)",
                                                            fontSize: "0.7rem",
                                                        }}
                                                    />
                                                ))}
                                            </Stack>
                                        )}
                                    </CardContent>
                                </Card>
                            </Grid>
                        ))}
                    </Grid>
                )}
            </Paper>

            {/* Create/Edit Dialog */}
            <Dialog
                open={dialogOpen}
                onClose={handleClose}
                maxWidth="sm"
                fullWidth
                PaperProps={{
                    sx: {
                        background: "rgba(7, 12, 30, 0.98)",
                        border: "1px solid rgba(93, 224, 255, 0.2)",
                    },
                }}
            >
                <DialogTitle>
                    <Stack direction="row" justifyContent="space-between" alignItems="center">
                        <Typography variant="h6">
                            {editingItem ? "Edit Project" : "Add Project"}
                        </Typography>
                        <IconButton onClick={handleClose} size="small">
                            <CloseIcon />
                        </IconButton>
                    </Stack>
                </DialogTitle>
                <DialogContent>
                    <Stack spacing={3} sx={{ mt: 1 }}>
                        <TextField
                            fullWidth
                            label="Project Title"
                            value={formData.title}
                            onChange={handleChange("title")}
                            required
                        />
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

                        {/* Tags */}
                        <Box>
                            <Typography variant="subtitle2" gutterBottom>
                                Tags
                            </Typography>
                            <Stack direction="row" flexWrap="wrap" gap={1} mb={1}>
                                {formData.tags.map((tag, index) => (
                                    <Chip
                                        key={index}
                                        label={tag}
                                        size="small"
                                        onDelete={() => handleRemoveTag(tag)}
                                    />
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
                                <Button variant="outlined" size="small" onClick={handleAddTag}>
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
