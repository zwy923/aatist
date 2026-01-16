import React, { useState, useMemo, useCallback } from 'react';
import { Box, Container, Chip } from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../features/auth/hooks/useAuth';
import DashboardHeader from '../../components/dashboard/DashboardHeader';

const NAV_LINKS = [
    { label: "Dashboard", path: "/dashboard" },
    { label: "Opportunities", path: "/opportunities" },
    { label: "Community", path: "/community" },
    { label: "Events", path: "/dashboard?section=events" },
];

const PageLayout = ({ children, maxWidth = "xl" }) => {
    const navigate = useNavigate();
    const { user, logout } = useAuth();
    const [menuAnchorEl, setMenuAnchorEl] = useState(null);

    const isMenuOpen = Boolean(menuAnchorEl);
    const userDisplayName = useMemo(
        () => user?.name || user?.username || user?.email || "User",
        [user]
    );
    const isStudentRole = useMemo(
        () => user?.role?.toLowerCase?.() === "student",
        [user?.role]
    );
    const isVerifiedStudent = isStudentRole && Boolean(user?.is_verified_email);

    const handleMenuOpen = (event) => setMenuAnchorEl(event.currentTarget);
    const handleMenuClose = () => setMenuAnchorEl(null);
    const handleLogout = async () => {
        await logout();
        handleMenuClose();
        navigate("/auth/login");
    };

    const verificationChip = isStudentRole ? (
        <Chip
            size="small"
            label={isVerifiedStudent ? "Student verified" : "Verification pending"}
            color={isVerifiedStudent ? "success" : "warning"}
            variant="outlined"
            sx={{ fontSize: "0.75rem" }}
        />
    ) : (
        <Chip
            size="small"
            label="Organization workspace"
            color="info"
            variant="outlined"
            sx={{ fontSize: "0.75rem" }}
        />
    );

    return (
        <Box
            sx={{
                minHeight: "100vh",
                background: "radial-gradient(ellipse at top left, #101820, #050505)",
                py: 4,
            }}
        >
            <Container maxWidth={maxWidth}>
                <DashboardHeader
                    navItems={NAV_LINKS}
                    onNavClick={(path) => navigate(path)}
                    verificationChip={verificationChip}
                    menuAnchorEl={menuAnchorEl}
                    isMenuOpen={isMenuOpen}
                    onMenuOpen={handleMenuOpen}
                    onMenuClose={handleMenuClose}
                    onLogout={handleLogout}
                    userDisplayName={userDisplayName}
                    userEmail={user?.email}
                    isStudentRole={isStudentRole}
                    onNavigate={(path) => navigate(path)}
                />
                <Box sx={{ mt: 4 }}>
                    {children}
                </Box>
            </Container>
        </Box>
    );
};

export default PageLayout;
