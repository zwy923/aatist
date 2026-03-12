import React, { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import {
  Box,
  Button,
  Card,
  CardContent,
  Checkbox,
  Grid,
  InputBase,
  Paper,
  Stack,
  Typography,
  IconButton,
} from "@mui/material";
import SearchIcon from "@mui/icons-material/Search";
import AddIcon from "@mui/icons-material/Add";
import PageLayout from "../shared/components/PageLayout";
import { StateContainer } from "../shared/components/ui/StateContainer";
import { opportunitiesApi } from "../features/opportunities/api/opportunities";
import OpportunityCard from "../features/opportunities/components/OpportunityCard";
import PostProjectBriefDialog from "../features/opportunities/components/PostProjectBriefDialog";
import { useAuth } from "../features/auth/hooks/useAuth";

const CATEGORY_TREE = [
  {
    title: "Graphics & Design",
    items: ["Logo & Brand Identity", "Illustration & Drawing", "Print Design", "Presentation Design", "Infographics"],
  },
  {
    title: "Website & Digital Design",
    items: ["Web Design", "UI/UX Design", "Digital Product"],
  },
  {
    title: "Video & Animation",
    items: ["Video Editing", "Animation", "Motion Graphics", "Explainer Videos"],
  },
  {
    title: "Photography",
    items: ["Event Photography", "Portrait Photography", "Product Photography", "Food Photography"],
  },
];

const isClientRole = (role) => {
  const r = (role || "").toLowerCase();
  return r === "org_person" || r === "org_team";
};

export default function OpportunitiesPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [opportunities, setOpportunities] = useState([]);
  const [query, setQuery] = useState(searchParams.get("q") || "");
  const [postDialogOpen, setPostDialogOpen] = useState(false);
  const { user, isAuthenticated } = useAuth();

  const activeCategory = searchParams.get("category") || "";
  const canPostOpportunity = isAuthenticated && isClientRole(user?.role);

  const loadOpportunities = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const params = {
        page: 1,
        limit: 50,
        status: "active",
      };
      if (activeCategory) params.category = activeCategory;
      const response = await opportunitiesApi.getOpportunities(params);
      const data = response.data?.data;
      const list = data?.data || data || [];
      setOpportunities(Array.isArray(list) ? list : []);
    } catch (err) {
      console.error("Failed to load opportunities:", err);
      setError("Failed to load opportunities. Please try again.");
      setOpportunities([]);
    } finally {
      setLoading(false);
    }
  }, [activeCategory]);

  useEffect(() => {
    loadOpportunities();
  }, [loadOpportunities]);

  const filteredOpportunities = useMemo(() => {
    let list = opportunities;
    const q = (searchParams.get("q") || query || "").toLowerCase().trim();
    if (q) {
      list = list.filter((o) => {
        const text = `${o.title || ""} ${o.description || ""} ${(o.tags || []).join(" ")} ${o.category || ""}`.toLowerCase();
        return text.includes(q);
      });
    }
    return list;
  }, [opportunities, searchParams, query]);

  const handleSearchSubmit = () => {
    const next = new URLSearchParams(searchParams);
    if (query.trim()) next.set("q", query.trim());
    else next.delete("q");
    setSearchParams(next);
  };

  return (
    <PageLayout maxWidth="xl" variant="light">
      <Grid container spacing={3}>
        <Grid item xs={12} md={3}>
          <Paper sx={{ p: 2.5, border: "1px solid #e5e7eb", borderRadius: 3 }}>
            <Typography fontWeight={700} sx={{ mb: 2 }}>
              Browse by Category
            </Typography>
            <Stack spacing={2}>
              {CATEGORY_TREE.map((group) => (
                <Box key={group.title}>
                  <Stack direction="row" alignItems="center" spacing={1} sx={{ mb: 0.5 }}>
                    <Checkbox
                      size="small"
                      checked={activeCategory === group.title}
                      onChange={() => setSearchParams(activeCategory === group.title ? {} : { category: group.title })}
                    />
                    <Typography fontWeight={600} variant="body2">
                      {group.title}
                    </Typography>
                  </Stack>
                  <Stack spacing={0.5} sx={{ pl: 3.5 }}>
                    {group.items.map((item) => (
                      <Stack key={item} direction="row" alignItems="center" spacing={1}>
                        <Checkbox
                          size="small"
                          checked={activeCategory === item}
                          onChange={() => setSearchParams(activeCategory === item ? {} : { category: item })}
                        />
                        <Typography variant="body2" color="text.secondary">
                          {item}
                        </Typography>
                      </Stack>
                    ))}
                  </Stack>
                </Box>
              ))}
            </Stack>
          </Paper>
        </Grid>

        <Grid item xs={12} md={9}>
          <Stack spacing={2.5}>
            <Box sx={{ display: "flex", flexDirection: { xs: "column", sm: "row" }, gap: 2, alignItems: "center" }}>
              <Paper
                sx={{
                  display: "flex",
                  alignItems: "center",
                  border: "1px solid #e5e7eb",
                  borderRadius: 3,
                  overflow: "hidden",
                  flex: 1,
                  maxWidth: 620,
                }}
              >
                <InputBase
                  placeholder="Search opportunities..."
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleSearchSubmit()}
                  sx={{ px: 2, py: 1.2, flex: 1 }}
                />
                <IconButton onClick={handleSearchSubmit} sx={{ borderRadius: 2, bgcolor: "#1976d2", color: "#fff" }}>
                  <SearchIcon />
                </IconButton>
              </Paper>
              {canPostOpportunity && (
                <Button
                  variant="contained"
                  startIcon={<AddIcon />}
                  onClick={() => setPostDialogOpen(true)}
                  sx={{
                    bgcolor: "#22c55e",
                    "&:hover": { bgcolor: "#16a34a" },
                    textTransform: "none",
                    fontWeight: 600,
                  }}
                >
                  Post a Project Brief
                </Button>
              )}
            </Box>

            <Box>
              <Typography variant="h4" fontWeight={700}>
                Opportunities
              </Typography>
              <Typography color="text.secondary">
                {loading ? "Loading..." : `${filteredOpportunities.length} opportunities`}
              </Typography>
            </Box>

            <StateContainer loading={loading} error={error} empty={filteredOpportunities.length === 0}>
              <Grid container spacing={2.5}>
                {filteredOpportunities.map((opp) => (
                  <Grid item xs={12} sm={6} md={4} key={opp.id}>
                    <OpportunityCard opportunity={opp} />
                  </Grid>
                ))}
              </Grid>
            </StateContainer>
          </Stack>
        </Grid>
      </Grid>

      <PostProjectBriefDialog
        open={postDialogOpen}
        onClose={() => setPostDialogOpen(false)}
        onSuccess={loadOpportunities}
      />
    </PageLayout>
  );
}
