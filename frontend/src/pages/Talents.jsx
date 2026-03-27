import React, { useState, useCallback, useRef, useEffect } from 'react';
import {
    Box,
    Typography,
    Grid,
    Card,
    Avatar,
    Chip,
    Stack,
    TextField,
    Button,
    IconButton,
    CircularProgress,
    Alert,
    Collapse,
    Paper,
    Divider,
    FormControl,
    InputLabel,
    Select,
    MenuItem
} from '@mui/material';
import {
    Search as SearchIcon,
    FilterList as FilterIcon,
    CheckCircle as VerifiedIcon,
    Help as HelpIcon,
    AttachMoney as AttachMoneyIcon,
    Message as MessageIcon,
    KeyboardArrowUp as ArrowUpIcon,
    ArrowForward as ArrowForwardIcon
} from '@mui/icons-material';
import { useNavigate, Link } from 'react-router-dom';
import PageLayout from '../shared/components/PageLayout';
import { profileApi, portfolioApi } from '../features/profile/api/profile';
import { useAuth } from '../features/auth/hooks/useAuth';

import './Talents.css';

const FACULTIES = [
    "Arts, Design and Architecture",
    "Business",
    "Chemical Engineering",
    "Electrical Engineering",
    "Engineering",
    "Science"
];

const SCHOOLS = ["Aalto University", ...FACULTIES, "Other"];

const PROGRAMS = ["Design", "Engineering", "Business", "Arts", "Science", "Other"];

const BUDGET_RANGES = [
    { value: "", label: "Any" },
    { value: "0-500", label: "Under €500" },
    { value: "500-1500", label: "€500 - €1,500" },
    { value: "1500-5000", label: "€1,500 - €5,000" },
    { value: "5000+", label: "€5,000+" }
];

const CATEGORIES = [
    {
        main: "Graphic & Illustration",
        items: ["Graphic & Illustration", "Pitch deck & presentation", "Poster Design", "Social media post design", "Print & Packaging"]
    },
    {
        main: "Branding",
        items: ["LOGO Design", "Brand Design", "LOGO + Brand Design"]
    },
    {
        main: "Web Design",
        items: ["Website Design", "APP Design"]
    },
    {
        main: "Photography",
        items: ["Product Photography", "Team Photography", "Event Photography"]
    },
    {
        main: "Video & Motion",
        items: ["Startup Promo Video", "Animated explainer Video", "Video Editing", "Social Media Short Video", "Event Video Documentation"]
    },
    {
        main: "Creative Styling",
        items: ["Team Outfit Design", "Makeup for photoshoots", "Creative/Artistic Makeup"]
    },
    {
        main: "Exhibition & Spatial Design",
        items: ["Exhibition Design", "Booth Design"]
    }
];

const ALL_SUGGESTIONS = [...new Set(CATEGORIES.flatMap((g) => g.items))];

const highlightMatch = (text, query) => {
    if (!query.trim()) return <span style={{ color: '#6b7280' }}>{text}</span>;
    const q = query.trim().replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    const re = new RegExp(`(${q})`, 'gi');
    const parts = text.split(re);
    return parts.map((part, i) =>
        i % 2 === 1 ? (
            <strong key={i} style={{ fontWeight: 700, color: '#374151' }}>{part}</strong>
        ) : (
            <span key={i} style={{ color: '#6b7280' }}>{part}</span>
        )
    );
};

const formatPrice = (s) => {
    if (s.price_min != null && s.price_max != null) return `€${s.price_min}–€${s.price_max}`;
    if (s.price_min != null) {
        if (s.price_type === 'project') return `€${s.price_min} / project`;
        if (s.price_type === 'hourly') return `€${s.price_min} / hr`;
        return `From €${s.price_min}`;
    }
    if (s.price_type === 'hourly') return '€/hr';
    if (s.price_type === 'project') return '€/project';
    return s.price_type || 'Negotiable';
};

