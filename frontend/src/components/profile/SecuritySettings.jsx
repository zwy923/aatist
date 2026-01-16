import React, { useState } from "react";
import {
    Box,
    Button,
    FormControl,
    Grid,
    InputLabel,
    MenuItem,
    Paper,
    Select,
    Stack,
    TextField,
    Typography,
    Alert,
    Snackbar,
    InputAdornment,
    IconButton,
} from "@mui/material";
import LockIcon from "@mui/icons-material/Lock";
import VisibilityIcon from "@mui/icons-material/Visibility";
import VisibilityOffIcon from "@mui/icons-material/VisibilityOff";
import SaveIcon from "@mui/icons-material/Save";
import SecurityIcon from "@mui/icons-material/Security";

// Visibility options
const VISIBILITY_OPTIONS = [
    { value: "public", label: "Public", description: "Visible to everyone" },
    { value: "aalto_only", label: "Aalto Only", description: "Only visible to Aalto members" },
    { value: "private", label: "Private", description: "Only visible to you" },
];

export default function SecuritySettings({ profile, onPasswordChange, onProfileUpdate }) {
    const [passwordForm, setPasswordForm] = useState({
        currentPassword: "",
        newPassword: "",
        confirmPassword: "",
    });
    const [showPasswords, setShowPasswords] = useState({
        current: false,
        new: false,
        confirm: false,
    });
    const [visibilitySettings, setVisibilitySettings] = useState({
        profile_visibility: profile?.profile_visibility || "public",
        portfolio_visibility: profile?.portfolio_visibility || "public",
    });
    const [saving, setSaving] = useState(false);
    const [snackbar, setSnackbar] = useState({ open: false, message: "", severity: "success" });

    const handlePasswordFormChange = (field) => (event) => {
        setPasswordForm((prev) => ({ ...prev, [field]: event.target.value }));
    };

    const toggleShowPassword = (field) => () => {
        setShowPasswords((prev) => ({ ...prev, [field]: !prev[field] }));
    };

    const handlePasswordSubmit = async (e) => {
        e.preventDefault();

        // Validation
        if (!passwordForm.currentPassword || !passwordForm.newPassword) {
            setSnackbar({ open: true, message: "Please fill in all password fields", severity: "error" });
            return;
        }

        if (passwordForm.newPassword !== passwordForm.confirmPassword) {
            setSnackbar({ open: true, message: "New passwords do not match", severity: "error" });
            return;
        }

        if (passwordForm.newPassword.length < 8) {
            setSnackbar({ open: true, message: "Password must be at least 8 characters", severity: "error" });
            return;
        }

        setSaving(true);
        try {
            const result = await onPasswordChange(
                passwordForm.currentPassword,
                passwordForm.newPassword
            );

            if (result.success) {
                setSnackbar({ open: true, message: "Password changed successfully", severity: "success" });
                setPasswordForm({ currentPassword: "", newPassword: "", confirmPassword: "" });
            } else {
                setSnackbar({ open: true, message: result.error || "Password change failed", severity: "error" });
            }
        } finally {
            setSaving(false);
        }
    };

    const handleVisibilityChange = (field) => (event) => {
        setVisibilitySettings((prev) => ({ ...prev, [field]: event.target.value }));
    };

    const handleSaveVisibility = async () => {
        setSaving(true);
        try {
            const result = await onProfileUpdate(visibilitySettings);
            if (result.success) {
                setSnackbar({ open: true, message: "Privacy settings updated", severity: "success" });
            } else {
                setSnackbar({ open: true, message: result.error || "Update failed", severity: "error" });
            }
        } finally {
            setSaving(false);
        }
    };

    return (
        <>
            <Stack spacing={3}>
                {/* Password Change */}
                <Paper
                    sx={{
                        background: "rgba(7, 12, 30, 0.8)",
                        backdropFilter: "blur(20px)",
                        border: "1px solid rgba(93, 224, 255, 0.15)",
                        borderRadius: 3,
                        p: 4,
                    }}
                >
                    <Stack direction="row" alignItems="center" spacing={1} mb={3}>
                        <LockIcon sx={{ color: "primary.main" }} />
                        <Typography variant="h6" fontWeight={600}>
                            Change Password
                        </Typography>
                    </Stack>

                    {profile?.oauth_provider && (
                        <Alert severity="info" sx={{ mb: 3 }}>
                            You signed up with {profile.oauth_provider}. Password change may not be available.
                        </Alert>
                    )}

                    <Box component="form" onSubmit={handlePasswordSubmit}>
                        <Grid container spacing={3}>
                            <Grid item xs={12}>
                                <TextField
                                    fullWidth
                                    label="Current Password"
                                    type={showPasswords.current ? "text" : "password"}
                                    value={passwordForm.currentPassword}
                                    onChange={handlePasswordFormChange("currentPassword")}
                                    InputProps={{
                                        endAdornment: (
                                            <InputAdornment position="end">
                                                <IconButton onClick={toggleShowPassword("current")} edge="end">
                                                    {showPasswords.current ? <VisibilityOffIcon /> : <VisibilityIcon />}
                                                </IconButton>
                                            </InputAdornment>
                                        ),
                                    }}
                                />
                            </Grid>
                            <Grid item xs={12} md={6}>
                                <TextField
                                    fullWidth
                                    label="New Password"
                                    type={showPasswords.new ? "text" : "password"}
                                    value={passwordForm.newPassword}
                                    onChange={handlePasswordFormChange("newPassword")}
                                    helperText="At least 8 characters"
                                    InputProps={{
                                        endAdornment: (
                                            <InputAdornment position="end">
                                                <IconButton onClick={toggleShowPassword("new")} edge="end">
                                                    {showPasswords.new ? <VisibilityOffIcon /> : <VisibilityIcon />}
                                                </IconButton>
                                            </InputAdornment>
                                        ),
                                    }}
                                />
                            </Grid>
                            <Grid item xs={12} md={6}>
                                <TextField
                                    fullWidth
                                    label="Confirm New Password"
                                    type={showPasswords.confirm ? "text" : "password"}
                                    value={passwordForm.confirmPassword}
                                    onChange={handlePasswordFormChange("confirmPassword")}
                                    error={
                                        passwordForm.confirmPassword &&
                                        passwordForm.newPassword !== passwordForm.confirmPassword
                                    }
                                    helperText={
                                        passwordForm.confirmPassword &&
                                            passwordForm.newPassword !== passwordForm.confirmPassword
                                            ? "Passwords do not match"
                                            : ""
                                    }
                                    InputProps={{
                                        endAdornment: (
                                            <InputAdornment position="end">
                                                <IconButton onClick={toggleShowPassword("confirm")} edge="end">
                                                    {showPasswords.confirm ? <VisibilityOffIcon /> : <VisibilityIcon />}
                                                </IconButton>
                                            </InputAdornment>
                                        ),
                                    }}
                                />
                            </Grid>
                            <Grid item xs={12}>
                                <Button
                                    type="submit"
                                    variant="contained"
                                    startIcon={<SaveIcon />}
                                    disabled={saving}
                                >
                                    {saving ? "Updating..." : "Update Password"}
                                </Button>
                            </Grid>
                        </Grid>
                    </Box>
                </Paper>

                {/* Privacy Settings */}
                <Paper
                    sx={{
                        background: "rgba(7, 12, 30, 0.8)",
                        backdropFilter: "blur(20px)",
                        border: "1px solid rgba(93, 224, 255, 0.15)",
                        borderRadius: 3,
                        p: 4,
                    }}
                >
                    <Stack direction="row" alignItems="center" spacing={1} mb={3}>
                        <SecurityIcon sx={{ color: "primary.main" }} />
                        <Typography variant="h6" fontWeight={600}>
                            Privacy Settings
                        </Typography>
                    </Stack>

                    <Grid container spacing={3}>
                        <Grid item xs={12} md={6}>
                            <FormControl fullWidth>
                                <InputLabel>Profile Visibility</InputLabel>
                                <Select
                                    value={visibilitySettings.profile_visibility}
                                    label="Profile Visibility"
                                    onChange={handleVisibilityChange("profile_visibility")}
                                >
                                    {VISIBILITY_OPTIONS.map((opt) => (
                                        <MenuItem key={opt.value} value={opt.value}>
                                            <Stack>
                                                <Typography>{opt.label}</Typography>
                                                <Typography variant="caption" color="text.secondary">
                                                    {opt.description}
                                                </Typography>
                                            </Stack>
                                        </MenuItem>
                                    ))}
                                </Select>
                            </FormControl>
                        </Grid>

                        <Grid item xs={12} md={6}>
                            <FormControl fullWidth>
                                <InputLabel>Portfolio Visibility</InputLabel>
                                <Select
                                    value={visibilitySettings.portfolio_visibility}
                                    label="Portfolio Visibility"
                                    onChange={handleVisibilityChange("portfolio_visibility")}
                                >
                                    {VISIBILITY_OPTIONS.map((opt) => (
                                        <MenuItem key={opt.value} value={opt.value}>
                                            <Stack>
                                                <Typography>{opt.label}</Typography>
                                                <Typography variant="caption" color="text.secondary">
                                                    {opt.description}
                                                </Typography>
                                            </Stack>
                                        </MenuItem>
                                    ))}
                                </Select>
                            </FormControl>
                        </Grid>

                        <Grid item xs={12}>
                            <Button
                                variant="outlined"
                                startIcon={<SaveIcon />}
                                onClick={handleSaveVisibility}
                                disabled={saving}
                            >
                                {saving ? "Saving..." : "Save Privacy Settings"}
                            </Button>
                        </Grid>
                    </Grid>
                </Paper>
            </Stack>

            <Snackbar
                open={snackbar.open}
                autoHideDuration={4000}
                onClose={() => setSnackbar((prev) => ({ ...prev, open: false }))}
                anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
            >
                <Alert
                    severity={snackbar.severity}
                    onClose={() => setSnackbar((prev) => ({ ...prev, open: false }))}
                >
                    {snackbar.message}
                </Alert>
            </Snackbar>
        </>
    );
}
