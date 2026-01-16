import React from 'react';
import {
    Box,
    Paper,
    Typography,
    Avatar,
    Stack,
    Chip,
    IconButton,
    Divider,
    Button
} from '@mui/material';
import {
    FavoriteBorder as LikeIcon,
    Favorite as LikeIconFilled,
    ChatBubbleOutline as CommentIcon,
    ShareOutlined as ShareIcon,
    MoreHoriz as MoreIcon,
    TrendingUp as TrendingIcon
} from '@mui/icons-material';

const CATEGORY_MAP = {
    'food_cafes': 'Food & Cafes',
    'study_tips': 'Study Tips',
    'events': 'Events',
    'housing': 'Housing',
    'sports_hobbies': 'Sports & Hobbies',
    'random': 'Random',
    'general': 'General',
    'projects': 'Projects'
};

const PostItem = ({ post, onClick }) => {
    const isTrending = post.like_count > 10 || post.comment_count > 5;

    return (
        <Paper
            onClick={() => onClick(post)}
            sx={{
                p: 3,
                mb: 2,
                borderRadius: 4,
                background: "rgba(255, 255, 255, 0.02)",
                border: "1px solid rgba(255, 255, 255, 0.05)",
                transition: 'all 0.2s',
                cursor: 'pointer',
                '&:hover': {
                    background: "rgba(255, 255, 255, 0.04)",
                    borderColor: "rgba(93, 224, 255, 0.2)"
                }
            }}
        >
            <Stack direction="row" spacing={2} alignItems="flex-start">
                <Avatar
                    src={post.author_avatar}
                    sx={{ width: 48, height: 48, border: '2px solid rgba(93, 224, 255, 0.3)' }}
                >
                    {post.author_name?.[0] || 'U'}
                </Avatar>

                <Box sx={{ flexGrow: 1 }}>
                    <Stack direction="row" justifyContent="space-between" alignItems="center">
                        <Stack direction="row" spacing={1} alignItems="center">
                            <Typography variant="subtitle1" fontWeight="700">
                                {post.author_name || 'Anonymous'}
                            </Typography>
                            <Typography variant="caption" color="text.secondary">
                                {post.author_faculty || 'Aalto Student'} • {new Date(post.created_at).toLocaleDateString()}
                            </Typography>
                        </Stack>
                        <IconButton size="small" onClick={(e) => e.stopPropagation()}>
                            <MoreIcon fontSize="small" />
                        </IconButton>
                    </Stack>

                    <Typography variant="h5" fontWeight="700" sx={{ mt: 1, mb: 1, color: '#f5fbff' }}>
                        {post.title}
                    </Typography>

                    <Typography variant="body1" color="text.secondary" sx={{ mb: 2, lineHeight: 1.6 }}>
                        {post.content}
                    </Typography>

                    <Stack direction="row" spacing={1} sx={{ mb: 2 }}>
                        <Chip
                            label={CATEGORY_MAP[post.category] || 'General'}
                            size="small"
                            variant="outlined"
                            sx={{ borderRadius: 1, fontSize: '0.75rem' }}
                        />
                        {isTrending && (
                            <Chip
                                icon={<TrendingIcon sx={{ fontSize: '1rem !important' }} />}
                                label="Trending"
                                size="small"
                                color="secondary"
                                sx={{ borderRadius: 1, fontSize: '0.75rem', fontWeight: 700 }}
                            />
                        )}
                        {post.tags?.map(tag => (
                            <Typography key={tag} variant="caption" color="primary.main" sx={{ cursor: 'pointer', '&:hover': { textDecoration: 'underline' } }}>
                                #{tag}
                            </Typography>
                        ))}
                    </Stack>

                    <Divider sx={{ my: 2, borderColor: 'rgba(255,255,255,0.05)' }} />

                    <Stack direction="row" spacing={4}>
                        <Button
                            startIcon={post.has_liked ? <LikeIconFilled sx={{ color: '#e91e63' }} /> : <LikeIcon sx={{ fontSize: 'small' }} />}
                            size="small"
                            color="inherit"
                            sx={{
                                opacity: post.has_liked ? 1 : 0.7,
                                color: post.has_liked ? 'primary.main' : 'inherit',
                                '&:hover': { opacity: 1, color: 'primary.main' }
                            }}
                            onClick={(e) => {
                                e.stopPropagation();
                                // Optional: simple like from here could be added later
                            }}
                        >
                            {post.like_count || 0}
                        </Button>
                        <Button
                            startIcon={<CommentIcon fontSize="small" />}
                            size="small"
                            color="inherit"
                            sx={{ opacity: 0.7, '&:hover': { opacity: 1, color: 'primary.main' } }}
                            onClick={(e) => e.stopPropagation()}
                        >
                            {post.comment_count || 0}
                        </Button>
                        <Button
                            startIcon={<ShareIcon fontSize="small" />}
                            size="small"
                            color="inherit"
                            sx={{ opacity: 0.7, '&:hover': { opacity: 1, color: 'primary.main' } }}
                            onClick={(e) => e.stopPropagation()}
                        >
                            Share
                        </Button>
                    </Stack>
                </Box>
            </Stack>
        </Paper>
    );
};

const PostStream = ({ posts = [], onRefresh, onPostClick }) => {
    return (
        <Box sx={{ maxWidth: 800, mx: 'auto' }}>
            {posts && Array.isArray(posts) && posts.map((post) => (
                <PostItem key={post.id} post={post} onClick={onPostClick} />
            ))}
        </Box>
    );
};

export default PostStream;
