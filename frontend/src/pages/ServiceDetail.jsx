import React, { useCallback, useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import {
  Box,
  Button,
  CircularProgress,
  Container,
  IconButton,
  TextField,
  Tooltip,
} from "@mui/material";
import VerifiedUserIcon from "@mui/icons-material/VerifiedUser";
import FavoriteBorderIcon from "@mui/icons-material/FavoriteBorder";
import FavoriteIcon from "@mui/icons-material/Favorite";
import SendIcon from "@mui/icons-material/Send";
import PageLayout from "../shared/components/PageLayout";
import { StateContainer } from "../shared/components/ui/StateContainer";
import { profileApi } from "../features/profile/api/profile";
import useAuthStore from "../shared/stores/authStore";
import "./ServiceDetail.css";
import { parsePriceTypeTokens } from "../shared/utils/priceType";

const PREFILL_KEY = "aatist_chat_prefill";

function formatEducationLine(profile) {
  const parts = [profile?.school, profile?.faculty || profile?.major].filter(Boolean);
  const line = parts.join(" / ");
  if (profile?.year) {
    return `${profile.year}, ${line}`;
  }
  return line;
}

function formatHourly(service) {
  if (!parsePriceTypeTokens(service?.price_type).includes("hourly")) return "—";
  const min = service.price_min;
  const max = service.price_max;
  if (min == null && max == null) return "—";
  if (min != null && max != null && min !== max) return `€${min}–€${max} /hour`;
  return `€${min ?? max} /hour`;
}

function formatProject(service) {
  if (!parsePriceTypeTokens(service?.price_type).includes("project")) return "—";
  const min = service.price_min;
  const max = service.price_max;
  if (min == null && max == null) return "—";
  if (min != null && max != null && min !== max) return `€${min}–€${max} /project`;
  return `€${min ?? max} /project`;
}

function formatFlexible(service) {
  if (parsePriceTypeTokens(service?.price_type).includes("negotiable")) {
    return "Negotiable";
  }
  return "—";
}

export default function ServiceDetailPage() {
  const { userId, serviceId } = useParams();
  const navigate = useNavigate();
  const { isAuthenticated } = useAuthStore();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [profile, setProfile] = useState(null);
  const [service, setService] = useState(null);
  const [activeMediaIndex, setActiveMediaIndex] = useState(0);
  const [messageDraft, setMessageDraft] = useState("");
  const [favLoading, setFavLoading] = useState(false);
  const [favSaved, setFavSaved] = useState(false);

  const uid = userId ? Number(userId) : NaN;
  const sid = serviceId ? String(serviceId) : "";

  const load = useCallback(async () => {
    if (!Number.isFinite(uid)) {
      setError("Invalid profile");
      setLoading(false);
      return;
    }
    try {
      setLoading(true);
      setError(null);
      const res = await profileApi.getPublicProfile(uid);
      const data = res.data?.data;
      setProfile(data || null);
      const list = data?.services || [];
      const found = list.find((s) => String(s.id) === sid);
      setService(found || null);
      const urls = found?.media_urls?.length ? found.media_urls : [];
      setActiveMediaIndex(0);
      const name = data?.name || "there";
      const svcTitle = found?.title || found?.category || "your service";
      setMessageDraft(
        `Hi ${name.split(" ")[0]}, I'm interested in your ${svcTitle} service. I'm working on `
      );
    } catch (e) {
      console.error(e);
      setError("Could not load this service.");
      setProfile(null);
      setService(null);
    } finally {
      setLoading(false);
    }
  }, [uid, sid]);

  useEffect(() => {
    load();
  }, [load]);

  const mediaUrls = useMemo(() => {
    if (!service?.media_urls?.length) return [];
    return service.media_urls.filter(Boolean);
  }, [service]);

  const displayName = profile?.name || profile?.organization_name || "Talent";
  const categoryLabel = (service?.category || service?.title || "Service").toUpperCase();
  const verified = profile?.role_verified === true || profile?.is_verified_email === true;
  const educationLine = formatEducationLine(profile);
  const aboutText =
    service?.description?.trim() ||
    service?.short_description?.trim() ||
    service?.experience_summary?.trim() ||
    "This creator has not added a long description yet. Message them to learn more about their process and availability.";

  const mainSrc = mediaUrls[activeMediaIndex] || null;
  const totalMedia = Math.max(mediaUrls.length, 1);

  const sendMessage = () => {
    if (!isAuthenticated) {
      navigate(`/auth/login?redirect=${encodeURIComponent(window.location.pathname)}`);
      return;
    }
    try {
      sessionStorage.setItem(
        PREFILL_KEY,
        JSON.stringify({ forUser: uid, text: messageDraft.trim() })
      );
    } catch (_) {}
    navigate(`/messages?user=${uid}`);
  };

  const toggleFavorite = async () => {
    if (!isAuthenticated) {
      navigate(`/auth/login?redirect=${encodeURIComponent(window.location.pathname)}`);
      return;
    }
    setFavLoading(true);
    try {
      if (favSaved) {
        await profileApi.removeSavedItemByTarget("user", uid);
        setFavSaved(false);
      } else {
        await profileApi.saveItem("user", uid);
        setFavSaved(true);
      }
    } catch (err) {
      console.error(err);
    } finally {
      setFavLoading(false);
    }
  };

  return (
    <PageLayout noContainer>
      <div className="service-detail-page">
        <Container maxWidth="lg">
          <StateContainer
            loading={loading}
            error={error}
            empty={!service}
            emptyMessage="Service not found."
            onRetry={load}
          >
            <nav className="service-detail-breadcrumb" aria-label="Breadcrumb">
              <Link to="/talents">Hire Talent</Link>
              <span className="service-detail-breadcrumb-sep">/</span>
              <span>{categoryLabel}</span>
              <span className="service-detail-breadcrumb-sep">/</span>
              <span>{displayName}</span>
            </nav>

            <div className="service-detail-card">
              <div className="service-detail-card-header">
                <div className="service-detail-provider">
                  {profile?.avatar_url ? (
                    <img className="service-detail-avatar" src={profile.avatar_url} alt="" />
                  ) : (
                    <div className="service-detail-avatar-placeholder" aria-hidden />
                  )}
                  <div>
                    <div className="service-detail-name-row">
                      <h1 className="service-detail-name">{displayName}</h1>
                      {verified && (
                        <VerifiedUserIcon sx={{ color: "#137333", fontSize: 26 }} titleAccess="Verified" />
                      )}
                    </div>
                    {educationLine ? (
                      <p className="service-detail-subtitle">{educationLine}</p>
                    ) : (
                      <p className="service-detail-subtitle">Aalto talent</p>
                    )}
                    <span className="service-detail-badge-available">Available</span>
                  </div>
                </div>
                <div className="service-detail-actions">
                  <Button
                    variant="contained"
                    onClick={() => navigate(`/users/${uid}`)}
                    sx={{
                      bgcolor: "#111",
                      textTransform: "none",
                      borderRadius: "10px",
                      px: 2.5,
                      fontWeight: 600,
                      boxShadow: "none",
                      "&:hover": { bgcolor: "#333", boxShadow: "none" },
                    }}
                  >
                    View Profile
                  </Button>
                  <Tooltip title={favSaved ? "Remove from saved" : "Save talent"}>
                    <span>
                      <IconButton
                        onClick={toggleFavorite}
                        disabled={favLoading}
                        sx={{
                          border: "1px solid #dadce0",
                          borderRadius: "10px",
                          bgcolor: "#fff",
                        }}
                        aria-label={favSaved ? "Remove from saved" : "Save talent"}
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
              </div>

              <div className="service-detail-body">
                <div className="service-detail-gallery-col">
                  <div className="service-detail-main-media">
                    {mainSrc ? (
                      <img src={mainSrc} alt="" />
                    ) : (
                      <Box sx={{ width: "100%", height: "100%", bgcolor: "#dadce0" }} />
                    )}
                    <span className="service-detail-media-counter">
                      {mediaUrls.length ? `${activeMediaIndex + 1}/${mediaUrls.length}` : "—"}
                    </span>
                  </div>
                  {mediaUrls.length > 1 && (
                    <div className="service-detail-thumbs">
                      {mediaUrls.map((url, i) => (
                        <button
                          key={url + i}
                          type="button"
                          className={`service-detail-thumb ${i === activeMediaIndex ? "active" : ""}`}
                          onClick={() => setActiveMediaIndex(i)}
                          aria-label={`Image ${i + 1}`}
                        >
                          <img src={url} alt="" />
                        </button>
                      ))}
                    </div>
                  )}
                </div>

                <div className="service-detail-info-col">
                  <h2 className="service-detail-service-title">
                    {service?.title || service?.category || "Service"}
                  </h2>

                  <h3 className="service-detail-section-title">About this service</h3>
                  <div className="service-detail-about">{aboutText}</div>

                  <div className="service-detail-pricing-row">
                    <div className="service-detail-price-card">
                      <div className="service-detail-price-label">Hourly</div>
                      <div className="service-detail-price-value">{formatHourly(service)}</div>
                    </div>
                    <div className="service-detail-price-card">
                      <div className="service-detail-price-label">Project</div>
                      <div className="service-detail-price-value">{formatProject(service)}</div>
                    </div>
                    <div className="service-detail-price-card">
                      <div className="service-detail-price-label">Flexible</div>
                      <div className="service-detail-price-negotiable">{formatFlexible(service)}</div>
                    </div>
                  </div>

                  <div className="service-detail-message-block">
                    <h3 className="service-detail-message-title">Send a message to {displayName.split(" ")[0]}</h3>
                    <p className="service-detail-message-sub">ask about pricing, timeline, or project ideas.</p>
                    <TextField
                      fullWidth
                      multiline
                      minRows={4}
                      value={messageDraft}
                      onChange={(e) => setMessageDraft(e.target.value)}
                      placeholder="Write your message…"
                      sx={{
                        "& .MuiOutlinedInput-root": {
                          bgcolor: "#fff",
                          borderRadius: "10px",
                        },
                      }}
                    />
                    <div className="service-detail-message-actions">
                      <Button
                        variant="contained"
                        endIcon={<SendIcon />}
                        onClick={sendMessage}
                        sx={{
                          textTransform: "none",
                          borderRadius: "10px",
                          px: 2.5,
                          fontWeight: 600,
                        }}
                      >
                        Send Message
                      </Button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </StateContainer>
        </Container>
      </div>
    </PageLayout>
  );
}
