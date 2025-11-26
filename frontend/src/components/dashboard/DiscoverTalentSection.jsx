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
} from "@mui/material";

export default function DiscoverTalentSection({ portfolios, onSelectUser }) {
  return (
    <Box>
      <Typography variant="h5" fontWeight={700} mb={1} sx={{ color: "text.primary" }}>
        Discover Talent
      </Typography>
      <Typography variant="body1" color="text.secondary" mb={3}>
        Explore portfolios from talented students across design, engineering, business, and arts
      </Typography>

      {portfolios.length > 0 ? (
        <Grid container spacing={3}>
          {portfolios.map((portfolio) => (
            <Grid item xs={12} sm={6} md={4} lg={3} key={portfolio.id}>
              <Card
                sx={{
                  background: "rgba(7, 12, 30, 0.96)",
                  border: "1px solid rgba(93, 224, 255, 0.25)",
                  borderRadius: 3,
                  cursor: "pointer",
                  transition: "all 0.2s",
                  "&:hover": {
                    transform: "translateY(-4px)",
                    borderColor: "rgba(93, 224, 255, 0.5)",
                    boxShadow: "0 8px 24px rgba(93, 224, 255, 0.2)",
                  },
                }}
                onClick={() => onSelectUser(portfolio.user_id)}
              >
                {portfolio.image_url ? (
                  <CardMedia
                    component="img"
                    height="200"
                    image={portfolio.image_url}
                    alt={portfolio.title}
                    sx={{ objectFit: "cover" }}
                  />
                ) : (
                  <Box
                    sx={{
                      height: 200,
                      background:
                        "linear-gradient(135deg, rgba(93, 224, 255, 0.2), rgba(127, 93, 255, 0.2))",
                      display: "flex",
                      alignItems: "center",
                      justifyContent: "center",
                    }}
                  >
                    <Typography variant="h6" color="text.secondary">
                      {portfolio.title?.[0] || "P"}
                    </Typography>
                  </Box>
                )}
                <CardContent>
                  <Typography
                    variant="subtitle1"
                    fontWeight={600}
                    sx={{ color: "text.primary", mb: 0.5 }}
                    noWrap
                  >
                    {portfolio.title}
                  </Typography>
                  <Typography
                    variant="body2"
                    color="text.secondary"
                    sx={{
                      display: "-webkit-box",
                      WebkitLineClamp: 2,
                      WebkitBoxOrient: "vertical",
                      overflow: "hidden",
                    }}
                  >
                    {portfolio.description}
                  </Typography>
                  {portfolio.technologies && portfolio.technologies.length > 0 && (
                    <Stack direction="row" spacing={0.5} mt={1} flexWrap="wrap">
                      {portfolio.technologies.slice(0, 3).map((tech, idx) => (
                        <Chip
                          key={idx}
                          label={tech}
                          size="small"
                          sx={{
                            background: "rgba(93, 224, 255, 0.1)",
                            color: "#5de0ff",
                            fontSize: "0.7rem",
                          }}
                        />
                      ))}
                    </Stack>
                  )}
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>
      ) : (
        <Paper
          sx={{
            p: 4,
            textAlign: "center",
            background: "rgba(7, 12, 30, 0.96)",
            border: "1px solid rgba(93, 224, 255, 0.25)",
            borderRadius: 3,
          }}
        >
          <Typography variant="body1" color="text.secondary">
            No portfolios available yet. Check back soon!
          </Typography>
        </Paper>
      )}
    </Box>
  );
}

