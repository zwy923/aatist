import React from 'react';
import { Box, CircularProgress, Typography, Button, Skeleton, Stack } from '@mui/material';
import ErrorOutlineIcon from '@mui/icons-material/ErrorOutline';
import InboxIcon from '@mui/icons-material/Inbox';

export const LoadingSkeleton = ({ count = 3, height = 100 }) => (
    <Stack spacing={2} sx={{ width: '100%' }}>
        {[...Array(count)].map((_, i) => (
            <Skeleton key={i} variant="rounded" width="100%" height={height} sx={{ bgcolor: 'rgba(255, 255, 255, 0.05)' }} />
        ))}
    </Stack>
);

export const EmptyState = ({ message = 'No data found', icon: Icon = InboxIcon, sx = {} }) => (
    <Box sx={{ py: 8, textAlign: 'center', opacity: 0.5, ...sx }}>
        <Icon sx={{ fontSize: 64, mb: 2 }} />
        <Typography variant="h6">{message}</Typography>
    </Box>
);

export const ErrorState = ({ error, onRetry }) => (
    <Box sx={{ py: 8, textAlign: 'center', color: 'error.main' }}>
        <ErrorOutlineIcon sx={{ fontSize: 64, mb: 2 }} />
        <Typography variant="h6" gutterBottom>
            {error || 'Something went wrong'}
        </Typography>
        {onRetry && (
            <Button variant="outlined" color="primary" onClick={onRetry} sx={{ mt: 2 }}>
                Retry
            </Button>
        )}
    </Box>
);

export const StateContainer = ({
    loading,
    error,
    empty,
    onRetry,
    children,
    loadingComponent,
    emptyMessage,
    skeletonCount = 3,
    skeletonHeight = 100,
    emptyStateSx = {}
}) => {
    if (loading) {
        return loadingComponent || <LoadingSkeleton count={skeletonCount} height={skeletonHeight} />;
    }

    if (error) {
        return <ErrorState error={error} onRetry={onRetry} />;
    }

    if (empty) {
        return <EmptyState message={emptyMessage} sx={emptyStateSx} />;
    }

    return <>{children}</>;
};
