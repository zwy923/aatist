import React, { useCallback, useEffect, useState, useMemo } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import {
  Box,
  Paper,
  Stack,
  Tab,
  Tabs,
  Typography,
} from "@mui/material";
import PersonIcon from "@mui/icons-material/Person";
import WorkIcon from "@mui/icons-material/Work";
import DesignServicesIcon from "@mui/icons-material/DesignServices";
import BookmarkIcon from "@mui/icons-material/Bookmark";
import SettingsIcon from "@mui/icons-material/Settings";
import AssignmentIcon from "@mui/icons-material/Assignment";
import { profileApi, portfolioApi } from "../features/profile/api/profile";
import { opportunitiesApi } from "../features/opportunities/api/opportunities";
import { useAuth } from "../features/auth/hooks/useAuth";
import { useProfile } from "../features/profile/hooks/useProfile";
import { StateContainer } from "../shared/components/ui/StateContainer";
import ProfileEditHeader from "../components/profile/ProfileEditHeader";
import BasicInfoSection from "../components/profile/BasicInfoSection";
import ServicesSection from "../components/profile/ServicesSection";
import PortfolioSection from "../components/profile/PortfolioSection";
import MyProjectsSection from "../components/profile/MyProjectsSection";
import SavedItemsSection from "../components/profile/SavedItemsSection";
import SecuritySettings from "../components/profile/SecuritySettings";
import PageLayout from "../shared/components/PageLayout";

const STUDENT_TABS = [
  { id: "basic", label: "Basic Info", icon: <PersonIcon /> },
  { id: "services", label: "Services", icon: <DesignServicesIcon /> },
  { id: "portfolio", label: "Portfolio", icon: <WorkIcon /> },
];

const CLIENT_TABS = [
  { id: "basic", label: "Basic Info", icon: <PersonIcon /> },
  { id: "my-projects", label: "My Projects", icon: <AssignmentIcon /> },
];

const OTHER_TABS = [
  { id: "saved", label: "Saved", icon: <BookmarkIcon /> },
  { id: "settings", label: "Settings", icon: <SettingsIcon /> },
];

const getAllTabs = (isStudent) => {
  const editTabs = isStudent ? STUDENT_TABS : CLIENT_TABS;
  return [...editTabs, ...OTHER_TABS];
};

