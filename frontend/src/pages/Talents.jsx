import React, { useState, useEffect, useCallback } from 'react';
import {
    Box,
    Typography,
    Grid,
    Card,
    CardContent,
    Avatar,
    Chip,
    Stack,
    TextField,
    InputAdornment,
    Button,
    IconButton,
    CircularProgress,
    Alert,
    Tooltip,
    Select,
    MenuItem,
    FormControl,
    InputLabel,
    Collapse,
    Paper,
    Divider
} from '@mui/material';
import {
    Search as SearchIcon,
    FilterList as FilterIcon,
    WorkOutline as WorkIcon,
    AccessTime as TimeIcon,
    School as SchoolIcon,
    Favorite as FavoriteIcon,
    FavoriteBorder as FavoriteBorderIcon,
    NavigateNext as ViewProfileIcon,
    Clear as ClearIcon
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import PageLayout from '../shared/components/PageLayout';
import { profileApi } from '../features/profile/api/profile';
import { useAuth } from '../features/auth/hooks/useAuth';

const FACULTIES = [
    "Arts, Design and Architecture",
    "Business",
    "Chemical Engineering",
    "Electrical Engineering",
    "Engineering",
    "Science"
];

const TalentCard = ({ student, onSave, isSaved }) => {
    const navigate = useNavigate();

    // Status color logic (mock availability status based on weekly hours or custom field if added)
    const isAvailable = student.weekly_hours > 0;
    const statusColor = isAvailable ? "#4caf50" : "#f44336";

    return (
        <Card sx={{
            height: '100%',
            display: 'flex',
            flexDirection: 'column',
            position: 'relative',
            borderRadius: 4,
            background: "rgba(255, 255, 255, 0.03)",
            border: "1px solid rgba(255, 255, 255, 0.05)",
            transition: "transform 0.2s, box-shadow 0.2s",
            '&:hover': {
                transform: "translateY(-4px)",
                boxShadow: "0 12px 24px rgba(0,0,0,0.3)",
                border: "1px solid rgba(93, 224, 255, 0.2)"
            }
        }}>
            <Box sx={{ p: 3, display: 'flex', alignItems: 'flex-start', gap: 2 }}>
                <Box sx={{ position: 'relative' }}>
                    <Avatar
                        src={student.avatar_url}
                        sx={{ width: 80, height: 80, borderRadius: 3, fontSize: '2rem' }}
                    >
                        {student.name?.charAt(0)}
                    </Avatar>
                    <Box sx={{
                        position: 'absolute',
                        bottom: -2,
                        right: -2,
                        width: 16,
                        height: 16,
                        borderRadius: '50%',
                        bgcolor: statusColor,
                        border: '3px solid #030617'
                    }} />
                </Box>
                <Box sx={{ flex: 1 }}>
                    <Typography variant="h6" fontWeight="700" gutterBottom>
                        {student.name}
                    </Typography>
                    <Stack direction="row" spacing={0.5} alignItems="center" color="text.secondary" sx={{ mb: 1 }}>
                        <SchoolIcon sx={{ fontSize: 16 }} />
                        <Typography variant="body2" noWrap sx={{ maxWidth: 150 }}>
                            {student.faculty || student.major || "No Field"}
                        </Typography>
                    </Stack>
                    <Stack direction="row" spacing={0.5} alignItems="center" color="primary.main">
                        <TimeIcon sx={{ fontSize: 16 }} />
                        <Typography variant="caption" fontWeight="600">
                            {student.weekly_hours ? `${student.weekly_hours} hrs/week` : "Not available"}
                        </Typography>
                    </Stack>
                </Box>
                <IconButton
                    size="small"
                    onClick={(e) => { e.stopPropagation(); onSave(student.id); }}
                    sx={{ color: isSaved ? 'secondary.main' : 'text.disabled' }}
                >
                    {isSaved ? <FavoriteIcon /> : <FavoriteBorderIcon />}
                </IconButton>
            </Box>

            <Divider sx={{ mx: 2, opacity: 0.1 }} />

            <CardContent sx={{ flexGrow: 1, pt: 2 }}>
                <Typography variant="body2" color="text.secondary" sx={{
                    mb: 2,
                    display: '-webkit-box',
                    WebkitLineClamp: 2,
                    WebkitBoxOrient: 'vertical',
                    overflow: 'hidden',
                    height: 40
                }}>
                    {student.bio || "No bio description provided."}
                </Typography>

                <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.8 }}>
                    {student.skills?.slice(0, 4).map((skill, index) => (
                        <Chip
                            key={index}
                            label={typeof skill === 'string' ? skill : skill.name}
                            size="small"
                            sx={{
                                bgcolor: "rgba(93, 224, 255, 0.1)",
                                color: "#5de0ff",
                                fontWeight: 500,
                                fontSize: '0.7rem'
                            }}
                        />
                    ))}
                    {student.skills?.length > 4 && (
                        <Typography variant="caption" color="text.disabled" sx={{ mt: 0.5 }}>
                            +{student.skills.length - 4} more
                        </Typography>
                    )}
                </Box>
            </CardContent>

            <Box sx={{ p: 2, mt: 'auto' }}>
                <Button
                    fullWidth
                    variant="outlined"
                    endIcon={<ViewProfileIcon />}
                    onClick={() => navigate(`/profile?id=${student.id}`)}
                    sx={{
                        borderRadius: 2,
                        textTransform: 'none',
                        borderColor: "rgba(93, 224, 255, 0.3)",
                        '&:hover': {
                            borderColor: "primary.main",
                            bgcolor: "rgba(93, 224, 255, 0.05)"
                        }
                    }}
                >
                    View Profile
                </Button>
            </Box>
        </Card>
    );
};

