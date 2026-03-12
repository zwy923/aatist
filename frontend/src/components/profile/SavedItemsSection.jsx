import React, { useState } from "react";
import {
    Box,
    Button,
    Card,
    CardContent,
    Chip,
    FormControl,
    Grid,
    IconButton,
    InputLabel,
    MenuItem,
    Paper,
    Select,
    Stack,
    Typography,
} from "@mui/material";
import BookmarkIcon from "@mui/icons-material/Bookmark";
import DeleteIcon from "@mui/icons-material/Delete";
import OpenInNewIcon from "@mui/icons-material/OpenInNew";
import WorkIcon from "@mui/icons-material/Work";
import EventIcon from "@mui/icons-material/Event";
import ArticleIcon from "@mui/icons-material/Article";
import PersonIcon from "@mui/icons-material/Person";

// Item type configurations
const ITEM_TYPES = {
    opportunity: { label: "Opportunity", icon: WorkIcon, color: "#5de0ff" },
    event: { label: "Event", icon: EventIcon, color: "#ffb877" },
    post: { label: "Post", icon: ArticleIcon, color: "#a78bfa" },
    user: { label: "User", icon: PersonIcon, color: "#4ade80" },
};

export default function SavedItemsSection({ items, onRemove, onNavigate }) {
    const [filter, setFilter] = useState("all");

    const filteredItems = filter === "all"
        ? items
        : items.filter((item) => item.type === filter);

    const handleRemove = async (id) => {
        if (!window.confirm("Are you sure you want to remove this item from saved?")) {
            return;
        }
        await onRemove(id);
    };

    const getItemConfig = (type) => {
        return ITEM_TYPES[type] || { label: type, icon: BookmarkIcon, color: "#888" };
    };

    return (
        <Paper
            sx={{
                background: "#ffffff",
                border: "1px solid #e5e7eb",
                borderRadius: 3,
                p: 4,
            }}
        >
            <Stack direction="row" justifyContent="space-between" alignItems="center" mb={3}>
                <Stack direction="row" alignItems="center" spacing={1}>
                    <BookmarkIcon sx={{ color: "primary.main" }} />
                    <Typography variant="h5" fontWeight={600}>
                        Saved Items
                    </Typography>
                    <Chip label={items.length} size="small" color="primary" />
                </Stack>

                <FormControl size="small" sx={{ minWidth: 150 }}>
                    <InputLabel>Filter</InputLabel>
                    <Select
                        value={filter}
                        label="Filter"
                        onChange={(e) => setFilter(e.target.value)}
                    >
                        <MenuItem value="all">All Types</MenuItem>
                        {Object.entries(ITEM_TYPES).map(([key, config]) => (
                            <MenuItem key={key} value={key}>
                                {config.label}
                            </MenuItem>
                        ))}
                    </Select>
                </FormControl>
            </Stack>

            {filteredItems.length === 0 ? (
                <Box
                    sx={{
                        textAlign: "center",
                        py: 6,
                        color: "text.secondary",
                    }}
                >
                    <BookmarkIcon sx={{ fontSize: 48, opacity: 0.5, mb: 2 }} />
                    <Typography variant="h6">No saved items</Typography>
                    <Typography variant="body2">
                        {filter === "all"
                            ? "Items you save will appear here"
                            : `No saved ${ITEM_TYPES[filter]?.label.toLowerCase() || filter}s found`}
                    </Typography>
                </Box>
            ) : (
                <Grid container spacing={2}>
                    {filteredItems.map((item) => {
                        const config = getItemConfig(item.type);
                        const IconComponent = config.icon;

                        return (
                            <Grid item xs={12} sm={6} md={4} key={item.id}>
                                <Card
                                    sx={{
                                        backgroundColor: "#ffffff",
                                        border: "1px solid #e5e7eb",
                                        borderRadius: 2,
                                        height: "100%",
                                        transition: "all 0.2s ease",
                                        "&:hover": {
                                            borderColor: config.color,
                                            transform: "translateY(-2px)",
                                        },
                                    }}
                                >
                                    <CardContent>
                                        <Stack spacing={2}>
                                            {/* Header */}
                                            <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
                                                <Stack direction="row" alignItems="center" spacing={1}>
                                                    <Box
                                                        sx={{
                                                            width: 32,
                                                            height: 32,
                                                            borderRadius: 1,
                                                            backgroundColor: `${config.color}20`,
                                                            display: "flex",
                                                            alignItems: "center",
                                                            justifyContent: "center",
                                                        }}
                                                    >
                                                        <IconComponent sx={{ color: config.color, fontSize: 18 }} />
                                                    </Box>
                                                    <Chip
                                                        label={config.label}
                                                        size="small"
                                                        sx={{
                                                            backgroundColor: `${config.color}15`,
                                                            color: config.color,
                                                            fontSize: "0.7rem",
                                                        }}
                                                    />
                                                </Stack>
                                                <Stack direction="row" spacing={0.5}>
                                                    <IconButton
                                                        size="small"
                                                        onClick={() => onNavigate(item.type, item.target_id || item.targetId)}
                                                        sx={{ color: "text.secondary", "&:hover": { color: "primary.main" } }}
                                                    >
                                                        <OpenInNewIcon fontSize="small" />
                                                    </IconButton>
                                                    <IconButton
                                                        size="small"
                                                        onClick={() => handleRemove(item.id)}
                                                        sx={{ color: "text.secondary", "&:hover": { color: "error.main" } }}
                                                    >
                                                        <DeleteIcon fontSize="small" />
                                                    </IconButton>
                                                </Stack>
                                            </Stack>

                                            {/* Content */}
                                            <Box>
                                                <Typography variant="subtitle1" fontWeight={600} noWrap>
                                                    {item.title || item.name || `Saved ${config.label}`}
                                                </Typography>
                                                {item.description && (
                                                    <Typography
                                                        variant="body2"
                                                        color="text.secondary"
                                                        sx={{
                                                            display: "-webkit-box",
                                                            WebkitLineClamp: 2,
                                                            WebkitBoxOrient: "vertical",
                                                            overflow: "hidden",
                                                        }}
                                                    >
                                                        {item.description}
                                                    </Typography>
                                                )}
                                            </Box>

                                            {/* Footer */}
                                            {item.saved_at && (
                                                <Typography variant="caption" color="text.secondary">
                                                    Saved {new Date(item.saved_at).toLocaleDateString()}
                                                </Typography>
                                            )}
                                        </Stack>
                                    </CardContent>
                                </Card>
                            </Grid>
                        );
                    })}
                </Grid>
            )}
        </Paper>
    );
}
