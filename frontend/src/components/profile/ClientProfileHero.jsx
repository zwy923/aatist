import React from "react";
import { Avatar } from "@mui/material";
import CalendarTodayOutlinedIcon from "@mui/icons-material/CalendarTodayOutlined";
import EditIcon from "@mui/icons-material/Edit";
import ImageOutlinedIcon from "@mui/icons-material/ImageOutlined";
import VerifiedUserIcon from "@mui/icons-material/VerifiedUser";
import "./clientProfile.css";

function formatJoined(createdAt) {
  if (!createdAt) return null;
  const d = new Date(createdAt);
  if (Number.isNaN(d.getTime())) return null;
  const month = d.toLocaleString("en-US", { month: "long" });
  const year = d.getFullYear();
  return `Joined in ${month} ${year}`;
}

/**
 * Organization / client profile hero — different from student talent layout.
 */
export default function ClientProfileHero({
  profile,
  bannerUrl,
  isOwnProfile,
  onChangeBackground,
  onEditProfile,
  onMessage,
}) {
  const contactName = (profile?.name || "").trim() || "Client";
  const email = (profile?.email || "").trim();
  const orgName = (profile?.organization_name || "").trim();
  const industry = (profile?.professional_interests || "").trim().split(",")[0]?.trim() || "";
  const roleTitle = (profile?.contact_title || "").trim();
  const bioText = (profile?.organization_bio || profile?.bio || "").trim();
  const isVerified = profile?.is_verified_email || profile?.role_verified;
  const joined = formatJoined(profile?.created_at);

  return (
    <div className="profile-hero profile-hero--client">
      <div className="profile-hero--client-watermark" aria-hidden>
        CLIENT
      </div>
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
          {contactName.charAt(0).toUpperCase() || "?"}
        </Avatar>

        <div className="profile-hero--client-info">
          <div className="profile-hero-name-row">
            <span className="profile-hero--client-name">{contactName}</span>
            {isVerified && <VerifiedUserIcon className="profile-hero-verified" sx={{ fontSize: 28 }} />}
          </div>
          {email ? <div className="profile-hero--client-email">{email}</div> : null}

          {joined ? (
            <div className="profile-hero--client-meta-row">
              <span className="profile-hero--client-meta-item">
                <CalendarTodayOutlinedIcon sx={{ fontSize: 18, opacity: 0.85 }} />
                {joined}
              </span>
            </div>
          ) : null}

          <div className="profile-hero--client-columns">
            <div>
              <div className="profile-hero--client-col-label">Company / Organization</div>
              <div className="profile-hero--client-col-value">{orgName || "—"}</div>
            </div>
            <div>
              <div className="profile-hero--client-col-label">Industry</div>
              <div className="profile-hero--client-col-value">{industry || "—"}</div>
            </div>
            <div>
              <div className="profile-hero--client-col-label">Role</div>
              <div className="profile-hero--client-col-value">{roleTitle || "—"}</div>
            </div>
          </div>

          <div className="profile-hero--client-about-label">About us</div>
          {bioText ? (
            <div className="profile-hero--client-about">{bioText}</div>
          ) : (
            <div className="profile-hero--client-about profile-hero--client-about-empty">No description yet.</div>
          )}
        </div>

        <div className="profile-hero-edit-actions">
          {isOwnProfile && onEditProfile && (
            <button type="button" className="profile-hero-btn-edit-client" onClick={onEditProfile}>
              <EditIcon sx={{ fontSize: 18 }} />
              Edit Profile
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
