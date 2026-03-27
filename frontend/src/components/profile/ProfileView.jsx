import React, { useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Avatar, Box, IconButton, Menu, MenuItem, Typography } from "@mui/material";
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
import "../../pages/PublicProfile.css";

export default function ProfileView({
  profile,
  portfolio = [],
  services = [],
  isOwnProfile,
  bannerUrl,
  onMessage,
  onChangeBackground,
  onEditProfile,
  onAddService,
  onAddProject,
  onPortfolioEdit,
  alternateContent = null,
  showServicesAndPortfolio = true,
  profileUserId: profileUserIdProp,
}) {
  const navigate = useNavigate();
  const portfolioSectionRef = useRef(null);
  const profileOwnerId = profileUserIdProp ?? profile?.id;
  const [servicesView, setServicesView] = useState("grid");
  const [portfolioView, setPortfolioView] = useState("grid");
  const [portfolioMenuAnchor, setPortfolioMenuAnchor] = useState(null);
  const [portfolioMenuId, setPortfolioMenuId] = useState(null);

  const educationParts = [profile?.school, profile?.faculty || profile?.major].filter(Boolean);
  const educationLine = profile?.year
    ? `${profile.year}, ${educationParts.join(" / ")}`
    : educationParts.join(" / ");

  const orgSubtitle = [profile?.organization_name, profile?.contact_title].filter(Boolean).join(" · ");

  const skills = profile?.skills || [];
  const skillNames = skills.map((s) => (typeof s === "string" ? s : s?.name)).filter(Boolean);
  const interests = (profile?.professional_interests || "")
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);

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
      return `Fee: €${s.price_min}${s.price_max != null ? `–€${s.price_max}` : ""} / project`;
    }
    if (s.price_type === "hourly" && s.price_min != null) {
      return `Fee: €${s.price_min}${s.price_max != null ? `–€${s.price_max}` : ""} / hr`;
    }
    return "Fee: Negotiable";
  };

  const formatProjectDate = (p) => {
    if (p.created_at) {
      const d = new Date(p.created_at);
      return d.toLocaleDateString("en-GB", { day: "2-digit", month: "2-digit", year: "numeric" });
    }
    return p.year ? `${p.year}` : "";
  };

  const scrollToPortfolio = () => {
    portfolioSectionRef.current?.scrollIntoView({ behavior: "smooth", block: "start" });
  };

  const scrollToTop = () => {
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const displayName = profile?.name || profile?.organization_name || "User";

  return (
    <div className="profile-page">
      <div className="profile-hero">
        <div
          className="profile-hero-bg"
          style={{
            backgroundImage: bannerUrl ? `url(${bannerUrl})` : "none",
          }}
        />
        <div className="profile-hero-bg-placeholder" />

        {isOwnProfile ? (
          <div className="profile-hero-actions">
            <button type="button" className="profile-hero-btn-edit" onClick={onEditProfile}>
              Edit Profile
            </button>
            <button type="button" className="profile-hero-btn-change-bg" onClick={onChangeBackground}>
              Change Background
            </button>
          </div>
        ) : (
          onMessage && (
            <div className="profile-hero-actions">
              <button type="button" className="profile-hero-btn-edit" onClick={onMessage}>
                Message
              </button>
            </div>
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
            {displayName?.charAt(0)?.toUpperCase() || "?"}
          </Avatar>

          <div className="profile-hero-info">
            <div className="profile-hero-name-row">
              <span className="profile-hero-name">{displayName}</span>
              {profile?.is_verified_email && (
                <VerifiedUserIcon className="profile-hero-verified" sx={{ fontSize: 28 }} />
              )}
              <span className="profile-hero-badge-available">Available</span>
            </div>
            {(educationLine || orgSubtitle) && (
              <div className="profile-hero-education">{educationLine || orgSubtitle}</div>
            )}
            {(profile?.bio || profile?.organization_bio) && (
              <div className="profile-hero-about">{profile.bio || profile.organization_bio}</div>
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
                <div className="profile-hero-meta-label">Languages</div>
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
                  {skillNames.slice(0, 8).map((sk, i) => (
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
                  {interests.slice(0, 8).map((int, i) => (
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

      <div className="profile-content">
        {alternateContent}

        {showServicesAndPortfolio && !alternateContent && (
          <>
            <Box sx={{ mb: 6 }}>
              <div className="profile-section-header">
                <div className="profile-section-title">
                  <DesignServicesIcon sx={{ color: "#0095D9", fontSize: 24 }} />
                  Services
                </div>
                <div className="profile-section-actions">
                  <div className="profile-section-toggle">
                    <button
                      type="button"
                      className={`profile-section-toggle-btn ${servicesView === "grid" ? "active" : ""}`}
                      onClick={() => setServicesView("grid")}
                    >
                      <GridViewIcon fontSize="small" /> Grid
                    </button>
                    <button
                      type="button"
                      className={`profile-section-toggle-btn ${servicesView === "list" ? "active" : ""}`}
                      onClick={() => setServicesView("list")}
                    >
                      <ViewListIcon fontSize="small" /> List
                    </button>
                  </div>
                  {isOwnProfile && onAddService && (
                    <button type="button" className="profile-section-add-btn" onClick={onAddService}>
                      <AddIcon sx={{ fontSize: 18 }} /> Add Service
                    </button>
                  )}
                </div>
              </div>

              <div className={`profile-services ${servicesView === "list" ? "profile-services-list" : ""}`}>
                {services.map((s) => (
                  <div
                    key={s.id}
                    className="profile-service-card"
                    role={profileOwnerId && s.id ? "button" : undefined}
                    tabIndex={profileOwnerId && s.id ? 0 : undefined}
                    onClick={() => {
                      if (profileOwnerId && s.id) navigate(`/users/${profileOwnerId}/services/${s.id}`);
                    }}
                    onKeyDown={(e) => {
                      if (!profileOwnerId || !s.id) return;
                      if (e.key === "Enter" || e.key === " ") {
                        e.preventDefault();
                        navigate(`/users/${profileOwnerId}/services/${s.id}`);
                      }
                    }}
                    style={{ cursor: profileOwnerId && s.id ? "pointer" : undefined }}
                  >
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
                      <div className="profile-service-title">{s.title || s.category || "Service"}</div>
                      <div className="profile-service-desc">
                        {s.short_description || s.description || s.experience_summary || ""}
                      </div>
                      <div className="profile-service-fee">{formatPrice(s)}</div>
                    </div>
                  </div>
                ))}
              </div>
              {services.length === 0 && (
                <Typography sx={{ py: 2, color: "#999", textAlign: "center" }} variant="body2">
                  No services listed yet.
                </Typography>
              )}
              <div className="profile-section-footer">
                <button type="button" className="profile-section-chevron-btn" onClick={scrollToPortfolio} aria-label="Scroll to portfolio">
                  <ExpandMoreIcon className="profile-section-chevron" />
                </button>
              </div>
            </Box>

            <Box ref={portfolioSectionRef}>
              <div className="profile-section-header">
                <div className="profile-section-title">
                  <WorkIcon sx={{ color: "#0095D9", fontSize: 24 }} />
                  Portfolio
                </div>
                <div className="profile-section-actions">
                  <div className="profile-section-toggle">
                    <button
                      type="button"
                      className={`profile-section-toggle-btn ${portfolioView === "grid" ? "active" : ""}`}
                      onClick={() => setPortfolioView("grid")}
                    >
                      <GridViewIcon fontSize="small" /> Grid
                    </button>
                    <button
                      type="button"
                      className={`profile-section-toggle-btn ${portfolioView === "list" ? "active" : ""}`}
                      onClick={() => setPortfolioView("list")}
                    >
                      <ViewListIcon fontSize="small" /> List
                    </button>
                  </div>
                  {isOwnProfile && onAddProject && (
                    <button type="button" className="profile-section-add-btn" onClick={onAddProject}>
                      <AddIcon sx={{ fontSize: 18 }} /> Add Project
                    </button>
                  )}
                </div>
              </div>

              {portfolio.length > 0 ? (
                <>
                  <div className={`profile-portfolio ${portfolioView === "list" ? "profile-portfolio-list" : ""}`}>
                    {portfolio.map((project) => (
                      <div
                        key={project.id}
                        className="profile-portfolio-card"
                        role="button"
                        tabIndex={0}
                        onClick={() => navigate(`/portfolio/${project.id}`)}
                        onKeyDown={(e) => {
                          if (e.key === "Enter" || e.key === " ") {
                            e.preventDefault();
                            navigate(`/portfolio/${project.id}`);
                          }
                        }}
                        style={{ cursor: "pointer" }}
                      >
                        <div className="profile-portfolio-image">
                          {project.cover_image_url ? (
                            <img src={project.cover_image_url} alt={project.title} />
                          ) : null}
                          {isOwnProfile && onPortfolioEdit && (
                            <IconButton
                              className="profile-portfolio-menu"
                              size="small"
                              onClick={(e) => {
                                e.stopPropagation();
                                handlePortfolioMenuOpen(e, project.id);
                              }}
                            >
                              <MoreVertIcon fontSize="small" />
                            </IconButton>
                          )}
                          <span className="profile-portfolio-date">{formatProjectDate(project)}</span>
                          <span className="profile-portfolio-type">
                            {project.service_category || "Personal"}
                          </span>
                        </div>
                        <div className="profile-portfolio-body">
                          <div className="profile-portfolio-title">{project.title}</div>
                          <div className="profile-portfolio-desc">{project.short_caption || project.description || ""}</div>
                        </div>
                      </div>
                    ))}
                  </div>
                  <div className="profile-section-footer">
                    <button type="button" className="profile-section-chevron-btn" onClick={scrollToTop} aria-label="Back to top">
                      <ExpandLessIcon className="profile-section-chevron" />
                    </button>
                  </div>
                </>
              ) : (
                <Box sx={{ py: 6, textAlign: "center", color: "#999" }}>
                  <Typography>No portfolio projects to display.</Typography>
                </Box>
              )}
            </Box>
          </>
        )}

        {alternateContent && (
          <div className="profile-section-footer">
            <button type="button" className="profile-section-chevron-btn" onClick={scrollToTop} aria-label="Back to top">
              <ExpandLessIcon className="profile-section-chevron" />
            </button>
          </div>
        )}
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
            if (portfolioMenuId != null && onPortfolioEdit) onPortfolioEdit(portfolioMenuId);
          }}
        >
          Edit
        </MenuItem>
      </Menu>
    </div>
  );
}
