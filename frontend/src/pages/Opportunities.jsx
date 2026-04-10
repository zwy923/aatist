import React, { useCallback, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import SearchIcon from "@mui/icons-material/Search";
import KeyboardArrowUpIcon from "@mui/icons-material/KeyboardArrowUp";
import { CircularProgress, IconButton, Tooltip } from "@mui/material";
import PageLayout from "../shared/components/PageLayout";
import { StateContainer } from "../shared/components/ui/StateContainer";
import { opportunitiesApi } from "../features/opportunities/api/opportunities";
import OpportunityCard from "../features/opportunities/components/OpportunityCard";
import PostProjectBriefDialog from "../features/opportunities/components/PostProjectBriefDialog";
import { useAuth } from "../features/auth/hooks/useAuth";
import "./Opportunities.css";

const isClientRole = (role) => {
  const r = (role || "").toLowerCase();
  return r === "org_person" || r === "org_team";
};

export default function OpportunitiesPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [opportunities, setOpportunities] = useState([]);
  const [locationOptions, setLocationOptions] = useState([]);
  const [postDialogOpen, setPostDialogOpen] = useState(false);
  const [showScrollTop, setShowScrollTop] = useState(false);
  const { user, isAuthenticated } = useAuth();

  const canPostOpportunity = isAuthenticated && isClientRole(user?.role);

  const queryInput = searchParams.get("q") || "";
  const [draftQ, setDraftQ] = useState(queryInput);

  useEffect(() => {
    setDraftQ(queryInput);
  }, [queryInput]);

  const locationFilter = searchParams.get("location") || "";
  const urgentOnly = searchParams.get("urgent") === "true";
  const sortKey = searchParams.get("sort") || "latest";
  const order = searchParams.get("order") || "desc";

  useEffect(() => {
    const onScroll = () => setShowScrollTop(window.scrollY > 400);
    window.addEventListener("scroll", onScroll, { passive: true });
    onScroll();
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const res = await opportunitiesApi.getOpportunityLocations();
        const list = res.data?.data;
        if (!cancelled && Array.isArray(list)) {
          setLocationOptions(list);
        }
      } catch (e) {
        if (!cancelled) setLocationOptions([]);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const fetchList = useCallback(async (signal) => {
    const params = {
      page: 1,
      limit: 50,
      status: "active",
      sort: sortKey,
      order,
    };
    const q = (searchParams.get("q") || "").trim();
    if (q) params.q = q;
    if ((searchParams.get("location") || "").trim()) {
      params.location = searchParams.get("location").trim();
    }
    if (searchParams.get("urgent") === "true") {
      params.urgent = true;
    }
    const reqCfg = signal ? { signal } : {};
    const response = await opportunitiesApi.getOpportunities(params, reqCfg);
    const payload = response.data?.data;
    const list = payload?.data ?? payload;
    return Array.isArray(list) ? list : [];
  }, [searchParams, sortKey, order]);

  useEffect(() => {
    const ac = new AbortController();
    (async () => {
      try {
        setLoading(true);
        setError(null);
        const list = await fetchList(ac.signal);
        if (ac.signal.aborted) return;
        setOpportunities(list);
      } catch (err) {
        if (err.code === "ERR_CANCELED" || err.name === "CanceledError") return;
        if (ac.signal.aborted) return;
        console.error("Failed to load opportunities:", err);
        setError(err.message || "Failed to load opportunities. Please try again.");
        setOpportunities([]);
      } finally {
        if (!ac.signal.aborted) setLoading(false);
      }
    })();
    return () => ac.abort();
  }, [fetchList]);

  const applySearchParams = (mutator) => {
    const next = new URLSearchParams(searchParams);
    mutator(next);
    setSearchParams(next);
  };

  const handleSearchSubmit = () => {
    applySearchParams((next) => {
      const t = draftQ.trim();
      if (t) next.set("q", t);
      else next.delete("q");
    });
  };

  const setLocation = (value) => {
    applySearchParams((next) => {
      if (value) next.set("location", value);
      else next.delete("location");
    });
  };

  const setUrgentMode = (mode) => {
    applySearchParams((next) => {
      if (mode === "urgent") next.set("urgent", "true");
      else next.delete("urgent");
    });
  };

  const setSortSelect = (value) => {
    applySearchParams((next) => {
      if (value === "latest") {
        next.set("sort", "latest");
        next.set("order", "desc");
      } else if (value === "oldest") {
        next.set("sort", "published_at");
        next.set("order", "asc");
      } else if (value === "budget_asc") {
        next.set("sort", "budget");
        next.set("order", "asc");
      } else if (value === "budget_desc") {
        next.set("sort", "budget");
        next.set("order", "desc");
      }
    });
  };

  const sortSelectValue = useMemo(() => {
    if (sortKey === "latest" || (sortKey === "published_at" && order === "desc")) return "latest";
    if (sortKey === "published_at" && order === "asc") return "oldest";
    if (sortKey === "budget" && order === "asc") return "budget_asc";
    if (sortKey === "budget" && order === "desc") return "budget_desc";
    return "latest";
  }, [sortKey, order]);

  return (
    <PageLayout
      noContainer
      rootClassName="opp-page-root"
      rootSx={{ bgcolor: "#c4ff3e" }}
      contentSx={{ py: 0, bgcolor: "transparent" }}
    >
      <div className="opp-shell">
        <section className="opp-hero" aria-label="Open opportunities">
          <div className="opp-hero-watermark" aria-hidden>
            CLIENT
          </div>
          <div className="opp-hero-inner">
            <h1 className="opp-hero-title">Open Opportunities</h1>
            <div className="opp-hero-row">
              <div className="opp-search-wrap">
                <input
                  className="opp-search-input"
                  placeholder="Search requests, skills, or tags..."
                  value={draftQ}
                  onChange={(e) => setDraftQ(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleSearchSubmit()}
                  aria-label="Search opportunities"
                />
                <button type="button" className="opp-search-btn" onClick={handleSearchSubmit} aria-label="Search">
                  <SearchIcon sx={{ fontSize: 26 }} />
                </button>
              </div>
              <Tooltip
                title={
                  canPostOpportunity
                    ? ""
                    : "Sign in with a client account to post an opportunity."
                }
              >
                <span style={{ display: "inline-flex" }}>
                  <button
                    type="button"
                    className="opp-post-btn"
                    disabled={!canPostOpportunity}
                    onClick={() => canPostOpportunity && setPostDialogOpen(true)}
                  >
                    Post an opportunity →
                  </button>
                </span>
              </Tooltip>
            </div>
          </div>
        </section>

        <section className="opp-filters" aria-label="Filters">
          <div className="opp-filter-group">
            <span className="opp-filter-label">Location</span>
            <select
              className="opp-filter-select"
              value={locationFilter}
              onChange={(e) => setLocation(e.target.value)}
              aria-label="Filter by location"
            >
              <option value="">All locations</option>
              {locationOptions.map((loc) => (
                <option key={loc} value={loc}>
                  {loc}
                </option>
              ))}
            </select>
          </div>

          <div className="opp-filter-group">
            <span className="opp-filter-label">Urgency</span>
            <div className="opp-urgency-toggle" role="group" aria-label="Urgency">
              <button
                type="button"
                className={`opp-urgency-btn${!urgentOnly ? " active" : ""}`}
                onClick={() => setUrgentMode("all")}
              >
                All
              </button>
              <button
                type="button"
                className={`opp-urgency-btn${urgentOnly ? " active" : ""}`}
                onClick={() => setUrgentMode("urgent")}
              >
                Urgent
              </button>
            </div>
          </div>

          <div className="opp-filters-right">
            <div className="opp-filter-group">
              <span className="opp-filter-label">Sort by</span>
              <select
                className="opp-filter-select"
                value={sortSelectValue}
                onChange={(e) => setSortSelect(e.target.value)}
                aria-label="Sort opportunities"
              >
                <option value="latest">Latest</option>
                <option value="oldest">Oldest</option>
                <option value="budget_asc">Budget: low to high</option>
                <option value="budget_desc">Budget: high to low</option>
              </select>
            </div>
          </div>
        </section>

        <section className="opp-grid-section">
          <div className="opp-grid-inner">
            <StateContainer
              loading={loading}
              error={error}
              empty={!loading && !error && opportunities.length === 0}
              emptyMessage="No opportunities match your filters."
              emptyStateSx={{ py: 4, opacity: 0.65 }}
              onRetry={() => setSearchParams(new URLSearchParams(searchParams))}
              loadingComponent={
                <div className="opp-loading">
                  <CircularProgress sx={{ color: "#0a5c5c" }} />
                </div>
              }
            >
              <div className="opp-cards-grid">
                {opportunities.map((opp) => (
                  <OpportunityCard key={opp.id} opportunity={opp} />
                ))}
              </div>
            </StateContainer>
          </div>
        </section>
      </div>

      <PostProjectBriefDialog
        open={postDialogOpen}
        onClose={() => setPostDialogOpen(false)}
        onSuccess={() => {
          setSearchParams(new URLSearchParams(searchParams));
        }}
        defaultOrganization={user?.organization_name || user?.name || ""}
      />

      {showScrollTop ? (
        <Tooltip title="回到顶部" placement="left">
          <IconButton
            className="opp-scroll-top-btn"
            onClick={() => window.scrollTo({ top: 0, behavior: "smooth" })}
            aria-label="回到顶部"
            size="large"
          >
            <KeyboardArrowUpIcon sx={{ fontSize: 32 }} />
          </IconButton>
        </Tooltip>
      ) : null}
    </PageLayout>
  );
}