const TalentCard = ({ student }) => {
    const navigate = useNavigate();
    const [profile, setProfile] = useState(null);

    useEffect(() => {
        let cancelled = false;
        const load = async () => {
            try {
                const [profileRes] = await Promise.allSettled([
                    profileApi.getPublicProfile(student.id),
                ]);
                if (cancelled) return;
                if (profileRes.status === 'fulfilled') {
                    setProfile(profileRes.value.data?.data || null);
                }
            } catch {
                if (!cancelled) setProfile(null);
            }
        };
        load();
        return () => { cancelled = true; };
    }, [student.id]);

    const services = profile?.services || [];
    const primaryService = services[0];
    const offers = [...(student.skills || []).map(s => typeof s === 'string' ? s : s?.name).filter(Boolean), ...services.map(s => s.category || s.title).filter(Boolean)];
    const uniqueOffers = [...new Set(offers)];
    const displayedOffers = uniqueOffers.slice(0, 5);
    const moreOffers = uniqueOffers.length - 5;

    const educationLine = [student.school, student.faculty || student.major].filter(Boolean).join(', ') || 'Student';

    return (
        <Card className="talent-result-card" sx={{
            borderRadius: 2,
            background: '#fff',
            border: '1px solid #e5e7eb',
            overflow: 'hidden',
            boxShadow: '0 1px 3px rgba(0,0,0,0.08)'
        }}>
            {/* Left: Profile */}
            <div className="talent-card-left">
                <Avatar
                    src={student.avatar_url}
                    className="talent-card-avatar"
                    sx={{ width: 100, height: 100, borderRadius: 1, bgcolor: '#e0e0e0' }}
                >
                    {student.name?.charAt(0)}
                </Avatar>
                <div className="talent-card-name">
                    {student.name}
                    {(profile?.is_verified_email || profile?.role_verified) && (
                        <VerifiedIcon sx={{ fontSize: 20, color: '#22c55e' }} />
                    )}
                </div>
                <div className="talent-card-education">{educationLine}</div>
                <div className="talent-card-available">
                    <span className="talent-card-available-dot" />
                    Available
                </div>
                <div className="talent-card-buttons">
                    <button
                        type="button"
                        className="talent-card-btn-contact"
                        onClick={() => navigate(`/messages?user=${student.id}`)}
                    >
                        Contact
                    </button>
                    <button
                        type="button"
                        className="talent-card-btn-profile"
                        onClick={() => navigate(`/users/${student.id}`)}
                    >
                        View Profile
                    </button>
                </div>
            </div>

            {/* Right: Offers + Service */}
            <div className="talent-card-right">
                {/* Offers */}
                <div className="talent-card-offers-section">
                    <div className="talent-card-offers-label">
                        Offers <HelpIcon sx={{ fontSize: 16, color: '#999' }} />
                    </div>
                    <div className="talent-card-offers-pills">
                        {displayedOffers.map((o, i) => (
                            <span key={i} className="talent-card-offers-pill">
                                {typeof o === 'string' ? o : o}
                            </span>
                        ))}
                        {moreOffers > 0 && (
                            <span className="talent-card-offers-pill">+{moreOffers}</span>
                        )}
                    </div>
                </div>

                {/* Service */}
                <div
                    className="talent-card-service-section"
                    role={primaryService?.id ? "button" : undefined}
                    tabIndex={primaryService?.id ? 0 : undefined}
                    onClick={() => {
                        if (primaryService?.id) navigate(`/users/${student.id}/services/${primaryService.id}`);
                    }}
                    onKeyDown={(e) => {
                        if (!primaryService?.id) return;
                        if (e.key === "Enter" || e.key === " ") {
                            e.preventDefault();
                            navigate(`/users/${student.id}/services/${primaryService.id}`);
                        }
                    }}
                    style={{ cursor: primaryService?.id ? "pointer" : undefined }}
                >
                    <div className="talent-card-service-label">
                        Service <AttachMoneyIcon sx={{ fontSize: 16, color: '#999' }} />
                    </div>
                    {primaryService ? (
                        <>
                            <div className="talent-card-service-gallery">
                                {(primaryService.media_urls || []).slice(0, 3).map((url, i) => (
                                    <Box key={i} component="img" src={url} alt="" sx={{ width: 80, height: 100, objectFit: 'cover', borderRadius: 1 }} />
                                ))}
                                {(!primaryService.media_urls || primaryService.media_urls.length === 0) && (
                                    <>
                                        <div className="talent-card-service-thumb" />
                                        <div className="talent-card-service-thumb" />
                                        <div className="talent-card-service-thumb" />
                                    </>
                                )}
                            </div>
                            <div className="talent-card-service-title">
                                {primaryService.title || primaryService.category || 'Service'}
                            </div>
                            <div className="talent-card-service-desc">
                                {primaryService.short_description || primaryService.description || primaryService.experience_summary || ''}
                            </div>
                            <div className="talent-card-service-price">
                                {formatPrice(primaryService)}
                            </div>
                        </>
                    ) : (
                        <>
                            <div className="talent-card-service-gallery">
                                <div className="talent-card-service-thumb" />
                                <div className="talent-card-service-thumb" />
                                <div className="talent-card-service-thumb" />
                            </div>
                            <div className="talent-card-service-title">No service listed</div>
                            <div className="talent-card-service-desc">This talent has not added services yet.</div>
                        </>
                    )}
                </div>
            </div>
        </Card>
    );
};

