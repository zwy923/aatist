import React from 'react';
import { Card, CardContent, Typography, Box, Chip, Stack, Button } from '@mui/material';
import {
    LocationOn,
    AccessTime,
    AttachMoney,
    Language,
    CalendarToday,
    Timer,
    FlashOn
} from '@mui/icons-material';
import { Link } from 'react-router-dom';
import SavedButton from './SavedButton';

// Maps API response (snake_case) to card props
const mapOpportunity = (o) => ({
    id: o.id,
    title: o.title,
    description: o.description,
    budget: o.budget_value != null ? `€${o.budget_value}` : 'Negotiable',
    payType: o.budget_type || 'fixed',
    location: o.location,
    workLanguage: Array.isArray(o.languages) ? o.languages.join(', ') : (o.languages || 'N/A'),
    startDate: o.start_date,
    duration: o.duration_months != null ? `${o.duration_months} month(s)` : 'Flexible',
    publishedAt: o.published_at,
    tags: o.tags || [],
    urgent: o.urgent,
    savedByMe: o.is_favorite,
    appliedByMe: o.applied_by_me,
});

const OpportunityCard = ({ opportunity }) => {
    const opp = mapOpportunity(opportunity);
    const {
        id,
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
        appliedByMe
    } = opp;

    const formatDate = (dateString) => {
        if (!dateString) return 'N/A';
        return new Date(dateString).toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        });
    };

    const formatTimeAgo = (dateString) => {
        if (!dateString) return '';
        const now = new Date();
        const past = new Date(dateString);
        const diffInMs = now - past;
        const diffInDays = Math.floor(diffInMs / (1000 * 60 * 60 * 24));

        if (diffInDays === 0) return 'Today';
        if (diffInDays === 1) return 'Yesterday';
        if (diffInDays < 30) return `${diffInDays} days ago`;
        return formatDate(dateString);
    };

    return (
        <Card
            sx={{
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
                position: 'relative',
                transition: 'transform 0.2s, box-shadow 0.2s',
                '&:hover': {
                    transform: 'translateY(-4px)',
                    boxShadow: '0 12px 24px rgba(0,0,0,0.2)',
                    borderColor: 'primary.main',
                },
                border: '1px solid #e5e7eb',
                background: '#ffffff',
            }}
        >
            {urgent && (
                <Box
                    sx={{
                        position: 'absolute',
                        top: 12,
                        right: 50,
                        backgroundColor: 'error.main',
                        color: 'white',
                        px: 1,
                        py: 0.5,
                        borderRadius: 1,
                        display: 'flex',
                        alignItems: 'center',
                        gap: 0.5,
                        zIndex: 1,
                        fontSize: '0.75rem',
                        fontWeight: 'bold',
                        textTransform: 'uppercase'
                    }}
                >
                    <FlashOn fontSize="inherit" /> Urgent
                </Box>
            )}

            <Box sx={{ position: 'absolute', top: 8, right: 8, zIndex: 1 }}>
                <SavedButton targetId={id} initialSaved={savedByMe} />
            </Box>

            <CardContent sx={{ flexGrow: 1, pt: 3 }}>
                <Typography variant="h6" component="div" gutterBottom sx={{ fontWeight: 'bold', pr: 8 }}>
                    {title}
                </Typography>

                <Stack direction="row" spacing={2} sx={{ mb: 2, flexWrap: 'wrap', gap: 1 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', color: 'primary.main' }}>
                        <AttachMoney fontSize="small" />
                        <Typography variant="subtitle2" sx={{ fontWeight: 'bold' }}>
                            {budget} ({payType})
                        </Typography>
                    </Box>
                    <Box sx={{ display: 'flex', alignItems: 'center', color: 'text.secondary' }}>
                        <LocationOn fontSize="small" sx={{ mr: 0.5 }} />
                        <Typography variant="body2">{location}</Typography>
                    </Box>
                </Stack>

                <Typography
                    variant="body2"
                    color="text.secondary"
                    sx={{
                        mb: 2,
                        display: '-webkit-box',
                        WebkitLineClamp: 3,
                        WebkitBoxOrient: 'vertical',
                        overflow: 'hidden',
                        minHeight: '3em'
                    }}
                >
                    {description}
                </Typography>

                <Stack direction="row" spacing={1} sx={{ mb: 2, flexWrap: 'wrap', gap: 0.5 }}>
                    {tags && tags.map((tag, index) => (
                        <Chip
                            key={index}
                            label={tag}
                            size="small"
                            variant="outlined"
                            sx={{ borderColor: '#e5e7eb', color: 'text.secondary' }}
                        />
                    ))}
                </Stack>

                <Stack spacing={1}>
                    <Box sx={{ display: 'flex', alignItems: 'center', color: 'text.secondary', gap: 1 }}>
                        <Language fontSize="inherit" />
                        <Typography variant="caption">Language: {workLanguage}</Typography>
                    </Box>
                    <Box sx={{ display: 'flex', alignItems: 'center', color: 'text.secondary', gap: 1 }}>
                        <CalendarToday fontSize="inherit" />
                        <Typography variant="caption">Starts: {formatDate(startDate)}</Typography>
                    </Box>
                    <Box sx={{ display: 'flex', alignItems: 'center', color: 'text.secondary', gap: 1 }}>
                        <Timer fontSize="inherit" />
                        <Typography variant="caption">Duration: {duration}</Typography>
                    </Box>
                </Stack>
            </CardContent>

            <Box sx={{ p: 2, pt: 0, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Typography variant="caption" color="text.secondary">
                    Posted {formatTimeAgo(publishedAt)}
                </Typography>
                <Button
                    component={Link}
                    to={`/opportunities/${id}`}
                    variant={appliedByMe ? "outlined" : "contained"}
                    size="small"
                    color={appliedByMe ? "success" : "primary"}
                >
                    {appliedByMe ? "Applied" : "View Details"}
                </Button>
            </Box>
        </Card>
    );
};

export default OpportunityCard;
