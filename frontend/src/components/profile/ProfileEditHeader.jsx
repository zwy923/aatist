import React from "react";
import { Box, Button, LinearProgress, Switch, Typography } from "@mui/material";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import VisibilityIcon from "@mui/icons-material/Visibility";
import { useNavigate } from "react-router-dom";

const PROFILE_FIELDS = [
  "name",
  "bio",
  "avatar_url",
  "website",
  "linkedin",
  "languages",
  "skills",
  "professional_interests",
  "services",
  "portfolio",
];

function computeCompletion(profile, services = [], portfolio = []) {
  let filled = 0;
  if (profile?.name) filled++;
  if (profile?.bio) filled++;
  if (profile?.avatar_url) filled++;
  if (profile?.website) filled++;
  if (profile?.linkedin) filled++;
  if (profile?.languages) filled++;
  if (profile?.skills?.length > 0) filled++;
  if (profile?.professional_interests) filled++;
  if (services?.length > 0) filled++;
  if (portfolio?.length > 0) filled++;
  return Math.round((filled / PROFILE_FIELDS.length) * 100);
}

export default function ProfileEditHeader({
  profile,
  services = [],
  portfolio = [],
  onNavigateBack,
  onPreview,
  showPreview = true,
}) {
  const navigate = useNavigate();
  const completion = computeCompletion(profile, services, portfolio);

  return (
    <Box sx={{ mb: 4 }}>
      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 3 }}>
        <Button
          startIcon={<ArrowBackIcon />}
          onClick={() => (onNavigateBack ? onNavigateBack() : navigate("/dashboard"))}
          sx={{ color: "#666", textTransform: "none" }}
        >
          Back to browse
        </Button>
        <Box sx={{ display: "flex", alignItems: "center", gap: 2 }}>
          {showPreview && (
            <Button
              variant="outlined"
              startIcon={<VisibilityIcon />}
              onClick={() => (onPreview ? onPreview() : navigate(`/users/${profile?.id}`))}
              sx={{ borderColor: "#1976d2", color: "#1976d2", textTransform: "none" }}
            >
              Preview Public Profile
            </Button>
          )}
          <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
            <Switch size="small" color="primary" defaultChecked />
            <Typography variant="body2" color="#666">
              {profile?.name || "Aatist"}
            </Typography>
          </Box>
        </Box>
      </Box>

      <Box sx={{ mb: 3 }}>
        <Typography variant="subtitle1" fontWeight={600} color="#1a1a1a" gutterBottom>
          Profile Completion
        </Typography>
        <LinearProgress
          variant="determinate"
          value={completion}
          sx={{
            height: 8,
            borderRadius: 4,
            bgcolor: "#e0e0e0",
            "& .MuiLinearProgress-bar": {
              background: "linear-gradient(90deg, #1976d2 0%, #7b1fa2 100%)",
            },
          }}
        />
        <Typography variant="body2" color="#666" sx={{ mt: 1 }}>
          Complete your profile to get more visibility from clients. Profiles above 80% completion
          get 3x more views!
        </Typography>
      </Box>
    </Box>
  );
}
