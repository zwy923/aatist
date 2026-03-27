import React, { useState, useEffect } from "react";
import { Box, Grid, IconButton, Stack, Typography, Alert, Snackbar } from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import EditIcon from "@mui/icons-material/Edit";
import DeleteIcon from "@mui/icons-material/Delete";
import ImageIcon from "@mui/icons-material/Image";
import AddPortfolioProjectDialog from "./AddPortfolioProjectDialog";

const PORTFOLIO_TIPS = [
  "Use high-quality images (at least 1200px wide)",
  "Write clear descriptions explaining the context",
  "Tag your work with relevant skills (max 10 tags)",
  "Link each work to a service category",
];

export default function PortfolioSection({
  items,
  onPortfolioRefresh,
  onDelete,
  triggerEditForId = null,
  onTriggerEditConsumed,
  hideIntro = false,
}) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingItem, setEditingItem] = useState(null);
  const [snackbar, setSnackbar] = useState({ open: false, message: "", severity: "success" });

  const handleOpenCreate = () => {
    setEditingItem(null);
    setDialogOpen(true);
  };

  const handleOpenEdit = (item) => {
    setEditingItem(item);
    setDialogOpen(true);
  };

  useEffect(() => {
    if (triggerEditForId == null || !items?.length) return;
    const item = items.find((i) => String(i.id) === String(triggerEditForId));
    if (!item) return;
    setEditingItem(item);
    setDialogOpen(true);
    onTriggerEditConsumed?.();
  }, [triggerEditForId, items, onTriggerEditConsumed]);

  const handleClose = () => {
    setDialogOpen(false);
    setEditingItem(null);
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
      {!hideIntro && (
        <>
          <Typography variant="h6" fontWeight={600} color="#1a1a1a" gutterBottom>
            Portfolio Works
          </Typography>
          <Typography variant="body2" color="#666" sx={{ mb: 3 }}>
            Upload your best work. Each piece should have a title, description, and tags.
          </Typography>
        </>
      )}

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
                  {(item.short_caption || item.description) && (
                    <Typography variant="caption" color="text.secondary" noWrap sx={{ display: "block", mt: 0.25 }}>
                      {item.short_caption || item.description}
                    </Typography>
                  )}
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

      <AddPortfolioProjectDialog
        open={dialogOpen}
        onClose={handleClose}
        editingItem={editingItem}
        onSaved={() => onPortfolioRefresh?.()}
      />

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
