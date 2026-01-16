import React, { useState } from "react";
import {
    Box,
    Button,
    Grid,
    Paper,
    Slider,
    Stack,
    Typography,
    Chip,
    Alert,
    Snackbar,
} from "@mui/material";
import AccessTimeIcon from "@mui/icons-material/AccessTime";
import SaveIcon from "@mui/icons-material/Save";

// Availability status options
const AVAILABILITY_STATUS = [
    { value: "open", label: "Open", color: "#4caf50" },
    { value: "limited", label: "Limited", color: "#ff9800" },
    { value: "busy", label: "Busy", color: "#f44336" },
];

export default function AvailabilitySettings({
    availability,
    weeklyHours,
    onUpdate,
    onProfileUpdate,
}) {
    const [hours, setHours] = useState(weeklyHours || 10);
    const [weeklyStatus, setWeeklyStatus] = useState(availability || []);
    const [saving, setSaving] = useState(false);
    const [snackbar, setSnackbar] = useState({ open: false, message: "", severity: "success" });

    // Get current week number
    const getCurrentWeek = () => {
        const now = new Date();
        const start = new Date(now.getFullYear(), 0, 1);
        const diff = now - start;
        const oneWeek = 1000 * 60 * 60 * 24 * 7;
        return Math.ceil(diff / oneWeek);
    };

    const currentWeek = getCurrentWeek();
    const currentYear = new Date().getFullYear();

    // Generate next 8 weeks for selection
    const weeks = Array.from({ length: 8 }, (_, i) => {
        const week = ((currentWeek + i - 1) % 52) + 1;
        const year = currentWeek + i > 52 ? currentYear + 1 : currentYear;
        return { week, year };
    });

    const getWeekStatus = (week, year) => {
        const found = weeklyStatus.find((w) => w.week === week && w.year === year);
        return found?.status || "open";
    };

    const handleStatusChange = (week, year, status) => {
        setWeeklyStatus((prev) => {
            const existing = prev.findIndex((w) => w.week === week && w.year === year);
            if (existing >= 0) {
                const updated = [...prev];
                updated[existing] = { week, year, status };
                return updated;
            }
            return [...prev, { week, year, status }];
        });
    };

    const handleSaveHours = async () => {
        setSaving(true);
        try {
            const result = await onProfileUpdate({ weekly_hours: hours });
            if (result.success) {
                setSnackbar({ open: true, message: "Weekly hours updated", severity: "success" });
            } else {
                setSnackbar({ open: true, message: result.error || "Update failed", severity: "error" });
            }
        } finally {
            setSaving(false);
        }
    };

    const handleSaveAvailability = async () => {
        setSaving(true);
        try {
            const result = await onUpdate({ weekly_availability: weeklyStatus });
            if (result.success) {
                setSnackbar({ open: true, message: "Availability updated", severity: "success" });
            } else {
                setSnackbar({ open: true, message: result.error || "Update failed", severity: "error" });
            }
        } finally {
            setSaving(false);
        }
    };

    return (
        <>
            <Paper
                sx={{
                    background: "rgba(7, 12, 30, 0.8)",
                    backdropFilter: "blur(20px)",
                    border: "1px solid rgba(93, 224, 255, 0.15)",
                    borderRadius: 3,
                    p: 4,
                }}
            >
                <Stack spacing={4}>
                    {/* Weekly Hours */}
                    <Box>
                        <Stack direction="row" alignItems="center" spacing={1} mb={2}>
                            <AccessTimeIcon sx={{ color: "primary.main" }} />
                            <Typography variant="h6" fontWeight={600}>
                                Weekly Availability
                            </Typography>
                        </Stack>

                        <Typography variant="body2" color="text.secondary" mb={3}>
                            How many hours per week are you available for projects?
                        </Typography>

                        <Box sx={{ px: 2 }}>
                            <Stack direction="row" alignItems="center" spacing={3}>
                                <Slider
                                    value={hours}
                                    onChange={(e, value) => setHours(value)}
                                    min={0}
                                    max={40}
                                    step={1}
                                    valueLabelDisplay="on"
                                    sx={{
                                        flex: 1,
                                        "& .MuiSlider-thumb": {
                                            backgroundColor: "primary.main",
                                        },
                                        "& .MuiSlider-track": {
                                            backgroundColor: "primary.main",
                                        },
                                        "& .MuiSlider-rail": {
                                            backgroundColor: "rgba(93, 224, 255, 0.2)",
                                        },
                                    }}
                                />
                                <Typography variant="h6" sx={{ minWidth: 60 }}>
                                    {hours}h/week
                                </Typography>
                            </Stack>

                            <Stack direction="row" justifyContent="space-between" mt={1}>
                                <Typography variant="caption" color="text.secondary">0h</Typography>
                                <Typography variant="caption" color="text.secondary">40h</Typography>
                            </Stack>
                        </Box>

                        <Button
                            startIcon={<SaveIcon />}
                            onClick={handleSaveHours}
                            variant="outlined"
                            size="small"
                            disabled={saving}
                            sx={{ mt: 2 }}
                        >
                            Save Hours
                        </Button>
                    </Box>

                    {/* Weekly Status Calendar */}
                    <Box>
                        <Typography variant="h6" fontWeight={600} mb={2}>
                            Upcoming Weeks Status
                        </Typography>

                        <Typography variant="body2" color="text.secondary" mb={3}>
                            Set your availability status for upcoming weeks to let recruiters know when you're free.
                        </Typography>

                        <Grid container spacing={2}>
                            {weeks.map(({ week, year }) => {
                                const status = getWeekStatus(week, year);
                                const statusConfig = AVAILABILITY_STATUS.find((s) => s.value === status);

                                return (
                                    <Grid item xs={6} sm={3} key={`${week}-${year}`}>
                                        <Paper
                                            sx={{
                                                p: 2,
                                                backgroundColor: "rgba(7, 12, 30, 0.6)",
                                                border: "1px solid rgba(93, 224, 255, 0.1)",
                                                borderRadius: 2,
                                            }}
                                        >
                                            <Typography variant="caption" color="text.secondary">
                                                Week {week}, {year}
                                            </Typography>

                                            <Stack direction="row" spacing={0.5} mt={1} flexWrap="wrap" gap={0.5}>
                                                {AVAILABILITY_STATUS.map((s) => (
                                                    <Chip
                                                        key={s.value}
                                                        label={s.label}
                                                        size="small"
                                                        onClick={() => handleStatusChange(week, year, s.value)}
                                                        sx={{
                                                            backgroundColor:
                                                                status === s.value
                                                                    ? `${s.color}30`
                                                                    : "transparent",
                                                            border: `1px solid ${status === s.value ? s.color : "rgba(255,255,255,0.1)"}`,
                                                            color: status === s.value ? s.color : "text.secondary",
                                                            fontSize: "0.7rem",
                                                            cursor: "pointer",
                                                            "&:hover": {
                                                                backgroundColor: `${s.color}20`,
                                                            },
                                                        }}
                                                    />
                                                ))}
                                            </Stack>
                                        </Paper>
                                    </Grid>
                                );
                            })}
                        </Grid>

                        <Button
                            startIcon={<SaveIcon />}
                            onClick={handleSaveAvailability}
                            variant="outlined"
                            size="small"
                            disabled={saving}
                            sx={{ mt: 3 }}
                        >
                            Save Availability
                        </Button>
                    </Box>

                    {/* Legend */}
                    <Box>
                        <Typography variant="subtitle2" color="text.secondary" mb={1}>
                            Status Legend
                        </Typography>
                        <Stack direction="row" spacing={2}>
                            {AVAILABILITY_STATUS.map((s) => (
                                <Stack key={s.value} direction="row" alignItems="center" spacing={1}>
                                    <Box
                                        sx={{
                                            width: 12,
                                            height: 12,
                                            borderRadius: "50%",
                                            backgroundColor: s.color,
                                        }}
                                    />
                                    <Typography variant="caption">{s.label}</Typography>
                                </Stack>
                            ))}
                        </Stack>
                    </Box>
                </Stack>
            </Paper>

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
