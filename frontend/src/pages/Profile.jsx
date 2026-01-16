import React, { useCallback, useEffect, useState, useMemo } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router-dom";
import {
    Alert,
    Box,
    CircularProgress,
    Container,
    Paper,
    Stack,
    Tab,
    Tabs,
    Typography,
} from "@mui/material";
import PersonIcon from "@mui/icons-material/Person";
import WorkIcon from "@mui/icons-material/Work";
import BookmarkIcon from "@mui/icons-material/Bookmark";
import AssignmentIcon from "@mui/icons-material/Assignment";
import SettingsIcon from "@mui/icons-material/Settings";
import { profileApi, portfolioApi } from "../features/profile/api/profile";
import { useAuth } from "../features/auth/hooks/useAuth";
import { useProfile } from "../features/profile/hooks/useProfile";
import { StateContainer } from "../shared/components/ui/StateContainer";
import ProfileHeader from "../components/profile/ProfileHeader";
import ProfileDetails from "../components/profile/ProfileDetails";
import AvailabilitySettings from "../components/profile/AvailabilitySettings";
import PortfolioSection from "../components/profile/PortfolioSection";
import SavedItemsSection from "../components/profile/SavedItemsSection";
import MyApplicationsSection from "../components/profile/MyApplicationsSection";
import SecuritySettings from "../components/profile/SecuritySettings";

// Tab definitions
const TABS = [
    { id: "profile", label: "Profile", icon: <PersonIcon /> },
    { id: "portfolio", label: "Portfolio", icon: <WorkIcon /> },
    { id: "saved", label: "Saved", icon: <BookmarkIcon /> },
    { id: "applications", label: "Applications", icon: <AssignmentIcon /> },
    { id: "settings", label: "Settings", icon: <SettingsIcon /> },
];

