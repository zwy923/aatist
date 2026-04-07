import React, { useCallback, useEffect, useState, useMemo, useRef } from "react";
import { useNavigate, useLocation, useSearchParams } from "react-router-dom";
import {
  Box,
  Dialog,
  DialogContent,
  DialogTitle,
  Divider,
  IconButton,
  Typography,
} from "@mui/material";
import CloseIcon from "@mui/icons-material/Close";
import AssignmentIcon from "@mui/icons-material/Assignment";
import AddIcon from "@mui/icons-material/Add";
import { profileApi, portfolioApi } from "../features/profile/api/profile";
import { opportunitiesApi } from "../features/opportunities/api/opportunities";
import { useAuth } from "../features/auth/hooks/useAuth";
import { useProfile } from "../features/profile/hooks/useProfile";
import { StateContainer } from "../shared/components/ui/StateContainer";
import BasicInfoSection from "../components/profile/BasicInfoSection";
import ServicesSection from "../components/profile/ServicesSection";
import PortfolioSection from "../components/profile/PortfolioSection";
import MyProjectsSection from "../components/profile/MyProjectsSection";
import SecuritySettings from "../components/profile/SecuritySettings";
import PageLayout from "../shared/components/PageLayout";
import ProfileView from "../components/profile/ProfileView";

