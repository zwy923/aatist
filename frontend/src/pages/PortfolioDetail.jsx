import React, { useCallback, useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import {
  Box,
  CircularProgress,
  Container,
  IconButton,
  Tooltip,
} from "@mui/material";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import FavoriteBorderIcon from "@mui/icons-material/FavoriteBorder";
import FavoriteIcon from "@mui/icons-material/Favorite";
import PageLayout from "../shared/components/PageLayout";
import { StateContainer } from "../shared/components/ui/StateContainer";
import { portfolioApi, profileApi } from "../features/profile/api/profile";
import useAuthStore from "../shared/stores/authStore";
import "./PortfolioDetail.css";

function formatProjectDate(project) {
  if (project?.created_at) {
    const d = new Date(project.created_at);
    if (!Number.isNaN(d.getTime())) {
      return d.toLocaleDateString("en-GB", {
        day: "2-digit",
        month: "2-digit",
        year: "2-digit",
      });
    }
  }
  if (project?.year) return String(project.year);
  return "—";
}

function educationSubtitle(profile) {
  const parts = [profile?.school, profile?.faculty || profile?.major].filter(Boolean);
  const line = parts.join(" / ");
  if (profile?.year && line) return `${profile.year}, ${line}`;
  if (profile?.year) return String(profile.year);
  return line || "Aalto talent";
}

function coCreatorSubtitle(c) {
  if (c.subtitle) return c.subtitle;
  if (c.email) return c.email;
  return "Co-creator";
}

function normalizeTags(tags) {
  if (!tags) return [];
  if (Array.isArray(tags)) return tags.map((t) => (typeof t === "string" ? t : t?.name || String(t))).filter(Boolean);
  return [];
}

function collectMediaUrls(project) {
  if (!project) return [];
  const out = [];
  if (project.cover_image_url) out.push(project.cover_image_url);
  const rest = project.media_urls || [];
  for (const u of rest) {
    if (u && u !== project.cover_image_url) out.push(u);
  }
  return out;
}

export default function PortfolioDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { isAuthenticated } = useAuthStore();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [project, setProject] = useState(null);
  const [ownerProfile, setOwnerProfile] = useState(null);
  const [favLoading, setFavLoading] = useState(false);
  const [favSaved, setFavSaved] = useState(false);

  const projectId = id ? Number(id) : NaN;

  const load = useCallback(async () => {
    if (!Number.isFinite(projectId)) {
      setError("Invalid project");
      setLoading(false);
      return;
    }
    try {
      setLoading(true);
      setError(null);
      const res = await portfolioApi.getPortfolioById(projectId);
      const data = res.data?.data;
      if (!data?.id) {
        setProject(null);
        setOwnerProfile(null);
        setLoading(false);
        return;
      }
      setProject(data);
      const uid = data.user_id;
      if (uid) {
        try {
          const pr = await profileApi.getPublicProfile(uid);
          setOwnerProfile(pr.data?.data || null);
        } catch {
          setOwnerProfile(null);
        }
      } else {
        setOwnerProfile(null);
      }
    } catch (e) {
      console.error(e);
      setError("Could not load this project.");
      setProject(null);
      setOwnerProfile(null);
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => {
    load();
  }, [load]);

  const ownerName = ownerProfile?.name || ownerProfile?.organization_name || "Creator";
  const categoryCrumb = (project?.service_category || "Portfolio").toUpperCase();
  const tagList = normalizeTags(project?.tags);
  const relatedList = normalizeTags(project?.related_services);
  const mediaUrls = useMemo(() => collectMediaUrls(project), [project]);

  const creators = useMemo(() => {
    if (!project) return [];
    const list = [];
    if (ownerProfile) {
      list.push({
        key: `owner-${project.user_id}`,
        name: ownerProfile.name || ownerName,
        subtitle: educationSubtitle(ownerProfile),
        avatarUrl: ownerProfile.avatar_url,
      });
    }
    const cc = project.co_creators || [];
    for (let i = 0; i < cc.length; i++) {
      const c = cc[i];
      list.push({
        key: `co-${c.email || i}`,
        name: c.name || c.email || "Co-creator",
        subtitle: coCreatorSubtitle(c),
        avatarUrl: null,
      });
    }
    if (list.length === 0 && project.user_id) {
      list.push({
        key: "owner-fallback",
        name: ownerName,
        subtitle: "Project owner",
        avatarUrl: null,
      });
    }
    return list;
  }, [project, ownerProfile, ownerName]);

  const [hero, ...restMedia] = mediaUrls.length ? mediaUrls : [null];
  const stackMedia = restMedia.slice(0, 2);
  const moreMedia = restMedia.slice(2);

  const toggleFavorite = async () => {
    if (!isAuthenticated) {
      navigate(`/auth/login?redirect=${encodeURIComponent(window.location.pathname)}`);
      return;
    }
    setFavLoading(true);
    try {
      if (favSaved) {
        await profileApi.removeSavedItemByTarget("project", projectId);
        setFavSaved(false);
      } else {
        await profileApi.saveItem("project", projectId);
        setFavSaved(true);
      }
    } catch (err) {
      console.error(err);
    } finally {
      setFavLoading(false);
    }
  };

  const descriptionText =
    project?.description?.trim() || project?.short_caption?.trim() || "No description provided for this project yet.";

  return (
    <PageLayout noContainer>
      <div className="portfolio-detail-page">
        <Container maxWidth="lg">
          <StateContainer
            loading={loading}
            error={error}
            empty={!project}
            emptyMessage="Project not found."
            onRetry={load}
          >
            <div className="portfolio-detail-topbar">
              <nav className="portfolio-detail-back-row" aria-label="Breadcrumb">
                <IconButton
                  size="small"
                  aria-label="Go back"
                  onClick={() => navigate(-1)}
                  sx={{ color: "#5f6368", mr: 0.5 }}
                >
                  <ArrowBackIcon fontSize="small" />
                </IconButton>
                <Link to="/talents">Hire Talent</Link>
                <span className="sep">/</span>
                <span>{categoryCrumb}</span>
                <span className="sep">/</span>
                <span>{ownerName}</span>
              </nav>
            </div>

            <div className="portfolio-detail-header">
              <div className="portfolio-detail-title-row">
                <div style={{ display: "flex", flexWrap: "wrap", alignItems: "baseline", gap: "12px 24px", flex: 1 }}>
                  <h1 className="portfolio-detail-title">{project?.title || "Project"}</h1>
                  <span className="portfolio-detail-date">{formatProjectDate(project)}</span>
                </div>
                <Tooltip title={favSaved ? "Remove from saved" : "Save project"}>
                  <span className="portfolio-detail-fav-btn">
                    <IconButton
                      onClick={toggleFavorite}
                      disabled={favLoading}
                      sx={{
                        border: "1px solid #dadce0",
                        bgcolor: "#fff",
                        "&:hover": { bgcolor: "#f8f9fa" },
                      }}
                      aria-label={favSaved ? "Remove from saved" : "Save project"}
                    >
                      {favLoading ? (
                        <CircularProgress size={22} />
                      ) : favSaved ? (
                        <FavoriteIcon sx={{ color: "#e91e63" }} />
                      ) : (
                        <FavoriteBorderIcon sx={{ color: "#5f6368" }} />
                      )}
                    </IconButton>
                  </span>
                </Tooltip>
              </div>

              <h2 className="portfolio-detail-label">Created by</h2>
              <div className="portfolio-detail-creators">
                {creators.map((c) => (
                  <div key={c.key} className="portfolio-detail-creator">
                    {c.avatarUrl ? (
                      <img className="portfolio-detail-creator-avatar" src={c.avatarUrl} alt="" />
                    ) : (
                      <div className="portfolio-detail-creator-avatar-ph" aria-hidden />
                    )}
                    <div>
                      <p className="portfolio-detail-creator-name">{c.name}</p>
                      <p className="portfolio-detail-creator-sub">{c.subtitle}</p>
                    </div>
                  </div>
                ))}
              </div>

              <h2 className="portfolio-detail-label">Description</h2>
              <p className="portfolio-detail-description">{descriptionText}</p>
              {tagList.length > 0 && (
                <div className="portfolio-detail-tags">
                  {tagList.map((t) => (
                    <span key={t} className="portfolio-detail-tag">
                      {t}
                    </span>
                  ))}
                </div>
              )}

              {relatedList.length > 0 && (
                <>
                  <h2 className="portfolio-detail-label">Service involved</h2>
                  <div className="portfolio-detail-service-tags">
                    {relatedList.map((t) => (
                      <span key={t} className="portfolio-detail-service-tag">
                        {t}
                      </span>
                    ))}
                  </div>
                </>
              )}
            </div>

            <div className="portfolio-detail-media">
              <div className="portfolio-detail-media-row">
                <div className="portfolio-detail-media-hero">
                  {hero ? (
                    <img src={hero} alt="" />
                  ) : (
                    <Box sx={{ width: "100%", minHeight: 320, bgcolor: "#f8f9fa" }} />
                  )}
                </div>
                <div className="portfolio-detail-media-stack">
                  {stackMedia.map((url, i) => (
                    <div
                      key={url + i}
                      className={`portfolio-detail-media-cell ${i % 2 === 0 ? "light" : ""}`}
                    >
                      {url ? <img src={url} alt="" /> : null}
                    </div>
                  ))}
                  {hero && stackMedia.length === 0 && (
                    <div className="portfolio-detail-media-cell" aria-hidden />
                  )}
                  {!hero && (
                    <>
                      <div className="portfolio-detail-media-cell light" aria-hidden />
                      <div className="portfolio-detail-media-cell" aria-hidden />
                    </>
                  )}
                </div>
              </div>
              {moreMedia.length > 0 && (
                <div className="portfolio-detail-media-more">
                  {moreMedia.map((url, i) => (
                    <div key={url + i} className="portfolio-detail-media-cell light">
                      <img src={url} alt="" />
                    </div>
                  ))}
                </div>
              )}
            </div>
          </StateContainer>
        </Container>
      </div>
    </PageLayout>
  );
}
