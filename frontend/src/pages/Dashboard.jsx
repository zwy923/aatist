import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Box,
  Container,
  Typography,
  Grid,
  Card,
  CardContent,
  CardMedia,
  Stack,
  Chip,
  CircularProgress,
  Avatar,
  Paper,
} from "@mui/material";
import {
  WorkOutline,
  Assignment,
  Event,
  TrendingUp,
} from "@mui/icons-material";
import { opportunitiesAPI, eventsAPI, portfolioAPI } from "../services/api";
import { useUser } from "../store/userStore.jsx";

export default function Dashboard() {
  const navigate = useNavigate();
  const { user, isAuthenticated } = useUser();
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState({
    openOpportunities: 0,
    ongoingProjects: 0,
    upcomingEvents: 0,
  });
  const [opportunities, setOpportunities] = useState([]);
  const [popularSkills, setPopularSkills] = useState([]);
  const [studentPortfolios, setStudentPortfolios] = useState([]);

  useEffect(() => {
    if (!isAuthenticated) {
      navigate("/auth/login");
      return;
    }

    loadDashboardData();
  }, [isAuthenticated, navigate]);

  const loadDashboardData = async () => {
    try {
      setLoading(true);

      // Load opportunities
      const oppsData = await opportunitiesAPI.getOpportunities({
        page: 1,
        limit: 10,
        sort: "newest",
      });
      setOpportunities(oppsData.opportunities || []);
      setStats((prev) => ({
        ...prev,
        openOpportunities: oppsData.pagination?.total || 0,
      }));

      // Load events
      const eventsData = await eventsAPI.getEvents({
        page: 1,
        limit: 10,
        sort: "upcoming",
      });
      setStats((prev) => ({
        ...prev,
        upcomingEvents: eventsData.pagination?.total || 0,
      }));

      // Load my portfolio for ongoing projects count
      try {
        const myPortfolio = await portfolioAPI.getMyPortfolio();
        setStats((prev) => ({
          ...prev,
          ongoingProjects: myPortfolio.projects?.length || 0,
        }));
      } catch (err) {
        console.error("Failed to load portfolio", err);
      }

      // TODO: Load popular skills from API when available
      // For now, using mock data
      setPopularSkills([
        { skill: "React", count: 45 },
        { skill: "Python", count: 38 },
        { skill: "UI/UX Design", count: 32 },
        { skill: "Node.js", count: 28 },
        { skill: "JavaScript", count: 25 },
      ]);

      // TODO: Load student portfolios from API when available
      // For now, using mock data structure
      setStudentPortfolios([]);
    } catch (err) {
      console.error("Failed to load dashboard data", err);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <Box
        sx={{
          minHeight: "100vh",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          background: "radial-gradient(ellipse at top left, #101820, #050505)",
        }}
      >
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box
      sx={{
        minHeight: "100vh",
        background: "radial-gradient(ellipse at top left, #101820, #050505)",
        paddingTop: 4,
        paddingBottom: 4,
      }}
    >
      <Container maxWidth="xl">
        <Stack spacing={4}>
          {/* Welcome Section */}
          <Box>
            <Typography
              variant="h3"
              fontWeight={700}
              sx={{ color: "text.primary", mb: 1 }}
            >
              Welcome to Your Talent Hub
            </Typography>
            <Typography variant="h6" color="text.secondary">
              Discover opportunities and showcase your skills on the Aalto Talent
              Network
            </Typography>
          </Box>

          {/* Stats Cards */}
          <Grid container spacing={3}>
            <Grid item xs={12} sm={6} md={4}>
              <Card
                sx={{
                  background: "rgba(7, 12, 30, 0.96)",
                  border: "1px solid rgba(93, 224, 255, 0.25)",
                  borderRadius: 3,
                }}
              >
                <CardContent>
                  <Stack direction="row" spacing={2} alignItems="center">
                    <Box
                      sx={{
                        p: 1.5,
                        borderRadius: 2,
                        background: "rgba(93, 224, 255, 0.1)",
                      }}
                    >
                      <WorkOutline sx={{ color: "#5de0ff", fontSize: 32 }} />
                    </Box>
                    <Box>
                      <Typography variant="body2" color="text.secondary">
                        Open Opportunities
                      </Typography>
                      <Typography variant="h4" fontWeight={700} sx={{ color: "text.primary" }}>
                        {stats.openOpportunities}
                      </Typography>
                    </Box>
                  </Stack>
                </CardContent>
              </Card>
            </Grid>

            <Grid item xs={12} sm={6} md={4}>
              <Card
                sx={{
                  background: "rgba(7, 12, 30, 0.96)",
                  border: "1px solid rgba(93, 224, 255, 0.25)",
                  borderRadius: 3,
                }}
              >
                <CardContent>
                  <Stack direction="row" spacing={2} alignItems="center">
                    <Box
                      sx={{
                        p: 1.5,
                        borderRadius: 2,
                        background: "rgba(255, 184, 119, 0.1)",
                      }}
                    >
                      <Assignment sx={{ color: "#ffb877", fontSize: 32 }} />
                    </Box>
                    <Box>
                      <Typography variant="body2" color="text.secondary">
                        Ongoing Projects
                      </Typography>
                      <Typography variant="h4" fontWeight={700} sx={{ color: "text.primary" }}>
                        {stats.ongoingProjects}
                      </Typography>
                    </Box>
                  </Stack>
                </CardContent>
              </Card>
            </Grid>

            <Grid item xs={12} sm={6} md={4}>
              <Card
                sx={{
                  background: "rgba(7, 12, 30, 0.96)",
                  border: "1px solid rgba(93, 224, 255, 0.25)",
                  borderRadius: 3,
                }}
              >
                <CardContent>
                  <Stack direction="row" spacing={2} alignItems="center">
                    <Box
                      sx={{
                        p: 1.5,
                        borderRadius: 2,
                        background: "rgba(93, 224, 255, 0.1)",
                      }}
                    >
                      <Event sx={{ color: "#5de0ff", fontSize: 32 }} />
                    </Box>
                    <Box>
                      <Typography variant="body2" color="text.secondary">
                        Upcoming Events
                      </Typography>
                      <Typography variant="h4" fontWeight={700} sx={{ color: "text.primary" }}>
                        {stats.upcomingEvents}
                      </Typography>
                    </Box>
                  </Stack>
                </CardContent>
              </Card>
            </Grid>
          </Grid>

          {/* Popular Skills & Opportunities Overview */}
          <Grid container spacing={3}>
            {/* Popular Skills */}
            <Grid item xs={12} md={6}>
              <Paper
                sx={{
                  p: 3,
                  background: "rgba(7, 12, 30, 0.96)",
                  border: "1px solid rgba(93, 224, 255, 0.25)",
                  borderRadius: 3,
                }}
              >
                <Stack direction="row" spacing={1} alignItems="center" mb={2}>
                  <TrendingUp sx={{ color: "#5de0ff" }} />
                  <Typography variant="h6" fontWeight={600}>
                    Popular Skills
                  </Typography>
                </Stack>
                <Stack spacing={1.5}>
                  {popularSkills.map((item, index) => (
                    <Box key={index}>
                      <Stack
                        direction="row"
                        justifyContent="space-between"
                        alignItems="center"
                        mb={0.5}
                      >
                        <Typography variant="body1" sx={{ color: "text.primary" }}>
                          {item.skill}
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                          {item.count} students
                        </Typography>
                      </Stack>
                      <Box
                        sx={{
                          height: 6,
                          borderRadius: 3,
                          background: "rgba(93, 224, 255, 0.1)",
                          overflow: "hidden",
                        }}
                      >
                        <Box
                          sx={{
                            height: "100%",
                            width: `${(item.count / 50) * 100}%`,
                            background: "linear-gradient(90deg, #5de0ff, #7f5dff)",
                            borderRadius: 3,
                          }}
                        />
                      </Box>
                    </Box>
                  ))}
                </Stack>
              </Paper>
            </Grid>

            {/* Opportunities Overview */}
            <Grid item xs={12} md={6}>
              <Paper
                sx={{
                  p: 3,
                  background: "rgba(7, 12, 30, 0.96)",
                  border: "1px solid rgba(93, 224, 255, 0.25)",
                  borderRadius: 3,
                }}
              >
                <Typography variant="h6" fontWeight={600} mb={2}>
                  Opportunities Overview
                </Typography>
                <Stack spacing={2}>
                  {opportunities.slice(0, 5).map((opp) => (
                    <Box
                      key={opp.id}
                      sx={{
                        p: 2,
                        borderRadius: 2,
                        background: "rgba(93, 224, 255, 0.05)",
                        border: "1px solid rgba(93, 224, 255, 0.1)",
                        cursor: "pointer",
                        "&:hover": {
                          background: "rgba(93, 224, 255, 0.1)",
                          borderColor: "rgba(93, 224, 255, 0.3)",
                        },
                      }}
                      onClick={() => navigate(`/opportunities/${opp.id}`)}
                    >
                      <Stack direction="row" justifyContent="space-between" alignItems="start">
                        <Box sx={{ flex: 1 }}>
                          <Typography
                            variant="subtitle1"
                            fontWeight={600}
                            sx={{ color: "text.primary", mb: 0.5 }}
                          >
                            {opp.title}
                          </Typography>
                          <Typography variant="body2" color="text.secondary" noWrap>
                            {opp.description || opp.company || ""}
                          </Typography>
                        </Box>
                        <Chip
                          label={opp.type}
                          size="small"
                          sx={{
                            ml: 2,
                            background: "rgba(93, 224, 255, 0.2)",
                            color: "#5de0ff",
                          }}
                        />
                      </Stack>
                    </Box>
                  ))}
                  {opportunities.length === 0 && (
                    <Typography variant="body2" color="text.secondary" textAlign="center" py={2}>
                      No opportunities available
                    </Typography>
                  )}
                </Stack>
              </Paper>
            </Grid>
          </Grid>

          {/* Discover Talent Section */}
          <Box>
            <Typography variant="h5" fontWeight={700} mb={1} sx={{ color: "text.primary" }}>
              Discover Talent
            </Typography>
            <Typography variant="body1" color="text.secondary" mb={3}>
              Explore portfolios from talented students across design, engineering,
              business, and arts
            </Typography>

            {studentPortfolios.length > 0 ? (
              <Grid container spacing={3}>
                {studentPortfolios.map((portfolio) => (
                  <Grid item xs={12} sm={6} md={4} lg={3} key={portfolio.id}>
                    <Card
                      sx={{
                        background: "rgba(7, 12, 30, 0.96)",
                        border: "1px solid rgba(93, 224, 255, 0.25)",
                        borderRadius: 3,
                        cursor: "pointer",
                        transition: "all 0.2s",
                        "&:hover": {
                          transform: "translateY(-4px)",
                          borderColor: "rgba(93, 224, 255, 0.5)",
                          boxShadow: "0 8px 24px rgba(93, 224, 255, 0.2)",
                        },
                      }}
                      onClick={() => navigate(`/users/${portfolio.user_id}`)}
                    >
                      {portfolio.image_url ? (
                        <CardMedia
                          component="img"
                          height="200"
                          image={portfolio.image_url}
                          alt={portfolio.title}
                          sx={{ objectFit: "cover" }}
                        />
                      ) : (
                        <Box
                          sx={{
                            height: 200,
                            background:
                              "linear-gradient(135deg, rgba(93, 224, 255, 0.2), rgba(127, 93, 255, 0.2))",
                            display: "flex",
                            alignItems: "center",
                            justifyContent: "center",
                          }}
                        >
                          <Typography variant="h6" color="text.secondary">
                            {portfolio.title?.[0] || "P"}
                          </Typography>
                        </Box>
                      )}
                      <CardContent>
                        <Typography
                          variant="subtitle1"
                          fontWeight={600}
                          sx={{ color: "text.primary", mb: 0.5 }}
                          noWrap
                        >
                          {portfolio.title}
                        </Typography>
                        <Typography
                          variant="body2"
                          color="text.secondary"
                          sx={{
                            display: "-webkit-box",
                            WebkitLineClamp: 2,
                            WebkitBoxOrient: "vertical",
                            overflow: "hidden",
                          }}
                        >
                          {portfolio.description}
                        </Typography>
                        {portfolio.technologies && portfolio.technologies.length > 0 && (
                          <Stack direction="row" spacing={0.5} mt={1} flexWrap="wrap">
                            {portfolio.technologies.slice(0, 3).map((tech, idx) => (
                              <Chip
                                key={idx}
                                label={tech}
                                size="small"
                                sx={{
                                  background: "rgba(93, 224, 255, 0.1)",
                                  color: "#5de0ff",
                                  fontSize: "0.7rem",
                                }}
                              />
                            ))}
                          </Stack>
                        )}
                      </CardContent>
                    </Card>
                  </Grid>
                ))}
              </Grid>
            ) : (
              <Paper
                sx={{
                  p: 4,
                  textAlign: "center",
                  background: "rgba(7, 12, 30, 0.96)",
                  border: "1px solid rgba(93, 224, 255, 0.25)",
                  borderRadius: 3,
                }}
              >
                <Typography variant="body1" color="text.secondary">
                  No portfolios available yet. Check back soon!
                </Typography>
              </Paper>
            )}
          </Box>
        </Stack>
      </Container>
    </Box>
  );
}