export default function Profile() {
    const navigate = useNavigate();
    const location = useLocation();
    const [searchParams, setSearchParams] = useSearchParams();
    const { isAuthenticated } = useAuth();
    const { user, loading: profileLoading, updateProfile, uploadAvatar } = useProfile();

    // Read initial tab from URL params
    const tabFromUrl = searchParams.get("tab");
    const validTabs = TABS.map((t) => t.id);
    const initialTab = validTabs.includes(tabFromUrl) ? tabFromUrl : "profile";

    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [activeTab, setActiveTab] = useState(initialTab);
    const [profile, setProfile] = useState(null);
    const [portfolio, setPortfolio] = useState([]);
    const [savedItems, setSavedItems] = useState([]);
    const [applications, setApplications] = useState([]);
    const [availability, setAvailability] = useState(null);

    const isStudentRole = useMemo(
        () => user?.role?.toLowerCase?.() === "student",
        [user?.role]
    );

    // Sync tab with URL when URL changes
    useEffect(() => {
        const newTab = searchParams.get("tab");
        if (newTab && validTabs.includes(newTab) && newTab !== activeTab) {
            setActiveTab(newTab);
        }
    }, [searchParams]);

    // Load profile data
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

    // Load portfolio
    const loadPortfolio = useCallback(async () => {
        try {
            const response = await portfolioApi.getMyPortfolio();
            const data = response.data.data;
            setPortfolio(data?.projects || data?.items || []);
        } catch (err) {
            console.error("Failed to load portfolio:", err);
        }
    }, []);

    // Load saved items
    const loadSavedItems = useCallback(async () => {
        try {
            const response = await profileApi.getSavedItems();
            const data = response.data.data;
            setSavedItems(data?.items || []);
        } catch (err) {
            console.error("Failed to load saved items:", err);
        }
    }, []);

    // Load applications
    const loadApplications = useCallback(async () => {
        try {
            const response = await profileApi.getMyApplications();
            const data = response.data.data;
            setApplications(data?.applications || data?.items || []);
        } catch (err) {
            console.error("Failed to load applications:", err);
        }
    }, []);

    // Load availability
    const loadAvailability = useCallback(async () => {
        try {
            const response = await profileApi.getAvailability();
            setAvailability(response.data.data);
        } catch (err) {
            console.error("Failed to load availability:", err);
        }
    }, []);

    // Initial load
    useEffect(() => {
        if (!isAuthenticated) {
            navigate("/auth/login");
            return;
        }
        loadProfile();
    }, [isAuthenticated, navigate, loadProfile]);

    // Load tab-specific data
    useEffect(() => {
        if (!isAuthenticated) return;

        switch (activeTab) {
            case "portfolio":
                loadPortfolio();
                break;
            case "saved":
                loadSavedItems();
                break;
            case "applications":
                loadApplications();
                break;
            case "settings":
                loadAvailability();
                break;
            default:
                break;
        }
    }, [activeTab, isAuthenticated, loadPortfolio, loadSavedItems, loadApplications, loadAvailability]);

    // Handle profile update
    const handleProfileUpdate = async (updatedData) => {
        const result = await updateProfile(updatedData);
        if (result.success) {
            setProfile(result.data);
            return { success: true };
        } else {
            return { success: false, error: result.error };
        }
    };

    // Handle avatar upload
    const handleAvatarUpload = async (file) => {
        const result = await uploadAvatar(file);
        if (result.success) {
            setProfile((prev) => ({ ...prev, avatar_url: result.avatarUrl }));
            return { success: true };
        } else {
            return { success: false, error: result.error };
        }
    };

    // Handle password change
    const handlePasswordChange = async (currentPassword, newPassword) => {
        try {
            await profileApi.changePassword(currentPassword, newPassword);
            return { success: true };
        } catch (err) {
            console.error("Failed to change password:", err);
            return { success: false, error: err.message || "Password change failed" };
        }
    };

    // Handle availability update
    const handleAvailabilityUpdate = async (data) => {
        try {
            const response = await profileApi.updateAvailability(data);
            setAvailability(response.data.data);
            return { success: true };
        } catch (err) {
            console.error("Failed to update availability:", err);
            return { success: false, error: err.message || "Update failed" };
        }
    };

    // Handle portfolio CRUD
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

    // Handle saved items
    const handleRemoveSavedItem = async (savedItemId) => {
        try {
            await profileApi.removeSavedItem(savedItemId);
            setSavedItems((prev) => prev.filter((item) => item.id !== savedItemId));
            return { success: true };
        } catch (err) {
            return { success: false, error: err.message || "Remove failed" };
        }
    };

    // Tab change handler
    const handleTabChange = (event, newValue) => {
        setActiveTab(newValue);
        // Update URL without full navigation
        if (newValue === "profile") {
            setSearchParams({});
        } else {
            setSearchParams({ tab: newValue });
        }
    };

    return (
        <Box
            sx={{
                minHeight: "100vh",
                background: "radial-gradient(ellipse at top left, #101820, #050505)",
                py: 4,
            }}
        >
            <Container maxWidth="lg">
                <StateContainer
                    loading={loading || profileLoading}
                    error={error}
                    onRetry={loadProfile}
                    skeletonCount={1}
                    skeletonHeight={400}
                >
                    <Stack spacing={4}>
                        {/* Profile Header */}
                        <ProfileHeader
                            profile={profile}
                            onAvatarUpload={handleAvatarUpload}
                            onNavigateBack={() => navigate("/dashboard")}
                        />



                        {/* Tabs */}
                        <Paper
                            sx={{
                                background: "rgba(7, 12, 30, 0.8)",
                                backdropFilter: "blur(20px)",
                                border: "1px solid rgba(93, 224, 255, 0.15)",
                                borderRadius: 3,
                            }}
                        >
                            <Tabs
                                value={activeTab}
                                onChange={handleTabChange}
                                variant="scrollable"
                                scrollButtons="auto"
                                sx={{
                                    "& .MuiTab-root": {
                                        color: "text.secondary",
                                        minHeight: 64,
                                        "&.Mui-selected": {
                                            color: "primary.main",
                                        },
                                    },
                                    "& .MuiTabs-indicator": {
                                        backgroundColor: "primary.main",
                                    },
                                }}
                            >
                                {TABS.map((tab) => (
                                    <Tab
                                        key={tab.id}
                                        value={tab.id}
                                        label={tab.label}
                                        icon={tab.icon}
                                        iconPosition="start"
                                    />
                                ))}
                            </Tabs>
                        </Paper>

                        {/* Tab Content */}
                        <Box sx={{ minHeight: 400 }}>
                            {activeTab === "profile" && (
                                <ProfileDetails
                                    profile={profile}
                                    isStudentRole={isStudentRole}
                                    onUpdate={handleProfileUpdate}
                                />
                            )}

                            {activeTab === "portfolio" && (
                                <PortfolioSection
                                    items={portfolio}
                                    onCreate={handleCreatePortfolioItem}
                                    onUpdate={handleUpdatePortfolioItem}
                                    onDelete={handleDeletePortfolioItem}
                                />
                            )}

                            {activeTab === "saved" && (
                                <SavedItemsSection
                                    items={savedItems}
                                    onRemove={handleRemoveSavedItem}
                                    onNavigate={(type, id) => {
                                        const routes = {
                                            opportunity: `/opportunities/${id}`,
                                            event: `/events/${id}`,
                                            post: `/community/posts/${id}`,
                                            user: `/users/${id}`,
                                        };
                                        navigate(routes[type] || "/");
                                    }}
                                />
                            )}

                            {activeTab === "applications" && (
                                <MyApplicationsSection
                                    applications={applications}
                                    onNavigate={(id) => navigate(`/opportunities/${id}`)}
                                />
                            )}

                            {activeTab === "settings" && (
                                <Stack spacing={3}>
                                    {isStudentRole && (
                                        <AvailabilitySettings
                                            availability={availability}
                                            weeklyHours={profile?.weekly_hours}
                                            onUpdate={handleAvailabilityUpdate}
                                            onProfileUpdate={handleProfileUpdate}
                                        />
                                    )}
                                    <SecuritySettings
                                        profile={profile}
                                        onPasswordChange={handlePasswordChange}
                                        onProfileUpdate={handleProfileUpdate}
                                    />
                                </Stack>
                            )}
                        </Box>
                    </Stack>
                </StateContainer>
            </Container>
        </Box>
    );
}
