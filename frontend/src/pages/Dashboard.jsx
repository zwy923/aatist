import React, { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Box,
  Grid,
  Chip,
  IconButton,
  InputBase,
  Paper,
  Stack,
  Button,
  Typography,
} from "@mui/material";
import SearchIcon from "@mui/icons-material/Search";
import PaletteIcon from "@mui/icons-material/Palette";
import CameraAltIcon from "@mui/icons-material/CameraAlt";
import BarChartIcon from "@mui/icons-material/BarChart";
import BrushIcon from "@mui/icons-material/Brush";
import ComputerIcon from "@mui/icons-material/Computer";
import MovieIcon from "@mui/icons-material/Movie";
import AnimationIcon from "@mui/icons-material/Animation";
import PhoneAndroidIcon from "@mui/icons-material/PhoneAndroid";
import { useAuth } from "../features/auth/hooks/useAuth";
import { portfolioApi } from "../features/profile/api/profile";
import PageLayout from "../shared/components/PageLayout";
import { StateContainer } from "../shared/components/ui/StateContainer";

const CATEGORIES = [
  { id: "logo", label: "Logo Design", icon: PaletteIcon, color: "#e91e63" },
  { id: "photography", label: "Photography", icon: CameraAltIcon, color: "#2196f3" },
  { id: "ppt", label: "PPT Design", icon: BarChartIcon, color: "#ff9800" },
  { id: "illustration", label: "Illustration", icon: BrushIcon, color: "#f44336" },
  { id: "web", label: "Web Design", icon: ComputerIcon, color: "#4caf50" },
  { id: "video", label: "Video Editing", icon: MovieIcon, color: "#9c27b0" },
  { id: "animation", label: "Animation", icon: AnimationIcon, color: "#00bcd4" },
  { id: "uiux", label: "UI/UX Design", icon: PhoneAndroidIcon, color: "#607d8b" },
];

