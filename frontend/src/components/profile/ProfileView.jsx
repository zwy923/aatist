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
import EditIcon from "@mui/icons-material/Edit";
import ImageOutlinedIcon from "@mui/icons-material/ImageOutlined";
import "../../pages/PublicProfile.css";
import { formatServicePriceLine } from "../../shared/utils/priceType";
import { getProfileServiceHeading } from "../../constants/serviceCategories";
import { talentDisplayName } from "../../shared/utils/displayName";

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

  const educationLine = (profile?.school || "").trim();

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

  const formatPrice = (s) => `Fee: ${formatServicePriceLine(s)}`;

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

  const roleLower = profile?.role?.toLowerCase?.();
  const isOrgProfile = roleLower === "org_person" || roleLower === "org_team";
  const displayName = isOrgProfile
    ? profile?.organization_name || profile?.name || "User"
    : talentDisplayName(profile) || profile?.organization_name || profile?.name || "User";
  const isVerified = profile?.is_verified_email || profile?.role_verified;
  const bioText = (profile?.bio || profile?.organization_bio || "").trim();

  const renderServiceThumbs = (s) => {
    const urls = (s.media_urls || []).filter(Boolean);
    const extra = urls.length > 3 ? urls.length - 3 : 0;
    return [0, 1, 2].map((i) => {
      const url = urls[i];
      const showMore = i === 2 && extra > 0;
      return (
        <div key={i} className="profile-service-thumb">
          {url ? <img src={url} alt="" /> : null}
          {showMore && <span className="profile-service-thumb-more">+{extra}</span>}
        </div>
      );
    });
  };

  return (
    <div className="profile-page">
      <div className="profile-hero">
        <div
          className="profile-hero-bg"
          style={{
            backgroundImage: bannerUrl ? `url(${bannerUrl})` : "none",
          }}
        />
        {!bannerUrl && <div className="profile-hero-bg-placeholder" aria-hidden />}

        <div className="profile-hero-topbar">
          {isOwnProfile && onChangeBackground && (
            <button type="button" className="profile-hero-btn-change-bg" onClick={onChangeBackground}>
              <ImageOutlinedIcon sx={{ fontSize: 18, mr: 0.75 }} />
              Change Background
            </button>
          )}
          {!isOwnProfile && onMessage && (
            <button type="button" className="profile-hero-btn-message-top" onClick={onMessage}>
              Message
            </button>
          )}
        </div>

        <div className="profile-hero-flex-fill" aria-hidden />

        <div className="profile-hero-overlay-panel">
          <Avatar
            src={profile?.avatar_url}
            className="profile-hero-avatar"
            sx={{
              width: 148,
              height: 148,
              borderRadius: "10px",
              bgcolor: "rgba(255,255,255,0.92)",
              color: "#333",
            }}
          >
            {displayName?.charAt(0)?.toUpperCase() || "?"}
          </Avatar>

          <div className="profile-hero-info">
            <div className="profile-hero-name-row">
              <span className="profile-hero-name">{displayName}</span>
              {isVerified && (
                <VerifiedUserIcon className="profile-hero-verified" sx={{ fontSize: 28 }} />
              )}
            </div>
            {(educationLine || orgSubtitle) && (
              <div className="profile-hero-education">{educationLine || orgSubtitle}</div>
            )}
            <div className="profile-hero-about-heading">About me</div>
            {bioText ? (
              <div className="profile-hero-about">{bioText}</div>
            ) : (
              <div className="profile-hero-about profile-hero-about-empty">No bio yet.</div>
            )}
            <div className="profile-hero-links">
              {profile?.website && (
                <a
                  href={profile.website.startsWith("http") ? profile.website : `https://${profile.website}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="profile-hero-link"
                >
                  <LanguageIcon sx={{ fontSize: 18 }} />
                  {profile.website.replace(/^https?:\/\//, "").replace(/\/$/, "").slice(0, 36)}
                </a>
              )}
              {profile?.behance && (
                <a
                  href={profile.behance.startsWith("http") ? profile.behance : `https://${profile.behance}`}
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
                  href={profile.linkedin.startsWith("http") ? profile.linkedin : `https://${profile.linkedin}`}
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
            <div className="profile-hero-meta-block">
              <div className="profile-hero-meta-label">Language</div>
              <div className="profile-hero-meta-tags">
                {profile?.languages ? (
                  profile.languages.split(",").map((l, i) => (
                    <span key={i} className="profile-hero-meta-tag">
                      {l.trim()}
                    </span>
                  ))
                ) : (
                  <span className="profile-hero-meta-tag profile-hero-meta-tag-muted">—</span>
                )}
              </div>
            </div>
            <div className="profile-hero-meta-block">
              <div className="profile-hero-meta-label">Skills</div>
              <div className="profile-hero-meta-tags">
                {skillNames.length > 0 ? (
                  skillNames.slice(0, 12).map((sk, i) => (
                    <span key={i} className="profile-hero-meta-tag">
                      {sk}
                    </span>
                  ))
                ) : (
                  <span className="profile-hero-meta-tag profile-hero-meta-tag-muted">—</span>
                )}
              </div>
            </div>
            <div className="profile-hero-meta-block">
              <div className="profile-hero-meta-label">Interest areas</div>
              <div className="profile-hero-meta-tags">
                {interests.length > 0 ? (
                  interests.slice(0, 12).map((int, i) => (
                    <span key={i} className="profile-hero-meta-tag">
                      {int}
                    </span>
                  ))
                ) : (
                  <span className="profile-hero-meta-tag profile-hero-meta-tag-muted">—</span>
                )}
              </div>
            </div>
          </div>

          {isOwnProfile && onEditProfile && (
            <div className="profile-hero-edit-actions">
              <button type="button" className="profile-hero-btn-edit-panel" onClick={onEditProfile}>
                Edit Profile
              </button>
              <IconButton
                className="profile-hero-btn-edit-icon"
                onClick={onEditProfile}
                aria-label="Edit profile"
                size="small"
              >
                <EditIcon sx={{ fontSize: 22 }} />
              </IconButton>
            </div>
          )}
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
                    <div className="profile-service-thumbnails">{renderServiceThumbs(s)}</div>
                    <div className="profile-service-body">
                      <div className="profile-service-title">{getProfileServiceHeading(s)}</div>
                      <div className="profile-service-desc">
                        {s.short_description || s.description || s.experience_summary || ""}
                      </div>
                      <div className="profile-service-body-footer">
                        <div className="profile-service-fee">{formatPrice(s)}</div>
                        {profileOwnerId && s.id && (
                          <button
                            type="button"
                            className="profile-service-more-btn"
                            onClick={(e) => {
                              e.stopPropagation();
                              navigate(`/users/${profileOwnerId}/services/${s.id}`);
                            }}
                          >
                            More details
                          </button>
                        )}
                      </div>
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
