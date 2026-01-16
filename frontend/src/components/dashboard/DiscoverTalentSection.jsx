import React from "react";
import {
  Box,
  Card,
  CardContent,
  CardMedia,
  Chip,
  Grid,
  Paper,
  Stack,
  Typography,
  useTheme,
} from "@mui/material";
import ArrowForwardIcon from "@mui/icons-material/ArrowForward";

export default function DiscoverTalentSection({ portfolios, onSelectUser }) {
  const theme = useTheme();

  return (
    <Box sx={{ py: 2 }}>
      <Stack direction="row" justifyContent="space-between" alignItems="center" mb={3}>
        <Box>
          <Typography variant="h5" fontWeight={700} sx={{ color: "text.primary" }}>
            Discover Work
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Explore diverse projects from the Aalto community
          </Typography>
        </Box>
      </Stack>

      {portfolios.length > 0 ? (
        <Grid container spacing={3}>
          {portfolios.map((portfolio) => (
            <Grid item xs={12} sm={6} md={3} key={portfolio.id}>
              <Card
                onClick={() => onSelectUser(portfolio.user_id)}
                sx={{
                  height: "100%",
                  display: "flex",
                  flexDirection: "column",
                  background: "transparent",
                  border: "none",
                  boxShadow: "none",
                  cursor: "pointer",
                  "&:hover .portfolio-image": {
                    transform: "scale(1.03)",
                  },
                  "&:hover .portfolio-overlay": {
                    opacity: 1,
                  },
                }}
              >
                <Box
                  sx={{
                    position: "relative",
                    borderRadius: 4,
                    overflow: "hidden",
                    mb: 2,
                    aspectRatio: "4/3",
                    boxShadow: "0 4px 20px rgba(0,0,0,0.25)",
                    border: "1px solid rgba(255, 255, 255, 0.05)",
                  }}
                >
                  {portfolio.cover_image_url ? (
                    <CardMedia
                      component="img"
                      image={portfolio.cover_image_url}
                      alt={portfolio.title}
                      className="portfolio-image"
                      sx={{
                        width: "100%",
                        height: "100%",
                        objectFit: "cover",
                        transition: "transform 0.4s cubic-bezier(0.4, 0, 0.2, 1)",
                      }}
                    />
                  ) : (
                    <Box
                      className="portfolio-image"
                      sx={{
                        width: "100%",
                        height: "100%",
                        background: "linear-gradient(135deg, #1e293b, #0f172a)",
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                        transition: "transform 0.4s cubic-bezier(0.4, 0, 0.2, 1)",
                      }}
                    >
                      <Typography
                        variant="h3"
                        fontWeight={800}
                        sx={{ color: "rgba(255,255,255,0.05)" }}
                      >
                        {portfolio.title?.[0]?.toUpperCase() || "P"}
                      </Typography>
                    </Box>
                  )}

                  {/* Hover Overlay */}
                  <Box
                    className="portfolio-overlay"
                    sx={{
                      position: "absolute",
                      top: 0,
                      left: 0,
                      right: 0,
                      bottom: 0,
                      background: "rgba(0, 0, 0, 0.4)",
                      display: "flex",
                      alignItems: "center",
                      justifyContent: "center",
                      opacity: 0,
                      transition: "opacity 0.3s",
                    }}
                  >
                    <Box sx={{
                      bgcolor: "rgba(255, 255, 255, 0.1)",
                      backdropFilter: "blur(4px)",
                      borderRadius: "50%",
                      p: 1.5,
                      display: "flex"
                    }}>
                      <ArrowForwardIcon sx={{ color: "#fff" }} />
                    </Box>
                  </Box>
                </Box>

                <Box>
                  <Typography
                    variant="subtitle1"
                    fontWeight={700}
                    sx={{ color: "text.primary", lineHeight: 1.2, mb: 0.5 }}
                    noWrap
                  >
                    {portfolio.title}
                  </Typography>
                  <Stack direction="row" spacing={1} overflow="hidden">
                    {portfolio.tags && portfolio.tags.slice(0, 2).map((tag, idx) => (
                      <Typography
                        key={idx}
                        variant="caption"
                        sx={{ color: "text.secondary", fontSize: "0.75rem" }}
                      >
                        #{tag}
                      </Typography>
                    ))}
                  </Stack>
                </Box>
              </Card>
            </Grid>
          ))}
        </Grid>
      ) : (
        <Paper
          sx={{
            p: 6,
            textAlign: "center",
            background: "rgba(7, 12, 30, 0.4)",
            border: "1px dashed rgba(255, 255, 255, 0.1)",
            borderRadius: 4,
          }}
        >
          <Typography variant="body1" color="text.secondary">
            No active portfolios found.
          </Typography>
        </Paper>
      )}
    </Box>
  );
}

