import React, { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import {
    Avatar,
    Box,
    Button,
    Card,
    CardContent,
    CardMedia,
    Chip,
    Container,
    Grid,
    IconButton,
    Paper,
    Stack,
    Typography,
} from "@mui/material";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import EmailIcon from "@mui/icons-material/Email";
import MessageIcon from "@mui/icons-material/Message";
import SchoolIcon from "@mui/icons-material/School";
import useAuthStore from "../shared/stores/authStore";
import { profileApi, portfolioApi } from "../features/profile/api/profile";
import PageLayout from "../shared/components/PageLayout";
import { StateContainer } from "../shared/components/ui/StateContainer";

export default function PublicProfile() {
    const { id } = useParams();
    const navigate = useNavigate();
    const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
    const user = useAuthStore((s) => s.user);
    const myId = user?.id ?? user?.user_id;
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [profile, setProfile] = useState(null);
    const [portfolio, setPortfolio] = useState([]);

    useEffect(() => {
        const fetchData = async () => {
            try {
                setLoading(true);
                setError(null);

                // Fetch user profile and portfolio in parallel
                const [profileRes, portfolioRes] = await Promise.allSettled([
                    profileApi.getPublicProfile(id),
                    portfolioApi.getUserPortfolio(id),
                ]);

                if (profileRes.status === "fulfilled") {
                    setProfile(profileRes.value.data.data);
                } else {
                    throw new Error("Failed to load user profile");
                }

                if (portfolioRes.status === "fulfilled") {
                    const data = portfolioRes.value.data.data;
                    setPortfolio(data?.projects || data?.items || []);
                }
            } catch (err) {
                console.error("Public profile load error:", err);
                setError("Could not load user profile");
            } finally {
                setLoading(false);
            }
        };

        if (id) {
            fetchData();
        }
    }, [id]);

    return (
        <PageLayout>
            <StateContainer loading={loading} error={error}>
                <Container maxWidth="lg">
                    <IconButton
                        onClick={() => navigate(-1)}
                        sx={{ mb: 3, color: "text.secondary" }}
                    >
                        <ArrowBackIcon />
                    </IconButton>

                    {/* User Header Card */}
                    <Paper
                        sx={{
                            p: 4,
                            borderRadius: 4,
                            background: "rgba(255, 255, 255, 0.03)",
                            border: "1px solid rgba(255, 255, 255, 0.05)",
                            mb: 4,
                        }}
                    >
                        <Stack direction={{ xs: "column", md: "row" }} spacing={4} alignItems="center">
                            <Avatar
                                src={profile?.avatar_url}
                                sx={{ width: 120, height: 120, fontSize: "3rem" }}
                            >
                                {profile?.name?.charAt(0)}
                            </Avatar>
                            <Box sx={{ flex: 1, textAlign: { xs: "center", md: "left" } }}>
                                <Typography variant="h3" fontWeight={700} gutterBottom>
                                    {profile?.name}
                                </Typography>
                                <Typography variant="h6" color="primary" gutterBottom>
                                    {profile?.faculty || profile?.major || "Student"}
                                </Typography>
                                <Typography variant="body1" color="text.secondary" sx={{ maxWidth: 800 }}>
                                    {profile?.bio || "No bio available."}
                                </Typography>

                                <Stack
                                    direction="row"
                                    spacing={2}
                                    mt={3}
                                    alignItems="center"
                                    flexWrap="wrap"
                                    justifyContent={{ xs: "center", md: "flex-start" }}
                                >
                                    {isAuthenticated && id && Number(id) === Number(myId) && (
                                        <Button
                                            variant="contained"
                                            onClick={() => navigate("/profile")}
                                            sx={{ textTransform: "none", fontWeight: 600 }}
                                        >
                                            Edit Profile
                                        </Button>
                                    )}
                                    {isAuthenticated && id && Number(id) !== Number(myId) && (
                                        <Button
                                            variant="contained"
                                            startIcon={<MessageIcon />}
                                            onClick={() => navigate(`/messages?user=${id}`)}
                                            sx={{ textTransform: "none", fontWeight: 600 }}
                                        >
                                            Message
                                        </Button>
                                    )}
                                    {profile?.email && (
                                        <Chip icon={<EmailIcon />} label={profile.email} variant="outlined" />
                                    )}
                                    {profile?.school && (
                                        <Chip icon={<SchoolIcon />} label={profile.school} variant="outlined" />
                                    )}
                                </Stack>
                            </Box>
                        </Stack>
                    </Paper>

                    {/* Skills Section */}
                    {profile?.skills && profile.skills.length > 0 && (
                        <Box sx={{ mb: 6 }}>
                            <Typography variant="h5" fontWeight={700} gutterBottom sx={{ mb: 2 }}>
                                Skills
                            </Typography>
                            <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                                {profile.skills.map((skill, index) => (
                                    <Chip
                                        key={index}
                                        label={typeof skill === 'string' ? skill : skill.name}
                                        sx={{
                                            bgcolor: "rgba(93, 224, 255, 0.1)",
                                            color: "#5de0ff",
                                            fontWeight: 600
                                        }}
                                    />
                                ))}
                            </Stack>
                        </Box>
                    )}

                    {/* Services Section */}
                    {profile?.services && profile.services.length > 0 && (
                        <Box sx={{ mb: 6 }}>
                            <Typography variant="h5" fontWeight={700} gutterBottom sx={{ mb: 2 }}>
                                Services
                            </Typography>
                            <Stack spacing={2}>
                                {profile.services.map((s) => (
                                    <Paper
                                        key={s.id}
                                        sx={{
                                            p: 2.5,
                                            borderRadius: 2,
                                            background: "rgba(255, 255, 255, 0.03)",
                                            border: "1px solid rgba(255, 255, 255, 0.05)",
                                        }}
                                    >
                                        <Stack spacing={1}>
                                            <Typography variant="h6" fontWeight={600}>
                                                {s.title || s.category}
                                            </Typography>
                                            {s.short_description && (
                                                <Typography variant="body2" color="text.secondary">
                                                    {s.short_description}
                                                </Typography>
                                            )}
                                            {s.description && (
                                                <Typography variant="body2" sx={{
                                                    display: '-webkit-box',
                                                    WebkitLineClamp: 3,
                                                    WebkitBoxOrient: 'vertical',
                                                    overflow: 'hidden',
                                                }}>
                                                    {s.description}
                                                </Typography>
                                            )}
                                            <Stack direction="row" spacing={1} flexWrap="wrap" alignItems="center">
                                                {s.price_type && (
                                                    <Chip
                                                        size="small"
                                                        label={s.price_type}
                                                        variant="outlined"
                                                        sx={{ textTransform: 'capitalize' }}
                                                    />
                                                )}
                                                {s.price_min != null && s.price_max != null && (
                                                    <Typography variant="caption" color="text.secondary">
                                                        €{s.price_min}–€{s.price_max}
                                                    </Typography>
                                                )}
                                                {s.price_min != null && s.price_max == null && (
                                                    <Typography variant="caption" color="text.secondary">
                                                        From €{s.price_min}
                                                    </Typography>
                                                )}
                                            </Stack>
                                        </Stack>
                                    </Paper>
                                ))}
                            </Stack>
                        </Box>
                    )}

                    {/* Portfolio Section */}
                    <Box>
                        <Typography variant="h5" fontWeight={700} gutterBottom sx={{ mb: 3 }}>
                            Portfolio
                        </Typography>

                        {portfolio.length > 0 ? (
                            <Grid container spacing={3}>
                                {portfolio.map((project) => (
                                    <Grid item xs={12} sm={6} md={4} key={project.id}>
                                        <Card
                                            sx={{
                                                height: "100%",
                                                display: "flex",
                                                flexDirection: "column",
                                                background: "rgba(255, 255, 255, 0.03)",
                                                border: "1px solid rgba(255, 255, 255, 0.05)",
                                                borderRadius: 3,
                                                transition: "transform 0.2s",
                                                "&:hover": {
                                                    transform: "translateY(-4px)",
                                                    borderColor: "primary.main",
                                                },
                                            }}
                                        >
                                            {project.cover_image_url && (
                                                <CardMedia
                                                    component="img"
                                                    height="200"
                                                    image={project.cover_image_url}
                                                    alt={project.title}
                                                    sx={{ objectFit: "cover" }}
                                                />
                                            )}
                                            <CardContent sx={{ flex: 1 }}>
                                                <Typography variant="h6" fontWeight={700} gutterBottom>
                                                    {project.title}
                                                </Typography>
                                                <Stack direction="row" spacing={1} mb={2} overflow="hidden">
                                                    {project.tags?.slice(0, 3).map((tag, idx) => (
                                                        <Typography key={idx} variant="caption" color="primary">
                                                            #{tag}
                                                        </Typography>
                                                    ))}
                                                </Stack>
                                                <Typography variant="body2" color="text.secondary" sx={{
                                                    display: '-webkit-box',
                                                    WebkitLineClamp: 3,
                                                    WebkitBoxOrient: 'vertical',
                                                    overflow: 'hidden'
                                                }}>
                                                    {project.description}
                                                </Typography>
                                            </CardContent>
                                        </Card>
                                    </Grid>
                                ))}
                            </Grid>
                        ) : (
                            <Box sx={{ py: 4, textAlign: 'center', opacity: 0.5 }}>
                                <Typography>No portfolio projects to display.</Typography>
                            </Box>
                        )}
                    </Box>
                </Container>
            </StateContainer>
        </PageLayout>
    );
}
