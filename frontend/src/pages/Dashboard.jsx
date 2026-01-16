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
import { opportunitiesApi } from "../features/opportunities/api/opportunities";
import { eventsApi } from "../features/events/api/events";
import { portfolioApi } from "../features/profile/api/profile";
import { useAuth } from "../features/auth/hooks/useAuth";
import StatsCard from "../components/dashboard/StatsCard";
import DashboardHeader from "../components/dashboard/DashboardHeader";
import PopularSkillsCard from "../components/dashboard/PopularSkillsCard";
import OpportunitiesOverview from "../components/dashboard/OpportunitiesOverview";
import DiscoverTalentSection from "../components/dashboard/DiscoverTalentSection";
import { aggregateTags } from "../components/dashboard/utils";
import { StateContainer } from "../shared/components/ui/StateContainer";
import PageLayout from "../shared/components/PageLayout";

const NAV_LINKS = [
  { label: "Opportunities", path: "/opportunities" },
  { label: "My Applications", path: "/applications" },
  { label: "Events", path: "/dashboard?section=events" },
];

export default function Dashboard() {
  const navigate = useNavigate();
  const { user, isAuthenticated, logout } = useAuth();

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

      const oppsResponse = await opportunitiesApi.getOpportunities({
        page: 1,
        limit: 12,
        sort: "newest",
      });
      const oppsData = oppsResponse.data.data;
      const oppList = oppsData?.opportunities || oppsData?.items || [];
      setOpportunities(oppList);
      setStats((prev) => ({
        ...prev,
        openOpportunities: oppsData?.pagination?.total || oppList.length,
      }));
      setPopularSkills(aggregateTags(oppList));

      const eventsResponse = await eventsApi.getEvents({
        page: 1,
        limit: 6,
        sort: "upcoming",
      });
      const eventsData = eventsResponse.data.data;
      setStats((prev) => ({
        ...prev,
        upcomingEvents: eventsData?.pagination?.total || eventsData?.events?.length || eventsData?.items?.length || 0,
      }));

      try {
        const portfolioResponse = await portfolioApi.getMyPortfolio();
        const myPortfolio = portfolioResponse.data.data;
        setStats((prev) => ({
          ...prev,
          ongoingProjects: myPortfolio?.projects?.length || myPortfolio?.items?.length || 0,
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

  const hasFetched = React.useRef(false);
  useEffect(() => {
    if (!isAuthenticated) {
      navigate("/auth/login");
      return;
    }
    if (!hasFetched.current) {
      loadDashboardData();
      hasFetched.current = true;
    }
  }, [isAuthenticated, navigate, loadDashboardData]);

  const statsCards = [
    {
      id: "open",
      label: "Open Opportunities",
      value: stats.openOpportunities,
      icon: <WorkOutlineIcon sx={{ color: "#5de0ff", fontSize: 32 }} />,
      accent: { bg: "rgba(93, 224, 255, 0.1)" },
    },
    {
      id: "projects",
      label: "Ongoing Projects",
      value: stats.ongoingProjects,
      icon: <AssignmentIcon sx={{ color: "#ffb877", fontSize: 32 }} />,
      accent: { bg: "rgba(255, 184, 119, 0.1)" },
    },
    {
      id: "events",
      label: "Upcoming Events",
      value: stats.upcomingEvents,
      icon: <EventIcon sx={{ color: "#5de0ff", fontSize: 32 }} />,
      accent: { bg: "rgba(93, 224, 255, 0.1)" },
    },
  ];

  return (
    <PageLayout maxWidth="xl">
      <StateContainer loading={loading} onRetry={loadDashboardData}>
        <Stack spacing={4}>
          <Box>
            <Typography variant="h3" fontWeight={700} sx={{ color: "text.primary", mb: 1 }}>
              Welcome to Your Talent Hub
            </Typography>
            <Typography variant="h6" color="text.secondary">
              Discover opportunities and showcase your skills on the Aalto Talent Network
            </Typography>
          </Box>

          <Grid container spacing={3}>
            {statsCards.map(({ id, ...cardProps }) => (
              <Grid item xs={12} sm={6} md={4} key={id}>
                <StatsCard {...cardProps} />
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
      </StateContainer>
    </PageLayout>
  );
}

