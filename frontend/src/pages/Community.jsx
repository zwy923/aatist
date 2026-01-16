import React, { useState, useEffect, useCallback } from 'react';
import {
    Box,
    Typography,
    Tabs,
    Tab,
    Button,
    Stack,
    TextField,
    InputAdornment,
    IconButton,
    ToggleButton,
    ToggleButtonGroup,
    CircularProgress,
    Alert
} from '@mui/material';
import {
    Search as SearchIcon,
    Add as AddIcon,
    GridView as BoardIcon,
    ViewList as StreamIcon,
    FilterList as FilterIcon
} from '@mui/icons-material';
import PageLayout from '../shared/components/PageLayout';
import PostStream from '../features/community/components/PostStream';
import StickyNoteBoard from '../features/community/components/StickyNoteBoard';
import communityApi from '../features/community/api/communityApi';
import PostDetailDialog from '../features/community/components/PostDetailDialog';
import CreatePostDialog from '../features/community/components/CreatePostDialog';

const CATEGORIES = [
    { label: "All", value: "" },
    { label: "General", value: "general" },
    { label: "Study Tips", value: "study_tips" },
    { label: "Events", value: "events" },
    { label: "Housing", value: "housing" },
    { label: "Food & Cafes", value: "food_cafes" },
    { label: "Sports & Hobbies", value: "sports_hobbies" },
    { label: "Random", value: "random" }
];

