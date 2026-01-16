import React, { useState } from 'react';
import {
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    Button,
    TextField,
    Typography,
    Box,
    CircularProgress
} from '@mui/material';
import { opportunitiesApi } from '../api/opportunities';

const ApplyModal = ({ open, onClose, opportunityId, opportunityTitle, onSuccess }) => {
    const [message, setMessage] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    const handleApply = async () => {
        setLoading(true);
        setError(null);
        try {
            await opportunitiesApi.applyToOpportunity(opportunityId, { message });
            onSuccess();
            onClose();
        } catch (err) {
            setError(err.response?.data?.error || 'Failed to submit application');
        } finally {
            setLoading(false);
        }
    };

    return (
        <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
            <DialogTitle sx={{ pb: 1 }}>
                Apply for {opportunityTitle}
            </DialogTitle>
            <DialogContent>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
                    Introduce yourself and explain why you're a good fit for this opportunity.
                </Typography>
                <TextField
                    id="apply-message"
                    autoFocus
                    multiline
                    rows={4}
                    fullWidth
                    label="Cover Message (Optional)"
                    variant="outlined"
                    value={message}
                    onChange={(e) => setMessage(e.target.value)}
                    disabled={loading}
                    placeholder="Tell the client about your relevant experience..."
                />
                {error && (
                    <Typography color="error" variant="caption" sx={{ mt: 1, display: 'block' }}>
                        {error}
                    </Typography>
                )}
            </DialogContent>
            <DialogActions sx={{ px: 3, pb: 3 }}>
                <Button onClick={onClose} disabled={loading}>
                    Cancel
                </Button>
                <Button
                    onClick={handleApply}
                    variant="contained"
                    disabled={loading}
                    startIcon={loading && <CircularProgress size={20} />}
                >
                    {loading ? 'Submitting...' : 'Submit Application'}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default ApplyModal;
