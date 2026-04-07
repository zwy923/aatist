import React from "react";
import { Link } from "react-router-dom";
import PlaceOutlinedIcon from "@mui/icons-material/PlaceOutlined";
import ArrowForwardIcon from "@mui/icons-material/ArrowForward";
import FavoriteBorderIcon from "@mui/icons-material/FavoriteBorder";
import SavedButton from "./SavedButton";

function formatTimeAgo(dateString) {
  if (!dateString) return "";
  const past = new Date(dateString);
  const now = new Date();
  const diffMs = now - past;
  if (diffMs < 0) return "";
  const diffM = Math.floor(diffMs / 60000);
  const diffH = Math.floor(diffMs / 3600000);
  const diffD = Math.floor(diffMs / 86400000);
  if (diffM < 1) return "just now";
  if (diffM < 60) return `${diffM}m ago`;
  if (diffH < 48) return `${diffH}h ago`;
  if (diffD < 30) return `${diffD}d ago`;
  return "";
}

function formatCalendarDate(dateString) {
  if (!dateString) return "";
  const d = new Date(dateString);
  if (Number.isNaN(d.getTime())) return "";
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  const yyyy = d.getFullYear();
  return `${mm}/${dd}/${yyyy}`;
}

function formatBudgetLine(o) {
  const v = o.budget_value;
  const t = (o.budget_type || "").toLowerCase();
  if (v == null || v === "") return null;
  const n = Number(v);
  if (Number.isNaN(n)) return null;
  if (t === "hourly") return `€${n} /hr`;
  return `€${n}`;
}

export default function OpportunityCard({ opportunity: raw }) {
  const o = raw || {};
  const id = o.id;
  const title = o.title || "Untitled";
  const organization = o.organization || "Company";
  const category = o.category || "";
  const location = o.location || "—";
  const publishedAt = o.published_at;
  const urgent = Boolean(o.urgent);
  const savedByMe = o.is_favorite;

  const priceLine = formatBudgetLine(o);
  const negotiable = o.budget_value == null;

  return (
    <div className="opp-card">
      {urgent ? <div className="opp-card-urgent">• Urgent</div> : null}

      <Link to={`/opportunities/${id}`} style={{ textDecoration: "none", color: "inherit" }}>
        <div className="opp-card-head">
          <div className="opp-card-brand">
            <div className="opp-card-logo" aria-hidden />
            <div>
              <div className="opp-card-org">{organization}</div>
              {category ? <div className="opp-card-cat">{category}</div> : null}
            </div>
          </div>
          <div className="opp-card-meta">
            <div className="opp-card-ago">{formatTimeAgo(publishedAt) || "—"}</div>
            <div className="opp-card-date">{formatCalendarDate(publishedAt)}</div>
          </div>
        </div>

        <div className="opp-card-body">
          <div className="opp-card-loc">
            <PlaceOutlinedIcon sx={{ fontSize: 18, opacity: 0.75 }} />
            <span>{location}</span>
          </div>
          <h2 className="opp-card-title">{title}</h2>
          <div className="opp-card-badges">
            {priceLine ? <span className="opp-badge-price">{priceLine}</span> : null}
            {negotiable ? <span className="opp-badge-negotiable">Price Negotiable</span> : null}
          </div>
        </div>
      </Link>

      <div className="opp-card-foot">
        <Link to={`/opportunities/${id}`} className="opp-card-interested">
          <FavoriteBorderIcon sx={{ fontSize: 20 }} />
          I am Interested
        </Link>
        <div className="opp-card-foot-right">
          <SavedButton
            targetId={id}
            initialSaved={savedByMe}
            iconSet="favorite"
            size="small"
            sx={{ color: "#5f6368" }}
          />
          <ArrowForwardIcon sx={{ fontSize: 22 }} />
        </div>
      </div>
    </div>
  );
}
