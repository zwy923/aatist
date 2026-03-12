import React, { useRef, useState } from "react";
import {
    Avatar,
    Box,
    Button,
    Chip,
    CircularProgress,
    IconButton,
    Paper,
    Stack,
    Typography,
} from "@mui/material";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import CameraAltIcon from "@mui/icons-material/CameraAlt";
import VerifiedIcon from "@mui/icons-material/Verified";

export default function ProfileHeader({ profile, onAvatarUpload, onNavigateBack }) {
    const fileInputRef = useRef(null);
    const [uploading, setUploading] = useState(false);

    const handleAvatarClick = () => {
        fileInputRef.current?.click();
    };

    const handleFileChange = async (event) => {
        const file = event.target.files?.[0];
        if (!file) return;

        // Validate file type
        if (!file.type.startsWith("image/")) {
            alert("Please select an image file");
            return;
        }

        // Validate file size (max 5MB)
        if (file.size > 5 * 1024 * 1024) {
            alert("Image size should be less than 5MB");
            return;
        }

        setUploading(true);
        try {
            const result = await onAvatarUpload(file);
            if (!result.success) {
                alert(result.error || "Failed to upload avatar");
            }
        } finally {
            setUploading(false);
        }
    };

    const getRoleBadge = () => {
        switch (profile?.role) {
            case "student":
                return { label: "Student", color: "primary" };
            case "alumni":
                return { label: "Alumni", color: "secondary" };
            case "org_person":
                return { label: "Organization", color: "info" };
            case "org_team":
                return { label: "Team", color: "info" };
            default:
                return { label: "User", color: "default" };
        }
    };

    const roleBadge = getRoleBadge();

    return (
        <Paper
            sx={{
                background: "#ffffff",
                border: "1px solid #e5e7eb",
                borderRadius: 4,
                p: 4,
                position: "relative",
                overflow: "hidden",
            }}
        >
            {/* Background decoration */}
            <Box
                sx={{
                    position: "absolute",
                    top: -100,
                    right: -100,
                    width: 300,
                    height: 300,
                    background: "radial-gradient(circle, rgba(25, 118, 210, 0.08) 0%, transparent 70%)",
                    borderRadius: "50%",
                    pointerEvents: "none",
                }}
            />

            <Stack direction="row" spacing={3} alignItems="flex-start">
                {/* Back button */}
                <IconButton
                    onClick={onNavigateBack}
                    sx={{
                        color: "text.secondary",
                        "&:hover": { color: "primary.main" },
                    }}
                >
                    <ArrowBackIcon />
                </IconButton>

                {/* Avatar with upload */}
                <Box sx={{ position: "relative" }}>
                    <Avatar
                        src={profile?.avatar_url}
                        alt={profile?.name}
                        sx={{
                            width: 120,
                            height: 120,
                            border: "4px solid rgba(93, 224, 255, 0.3)",
                            fontSize: 48,
                            background: "linear-gradient(135deg, #5de0ff 0%, #1a5a6e 100%)",
                        }}
                    >
                        {profile?.name?.charAt(0)?.toUpperCase() || "U"}
                    </Avatar>

                    {/* Camera overlay */}
                    <IconButton
                        onClick={handleAvatarClick}
                        disabled={uploading}
                        sx={{
                            position: "absolute",
                            bottom: 0,
                            right: 0,
                            backgroundColor: "primary.main",
                            color: "#000",
                            width: 36,
                            height: 36,
                            "&:hover": {
                                backgroundColor: "primary.light",
                            },
                            "&.Mui-disabled": {
                                backgroundColor: "grey.700",
                            },
                        }}
                    >
                        {uploading ? (
                            <CircularProgress size={20} color="inherit" />
                        ) : (
                            <CameraAltIcon fontSize="small" />
                        )}
                    </IconButton>

                    <input
                        ref={fileInputRef}
                        type="file"
                        accept="image/*"
                        onChange={handleFileChange}
                        style={{ display: "none" }}
                    />
                </Box>

                {/* User info */}
                <Stack spacing={1} sx={{ flex: 1 }}>
                    <Stack direction="row" spacing={2} alignItems="center">
                        <Typography variant="h4" fontWeight={700} sx={{ color: "text.primary" }}>
                            {profile?.name || "User"}
                        </Typography>
                        {profile?.is_verified_email && (
                            <VerifiedIcon sx={{ color: "primary.main", fontSize: 28 }} />
                        )}
                    </Stack>

                    <Typography variant="body1" color="text.secondary">
                        {profile?.email}
                    </Typography>

                    <Stack direction="row" spacing={1} sx={{ mt: 1 }}>
                        <Chip
                            label={roleBadge.label}
                            color={roleBadge.color}
                            size="small"
                            sx={{ fontWeight: 600 }}
                        />
                        {profile?.role_verified && (
                            <Chip
                                label="Role Verified"
                                color="success"
                                size="small"
                                variant="outlined"
                            />
                        )}
                        {profile?.school && (
                            <Chip
                                label={profile.school}
                                size="small"
                                variant="outlined"
                                sx={{ borderColor: "rgba(93, 224, 255, 0.3)" }}
                            />
                        )}
                        {profile?.organization_name && (
                            <Chip
                                label={profile.organization_name}
                                size="small"
                                variant="outlined"
                                sx={{ borderColor: "rgba(255, 184, 119, 0.3)" }}
                            />
                        )}
                    </Stack>

                    {profile?.bio && (
                        <Typography
                            variant="body2"
                            color="text.secondary"
                            sx={{ mt: 2, maxWidth: 600 }}
                        >
                            {profile.bio}
                        </Typography>
                    )}
                </Stack>

                {/* Quick stats for students */}
                {(profile?.role === "student" || profile?.role === "alumni") && (
                    <Stack spacing={1} alignItems="flex-end">
                        {profile?.skills?.length > 0 && (
                            <Box sx={{ textAlign: "right" }}>
                                <Typography variant="h5" fontWeight={700} color="secondary.main">
                                    {profile.skills.length}
                                </Typography>
                                <Typography variant="caption" color="text.secondary">
                                    Skills
                                </Typography>
                            </Box>
                        )}
                    </Stack>
                )}
            </Stack>
        </Paper>
    );
}
