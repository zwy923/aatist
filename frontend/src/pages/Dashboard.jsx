import React, { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Box,
  Chip,
  CircularProgress,
  Container,
  Grid,
  Stack,
  Typography,
} from "@mui/material";
import AssignmentIcon from "@mui/icons-material/Assignment";
import EventIcon from "@mui/icons-material/Event";
import WorkOutlineIcon from "@mui/icons-material/WorkOutline";
import { eventsAPI, opportunitiesAPI, portfolioAPI } from "../services/api";
import { useUser } from "../store/userStore.jsx";
import StatsCard from "../components/dashboard/StatsCard";
import DashboardHeader from "../components/dashboard/DashboardHeader";
import PopularSkillsCard from "../components/dashboard/PopularSkillsCard";
import OpportunitiesOverview from "../components/dashboard/OpportunitiesOverview";
import DiscoverTalentSection from "../components/dashboard/DiscoverTalentSection";
import { aggregateTags } from "../components/dashboard/utils";

const NAV_LINKS = [
  { label: "Opportunities", path: "/dashboard?section=opportunities" },
  { label: "Events", path: "/dashboard?section=events" },
  { label: "Community", path: "/dashboard?section=community" },
];

export default function Dashboard() {
  const navigate = useNavigate();
  const { user, isAuthenticated, logout } = useUser();

  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState({
    openOpportunities: 0,
    ongoingProjects: 0,
    upcomingEvents: 0,
  });
  const [opportunities, setOpportunities] = useState([]);
  const [popularSkills, setPopularSkills] = useState([]);
  const [studentPortfolios, setStudentPortfolios] = useState([]);
  const [menuAnchorEl, setMenuAnchorEl] = useState(null);

  const isMenuOpen = Boolean(menuAnchorEl);
  const userDisplayName = useMemo(
    () => user?.name || user?.username || user?.email || "User",
    [user]
  );
  const isStudentRole = useMemo(
    () => user?.role?.toLowerCase?.() === "student",
    [user?.role]
  );
  const isVerifiedStudent = isStudentRole && Boolean(user?.is_verified_email);

  const loadDashboardData = useCallback(async () => {
    try {
      setLoading(true);

      const oppsData = await opportunitiesAPI.getOpportunities({
        page: 1,
        limit: 12,
        sort: "newest",
      });
      const oppList = oppsData.opportunities || [];
      setOpportunities(oppList);
      setStats((prev) => ({
        ...prev,
        openOpportunities: oppsData.pagination?.total || oppList.length,
      }));
      setPopularSkills(aggregateTags(oppList));

      const eventsData = await eventsAPI.getEvents({
        page: 1,
        limit: 6,
        sort: "upcoming",
      });
      setStats((prev) => ({
        ...prev,
        upcomingEvents: eventsData.pagination?.total || eventsData.events?.length || 0,
      }));

      try {
        const myPortfolio = await portfolioAPI.getMyPortfolio();
        setStats((prev) => ({
          ...prev,
          ongoingProjects: myPortfolio.projects?.length || 0,
        }));
      } catch (portfolioErr) {
        console.warn("Portfolio not available:", portfolioErr);
      }

      // @todo Replace placeholder when backend exposes public portfolios endpoint
      setStudentPortfolios([]);
    } catch (err) {
      console.error("Failed to load dashboard data:", err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!isAuthenticated) {
      navigate("/auth/login");
      return;
    }
    loadDashboardData();
  }, [isAuthenticated, navigate, loadDashboardData]);

  const handleMenuOpen = (event) => setMenuAnchorEl(event.currentTarget);
  const handleMenuClose = () => setMenuAnchorEl(null);
  const handleLogout = async () => {
    await logout();
    handleMenuClose();
    navigate("/auth/login");
  };
  const handleNavClick = useCallback(
    (path) => navigate(path),
    [navigate]
  );

  const verificationChip = isStudentRole ? (
    <Chip
      size="small"
      label={isVerifiedStudent ? "Student verified" : "Verification pending"}
      color={isVerifiedStudent ? "success" : "warning"}
      variant="outlined"
      sx={{ fontSize: "0.75rem" }}
    />
  ) : (
    <Chip
      size="small"
      label="Organization workspace"
      color="info"
      variant="outlined"
      sx={{ fontSize: "0.75rem" }}
    />
  );

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

  const statsCards = [
    {
      key: "open",
      label: "Open Opportunities",
      value: stats.openOpportunities,
      icon: <WorkOutlineIcon sx={{ color: "#5de0ff", fontSize: 32 }} />,
      accent: { bg: "rgba(93, 224, 255, 0.1)" },
    },
    {
      key: "projects",
      label: "Ongoing Projects",
      value: stats.ongoingProjects,
      icon: <AssignmentIcon sx={{ color: "#ffb877", fontSize: 32 }} />,
      accent: { bg: "rgba(255, 184, 119, 0.1)" },
    },
    {
      key: "events",
      label: "Upcoming Events",
      value: stats.upcomingEvents,
      icon: <EventIcon sx={{ color: "#5de0ff", fontSize: 32 }} />,
      accent: { bg: "rgba(93, 224, 255, 0.1)" },
    },
  ];

  return (
    <Box
      sx={{
        minHeight: "100vh",
        background: "radial-gradient(ellipse at top left, #101820, #050505)",
        py: 4,
      }}
    >
      <Container maxWidth="xl">
        <Stack spacing={4}>
          <DashboardHeader
            navItems={NAV_LINKS}
            onNavClick={handleNavClick}
            verificationChip={verificationChip}
            menuAnchorEl={menuAnchorEl}
            isMenuOpen={isMenuOpen}
            onMenuOpen={handleMenuOpen}
            onMenuClose={handleMenuClose}
            onLogout={handleLogout}
            userDisplayName={userDisplayName}
            userEmail={user?.email}
            isStudentRole={isStudentRole}
            onNavigate={(path) => navigate(path)}
          />

          <Box>
            <Typography variant="h3" fontWeight={700} sx={{ color: "text.primary", mb: 1 }}>
              Welcome to Your Talent Hub
            </Typography>
            <Typography variant="h6" color="text.secondary">
              Discover opportunities and showcase your skills on the Aalto Talent Network
            </Typography>
          </Box>

          <Grid container spacing={3}>
            {statsCards.map((card) => (
              <Grid item xs={12} sm={6} md={4} key={card.key}>
                <StatsCard {...card} />
              </Grid>
            ))}
          </Grid>

          <Grid container spacing={3}>
            <Grid item xs={12} md={6}>
              <PopularSkillsCard skills={popularSkills} />
            </Grid>
            <Grid item xs={12} md={6}>
              <OpportunitiesOverview
                opportunities={opportunities}
                onSelect={(id) => navigate(`/opportunities/${id}`)}
              />
            </Grid>
          </Grid>

          <DiscoverTalentSection
            portfolios={studentPortfolios}
            onSelectUser={(userId) => navigate(`/users/${userId}`)}
          />
        </Stack>
      </Container>
    </Box>
  );
}