const Talents = () => {
    const [students, setStudents] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const [search, setSearch] = useState("");
    const [faculty, setFaculty] = useState("");
    const [school, setSchool] = useState("");
    const [program, setProgram] = useState("");
    const [service, setService] = useState("");
    const [budget, setBudget] = useState("");
    const [showFilters, setShowFilters] = useState(false);
    const [talentType, setTalentType] = useState('service'); // 'service' | 'student'
    const [hasSearched, setHasSearched] = useState(false);
    const [showSuggestions, setShowSuggestions] = useState(false);
    const [showScrollTop, setShowScrollTop] = useState(false);
    const resultsRef = useRef(null);
    const searchWrapperRef = useRef(null);
    const { user } = useAuth();

    const suggestions = search.trim()
        ? ALL_SUGGESTIONS.filter((s) => s.toLowerCase().includes(search.trim().toLowerCase())).slice(0, 8)
        : [];

    const heroRef = useRef(null);

    useEffect(() => {
        const handleClickOutside = (e) => {
            if (searchWrapperRef.current && !searchWrapperRef.current.contains(e.target)) {
                setShowSuggestions(false);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    useEffect(() => {
        const onScroll = () => setShowScrollTop(window.scrollY > 400);
        window.addEventListener('scroll', onScroll);
        return () => window.removeEventListener('scroll', onScroll);
    }, []);

    useEffect(() => {
        if (showSuggestions && search.trim() && heroRef.current) {
            heroRef.current.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
    }, [showSuggestions, search.trim()]);

    const fetchTalents = useCallback(async (query, facultyFilter, schoolFilter, programFilter, serviceFilter) => {
        setLoading(true);
        setError(null);
        try {
            const qParts = [query].filter(Boolean);
            if (serviceFilter) qParts.push(serviceFilter);
            const params = {
                q: qParts.length ? qParts.join(" ") : undefined,
                faculty: facultyFilter || undefined,
                school: schoolFilter || undefined,
                major: programFilter || undefined,
                limit: 50
            };
            const response = await profileApi.searchUsers(params);
            let list = response.data.data || [];
            const myId = user?.id ?? user?.user_id;
            if (myId) {
                list = list.filter((s) => String(s.id) !== String(myId));
            }
            setStudents(list);
        } catch (err) {
            console.error("Failed to fetch talents:", err);
            const status = err?.response?.status;
            setError(status === 401
                ? "Please log in to search talents."
                : "Failed to load talent list. Please try again.");
        } finally {
            setLoading(false);
        }
    }, [user]);

    const handleSearch = () => {
        setHasSearched(true);
        fetchTalents(search, faculty, school, program, service);
        resultsRef.current?.scrollIntoView({ behavior: 'smooth' });
    };

    const handleCategoryClick = (tag) => {
        setSearch(tag);
        setHasSearched(true);
        fetchTalents(tag, faculty, school, program, service);
        resultsRef.current?.scrollIntoView({ behavior: 'smooth' });
    };

    const applyStudentFilters = (overrides = {}) => {
        const s = overrides.service ?? service;
        const sch = overrides.school ?? school;
        const p = overrides.program ?? program;
        fetchTalents(search, faculty, sch, p, s);
    };

    return (
        <PageLayout noContainer>
            <Box className={showSuggestions && search.trim() ? 'talents-page-compact' : ''} sx={{ background: '#fff', minHeight: 'calc(100vh - 56px)', display: 'flex', flexDirection: 'column' }}>
                {/* Hero - 60vh default, 90vh when suggestions open (9:1 ratio) */}
                <section className={`talents-hero ${showSuggestions && search.trim() ? 'talents-hero-expanded' : ''}`} ref={heroRef}>
                    <div className="talents-hero-content">
                        <h1 className="talents-hero-title">
                            Inside <span className="highlight">A</span>alto Find the right talent for your project
                        </h1>
                        <div className="talents-hero-toggle">
                            <span
                                className={talentType === 'service' ? 'active' : ''}
                                onClick={() => setTalentType('service')}
                            >
                                Service
                            </span>
                            <span>|</span>
                            <span
                                className={talentType === 'student' ? 'active' : ''}
                                onClick={() => setTalentType('student')}
                            >
                                Student
                            </span>
                        </div>
                        <div className="talents-hero-search-wrapper" ref={searchWrapperRef}>
                            {talentType === 'student' ? (
                                <div className="talents-hero-search talents-hero-search-with-filters">
                                    <div className="talents-hero-search-input-row">
                                        <input
                                            type="text"
                                            placeholder="Search by Major, Program, Name..."
                                            value={search}
                                            onChange={(e) => setSearch(e.target.value)}
                                            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
                                            autoComplete="off"
                                        />
                                        <button type="button" className="talents-hero-search-btn" onClick={handleSearch} aria-label="Search">
                                            <SearchIcon sx={{ fontSize: 22 }} />
                                        </button>
                                    </div>
                                    <div className="talents-hero-filters">
                                        <span className="talents-hero-filter-label">Filter:</span>
                                        <select
                                            value={service}
                                            onChange={(e) => { setService(e.target.value); applyStudentFilters({ service: e.target.value }); }}
                                            className="talents-hero-filter-select"
                                        >
                                            <option value="">Service</option>
                                            {ALL_SUGGESTIONS.slice(0, 12).map((s) => (
                                                <option key={s} value={s}>{s}</option>
                                            ))}
                                        </select>
                                        <select
                                            value={budget}
                                            onChange={(e) => setBudget(e.target.value)}
                                            className="talents-hero-filter-select"
                                        >
                                            {BUDGET_RANGES.map((b) => (
                                                <option key={b.value || 'any'} value={b.value}>{b.label}</option>
                                            ))}
                                        </select>
                                        <select
                                            value={school}
                                            onChange={(e) => { setSchool(e.target.value); applyStudentFilters({ school: e.target.value }); }}
                                            className="talents-hero-filter-select"
                                        >
                                            <option value="">School</option>
                                            {SCHOOLS.map((s) => (
                                                <option key={s} value={s}>{s}</option>
                                            ))}
                                        </select>
                                        <select
                                            value={program}
                                            onChange={(e) => { setProgram(e.target.value); applyStudentFilters({ program: e.target.value }); }}
                                            className="talents-hero-filter-select"
                                        >
                                            <option value="">Program</option>
                                            {PROGRAMS.map((p) => (
                                                <option key={p} value={p}>{p}</option>
                                            ))}
                                        </select>
                                    </div>
                                </div>
                            ) : (
                                <div className="talents-hero-search">
                                    <input
                                        type="text"
                                        placeholder="Search for ..."
                                        value={search}
                                        onChange={(e) => {
                                            setSearch(e.target.value);
                                            setShowSuggestions(true);
                                        }}
                                        onFocus={() => search.trim() && setShowSuggestions(true)}
                                        onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
                                        autoComplete="off"
                                    />
                                    <button type="button" className="talents-hero-search-btn" onClick={handleSearch} aria-label="Search">
                                        <SearchIcon sx={{ fontSize: 22 }} />
                                    </button>
                                </div>
                            )}
                            {talentType === 'service' && showSuggestions && search.trim() && (
                                <div className="talents-search-suggestions">
                                    <div className="talents-search-suggestions-header">Suggestions</div>
                                    {suggestions.length > 0 ? (
                                        <ul className="talents-search-suggestions-list">
                                            {suggestions.map((s) => (
                                                <li
                                                    key={s}
                                                    className="talents-search-suggestion-item"
                                                    onClick={() => {
                                                        setSearch(s);
                                                        setShowSuggestions(false);
                                                        setHasSearched(true);
                                                        fetchTalents(s, faculty, school, program, service);
                                                        resultsRef.current?.scrollIntoView({ behavior: 'smooth' });
                                                    }}
                                                >
                                                    {highlightMatch(s, search)}
                                                </li>
                                            ))}
                                        </ul>
                                    ) : (
                                        <div className="talents-search-suggestions-empty">
                                            No matching services
                                        </div>
                                    )}
                                    <div className="talents-search-suggestions-footer">
                                        <p>Can't find what you're looking for? Post a request and let the right talent reach out to you directly.</p>
                                        <Button
                                            component={Link}
                                            to="/opportunities"
                                            className="talents-post-request-btn"
                                            endIcon={<span aria-hidden>→</span>}
                                        >
                                            Post a request
                                        </Button>
                                    </div>
                                </div>
                            )}
                        </div>
                    </div>
                </section>

                {/* Categories - full grid, hidden when suggestions open */}
                <section className={`talents-categories ${!hasSearched ? 'talents-categories-fill' : ''} ${showSuggestions && search.trim() ? 'talents-categories-collapsed' : ''}`}>
                    <div className="talents-categories-grid">
                        {CATEGORIES.map((group) => (
                            <div key={group.main} className="talents-category-group">
                                <span className="talents-category-main">{group.main}</span>
                                <div className="talents-category-items">
                                    {group.items.map((tag) => (
                                        <span
                                            key={tag}
                                            className="talents-category-tag"
                                            onClick={() => handleCategoryClick(tag)}
                                            role="button"
                                            tabIndex={0}
                                            onKeyDown={(e) => e.key === 'Enter' && handleCategoryClick(tag)}
                                        >
                                            {tag}
                                        </span>
                                    ))}
                                </div>
                            </div>
                        ))}
                    </div>
                </section>

                {/* Results section - shown when user has searched */}
                {hasSearched && (
                    <Box
                        ref={resultsRef}
                        className="talents-results-area"
                        sx={{
                            pt: 3,
                            px: { xs: 2, md: 4 },
                            pb: 3,
                        }}
                    >
                        <Box sx={{ maxWidth: 1200, mx: 'auto' }}>
                            {/* Page header: category + result count */}
                            <div className="talents-page-header">
                                <div className="talents-page-header-tab">
                                    {search || 'All Services'}
                                    <span className="talents-page-header-count">{students.length}</span>
                                </div>
                            </div>

                            {talentType !== 'student' && (
                            <Collapse in={showFilters}>
                                <Paper sx={{ p: 3, mb: 4, borderRadius: 3 }}>
                                    <Stack direction={{ xs: 'column', md: 'row' }} spacing={2} alignItems="center">
                                        <TextField
                                            fullWidth
                                            size="small"
                                            placeholder="Search by name, skills..."
                                            value={search}
                                            onChange={(e) => setSearch(e.target.value)}
                                            sx={{ maxWidth: 320 }}
                                        />
                                        <FormControl size="small" sx={{ minWidth: 200 }}>
                                            <InputLabel>Faculty</InputLabel>
                                            <Select
                                                value={faculty}
                                                label="Faculty"
                                                onChange={(e) => setFaculty(e.target.value)}
                                            >
                                                <MenuItem value="">Any Faculty</MenuItem>
                                                {FACULTIES.map(f => (
                                                    <MenuItem key={f} value={f}>{f}</MenuItem>
                                                ))}
                                            </Select>
                                        </FormControl>
                                        <Button variant="contained" onClick={() => fetchTalents(search, faculty, school, program, service)}>
                                            Apply
                                        </Button>
                                    </Stack>
                                </Paper>
                            </Collapse>
                            )}

                            {error && (
                                <Alert severity="error" sx={{ mb: 4 }} action={
                                    error.includes("log in") ? (
                                        <Button component={Link} to="/auth/login" color="inherit" size="small">
                                            Log in
                                        </Button>
                                    ) : null
                                }>
                                    {error}
                                </Alert>
                            )}

                            {loading ? (
                                <Box sx={{ display: 'flex', justifyContent: 'center', py: 10 }}>
                                    <CircularProgress />
                                </Box>
                            ) : (
                                <Stack spacing={3} sx={{ mb: 4 }}>
                                    {students.map((student) => (
                                        <TalentCard
                                            key={student.id}
                                            student={student}
                                        />
                                    ))}
                                    {students.length === 0 && !loading && (
                                        <Box sx={{ textAlign: 'center', py: 10, opacity: 0.6 }}>
                                            <Typography variant="h6">No students match your search</Typography>
                                            <Typography variant="body2">Try different keywords or categories</Typography>
                                        </Box>
                                    )}
                                </Stack>
                            )}
                        </Box>
                        {/* Spacer for fixed bottom banner - ensures last card is not obscured */}
                        <Box sx={{ height: 220, minHeight: 220, flexShrink: 0 }} aria-hidden />
                    </Box>
                )}

                {/* Bottom overlay: Post request banner */}
                {hasSearched && (
                    <>
                        <button
                            type="button"
                            className="talents-upload-float"
                            aria-label="Scroll to top"
                            onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}
                        >
                            <ArrowUpIcon sx={{ fontSize: 28 }} />
                        </button>
                        <Link
                            to="/opportunities"
                            className="talents-post-request-banner"
                            style={{ textDecoration: 'none', color: 'inherit' }}
                        >
                            <div className="talents-post-request-text">
                                <strong>Can&apos;t find what you&apos;re looking for?</strong>
                                <span>Post a request and let the right talent reach out to you directly.</span>
                            </div>
                            <span className="talents-post-request-arrow">
                                <ArrowForwardIcon sx={{ fontSize: 24 }} />
                            </span>
                        </Link>
                    </>
                )}
            </Box>

            {showScrollTop && hasSearched && (
                <IconButton
                    onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}
                    sx={{
                        position: 'fixed',
                        bottom: 24,
                        right: 24,
                        bgcolor: '#22c55e',
                        color: '#fff',
                        '&:hover': { bgcolor: '#16a34a' }
                    }}
                    aria-label="Scroll to top"
                >
                    <ArrowUpIcon />
                </IconButton>
            )}
        </PageLayout>
    );
};

export default Talents;
