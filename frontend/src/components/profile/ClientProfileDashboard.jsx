import React, { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import AddIcon from "@mui/icons-material/Add";
import EditOutlinedIcon from "@mui/icons-material/EditOutlined";
import PeopleOutlineIcon from "@mui/icons-material/PeopleOutline";
import PauseCircleOutlineIcon from "@mui/icons-material/PauseCircleOutline";
import PlayCircleOutlineIcon from "@mui/icons-material/PlayCircleOutline";
import ChevronRightIcon from "@mui/icons-material/ChevronRight";
import { Box, CircularProgress, Typography } from "@mui/material";
import { profileApi, portfolioApi } from "../../features/profile/api/profile";
import { opportunitiesApi } from "../../features/opportunities/api/opportunities";
import { talentDisplayName } from "../../shared/utils/displayName";
import { getProfileServiceHeading } from "../../constants/serviceCategories";
import "./clientProfile.css";

function formatPostDate(dateString) {
  if (!dateString) return "—";
  const d = new Date(dateString);
  if (Number.isNaN(d.getTime())) return "—";
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}/${m}/${day}`;
}

function opportunityUiStatus(status) {
  const s = String(status || "").toLowerCase();
  if (s === "published" || s === "active") return { label: "Active", live: true };
  if (s === "closed" || s === "archived") return { label: "Closed", live: false };
  if (s === "draft") return { label: "Draft", live: false };
  return { label: s ? s : "—", live: false };
}

function messageSnippet(msg) {
  if (!msg || !String(msg).trim()) return "—";
  const t = String(msg).trim();
  return t.length > 100 ? `${t.slice(0, 100)}…` : t;
}

async function fetchPublicProfilesMap(userIds) {
  const uniq = [...new Set(userIds)].filter((id) => id != null);
  const map = {};
  await Promise.all(
    uniq.map(async (id) => {
      try {
        const res = await profileApi.getPublicProfile(id);
        map[id] = res.data?.data ?? null;
      } catch {
        map[id] = null;
      }
    })
  );
  return map;
}

function ClientPostCard({ opportunity, applications, profilesById, onOpen }) {
  const o = opportunity || {};
  const id = o.id;
  const title = o.title || "Untitled brief";
  const org = o.organization || "—";
  const dateSrc = o.published_at || o.created_at;
  const { label: statusLabel, live } = opportunityUiStatus(o.status);
  const apps = applications || [];
  const preview = apps.slice(0, 2);
  const more = apps.length > 2 ? apps.length - 2 : 0;

  return (
    <div className="client-post-card">
      <div className="client-post-card-top">
        <span className="client-post-card-org">{org}</span>
        <button type="button" className="client-post-card-edit" onClick={() => onOpen(id)}>
          <EditOutlinedIcon sx={{ fontSize: 16 }} />
          Edit
        </button>
      </div>
      <div className="client-post-card-body">
        <button
          type="button"
          className="client-post-card-title"
          onClick={() => onOpen(id)}
          style={{ background: "none", border: "none", padding: 0, cursor: "pointer", textAlign: "left" }}
        >
          {title}
        </button>
        <div className="client-post-card-badges">
          <span className="client-post-badge-date">{formatPostDate(dateSrc)}</span>
          <span className={`client-post-badge-status ${live ? "" : "inactive"}`}>
            {live ? (
              <PauseCircleOutlineIcon sx={{ fontSize: 16 }} />
            ) : (
              <PlayCircleOutlineIcon sx={{ fontSize: 16 }} />
            )}
            {statusLabel}
          </span>
          <span className="client-post-card-apps-count">
            <PeopleOutlineIcon sx={{ fontSize: 18 }} />
            {apps.length}
          </span>
        </div>
        {preview.length > 0 ? (
          <div className="client-post-applicants">
            {preview.map((app) => {
              const uid = app.user_id;
              const p = uid != null ? profilesById[uid] : null;
              const nm = p ? talentDisplayName(p) || p.name || `User ${uid}` : `Applicant`;
              const av = p?.avatar_url;
              return (
                <div key={app.id} className="client-post-applicant-row">
                  {av ? (
                    <img className="client-post-applicant-avatar" src={av} alt="" />
                  ) : (
                    <div className="client-post-applicant-avatar" aria-hidden />
                  )}
                  <div style={{ minWidth: 0 }}>
                    <div className="client-post-applicant-name">{nm}</div>
                    <div className="client-post-applicant-msg">{messageSnippet(app.message)}</div>
                  </div>
                </div>
              );
            })}
            {more > 0 ? (
              <button type="button" className="client-post-applicants-more" onClick={() => onOpen(id)}>
                + {more} more
              </button>
            ) : null}
          </div>
        ) : null}
      </div>
    </div>
  );
}

function SavedRow({ thumb, line1, line2, onClick }) {
  return (
    <button type="button" className="client-dash-saved-item" onClick={onClick}>
      {thumb ? <img className="client-dash-saved-thumb" src={thumb} alt="" /> : <div className="client-dash-saved-thumb" aria-hidden />}
      <div className="client-dash-saved-text">
        <div className="client-dash-saved-line1">{line1}</div>
        <div className="client-dash-saved-line2">{line2}</div>
      </div>
      <ChevronRightIcon className="client-dash-saved-chevron" fontSize="small" />
    </button>
  );
}

export default function ClientProfileDashboard() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [opportunities, setOpportunities] = useState([]);
  const [appsByOppId, setAppsByOppId] = useState({});
  const [profilesById, setProfilesById] = useState({});
  const [savedProjects, setSavedProjects] = useState([]);
  const [savedUsers, setSavedUsers] = useState([]);

  const openOpp = useCallback((id) => navigate(`/opportunities/${id}`), [navigate]);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [oppRes, savedRes] = await Promise.all([
        opportunitiesApi.getMyOpportunities({ limit: 50, page: 1 }),
        profileApi.getSavedItems(),
      ]);

      const rawOpps = oppRes.data?.data;
      const opps = Array.isArray(rawOpps) ? rawOpps : [];
      setOpportunities(opps);

      const savedItems = savedRes.data?.data?.items || [];
      const list = Array.isArray(savedItems) ? savedItems : [];
      const projectIds = list.filter((x) => x.item_type === "project").map((x) => x.item_id);
      const userIds = list.filter((x) => x.item_type === "user").map((x) => x.item_id);

      const appResults = await Promise.all(
        opps.map(async (opp) => {
          try {
            const r = await opportunitiesApi.getOpportunityApplications(opp.id, { limit: 8, page: 1 });
            const apps = r.data?.data?.applications || r.data?.data?.items || [];
            return [opp.id, Array.isArray(apps) ? apps : []];
          } catch {
            return [opp.id, []];
          }
        })
      );
      const appsMap = Object.fromEntries(appResults);
      setAppsByOppId(appsMap);

      const applicantIds = [...new Set(Object.values(appsMap).flat().map((a) => a.user_id))];
      const profileIds = [...new Set([...applicantIds, ...userIds])];
      const profiles = await fetchPublicProfilesMap(profileIds);
      setProfilesById(profiles);

      const projMeta = await Promise.all(
        projectIds.slice(0, 12).map(async (pid) => {
          try {
            const r = await portfolioApi.getPortfolioById(pid);
            const item = r.data?.data;
            const title = item?.title || `Project ${pid}`;
            const cover = item?.cover_image_url;
            const ownerId = item?.user_id;
            let ownerName = "";
            if (ownerId && !profiles[ownerId]) {
              try {
                const ur = await profileApi.getPublicProfile(ownerId);
                const u = ur.data?.data;
                ownerName = u ? talentDisplayName(u) || u.name || "" : "";
              } catch {
                ownerName = "";
              }
            } else if (ownerId && profiles[ownerId]) {
              const u = profiles[ownerId];
              ownerName = talentDisplayName(u) || u.name || "";
            }
            return { id: pid, title, cover, ownerName };
          } catch {
            return { id: pid, title: `Project ${pid}`, cover: null, ownerName: "" };
          }
        })
      );

      const userMeta = userIds.slice(0, 12).map((uid) => {
        const u = profiles[uid];
        const services = u?.services || [];
        const line1 =
          services.length > 0 ? getProfileServiceHeading(services[0]) : talentDisplayName(u) || u?.name || `Talent`;
        const line2 = talentDisplayName(u) || u?.name || "";
        return { id: uid, line1, line2, thumb: u?.avatar_url || null };
      });

      setSavedProjects(projMeta);
      setSavedUsers(userMeta);
    } catch (e) {
      console.error(e);
      setOpportunities([]);
      setAppsByOppId({});
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const nPosts = opportunities.length;
  const nPort = savedProjects.length;
  const nSvc = savedUsers.length;

  return (
    <div className="client-dashboard">
      <div>
        <div className="client-dash-section-head">
          <div>
            <span className="client-dash-section-head-title">
              My Posts
              <span className="client-dash-section-head-count">[{nPosts}]</span>
            </span>
          </div>
          <button type="button" className="client-dash-post-brief-btn" onClick={() => navigate("/opportunities")}>
            <AddIcon sx={{ fontSize: 18 }} />
            Post a Brief
          </button>
        </div>
        <div className="client-dash-panel">
          {loading ? (
            <Box sx={{ display: "flex", justifyContent: "center", py: 6 }}>
              <CircularProgress size={36} />
            </Box>
          ) : nPosts === 0 ? (
            <div className="client-dash-empty">
              <Typography variant="body1" color="text.secondary" gutterBottom>
                No briefs posted yet.
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                Post a project brief to reach Aalto talent.
              </Typography>
              <button type="button" className="client-dash-post-brief-btn" onClick={() => navigate("/opportunities")}>
                <AddIcon sx={{ fontSize: 18 }} />
                Post a Brief
              </button>
            </div>
          ) : (
            <div className="client-dash-post-grid">
              {opportunities.map((opp) => (
                <ClientPostCard
                  key={opp.id}
                  opportunity={opp}
                  applications={appsByOppId[opp.id] || []}
                  profilesById={profilesById}
                  onOpen={openOpp}
                />
              ))}
            </div>
          )}
        </div>
      </div>

      <aside className="client-dash-saved-wrap">
        <div className="client-dash-saved-head">Saved</div>
        <div className="client-dash-saved-body">
          <div className="client-dash-saved-sub">
            <div className="client-dash-saved-sub-title">
              Portfolio <span className="client-dash-section-head-count">[{nPort}]</span>
            </div>
            {nPort === 0 ? (
              <Typography variant="body2" color="text.secondary" sx={{ px: 1, py: 1 }}>
                Save portfolio projects while browsing to see them here.
              </Typography>
            ) : (
              <ul className="client-dash-saved-list">
                {savedProjects.slice(0, 2).map((p) => (
                  <li key={p.id}>
                    <SavedRow
                      thumb={p.cover}
                      line1={p.title}
                      line2={p.ownerName || " "}
                      onClick={() => navigate(`/portfolio/${p.id}`)}
                    />
                  </li>
                ))}
              </ul>
            )}
            {nPort > 2 ? (
              <button type="button" className="client-dash-saved-more" onClick={() => navigate("/talents")}>
                + {nPort - 2} more
              </button>
            ) : null}
          </div>

          <div className="client-dash-saved-sub" style={{ borderTop: "1px solid #e2e8e4" }}>
            <div className="client-dash-saved-sub-title">
              Service <span className="client-dash-section-head-count">[{nSvc}]</span>
            </div>
            {nSvc === 0 ? (
              <Typography variant="body2" color="text.secondary" sx={{ px: 1, py: 1 }}>
                Save talent profiles from service pages to see them here.
              </Typography>
            ) : (
              <ul className="client-dash-saved-list">
                {savedUsers.slice(0, 2).map((u) => (
                  <li key={u.id}>
                    <SavedRow
                      thumb={u.thumb}
                      line1={u.line1}
                      line2={u.line2}
                      onClick={() => navigate(`/users/${u.id}`)}
                    />
                  </li>
                ))}
              </ul>
            )}
            {nSvc > 2 ? (
              <button type="button" className="client-dash-saved-more" onClick={() => navigate("/talents")}>
                + {nSvc - 2} more
              </button>
            ) : null}
          </div>
        </div>
      </aside>
    </div>
  );
}