const Community = () => {
    const [view, setView] = useState('board'); // 'board' or 'stream'
    const [category, setCategory] = useState("");
    const [search, setSearch] = useState("");
    const [posts, setPosts] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [createDialogOpen, setCreateDialogOpen] = useState(false);
    const [createDialogType, setCreateDialogType] = useState('post'); // 'post' or 'note'
    const [selectedPost, setSelectedPost] = useState(null);
    const [detailDialogOpen, setDetailDialogOpen] = useState(false);

    const fetchPosts = useCallback(async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await communityApi.getPosts({
                category: category || undefined,
                search: search || undefined,
                limit: 50
            });
            const newPosts = data || [];
            setPosts(newPosts);

            // Update selectedPost if it's currently open
            if (selectedPost) {
                const refreshedPost = newPosts.find(p => p.id === selectedPost.id);
                if (refreshedPost) {
                    setSelectedPost(refreshedPost);
                }
            }
        } catch (err) {
            console.error("Failed to fetch community posts:", err);
            setError("Failed to load community content. Please try again.");
        } finally {
            setLoading(false);
        }
    }, [category, search]);

    useEffect(() => {
        fetchPosts();
    }, [fetchPosts]);

    const handleViewChange = (event, nextView) => {
        if (nextView !== null) {
            setView(nextView);
        }
    };

    const handleCategoryChange = (event, newValue) => {
        setCategory(newValue);
    };

    const handleOpenCreateDialog = (type) => {
        setCreateDialogType(type);
        setCreateDialogOpen(true);
    };

    const handleCreateSuccess = () => {
        fetchPosts();
    };

    const handlePostClick = (post) => {
        setSelectedPost(post);
        setDetailDialogOpen(true);
    };

    return (
        <PageLayout>
            <Box sx={{ mb: 6 }}>
                <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={2} sx={{ mb: 4 }}>
                    <Box>
                        <Typography variant="h3" fontWeight="800" gutterBottom sx={{
                            background: "linear-gradient(45deg, #5de0ff, #0072ff)",
                            WebkitBackgroundClip: "text",
                            WebkitTextFillColor: "transparent"
                        }}>
                            Community
                        </Typography>
                        <Typography variant="body1" color="text.secondary">
                            Connect with fellow students, share tips, and find your place.
                        </Typography>
                    </Box>
                    <Stack direction="row" spacing={2}>
                        <Button
                            variant="contained"
                            startIcon={<AddIcon />}
                            onClick={() => handleOpenCreateDialog('post')}
                            sx={{
                                borderRadius: 10,
                                px: 3,
                                py: 1.5,
                                background: "linear-gradient(45deg, #0072ff, #00c6ff)",
                                boxShadow: "0 4px 15px rgba(0,114,255,0.4)"
                            }}
                        >
                            Create Post
                        </Button>
                        <Button
                            variant="outlined"
                            startIcon={<AddIcon />}
                            onClick={() => handleOpenCreateDialog('note')}
                            sx={{ borderRadius: 10, px: 3 }}
                        >
                            Add Note
                        </Button>
                    </Stack>
                </Stack>

                <Box sx={{
                    display: 'flex',
                    flexDirection: { xs: 'column', md: 'row' },
                    gap: 3,
                    alignItems: 'center',
                    mb: 4,
                    p: 2,
                    borderRadius: 4,
                    background: "rgba(255, 255, 255, 0.03)",
                    border: "1px solid rgba(255, 255, 255, 0.05)"
                }}>
                    <Tabs
                        value={category}
                        onChange={handleCategoryChange}
                        variant="scrollable"
                        scrollButtons="auto"
                        sx={{
                            flexGrow: 1,
                            '& .MuiTabs-indicator': { height: 3, borderRadius: '3px 3px 0 0' }
                        }}
                    >
                        {CATEGORIES.map((cat) => (
                            <Tab
                                key={cat.value}
                                label={cat.label}
                                value={cat.value}
                                sx={{ textTransform: 'none', fontWeight: 600, minWidth: 'auto', px: 3 }}
                            />
                        ))}
                    </Tabs>

                    <Stack direction="row" spacing={2} sx={{ width: { xs: '100%', md: 'auto' } }}>
                        <TextField
                            size="small"
                            placeholder="Search content or author..."
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                            sx={{
                                width: { xs: '100%', md: 300 },
                                '& .MuiOutlinedInput-root': { borderRadius: 10 }
                            }}
                            InputProps={{
                                startAdornment: (
                                    <InputAdornment position="start">
                                        <SearchIcon fontSize="small" />
                                    </InputAdornment>
                                ),
                            }}
                        />
                        <ToggleButtonGroup
                            value={view}
                            exclusive
                            onChange={handleViewChange}
                            size="small"
                            sx={{ border: 'none', '& .MuiToggleButton-root': { border: 'none', borderRadius: '10px !important', px: 2 } }}
                        >
                            <ToggleButton value="board" aria-label="board view">
                                <BoardIcon />
                            </ToggleButton>
                            <ToggleButton value="stream" aria-label="stream view">
                                <StreamIcon />
                            </ToggleButton>
                        </ToggleButtonGroup>
                    </Stack>
                </Box>

                {error && (
                    <Alert severity="error" sx={{ mb: 4, borderRadius: 2 }}>{error}</Alert>
                )}

                {loading ? (
                    <Box sx={{ display: 'flex', justifyContent: 'center', py: 10 }}>
                        <CircularProgress />
                    </Box>
                ) : (
                    <Box>
                        {view === 'board' ? (
                            <StickyNoteBoard posts={posts} onRefresh={fetchPosts} onPostClick={handlePostClick} />
                        ) : (
                            <PostStream posts={posts} onRefresh={fetchPosts} onPostClick={handlePostClick} />
                        )}

                        {!loading && posts.length === 0 && (
                            <Box sx={{ textAlign: 'center', py: 10 }}>
                                <Typography variant="h6" color="text.secondary">
                                    No posts found in this category. Be the first to start the conversation!
                                </Typography>
                            </Box>
                        )}
                    </Box>
                )}

                <CreatePostDialog
                    open={createDialogOpen}
                    onClose={() => setCreateDialogOpen(false)}
                    onSuccess={handleCreateSuccess}
                    isSticky={createDialogType === 'note'}
                    initialCategory={category || "general"}
                />

                <PostDetailDialog
                    open={detailDialogOpen}
                    onClose={() => setDetailDialogOpen(false)}
                    post={selectedPost}
                    onRefresh={fetchPosts}
                />
            </Box>
        </PageLayout>
    );
};

export default Community;
