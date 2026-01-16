import React, { useState, useEffect, useCallback } from 'react';
import {
    Dialog,
    DialogContent,
    Box,
    Typography,
    Stack,
    Avatar,
    IconButton,
    Divider,
    Button,
    TextField,
    CircularProgress,
    Chip,
    Tooltip,
    Paper
} from '@mui/material';
import {
    Close as CloseIcon,
    Favorite as LikeIcon,
    FavoriteBorder as UnlikeIcon,
    ChatBubbleOutline as CommentIcon,
    Share as ShareIcon,
    Send as SendIcon,
    PushPin as PinIcon,
    DeleteOutline as DeleteIcon
} from '@mui/icons-material';
import communityApi from '../api/communityApi';
import { useAuth } from '../../../features/auth/hooks/useAuth';

const PostDetailDialog = ({ open, onClose, post, onRefresh }) => {
    const { user } = useAuth();
    const [comments, setComments] = useState([]);
    const [commentInput, setCommentInput] = useState('');
    const [loadingComments, setLoadingComments] = useState(false);
    const [submittingComment, setSubmittingComment] = useState(false);
    const [isLiked, setIsLiked] = useState(!!post?.has_liked);
    const [likesCount, setLikesCount] = useState(post?.like_count || 0);

    const isAuthor = user?.id === post?.user_id;
    const isAdmin = user?.role === 'admin' || user?.role === 'org_team';
    const canDelete = isAuthor || isAdmin;

    const fetchComments = useCallback(async () => {
        if (!post?.id) return;
        setLoadingComments(true);
        try {
            const data = await communityApi.getComments(post.id);
            setComments(data || []);
        } catch (err) {
            console.error("Failed to fetch comments:", err);
        } finally {
            setLoadingComments(false);
        }
    }, [post?.id]);

    useEffect(() => {
        if (open && post?.id) {
            fetchComments();
            setLikesCount(post.like_count || 0);
            setIsLiked(!!post.has_liked);
        }
    }, [open, post?.id, post?.like_count, post?.has_liked, fetchComments]);

    if (!post) return null;

    const handleLike = async () => {
        const previousLiked = isLiked;
        const previousCount = likesCount;

        try {
            // Optimistic update
            setIsLiked(!previousLiked);
            setLikesCount(prev => previousLiked ? prev - 1 : prev + 1);

            if (previousLiked) {
                await communityApi.unlikePost(post.id);
            } else {
                await communityApi.likePost(post.id);
            }
            onRefresh(); // Refresh parent to sync counts
        } catch (err) {
            console.error("Failed to toggle like:", err);
            // Revert on error
            setIsLiked(previousLiked);
            setLikesCount(previousCount);
        }
    };

    const handleDelete = async () => {
        if (!window.confirm("Are you sure you want to delete this post?")) return;
        try {
            await communityApi.deletePost(post.id);
            onRefresh();
            onClose();
        } catch (err) {
            console.error("Failed to delete post:", err);
            alert("Failed to delete post. Please try again.");
        }
    };

    const handleSubmitComment = async () => {
        if (!commentInput.trim()) return;
        setSubmittingComment(true);
        try {
            await communityApi.addComment(post.id, { content: commentInput });
            setCommentInput('');
            fetchComments();
            onRefresh(); // Update comment count in parent
        } catch (err) {
            console.error("Failed to add comment:", err);
        } finally {
            setSubmittingComment(false);
        }
    };

    return (
        <Dialog
            open={open}
            onClose={onClose}
            fullWidth
            maxWidth="md"
            PaperProps={{
                sx: {
                    borderRadius: 3,
                    bgcolor: 'background.paper',
                    overflow: 'hidden'
                }
            }}
        >
            <Box sx={{ position: 'absolute', right: 8, top: 8, zIndex: 1, display: 'flex', gap: 1 }}>
                {canDelete && (
                    <Tooltip title="Delete Post">
                        <IconButton onClick={handleDelete} size="small" color="error">
                            <DeleteIcon />
                        </IconButton>
                    </Tooltip>
                )}
                <IconButton onClick={onClose} size="small">
                    <CloseIcon />
                </IconButton>
            </Box>

            <DialogContent sx={{ p: 0, display: 'flex', flexDirection: { xs: 'column', md: 'row' }, minHeight: 500 }}>
                {/* Left Side: Post Content */}
                <Box sx={{
                    flex: 1,
                    p: 4,
                    borderRight: '1px solid',
                    borderColor: 'divider',
                    bgcolor: post.category === 'sticky' ? '#FFF9C4' : 'transparent', // Visual feedback for sticky
                    color: post.category === 'sticky' ? 'rgba(0,0,0,0.85)' : 'inherit'
                }}>
                    <Stack direction="row" spacing={2} alignItems="center" sx={{ mb: 3 }}>
                        <Avatar
                            src={post.author_avatar}
                            sx={{ width: 44, height: 44, border: '2px solid rgba(93, 224, 255, 0.3)' }}
                        >
                            {post.author_name?.[0] || 'U'}
                        </Avatar>
                        <Box>
                            <Typography variant="subtitle1" fontWeight="bold">
                                {post.author_name || 'Anonymous'}
                            </Typography>
                            <Typography variant="caption" color="text.secondary">
                                {post.author_faculty || 'Aalto Student'} • {new Date(post.created_at).toLocaleString()}
                            </Typography>
                        </Box>
                    </Stack>

                    <Typography variant="h4" fontWeight="800" sx={{ mb: 2, lineHeight: 1.2 }}>
                        {post.title}
                    </Typography>

                    <Typography variant="body1" sx={{ mb: 4, lineHeight: 1.8, whiteSpace: 'pre-wrap' }}>
                        {post.content}
                    </Typography>

                    <Stack direction="row" spacing={1} sx={{ mb: 4 }}>
                        <Chip
                            label={post.category.replace('_', ' ')}
                            size="small"
                            color="primary"
                            variant="outlined"
                        />
                        {post.tags?.map(tag => (
                            <Typography key={tag} variant="caption" color="primary.main">
                                #{tag}
                            </Typography>
                        ))}
                    </Stack>

                    <Divider sx={{ my: 3 }} />

                    <Stack direction="row" spacing={3}>
                        <Button
                            startIcon={isLiked ? <LikeIcon sx={{ color: '#e91e63' }} /> : <UnlikeIcon />}
                            onClick={handleLike}
                            color="inherit"
                        >
                            {likesCount} Likes
                        </Button>
                        <Button startIcon={<CommentIcon />} color="inherit">
                            {post.comment_count} Comments
                        </Button>
                        <Button startIcon={<ShareIcon />} color="inherit">
                            Share
                        </Button>
                    </Stack>
                </Box>

                {/* Right Side: Comments Section */}
                <Box sx={{ width: { xs: '100%', md: 350 }, display: 'flex', flexDirection: 'column', bgcolor: 'rgba(255,255,255,0.02)' }}>
                    <Box sx={{ p: 2, borderBottom: '1px solid', borderColor: 'divider' }}>
                        <Typography variant="h6" fontWeight="bold">Discussions</Typography>
                    </Box>

                    <Box sx={{ flexGrow: 1, overflowY: 'auto', p: 2, maxHeight: 400 }}>
                        {loadingComments ? (
                            <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
                                <CircularProgress size={24} />
                            </Box>
                        ) : comments.length === 0 ? (
                            <Box sx={{ textAlign: 'center', py: 8 }}>
                                <CommentIcon sx={{ fontSize: 40, opacity: 0.2, mb: 1 }} />
                                <Typography variant="body2" color="text.secondary">No comments yet. Start the conversation!</Typography>
                            </Box>
                        ) : (
                            <Stack spacing={2}>
                                {comments.map((comment) => (
                                    <Box key={comment.id}>
                                        <Stack direction="row" spacing={1} alignItems="flex-start">
                                            <Avatar src={comment.author_avatar} sx={{ width: 32, height: 32 }}>
                                                {comment.author_name?.[0] || 'U'}
                                            </Avatar>
                                            <Box sx={{ flexGrow: 1 }}>
                                                <Paper sx={{ p: 1.5, borderRadius: 2, bgcolor: 'background.default' }}>
                                                    <Typography variant="caption" fontWeight="bold" display="block">
                                                        {comment.author_name || 'Anonymous'}
                                                    </Typography>
                                                    <Typography variant="body2">{comment.content}</Typography>
                                                </Paper>
                                                <Typography variant="caption" color="text.secondary" sx={{ ml: 1 }}>
                                                    {new Date(comment.created_at).toLocaleDateString()}
                                                </Typography>
                                            </Box>
                                        </Stack>
                                    </Box>
                                ))}
                            </Stack>
                        )}
                    </Box>

                    <Box sx={{ p: 2, borderTop: '1px solid', borderColor: 'divider' }}>
                        <TextField
                            fullWidth
                            placeholder="Add a comment..."
                            size="small"
                            value={commentInput}
                            onChange={(e) => setCommentInput(e.target.value)}
                            onKeyDown={(e) => e.key === 'Enter' && handleSubmitComment()}
                            autoComplete="off"
                            InputProps={{
                                endAdornment: (
                                    <IconButton
                                        onClick={handleSubmitComment}
                                        disabled={!commentInput.trim() || submittingComment}
                                        size="small"
                                        color="primary"
                                    >
                                        {submittingComment ? <CircularProgress size={20} /> : <SendIcon fontSize="small" />}
                                    </IconButton>
                                )
                            }}
                            sx={{ '& .MuiOutlinedInput-root': { borderRadius: 4 } }}
                        />
                    </Box>
                </Box>
            </DialogContent>
        </Dialog>
    );
};

export default PostDetailDialog;
