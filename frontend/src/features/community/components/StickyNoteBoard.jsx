import React from 'react';
import {
    Box,
    Paper,
    Typography,
    Grid,
    Avatar,
    Stack,
    Chip,
    IconButton,
    Tooltip
} from '@mui/material';
import {
    Favorite as LikeIconFilled,
    FavoriteBorder as LikeIcon,
    ChatBubbleOutline as CommentIcon,
    Share as ShareIcon,
    PushPin as PinIcon
} from '@mui/icons-material';

const CATEGORY_COLORS = {
    'food_cafes': { bg: '#FFF9C4', accent: '#FBC02D', label: 'Food & Cafes' },
    'study_tips': { bg: '#E1F5FE', accent: '#0288D1', label: 'Study Tips' },
    'events': { bg: '#F3E5F5', accent: '#7B1FA2', label: 'Events' },
    'housing': { bg: '#E8F5E9', accent: '#388E3C', label: 'Housing' },
    'sports_hobbies': { bg: '#FFE0B2', accent: '#F57C00', label: 'Sports' },
    'random': { bg: '#F5F5F5', accent: '#616161', label: 'Random' },
    'general': { bg: '#FFEBEE', accent: '#D32F2F', label: 'General' },
    'projects': { bg: '#E0F2F1', accent: '#00796B', label: 'Projects' }
};

const StickyNote = ({ post, onClick }) => {
    const config = CATEGORY_COLORS[post.category] || CATEGORY_COLORS['general'];

    return (
        <Paper
            elevation={3}
            onClick={() => onClick(post)}
            sx={{
                p: 3,
                height: '100%',
                minHeight: 250,
                backgroundColor: config.bg,
                color: 'rgba(0,0,0,0.85)',
                position: 'relative',
                display: 'flex',
                flexDirection: 'column',
                transition: 'transform 0.2s, box-shadow 0.2s',
                '&:hover': {
                    transform: 'rotate(-1deg) scale(1.02)',
                    boxShadow: '0 8px 25px rgba(0,0,0,0.15)',
                    zIndex: 2
                },
                cursor: 'pointer',
                borderRadius: 1,
                borderTop: `4px solid ${config.accent}`,
                fontFamily: '"Architects Daughter", "Inter", sans-serif'
            }}
        >
            <Box sx={{ position: 'absolute', top: -10, left: '50%', transform: 'translateX(-50%)' }}>
                <PinIcon sx={{ color: '#ff5252', fontSize: 24, filter: 'drop-shadow(0 2px 2px rgba(0,0,0,0.2))' }} />
            </Box>

            <Typography variant="h6" fontWeight="bold" sx={{ mb: 1, lineHeight: 1.2 }}>
                {post.title}
            </Typography>

            <Typography
                variant="body2"
                sx={{
                    flexGrow: 1,
                    mb: 2,
                    overflow: 'hidden',
                    display: '-webkit-box',
                    WebkitLineClamp: 5,
                    WebkitBoxOrient: 'vertical',
                    lineHeight: 1.5
                }}
            >
                {post.content}
            </Typography>

            <Box>
                <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 2 }}>
                    <Avatar
                        src={post.author_avatar}
                        sx={{ width: 24, height: 24, fontSize: '0.75rem', bgcolor: config.accent }}
                    >
                        {post.author_name?.[0] || 'U'}
                    </Avatar>
                    <Typography variant="caption" fontWeight="bold">
                        {post.author_name || 'Anonymous'}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                        • {new Date(post.created_at).toLocaleDateString()}
                    </Typography>
                </Stack>

                <Stack direction="row" justifyContent="space-between" alignItems="center">
                    <Chip
                        label={config.label}
                        size="small"
                        sx={{
                            height: 20,
                            fontSize: '0.65rem',
                            bgcolor: config.accent,
                            color: 'white',
                            fontWeight: 'bold'
                        }}
                    />
                    <Stack direction="row" spacing={0.5}>
                        <Tooltip title={post.has_liked ? "Unlike" : "Like"}>
                            <Stack direction="row" alignItems="center" spacing={0.2}>
                                {post.has_liked ? (
                                    <LikeIconFilled sx={{ fontSize: 16, color: '#e91e63' }} />
                                ) : (
                                    <LikeIcon sx={{ fontSize: 16, color: 'rgba(0,0,0,0.4)' }} />
                                )}
                                <Typography variant="caption">{post.like_count || 0}</Typography>
                            </Stack>
                        </Tooltip>
                        <Tooltip title="Comments">
                            <Stack direction="row" alignItems="center" spacing={0.2}>
                                <CommentIcon sx={{ fontSize: 16, color: '#2196f3' }} />
                                <Typography variant="caption">{post.comment_count || 0}</Typography>
                            </Stack>
                        </Tooltip>
                    </Stack>
                </Stack>
            </Box>
        </Paper>
    );
};

const StickyNoteBoard = ({ posts = [], onRefresh, onPostClick }) => {
    return (
        <Grid container spacing={4}>
            {posts && Array.isArray(posts) && posts.map((post) => (
                <Grid item xs={12} sm={6} md={4} lg={3} key={post.id}>
                    <StickyNote post={post} onClick={onPostClick} />
                </Grid>
            ))}
        </Grid>
    );
};

export default StickyNoteBoard;