export default function Dashboard() {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();
  const [searchQuery, setSearchQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [recentWork, setRecentWork] = useState([]);
  const hasFetched = useRef(false);

  const loadRecentWork = useCallback(async () => {
    try {
      setLoading(true);
      const response = await portfolioApi.getPublicPortfolios({ limit: 3 });
      const data = response.data.data;
      setRecentWork(data?.projects || data?.items || []);
    } catch (err) {
      console.warn("Failed to load recent work:", err);
      setRecentWork([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!isAuthenticated) {
      navigate("/auth/login");
      return;
    }
    if (!hasFetched.current) {
      loadRecentWork();
      hasFetched.current = true;
    }
  }, [isAuthenticated, navigate, loadRecentWork]);

  const handleSearch = () => {
    if (searchQuery.trim()) {
      navigate(`/opportunities?q=${encodeURIComponent(searchQuery.trim())}`);
    } else {
      navigate("/opportunities");
    }
  };

  const handleBrowseAll = () => navigate("/opportunities");

  const handleCategoryClick = (categoryId) => {
    navigate(`/opportunities?category=${categoryId}`);
  };

  const recentWorkItems =
    recentWork.length > 0
      ? recentWork.map((item) => (
          <Grid item xs={12} md={4} key={item.id}>
            <Paper
              onClick={() => item.id && navigate(`/opportunities/${item.id}`)}
              sx={{
                cursor: item.id ? "pointer" : "default",
                borderRadius: 2,
                overflow: "hidden",
                border: "1px solid #e8e8e8",
                "&:hover": item.id
                  ? { boxShadow: "0 4px 16px rgba(0,0,0,0.1)" }
                  : {},
              }}
            >
              <Box
                sx={{
                  aspectRatio: "4/3",
                  bgcolor: "#1a1a1a",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                }}
              >
                {item.cover_image_url ? (
                  <Box
                    component="img"
                    src={item.cover_image_url}
                    alt={item.title}
                    sx={{ width: "100%", height: "100%", objectFit: "cover" }}
                  />
                ) : (
                  <Typography
                    variant="h4"
                    fontWeight={700}
                    sx={{ color: "rgba(255,255,255,0.15)" }}
                  >
                    {item.title?.[0]?.toUpperCase() || "P"}
                  </Typography>
                )}
              </Box>
              {item.title && (
                <Box sx={{ p: 2 }}>
                  <Typography fontWeight={600} color="#1a1a1a">
                    {item.title}
                  </Typography>
                </Box>
              )}
            </Paper>
          </Grid>
        ))
      : [1, 2, 3].map((i) => (
          <Grid item xs={12} md={4} key={i}>
            <Paper
              sx={{
                aspectRatio: "4/3",
                bgcolor: "#1a1a1a",
                borderRadius: 2,
                border: "1px solid #e8e8e8",
              }}
            />
          </Grid>
        ));

  return (
    <PageLayout maxWidth="xl" variant="light">
      <StateContainer loading={loading} onRetry={loadRecentWork}>
        <Stack spacing={{ xs: 4, md: 5 }} sx={{ color: "#1a1a1a" }}>
          {/* Hero & Search */}
          <Paper
            sx={{
              p: { xs: 2.5, md: 4 },
              borderRadius: 4,
              border: "1px solid #e5e7eb",
              boxShadow: "0 4px 16px rgba(15, 23, 42, 0.06)",
              background:
                "linear-gradient(145deg, rgba(25,118,210,0.06) 0%, rgba(255,255,255,1) 55%)",
            }}
          >
            <Grid container spacing={{ xs: 2.5, md: 4 }} alignItems="center">
              <Grid item xs={12} md={8}>
                <Typography
                  variant="h3"
                  fontWeight={700}
                  sx={{ color: "#1a1a1a", mb: 1.25, fontSize: { xs: "1.7rem", md: "2.3rem" } }}
                >
                  Find Aalto Arts talent for your next project.
                </Typography>
                <Typography variant="body1" sx={{ color: "#64748b", mb: 2.5, maxWidth: 640 }}>
                  Connect with talented Aalto students for design, photography, and creative
                  services.
                </Typography>

                <Paper
                  component="form"
                  onSubmit={(e) => {
                    e.preventDefault();
                    handleSearch();
                  }}
                  sx={{
                    display: "flex",
                    alignItems: "center",
                    borderRadius: 3,
                    border: "1px solid #dbe2ea",
                    boxShadow: "0 1px 2px rgba(0,0,0,0.04)",
                    maxWidth: 620,
                    bgcolor: "#fff",
                    p: 0.5,
                  }}
                >
                  <InputBase
                    placeholder="What service are you looking for?"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    sx={{
                      flex: 1,
                      px: 2,
                      py: 1,
                      fontSize: "0.98rem",
                    }}
                  />
                  <IconButton
                    type="submit"
                    sx={{
                      bgcolor: "#1976d2",
                      color: "white",
                      borderRadius: 2,
                      px: 1.5,
                      minWidth: 44,
                      minHeight: 44,
                      "&:hover": { bgcolor: "#1565c0" },
                    }}
                  >
                    <SearchIcon />
                  </IconButton>
                </Paper>

                <Button
                  onClick={handleBrowseAll}
                  variant="text"
                  sx={{ mt: 1, px: 0, fontWeight: 600, textTransform: "none" }}
                >
                  Browse all services
                </Button>
              </Grid>

              <Grid item xs={12} md={4}>
                <Paper
                  sx={{
                    p: 2.5,
                    borderRadius: 3,
                    border: "1px solid #e5e7eb",
                    background: "#ffffff",
                    height: "100%",
                  }}
                >
                  <Typography variant="subtitle1" fontWeight={700} sx={{ mb: 1.5 }}>
                    Popular categories
                  </Typography>
                  <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
                    {CATEGORIES.slice(0, 6).map((category) => (
                      <Chip
                        key={category.id}
                        label={category.label}
                        onClick={() => handleCategoryClick(category.id)}
                        clickable
                        sx={{
                          bgcolor: `${category.color}14`,
                          color: category.color,
                          fontWeight: 600,
                          borderRadius: 2,
                          "&:hover": { bgcolor: `${category.color}24` },
                        }}
                      />
                    ))}
                  </Stack>
                </Paper>
              </Grid>
            </Grid>
          </Paper>

          {/* Browse by Category */}
          <Box>
            <Stack
              direction={{ xs: "column", sm: "row" }}
              justifyContent="space-between"
              alignItems={{ xs: "flex-start", sm: "center" }}
              sx={{ mb: 2.5 }}
              spacing={1}
            >
              <Box>
                <Typography variant="h5" fontWeight={700} sx={{ color: "#1a1a1a" }}>
                  Browse by Category
                </Typography>
                <Typography variant="body2" sx={{ color: "#64748b" }}>
                  Pick a category to jump into matching opportunities.
                </Typography>
              </Box>
            </Stack>

            <Grid container spacing={{ xs: 1.5, md: 2 }}>
              {CATEGORIES.map(({ id, label, icon: Icon, color }) => (
                <Grid item xs={6} sm={4} md={3} key={id}>
                  <Paper
                    onClick={() => handleCategoryClick(id)}
                    sx={{
                      p: { xs: 2, md: 2.5 },
                      textAlign: "center",
                      cursor: "pointer",
                      border: "1px solid #e8e8e8",
                      borderRadius: 3,
                      bgcolor: "#fff",
                      height: "100%",
                      transition: "all 0.2s ease",
                      "&:hover": {
                        borderColor: "#1976d2",
                        transform: "translateY(-2px)",
                        boxShadow: "0 8px 16px rgba(25, 118, 210, 0.14)",
                      },
                    }}
                  >
                    <Box
                      sx={{
                        width: 52,
                        height: 52,
                        borderRadius: 2,
                        bgcolor: `${color}20`,
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                        mx: "auto",
                        mb: 1.25,
                      }}
                    >
                      <Icon sx={{ fontSize: 26, color }} />
                    </Box>
                    <Typography variant="body2" fontWeight={700} color="#1a1a1a">
                      {label}
                    </Typography>
                  </Paper>
                </Grid>
              ))}
            </Grid>
          </Box>

          {/* Recent Student Work */}
          <Box>
            <Stack
              direction={{ xs: "column", sm: "row" }}
              justifyContent="space-between"
              alignItems={{ xs: "flex-start", sm: "center" }}
              sx={{ mb: 2.5 }}
              spacing={1}
            >
              <Box>
                <Typography variant="h5" fontWeight={700} sx={{ color: "#1a1a1a" }}>
                  Recent Student Work
                </Typography>
                <Typography variant="body2" sx={{ color: "#64748b" }}>
                  Fresh portfolio highlights from talented students.
                </Typography>
              </Box>
            </Stack>

            <Grid container spacing={{ xs: 2, md: 3 }}>
              {recentWorkItems}
            </Grid>
          </Box>
        </Stack>
      </StateContainer>
    </PageLayout>
  );
}