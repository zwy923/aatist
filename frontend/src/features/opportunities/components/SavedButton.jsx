import React, { useEffect, useState } from 'react';
import { Alert, IconButton, Snackbar, Tooltip, CircularProgress } from '@mui/material';
import { Bookmark, BookmarkBorder, Favorite, FavoriteBorder } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { profileApi } from '../../profile/api/profile';
import useAuthStore from '../../../shared/stores/authStore';

const SavedButton = ({
    targetId,
    type = 'opportunity',
    initialSaved = false,
    onToggle,
    iconSet = 'bookmark',
    size = 'medium',
    sx = {},
}) => {
    const [isSaved, setIsSaved] = useState(initialSaved);
    const [loading, setLoading] = useState(false);
    const [errorMsg, setErrorMsg] = useState(null);
    const { isAuthenticated } = useAuthStore();
    const navigate = useNavigate();

    useEffect(() => {
        setIsSaved(initialSaved);
    }, [initialSaved]);

    const handleToggle = async (e) => {
        e.preventDefault();
        e.stopPropagation();

        if (!isAuthenticated) {
            navigate(`/auth/login?redirect=${encodeURIComponent(window.location.pathname + window.location.search)}`);
            return;
        }

        setLoading(true);
        setErrorMsg(null);
        try {
            const wasSaved = isSaved;
            if (isSaved) {
                await profileApi.removeSavedItemByTarget(type, targetId);
                setIsSaved(false);
            } else {
                await profileApi.saveItem(type, targetId);
                setIsSaved(true);
            }
            if (onToggle) onToggle(!wasSaved);
        } catch (error) {
            const msg =
                error?.message ||
                error?.response?.data?.error?.message ||
                'Could not update saved items. Please try again.';
            setErrorMsg(msg);
        } finally {
            setLoading(false);
        }
    };

    const Filled = iconSet === 'favorite' ? Favorite : Bookmark;
    const Outline = iconSet === 'favorite' ? FavoriteBorder : BookmarkBorder;
    const brandTeal = '#048B7F';
    const savedAccent = type === 'opportunity' ? brandTeal : 'primary.main';

    return (
        <>
            <Tooltip title={isSaved ? 'Remove from saved' : 'Save opportunity'}>
                <IconButton
                    size={size}
                    onClick={handleToggle}
                    disabled={loading}
                    sx={{
                        color: isSaved ? savedAccent : 'text.secondary',
                        '&:hover': {
                            color: savedAccent,
                            backgroundColor:
                                type === 'opportunity' ? 'rgba(4, 139, 127, 0.1)' : 'rgba(93, 224, 255, 0.1)',
                        },
                        ...sx,
                    }}
                >
                    {loading ? (
                        <CircularProgress size={iconSet === 'favorite' ? 20 : 24} color="inherit" />
                    ) : isSaved ? (
                        <Filled fontSize={size === 'small' ? 'small' : 'medium'} />
                    ) : (
                        <Outline fontSize={size === 'small' ? 'small' : 'medium'} />
                    )}
                </IconButton>
            </Tooltip>
            <Snackbar
                open={Boolean(errorMsg)}
                autoHideDuration={6000}
                onClose={() => setErrorMsg(null)}
                anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
            >
                <Alert onClose={() => setErrorMsg(null)} severity="error" variant="filled" sx={{ width: '100%' }}>
                    {errorMsg}
                </Alert>
            </Snackbar>
        </>
    );
};

export default SavedButton;
