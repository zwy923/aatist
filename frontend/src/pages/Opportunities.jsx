import React, { useState, useEffect, useCallback } from 'react';
import {
    Container,
    Grid,
    Typography,
    Box,
    TextField,
    MenuItem,
    Button,
    Stack,
    Pagination,
    InputAdornment,
    Drawer,
    IconButton,
    useMediaQuery,
    useTheme,
    Paper,
    Divider
} from '@mui/material';
import {
    Search,
    FilterList,
    Clear,
    Sort,
    Work,
    LocationOn,
    AttachMoney
} from '@mui/icons-material';
import { opportunitiesApi } from '../features/opportunities/api/opportunities';
import OpportunityCard from '../features/opportunities/components/OpportunityCard';
import { StateContainer } from '../shared/components/ui/StateContainer';
import PageLayout from '../shared/components/PageLayout';

const CATEGORIES = ['Translation', 'Interpretation', 'Voiceover', 'Subtitling', 'Transcription', 'Other'];
const LANGUAGES = ['English', 'Chinese', 'Japanese', 'Korean', 'French', 'German', 'Spanish', 'Other'];
const PAY_TYPES = ['Fixed', 'Hourly', 'Per Word', 'Per Minute'];
const SORT_OPTIONS = [
    { label: 'Newest', value: 'publishedAt', order: 'desc' },
    { label: 'Oldest', value: 'publishedAt', order: 'asc' },
    { label: 'Budget (High to Low)', value: 'budget', order: 'desc' },
    { label: 'Budget (Low to High)', value: 'budget', order: 'asc' },
    { label: 'Starting Soon', value: 'startDate', order: 'asc' },
];