export default function Profile() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { isAuthenticated } = useAuth();
  const { user, loading: profileLoading, updateProfile, uploadAvatar } = useProfile();

  const isStudentRole = useMemo(
    () => user?.role?.toLowerCase?.() === "student" || user?.role?.toLowerCase?.() === "alumni",
    [user?.role]
  );

  const tabFromUrl = searchParams.get("tab");
  const idFromUrl = searchParams.get("id");
  const ALL_TABS = useMemo(() => getAllTabs(isStudentRole), [isStudentRole]);
  const validTabs = ALL_TABS.map((t) => t.id);
  const initialTab = validTabs.includes(tabFromUrl) ? tabFromUrl : validTabs[0] || "basic";

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [activeTab, setActiveTab] = useState(initialTab);
  const [profile, setProfile] = useState(null);
  const [portfolio, setPortfolio] = useState([]);
  const [myProjects, setMyProjects] = useState([]);
  const [services, setServices] = useState([]);
  const [savedItems, setSavedItems] = useState([]);

  useEffect(() => {
    const newTab = searchParams.get("tab");
    if (newTab && validTabs.includes(newTab) && newTab !== activeTab) {
      setActiveTab(newTab);
    }
  }, [searchParams]);

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

  const loadSavedItems = useCallback(async () => {
    try {
      const response = await profileApi.getSavedItems();
      const items = response.data.data?.items || [];
      const enriched = await Promise.all(
        items.map(async (item) => {
          const type = item.item_type || item.type;
          const targetId = item.item_id ?? item.target_id ?? item.targetId;
          if (type === "user" && targetId) {
            try {
              const u = await profileApi.getPublicProfile(targetId);
              const d = u.data?.data;
              return { ...item, name: d?.name || d?.username || `User ${targetId}` };
            } catch (_) {
              return { ...item, name: `User ${targetId}` };
            }
          }
          return item;
        })
      );
      setSavedItems(enriched);
    } catch (err) {
      console.error("Failed to load saved items:", err);
    }
  }, []);

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

  // Strict: if URL has id param for another user, redirect to view-only public profile
  useEffect(() => {
    if (!profile || !idFromUrl) return;
    const targetId = Number(idFromUrl);
    const myId = profile.id ?? profile.user_id;
    if (!Number.isNaN(targetId) && myId != null && Number(myId) !== targetId) {
      navigate(`/users/${targetId}`, { replace: true });
    }
  }, [profile, idFromUrl, navigate]);

  useEffect(() => {
    if (!isAuthenticated) return;
    switch (activeTab) {
      case "portfolio":
        loadPortfolio();
        break;
      case "my-projects":
        loadMyProjects();
        break;
      case "services":
        loadServices();
        break;
      case "saved":
        loadSavedItems();
        break;
      default:
        break;
    }
  }, [activeTab, isAuthenticated, loadPortfolio, loadMyProjects, loadServices, loadSavedItems]);

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

  const handleCreatePortfolioItem = async (data) => {
    try {
      const response = await portfolioApi.createPortfolioItem(data);
      const newItem = response.data.data;
      setPortfolio((prev) => [...prev, newItem]);
      return { success: true };
    } catch (err) {
      return { success: false, error: err.message || "Create failed" };
    }
  };

  const handleUpdatePortfolioItem = async (id, data) => {
    try {
      const response = await portfolioApi.updatePortfolioItem(id, data);
      const updated = response.data.data;
      setPortfolio((prev) => prev.map((item) => (item.id === id ? updated : item)));
      return { success: true };
    } catch (err) {
      return { success: false, error: err.message || "Update failed" };
    }
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

  const handleRemoveSavedItem = async (savedItemId) => {
    try {
      await profileApi.removeSavedItem(savedItemId);
      setSavedItems((prev) => prev.filter((item) => item.id !== savedItemId));
      return { success: true };
    } catch (err) {
      return { success: false, error: err.message || "Remove failed" };
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

  const handleTabChange = (e, newValue) => {
    setActiveTab(newValue);
    setSearchParams(newValue === "basic" ? {} : { tab: newValue });
  };

  return (
    <PageLayout maxWidth="lg" variant="light">
        <StateContainer loading={loading || profileLoading} error={error} onRetry={loadProfile}>
          <Stack spacing={4}>
            <ProfileEditHeader
              profile={profile}
              services={services}
              portfolio={isStudentRole ? portfolio : myProjects}
              onNavigateBack={() => navigate("/dashboard")}
              onPreview={() => profile?.id && navigate(`/users/${profile.id}`)}
            />

            <Paper sx={{ borderRadius: 3, overflow: "hidden", boxShadow: "0 2px 8px rgba(0,0,0,0.06)" }}>
              <Tabs
                value={activeTab}
                onChange={handleTabChange}
                variant="scrollable"
                scrollButtons="auto"
                sx={{
                  borderBottom: "1px solid #e0e0e0",
                  "& .MuiTab-root": { color: "#666", textTransform: "none", fontWeight: 600 },
                  "& .Mui-selected": { color: "#1976d2" },
                  "& .MuiTabs-indicator": { bgcolor: "#1976d2", height: 3 },
                }}
              >
                {ALL_TABS.map((tab) => (
                  <Tab key={tab.id} value={tab.id} label={tab.label} icon={tab.icon} iconPosition="start" />
                ))}
              </Tabs>

              <Box sx={{ p: 4, minHeight: 400, bgcolor: "#fff" }}>
                {activeTab === "basic" && (
                  <BasicInfoSection
                    profile={profile}
                    isStudentRole={isStudentRole}
                    onUpdate={handleProfileUpdate}
                    onAvatarUpload={handleAvatarUpload}
                  />
                )}
                {activeTab === "services" && (
                  <ServicesSection onSave={loadServices} />
                )}
                {activeTab === "portfolio" && (
                  <PortfolioSection
                    items={portfolio}
                    onCreate={handleCreatePortfolioItem}
                    onUpdate={handleUpdatePortfolioItem}
                    onDelete={handleDeletePortfolioItem}
                  />
                )}
                {activeTab === "my-projects" && (
                  <MyProjectsSection items={myProjects} onRefresh={loadMyProjects} />
                )}
                {activeTab === "saved" && (
                  <SavedItemsSection
                    items={savedItems}
                    onRemove={handleRemoveSavedItem}
                    onNavigate={(type, id) => {
                      const routes = {
                        opportunity: `/opportunities/${id}`,
                        user: `/users/${id}`,
                      };
                      navigate(routes[type] || "/");
                    }}
                  />
                )}
                {activeTab === "settings" && (
                  <Stack spacing={3}>
                    <SecuritySettings
                      profile={profile}
                      onPasswordChange={handlePasswordChange}
                      onProfileUpdate={handleProfileUpdate}
                    />
                  </Stack>
                )}
              </Box>
            </Paper>
          </Stack>
        </StateContainer>
    </PageLayout>
  );
}
