import React from "react";
import {
    Box,
    Button,
    Card,
    CardContent,
    Chip,
    Paper,
    Stack,
    Typography,
} from "@mui/material";
import AssignmentIcon from "@mui/icons-material/Assignment";
import OpenInNewIcon from "@mui/icons-material/OpenInNew";
import AccessTimeIcon from "@mui/icons-material/AccessTime";
import BusinessIcon from "@mui/icons-material/Business";

// Application status configurations
const STATUS_CONFIGS = {
    submitted: { label: "Submitted", color: "#5de0ff", bgColor: "rgba(93, 224, 255, 0.15)" },
    reviewed: { label: "Reviewed", color: "#a78bfa", bgColor: "rgba(167, 139, 250, 0.15)" },
    accepted: { label: "Accepted", color: "#4ade80", bgColor: "rgba(74, 222, 128, 0.15)" },
    rejected: { label: "Rejected", color: "#f87171", bgColor: "rgba(248, 113, 113, 0.15)" },
    withdrawn: { label: "Withdrawn", color: "#94a3b8", bgColor: "rgba(148, 163, 184, 0.15)" },
};

export default function MyApplicationsSection({ applications, onNavigate }) {
    const getStatusConfig = (status) => {
        return STATUS_CONFIGS[status] || STATUS_CONFIGS.submitted;
    };

    const formatDate = (dateString) => {
        if (!dateString) return "";
        const date = new Date(dateString);
        return date.toLocaleDateString("en-US", {
            year: "numeric",
            month: "short",
            day: "numeric",
        });
    };

    // Group applications by status
    const groupedApplications = applications.reduce((acc, app) => {
        const status = app.status || "submitted";
        if (!acc[status]) acc[status] = [];
        acc[status].push(app);
        return acc;
    }, {});

    const statusOrder = ["submitted", "reviewed", "accepted", "rejected", "withdrawn"];

    return (
        <Paper
            sx={{
                background: "#ffffff",
                border: "1px solid #e5e7eb",
                borderRadius: 3,
                p: 4,
            }}
        >
            <Stack direction="row" alignItems="center" spacing={1} mb={3}>
                <AssignmentIcon sx={{ color: "primary.main" }} />
                <Typography variant="h5" fontWeight={600}>
                    My Applications
                </Typography>
                <Chip label={applications.length} size="small" color="primary" />
            </Stack>

            {applications.length === 0 ? (
                <Box
                    sx={{
                        textAlign: "center",
                        py: 6,
                        color: "text.secondary",
                    }}
                >
                    <AssignmentIcon sx={{ fontSize: 48, opacity: 0.5, mb: 2 }} />
                    <Typography variant="h6">No applications yet</Typography>
                    <Typography variant="body2">
                        Your job applications will appear here
                    </Typography>
                </Box>
            ) : (
                <Stack spacing={4}>
                    {statusOrder.map((status) => {
                        const apps = groupedApplications[status];
                        if (!apps || apps.length === 0) return null;

                        const config = getStatusConfig(status);

                        return (
                            <Box key={status}>
                                <Stack direction="row" alignItems="center" spacing={1} mb={2}>
                                    <Box
                                        sx={{
                                            width: 12,
                                            height: 12,
                                            borderRadius: "50%",
                                            backgroundColor: config.color,
                                        }}
                                    />
                                    <Typography variant="subtitle1" fontWeight={600}>
                                        {config.label}
                                    </Typography>
                                    <Chip label={apps.length} size="small" sx={{ height: 20 }} />
                                </Stack>

                                <Stack spacing={2}>
                                    {apps.map((app) => (
                                        <Card
                                            key={app.id}
                                            sx={{
                                                backgroundColor: "#ffffff",
                                                border: `1px solid ${config.color}30`,
                                                borderRadius: 2,
                                                transition: "all 0.2s ease",
                                                "&:hover": {
                                                    borderColor: config.color,
                                                    transform: "translateX(4px)",
                                                },
                                            }}
                                        >
                                            <CardContent>
                                                <Stack
                                                    direction={{ xs: "column", sm: "row" }}
                                                    justifyContent="space-between"
                                                    alignItems={{ xs: "flex-start", sm: "center" }}
                                                    spacing={2}
                                                >
                                                    <Box sx={{ flex: 1 }}>
                                                        <Typography variant="h6" fontWeight={600}>
                                                            {app.opportunity_title || app.title || "Opportunity"}
                                                        </Typography>

                                                        <Stack direction="row" spacing={2} mt={1} flexWrap="wrap">
                                                            {app.company_name && (
                                                                <Stack direction="row" alignItems="center" spacing={0.5}>
                                                                    <BusinessIcon sx={{ fontSize: 16, color: "text.secondary" }} />
                                                                    <Typography variant="body2" color="text.secondary">
                                                                        {app.company_name}
                                                                    </Typography>
                                                                </Stack>
                                                            )}
                                                            {app.applied_at && (
                                                                <Stack direction="row" alignItems="center" spacing={0.5}>
                                                                    <AccessTimeIcon sx={{ fontSize: 16, color: "text.secondary" }} />
                                                                    <Typography variant="body2" color="text.secondary">
                                                                        Applied {formatDate(app.applied_at)}
                                                                    </Typography>
                                                                </Stack>
                                                            )}
                                                        </Stack>

                                                        {app.cover_letter && (
                                                            <Typography
                                                                variant="body2"
                                                                color="text.secondary"
                                                                sx={{
                                                                    mt: 1,
                                                                    display: "-webkit-box",
                                                                    WebkitLineClamp: 2,
                                                                    WebkitBoxOrient: "vertical",
                                                                    overflow: "hidden",
                                                                }}
                                                            >
                                                                {app.cover_letter}
                                                            </Typography>
                                                        )}
                                                    </Box>

                                                    <Stack direction="row" spacing={1} alignItems="center">
                                                        <Chip
                                                            label={config.label}
                                                            size="small"
                                                            sx={{
                                                                backgroundColor: config.bgColor,
                                                                color: config.color,
                                                                fontWeight: 600,
                                                            }}
                                                        />
                                                        <Button
                                                            size="small"
                                                            endIcon={<OpenInNewIcon />}
                                                            onClick={() => onNavigate(app.opportunity_id || app.opportunityId)}
                                                            sx={{ color: "primary.main" }}
                                                        >
                                                            View
                                                        </Button>
                                                    </Stack>
                                                </Stack>

                                                {/* Status timeline */}
                                                {app.status_history && app.status_history.length > 0 && (
                                                    <Stack direction="row" spacing={1} mt={2} flexWrap="wrap">
                                                        {app.status_history.map((history, index) => (
                                                            <Chip
                                                                key={index}
                                                                label={`${history.status} - ${formatDate(history.changed_at)}`}
                                                                size="small"
                                                                variant="outlined"
                                                                sx={{ fontSize: "0.7rem" }}
                                                            />
                                                        ))}
                                                    </Stack>
                                                )}
                                            </CardContent>
                                        </Card>
                                    ))}
                                </Stack>
                            </Box>
                        );
                    })}
                </Stack>
            )}
        </Paper>
    );
}