const OpportunitiesPage = () => {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('md'));

    const [filters, setFilters] = useState({
        q: '',
        category: '',
        location: '',
        workLanguage: '',
        budgetMin: '',
        budgetMax: '',
        payType: '',
        sort: 'publishedAt',
        order: 'desc',
        page: 1,
        pageSize: 12
    });

    const [data, setData] = useState({ items: [], total: 0 });
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [drawerOpen, setDrawerOpen] = useState(false);

    const lastFetchedFilters = React.useRef(null);

    const fetchOpportunities = useCallback(async () => {
        // Prevent redundant fetches if filters haven't changed
        const currentFiltersStr = JSON.stringify(filters);
        if (lastFetchedFilters.current === currentFiltersStr) {
            return;
        }

        setLoading(true);
        setError(null);
        try {
            const response = await opportunitiesApi.getOpportunities(filters);
            const rawData = response.data.data || response.data;
            const items = rawData.items || rawData.opportunities || [];
            const total = rawData.pagination?.total || rawData.total || items.length;
            setData({ items, total });
            lastFetchedFilters.current = currentFiltersStr;
        } catch (err) {
            setError(err.response?.data?.error || 'Failed to fetch opportunities');
        } finally {
            setLoading(false);
        }
    }, [filters]);

    useEffect(() => {
        fetchOpportunities();
    }, [fetchOpportunities]);

    const handleFilterChange = (name, value) => {
        setFilters(prev => ({ ...prev, [name]: value, page: 1 }));
    };

    const handleClearFilters = () => {
        setFilters({
            q: '',
            category: '',
            location: '',
            workLanguage: '',
            budgetMin: '',
            budgetMax: '',
            payType: '',
            sort: 'publishedAt',
            order: 'desc',
            page: 1,
            pageSize: 12
        });
    };

    const handlePageChange = (event, value) => {
        setFilters(prev => ({ ...prev, page: value }));
        window.scrollTo({ top: 0, behavior: 'smooth' });
    };

    const FilterContent = ({ prefix = 'sidebar' }) => (
        <Stack spacing={3} sx={{ p: isMobile ? 3 : 0 }}>
            <Typography variant="h6" sx={{ mb: 1, display: 'flex', alignItems: 'center', gap: 1 }}>
                <FilterList /> Filters
            </Typography>

            <TextField
                id={`${prefix}-search`}
                fullWidth
                label="Search"
                value={filters.q}
                onChange={(e) => handleFilterChange('q', e.target.value)}
                InputProps={{
                    startAdornment: (
                        <InputAdornment position="start">
                            <Search fontSize="small" />
                        </InputAdornment>
                    ),
                }}
            />

            <TextField
                id={`${prefix}-category`}
                select
                fullWidth
                label="Category"
                value={filters.category}
                onChange={(e) => handleFilterChange('category', e.target.value)}
            >
                <MenuItem value="">All Categories</MenuItem>
                {CATEGORIES.map(cat => (
                    <MenuItem key={cat} value={cat}>{cat}</MenuItem>
                ))}
            </TextField>

            <TextField
                id={`${prefix}-location`}
                fullWidth
                label="Location"
                value={filters.location}
                onChange={(e) => handleFilterChange('location', e.target.value)}
                InputProps={{
                    startAdornment: (
                        <InputAdornment position="start">
                            <LocationOn fontSize="small" />
                        </InputAdornment>
                    ),
                }}
            />

            <TextField
                id={`${prefix}-language`}
                select
                fullWidth
                label="Work Language"
                value={filters.workLanguage}
                onChange={(e) => handleFilterChange('workLanguage', e.target.value)}
            >
                <MenuItem value="">All Languages</MenuItem>
                {LANGUAGES.map(lang => (
                    <MenuItem key={lang} value={lang}>{lang}</MenuItem>
                ))}
            </TextField>

            <Box>
                <Typography variant="caption" color="text.secondary" gutterBottom sx={{ display: 'block', mb: 1 }}>
                    Budget Range
                </Typography>
                <Stack direction="row" spacing={1}>
                    <TextField
                        id={`${prefix}-budget-min`}
                        size="small"
                        placeholder="Min"
                        type="number"
                        value={filters.budgetMin}
                        onChange={(e) => handleFilterChange('budgetMin', e.target.value)}
                    />
                    <TextField
                        id={`${prefix}-budget-max`}
                        size="small"
                        placeholder="Max"
                        type="number"
                        value={filters.budgetMax}
                        onChange={(e) => handleFilterChange('budgetMax', e.target.value)}
                    />
                </Stack>
            </Box>

            <TextField
                id={`${prefix}-pay-type`}
                select
                fullWidth
                label="Pay Type"
                value={filters.payType}
                onChange={(e) => handleFilterChange('payType', e.target.value)}
            >
                <MenuItem value="">Any Pay Type</MenuItem>
                {PAY_TYPES.map(type => (
                    <MenuItem key={type} value={type}>{type}</MenuItem>
                ))}
            </TextField>

            <Button
                fullWidth
                variant="outlined"
                startIcon={<Clear />}
                onClick={handleClearFilters}
                sx={{ mt: 2 }}
            >
                Clear All
            </Button>
        </Stack>
    );

    return (
        <PageLayout maxWidth="xl">
            <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 2 }}>
                <Box>
                    <Typography variant="h4" component="h1" gutterBottom sx={{ fontWeight: 'bold' }}>
                        Explore Opportunities
                    </Typography>
                    <Typography variant="body1" color="text.secondary">
                        Find the perfect project that matches your skills and interests.
                    </Typography>
                </Box>

                <Stack direction="row" spacing={2} alignItems="center">
                    {isMobile && (
                        <Button
                            variant="outlined"
                            startIcon={<FilterList />}
                            onClick={() => setDrawerOpen(true)}
                        >
                            Filters
                        </Button>
                    )}

                    <TextField
                        select
                        size="small"
                        label="Sort By"
                        value={`${filters.sort}-${filters.order}`}
                        onChange={(e) => {
                            const [sort, order] = e.target.value.split('-');
                            setFilters(prev => ({ ...prev, sort, order, page: 1 }));
                        }}
                        sx={{ minWidth: 200 }}
                        InputProps={{
                            startAdornment: (
                                <InputAdornment position="start">
                                    <Sort fontSize="small" />
                                </InputAdornment>
                            ),
                        }}
                    >
                        {SORT_OPTIONS.map(opt => (
                            <MenuItem key={`${opt.value}-${opt.order}`} value={`${opt.value}-${opt.order}`}>
                                {opt.label}
                            </MenuItem>
                        ))}
                    </TextField>
                </Stack>
            </Box>

            <Grid container spacing={4}>
                {!isMobile && (
                    <Grid item md={3}>
                        <Paper sx={{ p: 3, position: 'sticky', top: 24, background: 'rgba(7, 12, 30, 0.4)', backdropFilter: 'blur(10px)', border: '1px solid rgba(255,255,255,0.05)' }}>
                            <FilterContent prefix="sidebar" />
                        </Paper>
                    </Grid>
                )}

                <Grid item xs={12} md={9}>
                    <StateContainer loading={loading} error={error} empty={data.items.length === 0}>
                        <Grid container spacing={3}>
                            {data.items.map(opportunity => (
                                <Grid item xs={12} sm={6} lg={4} key={opportunity.id}>
                                    <OpportunityCard opportunity={opportunity} />
                                </Grid>
                            ))}
                        </Grid>

                        {data.total > filters.pageSize && (
                            <Box sx={{ mt: 6, display: 'flex', justifyContent: 'center' }}>
                                <Pagination
                                    count={Math.ceil(data.total / filters.pageSize)}
                                    page={filters.page}
                                    onChange={handlePageChange}
                                    color="primary"
                                    size="large"
                                />
                            </Box>
                        )}
                    </StateContainer>
                </Grid>
            </Grid>

            <Drawer
                anchor="left"
                open={drawerOpen}
                onClose={() => setDrawerOpen(false)}
                PaperProps={{ sx: { width: 300 } }}
            >
                <FilterContent prefix="drawer" />
            </Drawer>
        </PageLayout>
    );
};

export default OpportunitiesPage;
