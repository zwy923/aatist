import React, { useState } from 'react';
import { IconButton, Tooltip, CircularProgress } from '@mui/material';
import { Bookmark, BookmarkBorder } from '@mui/icons-material';
import { profileApi } from '../../profile/api/profile';
import useAuthStore from '../../../shared/stores/authStore';

const SavedButton = ({ targetId, type = 'opportunity', initialSaved = false, onToggle }) => {
    const [isSaved, setIsSaved] = useState(initialSaved);
    const [loading, setLoading] = useState(false);
    const { isAuthenticated } = useAuthStore();

    const handleToggle = async (e) => {
        e.preventDefault();
        e.stopPropagation();

        if (!isAuthenticated) {
            // Handle unauthenticated state (e.g., redirect to login or show message)
            alert('Please login to save opportunities');
            return;
        }

        setLoading(true);
        try {
            if (isSaved) {
                await profileApi.removeSavedItemByTarget(type, targetId);
                setIsSaved(false);
            } else {
                await profileApi.saveItem(type, targetId);
                setIsSaved(true);
            }
            if (onToggle) onToggle(!isSaved);
        } catch (error) {
            console.error('Failed to toggle save state:', error);
        } finally {
            setLoading(false);
        }
    };

    return (
        <Tooltip title={isSaved ? 'Remove from saved' : 'Save opportunity'}>
            <IconButton
                onClick={handleToggle}
                disabled={loading}
                sx={{
                    color: isSaved ? 'primary.main' : 'text.secondary',
                    '&:hover': {
                        color: 'primary.main',
                        backgroundColor: 'rgba(93, 224, 255, 0.1)',
                    }
                }}
            >
                {loading ? (
                    <CircularProgress size={24} color="inherit" />
                ) : isSaved ? (
                    <Bookmark />
                ) : (
                    <BookmarkBorder />
                )}
            </IconButton>
        </Tooltip>
    );
};

export default SavedButton;