const Talents = () => {
    const [students, setStudents] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [search, setSearch] = useState("");
    const [faculty, setFaculty] = useState("");
    const [minHours, setMinHours] = useState("");
    const [showFilters, setShowFilters] = useState(false);
    const [savedIds, setSavedIds] = useState(new Set());
    const { user } = useAuth();

    const fetchTalents = useCallback(async () => {
        setLoading(true);
        setError(null);
        try {
            const params = {
                q: search || undefined,
                faculty: faculty || undefined,
                min_hours: minHours ? parseInt(minHours) : undefined,
                limit: 50
            };
            const response = await profileApi.searchUsers(params);
            setStudents(response.data.data || []);

            // Also fetch saved items to show favorite state
            if (user) {
                const savedResp = await profileApi.getSavedItems({ type: 'user' });
                const ids = new Set((savedResp.data.data || []).map(item => item.item_id));
                setSavedIds(ids);
            }
        } catch (err) {
            console.error("Failed to fetch talents:", err);
            setError("Failed to load talent list. Please try again.");
        } finally {
            setLoading(false);
        }
    }, [search, faculty, minHours, user]);

    useEffect(() => {
        const timer = setTimeout(() => {
            fetchTalents();
        }, 300);
        return () => clearTimeout(timer);
    }, [fetchTalents]);

    const handleSaveToggle = async (studentId) => {
        if (!user) return;
        try {
            if (savedIds.has(studentId)) {
                await profileApi.removeSavedItemByTarget('user', studentId);
                setSavedIds(prev => {
                    const next = new Set(prev);
                    next.delete(studentId);
                    return next;
                });
            } else {
                await profileApi.saveItem('user', studentId);
                setSavedIds(prev => new Set(prev).add(studentId));
            }
        } catch (err) {
            console.error("Failed to toggle save student:", err);
        }
    };

    const clearFilters = () => {
        setSearch("");
        setFaculty("");
        setMinHours("");
    };

    return (
        <PageLayout>
            <Box sx={{ mb: 6 }}>
                <Box sx={{ mb: 4 }}>
                    <Typography variant="h3" fontWeight="800" gutterBottom sx={{
                        background: "linear-gradient(45deg, #5de0ff, #0072ff)",
                        WebkitBackgroundClip: "text",
                        WebkitTextFillColor: "transparent"
                    }}>
                        Talent Search
                    </Typography>
                    <Typography variant="body1" color="text.secondary">
                        Discover amazing students and collaborators for your next project.
                    </Typography>
                </Box>

                <Paper sx={{
                    p: 3,
                    mb: 4,
                    borderRadius: 4,
                    background: "rgba(255, 255, 255, 0.03)",
                    border: "1px solid rgba(255, 255, 255, 0.05)",
                    boxShadow: "none"
                }}>
                    <Stack direction={{ xs: 'column', md: 'row' }} spacing={2} alignItems="center">
                        <TextField
                            fullWidth
                            placeholder="Search by name, skills, or field..."
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                            InputProps={{
                                startAdornment: (
                                    <InputAdornment position="start">
                                        <SearchIcon color="primary" />
                                    </InputAdornment>
                                ),
                                endAdornment: search && (
                                    <InputAdornment position="end">
                                        <IconButton size="small" onClick={() => setSearch("")}>
                                            <ClearIcon />
                                        </IconButton>
                                    </InputAdornment>
                                )
                            }}
                            sx={{ '& .MuiOutlinedInput-root': { borderRadius: 3 } }}
                        />
                        <Button
                            variant={showFilters ? "contained" : "outlined"}
                            startIcon={<FilterIcon />}
                            onClick={() => setShowFilters(!showFilters)}
                            sx={{ borderRadius: 3, px: 3, height: 56, minWidth: 150 }}
                        >
                            Filters
                        </Button>
                    </Stack>

                    <Collapse in={showFilters}>
                        <Box sx={{ pt: 3 }}>
                            <Divider sx={{ mb: 3, opacity: 0.1 }} />
                            <Grid container spacing={3}>
                                <Grid item xs={12} md={6}>
                                    <FormControl fullWidth size="small">
                                        <InputLabel>Faculty / School</InputLabel>
                                        <Select
                                            value={faculty}
                                            label="Faculty / School"
                                            onChange={(e) => setFaculty(e.target.value)}
                                            sx={{ borderRadius: 2 }}
                                        >
                                            <MenuItem value="">Any Faculty</MenuItem>
                                            {FACULTIES.map(f => (
                                                <MenuItem key={f} value={f}>{f}</MenuItem>
                                            ))}
                                        </Select>
                                    </FormControl>
                                </Grid>
                                <Grid item xs={12} md={6}>
                                    <FormControl fullWidth size="small">
                                        <InputLabel>Minimum Availability (hrs/week)</InputLabel>
                                        <Select
                                            value={minHours}
                                            label="Minimum Availability (hrs/week)"
                                            onChange={(e) => setMinHours(e.target.value)}
                                            sx={{ borderRadius: 2 }}
                                        >
                                            <MenuItem value="">Any Availability</MenuItem>
                                            <MenuItem value="1">1+ hrs/week</MenuItem>
                                            <MenuItem value="5">5+ hrs/week</MenuItem>
                                            <MenuItem value="10">10+ hrs/week</MenuItem>
                                            <MenuItem value="20">20+ hrs/week</MenuItem>
                                        </Select>
                                    </FormControl>
                                </Grid>
                            </Grid>
                            <Box sx={{ mt: 3, display: 'flex', justifyContent: 'flex-end' }}>
                                <Button size="small" onClick={clearFilters} color="inherit">
                                    Clear all filters
                                </Button>
                            </Box>
                        </Box>
                    </Collapse>
                </Paper>

                <Box sx={{ mb: 3, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Typography variant="body2" color="text.secondary" fontWeight="500">
                        {loading ? "Searching..." : `${students.length} students found`}
                    </Typography>
                </Box>

                {error && (
                    <Alert severity="error" sx={{ mb: 4, borderRadius: 3 }}>{error}</Alert>
                )}

                {loading ? (
                    <Box sx={{ display: 'flex', justifyContent: 'center', py: 10 }}>
                        <CircularProgress />
                    </Box>
                ) : (
                    <Grid container spacing={3}>
                        {students.map((student) => (
                            <Grid item xs={12} sm={6} lg={4} key={student.id}>
                                <TalentCard
                                    student={student}
                                    isSaved={savedIds.has(student.id)}
                                    onSave={handleSaveToggle}
                                />
                            </Grid>
                        ))}
                        {students.length === 0 && !loading && (
                            <Grid item xs={12}>
                                <Box sx={{ textAlign: 'center', py: 10, opacity: 0.5 }}>
                                    <Typography variant="h5">No students match your search criteria</Typography>
                                    <Typography>Try adjusting your keywords or filters</Typography>
                                </Box>
                            </Grid>
                        )}
                    </Grid>
                )}
            </Box>
        </PageLayout>
    );
};

export default Talents;