export default function Profile() {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const { isAuthenticated } = useAuth();
  const { user, loading: profileLoading, updateProfile, uploadAvatar } = useProfile();

  const isStudentRole = useMemo(
    () => user?.role?.toLowerCase?.() === "student" || user?.role?.toLowerCase?.() === "alumni",
    [user?.role]
  );

  const idFromUrl = searchParams.get("id");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [profile, setProfile] = useState(null);
  const [portfolio, setPortfolio] = useState([]);
  const [myProjects, setMyProjects] = useState([]);
  const [services, setServices] = useState([]);

  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [servicesDialogOpen, setServicesDialogOpen] = useState(false);
  const [servicesEditTriggerId, setServicesEditTriggerId] = useState(null);
  const [portfolioDialogOpen, setPortfolioDialogOpen] = useState(false);
  const [portfolioEditTriggerId, setPortfolioEditTriggerId] = useState(null);
  const [bannerObjectUrl, setBannerObjectUrl] = useState(null);
  const bannerInputRef = useRef(null);

  const loadProfile = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await profileApi.getMyProfile();
      setProfile(response.data.data);
    } catch (err) {
      console.error("Failed to load profile:", err);
      setError("Failed to load profile data");
    } finally {
      setLoading(false);
    }
  }, []);

  const loadPortfolio = useCallback(async () => {
    try {
      const response = await portfolioApi.getMyPortfolio();
      const data = response.data.data;
      setPortfolio(data?.projects || data?.items || []);
    } catch (err) {
      console.error("Failed to load portfolio:", err);
    }
  }, []);

  const loadMyProjects = useCallback(async () => {
    try {
      const response = await opportunitiesApi.getMyOpportunities();
      const data = response.data.data;
      setMyProjects(Array.isArray(data) ? data : data?.items || []);
    } catch (err) {
      console.error("Failed to load my projects:", err);
      setMyProjects([]);
    }
  }, []);

  const loadServices = useCallback(async () => {
    try {
      const response = await profileApi.getServices();
      setServices(response.data?.data?.services || []);
    } catch (err) {
      setServices([]);
    }
  }, []);

  const consumePortfolioEditTrigger = useCallback(() => setPortfolioEditTriggerId(null), []);

  useEffect(() => {
    if (!isAuthenticated) {
      navigate("/auth/login");
      return;
    }
    loadProfile();
    loadServices();
    loadPortfolio();
    if (!isStudentRole) loadMyProjects();
  }, [isAuthenticated, navigate, loadProfile, loadServices, loadPortfolio, loadMyProjects, isStudentRole]);

  useEffect(() => {
    const st = location.state;
    if (!st || typeof st !== "object") return;
    const clear = () => navigate(location.pathname, { replace: true, state: {} });
    if (st.openProfileEdit) {
      setEditDialogOpen(true);
      clear();
    } else if (st.openBannerPicker) {
      setTimeout(() => bannerInputRef.current?.click(), 0);
      clear();
    } else if (st.openServices) {
      setServicesDialogOpen(true);
      if (st.editServiceId != null) setServicesEditTriggerId(st.editServiceId);
      clear();
    } else if (st.openPortfolio) {
      setPortfolioDialogOpen(true);
      clear();
    } else if (st.editPortfolioId != null) {
      setPortfolioDialogOpen(true);
      setPortfolioEditTriggerId(st.editPortfolioId);
      clear();
    }
  }, [location.state, location.pathname, navigate]);

  useEffect(() => {
    if (!profile || !idFromUrl) return;
    const targetId = Number(idFromUrl);
    const myId = profile.id ?? profile.user_id;
    if (!Number.isNaN(targetId) && myId != null && Number(myId) !== targetId) {
      navigate(`/users/${targetId}`, { replace: true });
    }
  }, [profile, idFromUrl, navigate]);

  useEffect(() => {
    return () => {
      if (bannerObjectUrl) URL.revokeObjectURL(bannerObjectUrl);
    };
  }, [bannerObjectUrl]);

  const handleProfileUpdate = async (updatedData) => {
    const result = await updateProfile(updatedData);
    if (result.success) {
      setProfile(result.data);
      return { success: true };
    }
    return { success: false, error: result.error };
  };

  const handleAvatarUpload = async (file) => {
    const result = await uploadAvatar(file);
    if (result.success) {
      setProfile((prev) => ({ ...prev, avatar_url: result.avatarUrl }));
      return { success: true };
    }
    return { success: false, error: result.error };
  };

  const handleDeletePortfolioItem = async (id) => {
    try {
      await portfolioApi.deletePortfolioItem(id);
      setPortfolio((prev) => prev.filter((item) => item.id !== id));
      return { success: true };
    } catch (err) {
      return { success: false, error: err.message || "Delete failed" };
    }
  };

  const handlePasswordChange = async (currentPassword, newPassword) => {
    try {
      await profileApi.changePassword(currentPassword, newPassword);
      return { success: true };
    } catch (err) {
      return { success: false, error: err.message || "Password change failed" };
    }
  };

  const onBannerFileChange = async (e) => {
    const file = e.target.files?.[0];
    if (!file) return;
    e.target.value = "";
    try {
      const res = await profileApi.uploadProfileBanner(file);
      const payload = res.data?.data;
      const url = payload?.banner_url ?? payload?.user?.banner_url;
      setBannerObjectUrl((prev) => {
        if (prev) URL.revokeObjectURL(prev);
        return null;
      });
      if (url) {
        setProfile((p) => (p ? { ...p, banner_url: url } : p));
      } else if (payload?.user) {
        setProfile(payload.user);
      }
    } catch (err) {
      console.error("Banner upload failed:", err);
    }
  };

  const bannerDisplayUrl = bannerObjectUrl || profile?.banner_url;

  const clientAlternate = !isStudentRole ? (
    <Box sx={{ mb: 6 }}>
      <div className="profile-section-header">
        <div className="profile-section-title">
          <AssignmentIcon sx={{ color: "#0095D9", fontSize: 24 }} />
          My Projects
        </div>
        <div className="profile-section-actions">
          <button
            type="button"
            className="profile-section-add-btn"
            onClick={() => navigate("/opportunities")}
          >
            <AddIcon sx={{ fontSize: 18 }} /> Post a Project Brief
          </button>
        </div>
      </div>
      <MyProjectsSection items={myProjects} hideOuterChrome />
    </Box>
  ) : null;

  return (
    <PageLayout noContainer>
      <input
        ref={bannerInputRef}
        type="file"
        accept="image/jpeg,image/png,image/webp"
        style={{ display: "none" }}
        onChange={onBannerFileChange}
      />

      <StateContainer loading={loading || profileLoading} error={error} onRetry={loadProfile}>
        <ProfileView
          profile={profile}
          profileUserId={profile?.id}
          portfolio={portfolio}
          services={services}
          isOwnProfile
          bannerUrl={bannerDisplayUrl}
          onChangeBackground={() => bannerInputRef.current?.click()}
          onEditProfile={() => setEditDialogOpen(true)}
          onAddService={isStudentRole ? () => setServicesDialogOpen(true) : undefined}
          onAddProject={isStudentRole ? () => setPortfolioDialogOpen(true) : undefined}
          onPortfolioEdit={
            isStudentRole
              ? (projectId) => {
                  setPortfolioEditTriggerId(projectId);
                  setPortfolioDialogOpen(true);
                }
              : undefined
          }
          alternateContent={clientAlternate}
          showServicesAndPortfolio={isStudentRole}
        />

        <Dialog
          open={editDialogOpen}
          onClose={() => setEditDialogOpen(false)}
          maxWidth="md"
          fullWidth
          scroll="paper"
          PaperProps={{ sx: { borderRadius: 3 } }}
        >
          <DialogTitle sx={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
            Edit profile
            <IconButton aria-label="Close" onClick={() => setEditDialogOpen(false)} size="small">
              <CloseIcon />
            </IconButton>
          </DialogTitle>
          <DialogContent dividers>
            <BasicInfoSection
              profile={profile}
              isStudentRole={isStudentRole}
              onUpdate={handleProfileUpdate}
              onAvatarUpload={handleAvatarUpload}
              compact
            />
            <Divider sx={{ my: 3 }} />
            <Typography variant="subtitle2" fontWeight={600} color="text.secondary" gutterBottom>
              Account security
            </Typography>
            <SecuritySettings
              profile={profile}
              onPasswordChange={handlePasswordChange}
              onProfileUpdate={handleProfileUpdate}
            />
          </DialogContent>
        </Dialog>

        <Dialog
          open={servicesDialogOpen}
          onClose={() => {
            setServicesDialogOpen(false);
            setServicesEditTriggerId(null);
          }}
          maxWidth="md"
          fullWidth
          scroll="paper"
          PaperProps={{ sx: { borderRadius: 3 } }}
        >
          <DialogTitle sx={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
            Manage services
            <IconButton
              aria-label="Close"
              onClick={() => {
                setServicesDialogOpen(false);
                setServicesEditTriggerId(null);
              }}
              size="small"
            >
              <CloseIcon />
            </IconButton>
          </DialogTitle>
          <DialogContent>
            <ServicesSection
              hideIntro
              triggerEditForId={servicesEditTriggerId}
              onTriggerEditConsumed={() => setServicesEditTriggerId(null)}
              onSave={() => {
                loadServices();
              }}
            />
          </DialogContent>
        </Dialog>

        <Dialog
          open={portfolioDialogOpen}
          onClose={() => {
            setPortfolioDialogOpen(false);
            setPortfolioEditTriggerId(null);
          }}
          maxWidth="lg"
          fullWidth
          scroll="paper"
          PaperProps={{ sx: { borderRadius: 3 } }}
        >
          <DialogTitle sx={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
            Manage portfolio
            <IconButton
              aria-label="Close"
              onClick={() => {
                setPortfolioDialogOpen(false);
                setPortfolioEditTriggerId(null);
              }}
              size="small"
            >
              <CloseIcon />
            </IconButton>
          </DialogTitle>
          <DialogContent>
            <PortfolioSection
              hideIntro
              items={portfolio}
              onPortfolioRefresh={loadPortfolio}
              onDelete={handleDeletePortfolioItem}
              triggerEditForId={portfolioEditTriggerId}
              onTriggerEditConsumed={consumePortfolioEditTrigger}
            />
          </DialogContent>
        </Dialog>
      </StateContainer>
    </PageLayout>
  );
}
