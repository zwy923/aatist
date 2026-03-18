import React, { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import {
  Avatar,
  Box,
  Chip,
  IconButton,
  Menu,
  MenuItem,
  Typography,
} from "@mui/material";
import DesignServicesIcon from "@mui/icons-material/DesignServices";
import WorkIcon from "@mui/icons-material/Work";
import VerifiedUserIcon from "@mui/icons-material/VerifiedUser";
import AddIcon from "@mui/icons-material/Add";
import MoreVertIcon from "@mui/icons-material/MoreVert";
import GridViewIcon from "@mui/icons-material/GridView";
import ViewListIcon from "@mui/icons-material/ViewList";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import ExpandLessIcon from "@mui/icons-material/ExpandLess";
import LanguageIcon from "@mui/icons-material/Language";
import LinkedInIcon from "@mui/icons-material/LinkedIn";
import PaletteIcon from "@mui/icons-material/Palette";
import useAuthStore from "../shared/stores/authStore";
import { profileApi, portfolioApi } from "../features/profile/api/profile";
import PageLayout from "../shared/components/PageLayout";
import { StateContainer } from "../shared/components/ui/StateContainer";
import "./PublicProfile.css";

export default function PublicProfile() {
  const { id } = useParams();
  const navigate = useNavigate();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const user = useAuthStore((s) => s.user);
  const myId = user?.id ?? user?.user_id;
  const isOwnProfile = isAuthenticated && id && Number(id) === Number(myId);

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [profile, setProfile] = useState(null);
  const [portfolio, setPortfolio] = useState([]);
  const [services, setServices] = useState([]);
  const [portfolioMenuAnchor, setPortfolioMenuAnchor] = useState(null);
  const [portfolioMenuId, setPortfolioMenuId] = useState(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        setError(null);

        const [profileRes, portfolioRes] = await Promise.allSettled([
          profileApi.getPublicProfile(id),
          portfolioApi.getUserPortfolio(id),
        ]);

        if (profileRes.status === "fulfilled") {
          const data = profileRes.value.data.data;
          setProfile(data);
          setServices(data?.services || []);
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

    if (id) fetchData();
  }, [id]);

  const educationParts = [profile?.school, profile?.faculty || profile?.major].filter(Boolean);
  const educationLine = profile?.year
    ? `${profile.year}, ${educationParts.join(" / ")}`
    : educationParts.join(" / ");

  const skills = profile?.skills || [];
  const skillNames = skills.map((s) => (typeof s === "string" ? s : s?.name)).filter(Boolean);
  const interests = (profile?.professional_interests || "")
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);

  const handleChangeBackground = () => {
    navigate("/profile?tab=basic");
  };

  const handleEditProfile = () => {
    navigate("/profile");
  };

  const handleAddService = () => {
    navigate("/profile?tab=services");
  };

  const handleAddProject = () => {
    navigate("/profile?tab=portfolio");
  };

  const handlePortfolioMenuOpen = (e, projectId) => {
    e.stopPropagation();
    setPortfolioMenuAnchor(e.currentTarget);
    setPortfolioMenuId(projectId);
  };

  const handlePortfolioMenuClose = () => {
    setPortfolioMenuAnchor(null);
    setPortfolioMenuId(null);
  };

  const formatPrice = (s) => {
    if (s.price_type === "project" && s.price_min != null) {
      return `€${s.price_min}${s.price_max != null ? `–€${s.price_max}` : ""} / project`;
    }
    if (s.price_type === "hourly" && s.price_min != null) {
      return `€${s.price_min}${s.price_max != null ? `–€${s.price_max}` : ""} / hr`;
    }
    return "Negotiable";
  };

  const formatProjectDate = (p) => {
    if (p.created_at) {
      const d = new Date(p.created_at);
      return d.toLocaleDateString("en-GB", { day: "2-digit", month: "2-digit", year: "numeric" });
    }
    return p.year ? `${p.year}` : "";
  };

  return (
    <PageLayout noContainer>
      <StateContainer loading={loading} error={error}>
        <div className="profile-page">
          {/* Profile Hero / Banner */}
          <div className="profile-hero">
            <div
              className="profile-hero-bg"
              style={{
                backgroundImage: profile?.banner_url
                  ? `url(${profile.banner_url})`
                  : "none",
              }}
            />
            <div className="profile-hero-bg-placeholder" />

            {isOwnProfile ? (
              <>
                <button
                  type="button"
                  className="profile-hero-btn-change-bg"
                  onClick={handleChangeBackground}
                >
                  Change Background
                </button>
                <button
                  type="button"
                  className="profile-hero-btn-edit"
                  onClick={handleEditProfile}
                >
                  Edit Profile
                </button>
              </>
            ) : (
              isAuthenticated && (
                <button
                  type="button"
                  className="profile-hero-btn-edit"
                  onClick={() => navigate(`/messages?user=${id}`)}
                  style={{ right: 24 }}
                >
                  Message
                </button>
              )
            )}

            <div className="profile-hero-overlay">
              <Avatar
                src={profile?.avatar_url}
                className="profile-hero-avatar"
                sx={{
                  width: 140,
                  height: 140,
                  borderRadius: 1,
                  bgcolor: "rgba(255,255,255,0.95)",
                  color: "#333",
                }}
              >
                {profile?.name?.charAt(0)?.toUpperCase() || "?"}
              </Avatar>

              <div className="profile-hero-info">
                <div className="profile-hero-name-row">
                  <span className="profile-hero-name">{profile?.name || "User"}</span>
                  {profile?.is_verified_email && (
                    <VerifiedUserIcon className="profile-hero-verified" sx={{ fontSize: 28 }} />
                  )}
                  <Chip
                    label="Available"
                    size="small"
                    sx={{
                      bgcolor: "#4caf50",
                      color: "#fff",
                      fontWeight: 600,
                      fontSize: "12px",
                    }}
                  />
                </div>
                {educationLine && (
                  <div className="profile-hero-education">{educationLine}</div>
                )}
                {profile?.bio && (
                  <div className="profile-hero-about">{profile.bio}</div>
                )}
                <div className="profile-hero-links">
                  {profile?.website && (
                    <a
                      href={profile.website}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="profile-hero-link"
                    >
                      <LanguageIcon sx={{ fontSize: 18 }} />
                      {profile.website.replace(/^https?:\/\//, "").slice(0, 30)}
                    </a>
                  )}
                  {profile?.behance && (
                    <a
                      href={profile.behance}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="profile-hero-link"
                    >
                      <PaletteIcon sx={{ fontSize: 18 }} />
                      Behance
                    </a>
                  )}
                  {profile?.linkedin && (
                    <a
                      href={profile.linkedin}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="profile-hero-link"
                    >
                      <LinkedInIcon sx={{ fontSize: 18 }} />
                      LinkedIn
                    </a>
                  )}
                </div>
              </div>

              <div className="profile-hero-meta">
                {profile?.languages && (
                  <div>
                    <div className="profile-hero-meta-label">Language</div>
                    <div className="profile-hero-meta-tags">
                      {profile.languages.split(",").map((l, i) => (
                        <span key={i} className="profile-hero-meta-tag">
                          {l.trim()}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
                {skillNames.length > 0 && (
                  <div>
                    <div className="profile-hero-meta-label">Skills</div>
                    <div className="profile-hero-meta-tags">
                      {skillNames.slice(0, 6).map((sk, i) => (
                        <span key={i} className="profile-hero-meta-tag">
                          {sk}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
                {interests.length > 0 && (
                  <div>
                    <div className="profile-hero-meta-label">Interest Areas</div>
                    <div className="profile-hero-meta-tags">
                      {interests.slice(0, 6).map((int, i) => (
                        <span key={i} className="profile-hero-meta-tag">
                          {int}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Content */}
          <div className="profile-content">
            {/* Services Section */}
            <Box sx={{ mb: 6 }}>
              <div className="profile-section-header">
                <div className="profile-section-title">
                  <DesignServicesIcon sx={{ color: "#0095D9", fontSize: 24 }} />
                  Services
                </div>
                <div className="profile-section-actions">
                  <div className="profile-section-toggle">
                    <button type="button" className="profile-section-toggle-btn active">
                      <GridViewIcon fontSize="small" /> Grid
                    </button>
                    <button type="button" className="profile-section-toggle-btn">
                      <ViewListIcon fontSize="small" /> List
                    </button>
                  </div>
                  {isOwnProfile && (
                    <button
                      type="button"
                      className="profile-section-add-btn"
                      onClick={handleAddService}
                    >
                      <AddIcon sx={{ fontSize: 18 }} /> Add Service
                    </button>
                  )}
                </div>
              </div>

              <div className="profile-services">
                {services.map((s) => (
                  <div key={s.id} className="profile-service-card">
                    <div className="profile-service-thumbnails">
                      {(s.media_urls || []).slice(0, 3).map((url, i) => (
                        <div key={i} className="profile-service-thumb">
                          <img src={url} alt="" />
                        </div>
                      ))}
                      {(!s.media_urls || s.media_urls.length === 0) && (
                        <>
                          <div className="profile-service-thumb" />
                          <div className="profile-service-thumb" />
                          <div className="profile-service-thumb" />
                        </>
                      )}
                    </div>
                    <div className="profile-service-body">
                      <div className="profile-service-title">
                        {s.title || s.category || "Service"}
                      </div>
                      <div className="profile-service-desc">
                        {s.short_description || s.description || s.experience_summary || ""}
                      </div>
                      <div className="profile-service-fee">{formatPrice(s)}</div>
                    </div>
                  </div>
                ))}
              </div>
              <div className="profile-section-footer">
                <ExpandMoreIcon className="profile-section-chevron" />
              </div>
            </Box>

            {/* Portfolio Section */}
            <Box>
              <div className="profile-section-header">
                <div className="profile-section-title">
                  <WorkIcon sx={{ color: "#0095D9", fontSize: 24 }} />
                  Portfolio
                </div>
                <div className="profile-section-actions">
                  <div className="profile-section-toggle">
                    <button type="button" className="profile-section-toggle-btn active">
                      <GridViewIcon fontSize="small" /> Grid
                    </button>
                    <button type="button" className="profile-section-toggle-btn">
                      <ViewListIcon fontSize="small" /> List
                    </button>
                  </div>
                  {isOwnProfile && (
                    <button
                      type="button"
                      className="profile-section-add-btn"
                      onClick={handleAddProject}
                    >
                      <AddIcon sx={{ fontSize: 18 }} /> Add Project
                    </button>
                  )}
                </div>
              </div>

              {portfolio.length > 0 ? (
                <>
                  <div className="profile-portfolio">
                    {portfolio.map((project) => (
                      <div key={project.id} className="profile-portfolio-card">
                        <div className="profile-portfolio-image">
                          {project.cover_image_url ? (
                            <img
                              src={project.cover_image_url}
                              alt={project.title}
                            />
                          ) : null}
                          {isOwnProfile && (
                            <IconButton
                              className="profile-portfolio-menu"
                              size="small"
                              onClick={(e) => handlePortfolioMenuOpen(e, project.id)}
                            >
                              <MoreVertIcon fontSize="small" />
                            </IconButton>
                          )}
                          <span className="profile-portfolio-date">
                            {formatProjectDate(project)}
                          </span>
                          <span className="profile-portfolio-type">
                            {project.service_category || "Personal"}
                          </span>
                        </div>
                        <div className="profile-portfolio-body">
                          <div className="profile-portfolio-title">
                            {project.title}
                          </div>
                          <div className="profile-portfolio-desc">
                            {project.description || ""}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                  <div className="profile-section-footer">
                    <ExpandLessIcon className="profile-section-chevron" />
                  </div>
                </>
              ) : (
                <Box sx={{ py: 6, textAlign: "center", color: "#999" }}>
                  <Typography>No portfolio projects to display.</Typography>
                </Box>
              )}
            </Box>
          </div>
        </div>

        <Menu
          anchorEl={portfolioMenuAnchor}
          open={Boolean(portfolioMenuAnchor)}
          onClose={handlePortfolioMenuClose}
          anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
          transformOrigin={{ vertical: "top", horizontal: "right" }}
        >
          <MenuItem
            onClick={() => {
              handlePortfolioMenuClose();
              navigate("/profile?tab=portfolio");
            }}
          >
            Edit
          </MenuItem>
        </Menu>
      </StateContainer>
    </PageLayout>
  );
}
