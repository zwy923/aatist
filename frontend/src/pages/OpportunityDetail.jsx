import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
    Container,
    Grid,
    Typography,
    Box,
    Button,
    Stack,
    Chip,
    Paper,
    Divider,
    Avatar,
    Breadcrumbs,
    Link as MuiLink
} from '@mui/material';
import {
    LocationOn,
    AccessTime,
    AttachMoney,
    Language,
    CalendarToday,
    Timer,
    FlashOn,
    ArrowBack,
    Business,
    Person,
    CheckCircle
} from '@mui/icons-material';
import { Link } from 'react-router-dom';
import { opportunitiesApi } from '../features/opportunities/api/opportunities';
import SavedButton from '../features/opportunities/components/SavedButton';
import ApplyModal from '../features/opportunities/components/ApplyModal';
import { StateContainer } from '../shared/components/ui/StateContainer';
import PageLayout from '../shared/components/PageLayout';
import useAuthStore from '../shared/stores/authStore';

const OpportunityDetailPage = () => {
    const { id } = useParams();
    const navigate = useNavigate();
    const { isAuthenticated } = useAuthStore();

    const [opportunity, setOpportunity] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [applyModalOpen, setApplyModalOpen] = useState(false);

    const fetchOpportunity = useCallback(async () => {
        setLoading(true);
        setError(null);
        try {
            const response = await opportunitiesApi.getOpportunity(id);
            setOpportunity(response.data.data || response.data);
        } catch (err) {
            setError(err.response?.data?.error || 'Failed to fetch opportunity details');
        } finally {
            setLoading(false);
        }
    }, [id]);

    useEffect(() => {
        fetchOpportunity();
    }, [fetchOpportunity]);

    const formatDate = (dateString) => {
        if (!dateString) return 'N/A';
        return new Date(dateString).toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'long',
            day: 'numeric'
        });
    };

    if (loading) return <StateContainer loading={true} />;
    if (error) return <StateContainer error={error} />;
    if (!opportunity) return <StateContainer empty={true} />;

    const {
        title,
        description,
        budget,
        payType,
        location,
        workLanguage,
        startDate,
        duration,
        publishedAt,
        tags,
        urgent,
        savedByMe,
        appliedByMe,
        category,
        client
    } = opportunity;

    return (
        <PageLayout maxWidth="lg">
            <Box sx={{ mb: 4 }}>
                <Breadcrumbs sx={{ mb: 2 }}>
                    <MuiLink component={Link} to="/opportunities" underline="hover" color="inherit">
                        Opportunities
                    </MuiLink>
                    <Typography color="text.primary">{category}</Typography>
                </Breadcrumbs>

                <Button
                    startIcon={<ArrowBack />}
                    onClick={() => navigate(-1)}
                    sx={{ mb: 2 }}
                >
                    Back to List
                </Button>
            </Box>

            <Grid container spacing={4}>
                <Grid item xs={12} md={8}>
                    <Paper sx={{ p: 4, background: 'rgba(7, 12, 30, 0.4)', backdropFilter: 'blur(10px)', border: '1px solid rgba(255,255,255,0.05)' }}>
                        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 3 }}>
                            <Box>
                                {urgent && (
                                    <Chip
                                        icon={<FlashOn />}
                                        label="Urgent"
                                        color="error"
                                        size="small"
                                        sx={{ mb: 1, fontWeight: 'bold' }}
                                    />
                                )}
                                <Typography variant="h4" component="h1" gutterBottom sx={{ fontWeight: 'bold' }}>
                                    {title}
                                </Typography>
                                <Typography variant="body2" color="text.secondary">
                                    Posted on {formatDate(publishedAt)}
                                </Typography>
                            </Box>
                            <SavedButton targetId={id} initialSaved={savedByMe} />
                        </Box>

                        <Divider sx={{ my: 3, opacity: 0.1 }} />

                        <Typography variant="h6" gutterBottom sx={{ fontWeight: 'bold' }}>
                            Description
                        </Typography>
                        <Typography variant="body1" sx={{ whiteSpace: 'pre-wrap', mb: 4, color: 'text.secondary', lineHeight: 1.7 }}>
                            {description}
                        </Typography>

                        <Typography variant="h6" gutterBottom sx={{ fontWeight: 'bold' }}>
                            Required Skills & Tags
                        </Typography>
                        <Stack direction="row" spacing={1} sx={{ mb: 4, flexWrap: 'wrap', gap: 1 }}>
                            {tags && tags.map((tag, index) => (
                                <Chip
                                    key={index}
                                    label={tag}
                                    variant="outlined"
                                    sx={{ background: 'rgba(93, 224, 255, 0.05)' }}
                                />
                            ))}
                        </Stack>
                    </Paper>
                </Grid>

                <Grid item xs={12} md={4}>
                    <Stack spacing={3}>
                        <Paper sx={{ p: 3, background: 'rgba(7, 12, 30, 0.6)', border: '1px solid rgba(93, 224, 255, 0.2)' }}>
                            <Typography variant="h6" gutterBottom sx={{ fontWeight: 'bold', display: 'flex', alignItems: 'center', gap: 1 }}>
                                <AttachMoney color="primary" /> Budget & Terms
                            </Typography>
                            <Typography variant="h4" color="primary.main" sx={{ fontWeight: 'bold', mb: 1 }}>
                                {budget}
                            </Typography>
                            <Typography variant="body2" color="text.secondary" gutterBottom>
                                {payType} Payment
                            </Typography>

                            <Divider sx={{ my: 2, opacity: 0.1 }} />

                            <Stack spacing={2}>
                                <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                                    <LocationOn color="action" />
                                    <Box>
                                        <Typography variant="caption" color="text.secondary" display="block">Location</Typography>
                                        <Typography variant="body2">{location}</Typography>
                                    </Box>
                                </Box>
                                <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                                    <Language color="action" />
                                    <Box>
                                        <Typography variant="caption" color="text.secondary" display="block">Work Language</Typography>
                                        <Typography variant="body2">{workLanguage}</Typography>
                                    </Box>
                                </Box>
                                <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                                    <CalendarToday color="action" />
                                    <Box>
                                        <Typography variant="caption" color="text.secondary" display="block">Start Date</Typography>
                                        <Typography variant="body2">{formatDate(startDate)}</Typography>
                                    </Box>
                                </Box>
                                <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                                    <Timer color="action" />
                                    <Box>
                                        <Typography variant="caption" color="text.secondary" display="block">Duration</Typography>
                                        <Typography variant="body2">{duration}</Typography>
                                    </Box>
                                </Box>
                            </Stack>

                            <Box sx={{ mt: 4 }}>
                                {appliedByMe ? (
                                    <Button
                                        fullWidth
                                        variant="contained"
                                        color="success"
                                        disabled
                                        startIcon={<CheckCircle />}
                                        size="large"
                                    >
                                        Applied
                                    </Button>
                                ) : (
                                    <Button
                                        fullWidth
                                        variant="contained"
                                        size="large"
                                        onClick={() => {
                                            if (!isAuthenticated) {
                                                navigate('/auth/login', { state: { from: `/opportunities/${id}` } });
                                            } else {
                                                setApplyModalOpen(true);
                                            }
                                        }}
                                    >
                                        Apply Now
                                    </Button>
                                )}
                            </Box>
                        </Paper>

                        {client && (
                            <Paper sx={{ p: 3, background: 'rgba(7, 12, 30, 0.4)', border: '1px solid rgba(255,255,255,0.05)' }}>
                                <Typography variant="subtitle1" gutterBottom sx={{ fontWeight: 'bold' }}>
                                    About the Client
                                </Typography>
                                <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
                                    <Avatar sx={{ bgcolor: 'secondary.main' }}>
                                        {client.name?.charAt(0) || <Business />}
                                    </Avatar>
                                    <Box>
                                        <Typography variant="body1" sx={{ fontWeight: 'bold' }}>{client.name}</Typography>
                                        <Typography variant="caption" color="text.secondary">
                                            {client.location || 'Location not specified'}
                                        </Typography>
                                    </Box>
                                </Box>
                                <Typography variant="body2" color="text.secondary">
                                    {client.description || 'No client description available.'}
                                </Typography>
                            </Paper>
                        )}
                    </Stack>
                </Grid>
            </Grid>

            <ApplyModal
                open={applyModalOpen}
                onClose={() => setApplyModalOpen(false)}
                opportunityId={id}
                opportunityTitle={title}
                onSuccess={() => {
                    setOpportunity(prev => ({ ...prev, appliedByMe: true }));
                }}
            />
        </PageLayout>
    );
};

export default OpportunityDetailPage;
