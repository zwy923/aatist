import React, { useState } from 'react';
import {
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    Button,
    TextField,
    MenuItem,
    Stack,
    Box,
    Typography,
    Chip
} from '@mui/material';
import communityApi from '../api/communityApi';

const CATEGORIES = [
    { label: "General", value: "general" },
    { label: "Study Tips", value: "study_tips" },
    { label: "Events", value: "events" },
    { label: "Housing", value: "housing" },
    { label: "Food & Cafes", value: "food_cafes" },
    { label: "Sports & Hobbies", value: "sports_hobbies" },
    { label: "Random", value: "random" }
];

const CreatePostDialog = ({ open, onClose, onSuccess, initialCategory = "general", isSticky = false }) => {
    const [formData, setFormData] = useState({
        title: '',
        content: '',
        category: initialCategory,
        tags: []
    });
    const [tagInput, setTagInput] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData(prev => ({ ...prev, [name]: value }));
    };

    const handleAddTag = (e) => {
        if (e.key === 'Enter' && tagInput.trim()) {
            e.preventDefault();
            if (!formData.tags.includes(tagInput.trim())) {
                setFormData(prev => ({ ...prev, tags: [...prev.tags, tagInput.trim()] }));
            }
            setTagInput('');
        }
    };

    const handleRemoveTag = (tagToRemove) => {
        setFormData(prev => ({ ...prev, tags: prev.tags.filter(t => t !== tagToRemove) }));
    };

    const handleSubmit = async () => {
        setLoading(true);
        setError(null);
        try {
            // For add note, we use "sticky" category internally if needed, 
            // or just use the selected category. The user request suggests:
            // "便签墙就是图像化的帖子", so they share the same model.
            // We'll use the selected category, but maybe flag it as sticky if the user clicked "Add Note".
            // However, the backend model uses Category for "sticky". 
            // Let's stick to the selected category unless it's a specific "Sticky" button.

            const submitData = {
                ...formData,
                category: isSticky ? "sticky" : formData.category
            };

            await communityApi.createPost(submitData);
            onSuccess();
            handleClose();
        } catch (err) {
            setError(err.message || "Failed to create post. Please check your input.");
        } finally {
            setLoading(false);
        }
    };

    const handleClose = () => {
        setFormData({ title: '', content: '', category: initialCategory, tags: [] });
        setTagInput('');
        setError(null);
        onClose();
    };

    return (
        <Dialog open={open} onClose={handleClose} fullWidth maxWidth="sm">
            <DialogTitle>
                <Typography variant="h5" fontWeight="bold">
                    {isSticky ? "Add a Sticky Note" : "Create New Post"}
                </Typography>
            </DialogTitle>
            <DialogContent dividers>
                <Stack spacing={3} sx={{ mt: 1 }}>
                    {error && <Typography color="error" variant="body2">{error}</Typography>}

                    <TextField
                        fullWidth
                        label="Title"
                        name="title"
                        value={formData.title}
                        onChange={handleChange}
                        placeholder="Give your post a clear title"
                        required
                    />

                    {!isSticky && (
                        <TextField
                            select
                            fullWidth
                            label="Category"
                            name="category"
                            value={formData.category}
                            onChange={handleChange}
                        >
                            {CATEGORIES.map((cat) => (
                                <MenuItem key={cat.value} value={cat.value}>
                                    {cat.label}
                                </MenuItem>
                            ))}
                        </TextField>
                    )}

                    <TextField
                        fullWidth
                        multiline
                        rows={isSticky ? 4 : 8}
                        label="Content"
                        name="content"
                        value={formData.content}
                        onChange={handleChange}
                        placeholder={isSticky ? "Write something quick..." : "What's on your mind? Share details, tips, or questions."}
                        required
                    />

                    <Box>
                        <TextField
                            fullWidth
                            label="Tags"
                            value={tagInput}
                            onChange={(e) => setTagInput(e.target.value)}
                            onKeyDown={handleAddTag}
                            placeholder="Type a tag and press Enter"
                            helperText="Add tags to help people find your post"
                        />
                        <Stack direction="row" spacing={1} flexWrap="wrap" sx={{ mt: 1 }}>
                            {formData.tags.map(tag => (
                                <Chip
                                    key={tag}
                                    label={`#${tag}`}
                                    onDelete={() => handleRemoveTag(tag)}
                                    size="small"
                                    color="primary"
                                    variant="outlined"
                                />
                            ))}
                        </Stack>
                    </Box>
                </Stack>
            </DialogContent>
            <DialogActions sx={{ p: 2, px: 3 }}>
                <Button onClick={handleClose} color="inherit">Cancel</Button>
                <Button
                    onClick={handleSubmit}
                    variant="contained"
                    loading={loading}
                    disabled={!formData.title || !formData.content || loading}
                    sx={{
                        borderRadius: 5,
                        px: 4,
                        background: "linear-gradient(45deg, #0072ff, #00c6ff)"
                    }}
                >
                    {isSticky ? "Pin it" : "Post"}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default CreatePostDialog;
