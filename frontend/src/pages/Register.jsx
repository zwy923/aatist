import React, { useState, useMemo } from "react";
import { useNavigate, Link } from "react-router-dom";
import {
  Box,
  Button,
  TextField,
  Typography,
  Alert,
  CircularProgress,
  Container,
  Stack,
  Paper,
  Card,
  CardActionArea,
  Chip,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  Grid,
} from "@mui/material";
import { authAPI } from "../services/api";

const createRegisterForm = () => ({
  name: "",
  email: "",
  password: "",
  confirmPassword: "",
  role: null,
  studentId: "",
  school: "",
  customSchool: "",
  faculty: "",
  customFaculty: "",
  organizationName: "",
  contactTitle: "",
});

const roleConfigs = {
  student: {
    title: "Student maker",
    accent: "#5de0ff",
    description: "Create & learn",
    badge: "Create & learn",
  },
  organization: {
    title: "Organization / partner",
    accent: "#ffb877",
    description: "Launch opportunities",
    badge: "Launch opportunities",
  },
};

const SCHOOL_OPTIONS = [
  {
    value: "aalto",
    label: "Aalto University",
    faculties: [
      "School of Arts, Design and Architecture",
      "School of Business",
      "School of Chemical Engineering",
      "School of Electrical Engineering",
      "School of Engineering",
      "School of Science",
    ],
  },
  {
    value: "helsinki",
    label: "University of Helsinki",
    faculties: [
      "Faculty of Arts",
      "Faculty of Science",
      "Faculty of Social Sciences",
      "Faculty of Medicine",
    ],
  },
];

const OTHER_OPTION = "other";

const strongPasswordRegex =
  /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[^\da-zA-Z]).{10,}$/;
const strongPasswordHint =
  "Use 10+ characters with uppercase, lowercase, a number, and a symbol.";
const isStrongPassword = (password) => strongPasswordRegex.test(password);

export default function Register() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [registerSuccess, setRegisterSuccess] = useState(false);
  const [registeredEmail, setRegisteredEmail] = useState("");
  const [registerStage, setRegisterStage] = useState("choice");
  const [registerForm, setRegisterForm] = useState(createRegisterForm());

  const selectedSchool = useMemo(
    () => SCHOOL_OPTIONS.find((option) => option.value === registerForm.school),
    [registerForm.school]
  );

  const handleRoleSelect = (role) => {
    setRegisterForm((prev) => ({
      ...prev,
      role,
      studentId: "",
      school: "",
      customSchool: "",
      faculty: "",
      customFaculty: "",
      organizationName: "",
      contactTitle: "",
    }));
    setRegisterStage("form");
    setError("");
  };

  const handleRegisterChange = (field, value) => {
    setRegisterForm((prev) => ({ ...prev, [field]: value }));
  };

  const handleRegister = async (e) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    if (!registerForm.role) {
      setError("Please choose how you'd like to use Aatist.");
      setLoading(false);
      return;
    }

    if (!isStrongPassword(registerForm.password)) {
      setError(strongPasswordHint);
      setLoading(false);
      return;
    }

    if (registerForm.password !== registerForm.confirmPassword) {
      setError("Passwords do not match.");
      setLoading(false);
      return;
    }

    const profile = {};

    if (registerForm.role === "student") {
      if (!registerForm.studentId.trim()) {
        setError("Student ID is required.");
        setLoading(false);
        return;
      }
      if (!registerForm.school) {
        setError("Please select your university.");
        setLoading(false);
        return;
      }

      const schoolOption = selectedSchool;
      let resolvedSchool = registerForm.school;

      if (registerForm.school === OTHER_OPTION) {
        if (!registerForm.customSchool.trim()) {
          setError("Please enter your university name.");
          setLoading(false);
          return;
        }
        resolvedSchool = registerForm.customSchool.trim();
      } else if (schoolOption?.label) {
        resolvedSchool = schoolOption.label;
      }

      let resolvedFaculty = "";
      if (schoolOption?.faculties?.length) {
        if (!registerForm.faculty) {
          setError("Please select your faculty.");
          setLoading(false);
          return;
        }
        if (
          registerForm.faculty === OTHER_OPTION &&
          !registerForm.customFaculty.trim()
        ) {
          setError("Please enter your faculty name.");
          setLoading(false);
          return;
        }
        resolvedFaculty =
          registerForm.faculty === OTHER_OPTION
            ? registerForm.customFaculty.trim()
            : registerForm.faculty;
      } else if (registerForm.customFaculty.trim()) {
        resolvedFaculty = registerForm.customFaculty.trim();
      }

      profile.studentId = registerForm.studentId.trim();
      profile.school = resolvedSchool;
      if (resolvedFaculty) {
        profile.faculty = resolvedFaculty;
      }
    } else if (registerForm.role === "organization") {
      if (!registerForm.organizationName.trim()) {
        setError("Organization / team name is required.");
        setLoading(false);
        return;
      }
      if (!registerForm.contactTitle.trim()) {
        setError("Please tell us your role / title.");
        setLoading(false);
        return;
      }
      profile.organizationName = registerForm.organizationName.trim();
      profile.contactTitle = registerForm.contactTitle.trim();
    }

    // Ensure profile is not empty (profile is required)
    if (Object.keys(profile).length === 0) {
      setError("Profile information is required.");
      setLoading(false);
      return;
    }

    const cleanProfile = Object.fromEntries(
      Object.entries(profile).filter(([, value]) => Boolean(value))
    );

    try {
      await authAPI.register({
        name: registerForm.name,
        email: registerForm.email,
        password: registerForm.password,
        role: registerForm.role,
        profile: cleanProfile,
      });
      setRegisterSuccess(true);
      setRegisteredEmail(registerForm.email);
      setRegisterForm(createRegisterForm());
      setRegisterStage("choice");
    } catch (err) {
      console.error("Registration error:", err); // Debug log
      
      let errorMessage = "Registration failed, try again.";
      
      // Check if we have a response object
      if (err.response) {
        const status = err.response.status;
        const data = err.response.data;
        
        // Handle 409 Conflict (email/username already exists)
        if (status === 409) {
          errorMessage = data?.error?.message || 
                        data?.message || 
                        "This email is already registered. Please use a different email address.";
        } 
        // Handle other error responses
        else if (data?.error?.message) {
          errorMessage = data.error.message;
        } else if (data?.message) {
          errorMessage = data.message;
        } else if (data?.error) {
          errorMessage = typeof data.error === 'string' ? data.error : data.error.message || "An error occurred";
        }
      } 
      // Handle network errors or other exceptions
      else if (err.request) {
        errorMessage = "Network error. Please check your connection and try again.";
      } 
      // Handle other errors
      else if (err.message) {
        errorMessage = err.message;
      }
      
      setError(errorMessage);
      setRegisterSuccess(false);
    } finally {
      setLoading(false);
    }
  };

  const emailPlaceholder =
    registerForm.role === "organization"
      ? "you@studio.com"
      : "firstname.lastname@studentmail.com";

  const renderRoleChoice = () => (
    <Stack spacing={3}>
      <Typography variant="body1" color="text.secondary">
        Choose the path that matches how you collaborate with the campus ecosystem.
      </Typography>
      <Grid container spacing={2}>
        {Object.entries(roleConfigs).map(([role, config]) => (
          <Grid item xs={12} md={6} key={role}>
            <Card
              variant="outlined"
              sx={{
                borderColor: config.accent + "55",
                background:
                  registerForm.role === role
                    ? "rgba(93, 224, 255, 0.09)"
                    : "rgba(255,255,255,0.02)",
              }}
            >
              <CardActionArea onClick={() => handleRoleSelect(role)} sx={{ p: 3 }}>
                <Stack spacing={1.5}>
                  <Chip
                    label={config.badge}
                    variant="outlined"
                    sx={{
                      borderColor: config.accent,
                      color: config.accent,
                      width: "fit-content",
                    }}
                  />
                  <Typography variant="h6" fontWeight={600}>
                    {config.title}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    {config.description}
                  </Typography>
                </Stack>
              </CardActionArea>
            </Card>
          </Grid>
        ))}
      </Grid>
      <Button
        variant="text"
        component={Link}
        to="/auth/login"
        sx={{ alignSelf: "flex-start", textTransform: "none" }}
      >
        Prefer to sign in instead?
      </Button>
    </Stack>
  );

  const renderFacultyControls = () => {
    if (!registerForm.school) {
      return null;
    }

    if (registerForm.school === OTHER_OPTION) {
      return (
        <TextField
          label="Faculty / School (optional)"
          placeholder="e.g. Department of Design"
          value={registerForm.customFaculty}
          onChange={(e) => handleRegisterChange("customFaculty", e.target.value)}
          fullWidth
        />
      );
    }

    if (!selectedSchool?.faculties?.length) {
      return null;
    }

    return (
      <Stack spacing={2}>
        <FormControl fullWidth required>
          <InputLabel id="faculty-select-label">Faculty / School</InputLabel>
          <Select
            labelId="faculty-select-label"
            label="Faculty / School"
            value={registerForm.faculty}
            onChange={(e) => handleRegisterChange("faculty", e.target.value)}
          >
            <MenuItem value="">
              <em>Select your faculty</em>
            </MenuItem>
            {selectedSchool.faculties.map((faculty) => (
              <MenuItem key={faculty} value={faculty}>
                {faculty}
              </MenuItem>
            ))}
            <MenuItem value={OTHER_OPTION}>Other faculty</MenuItem>
          </Select>
        </FormControl>
        {registerForm.faculty === OTHER_OPTION && (
          <TextField
            label="Faculty name"
            placeholder="Type your faculty name"
            value={registerForm.customFaculty}
            onChange={(e) =>
              handleRegisterChange("customFaculty", e.target.value)
            }
            required
            fullWidth
          />
        )}
      </Stack>
    );
  };

  const renderRegisterForm = () => (
    <Box component="form" onSubmit={handleRegister}>
      <Stack spacing={3}>
        <TextField
          label="Full name"
          placeholder="How should we call you?"
          value={registerForm.name}
          onChange={(e) => handleRegisterChange("name", e.target.value)}
          required
          fullWidth
        />
        <TextField
          label="Email"
          type="email"
          placeholder={emailPlaceholder}
          value={registerForm.email}
          onChange={(e) => handleRegisterChange("email", e.target.value)}
          required
          fullWidth
        />
        <Stack direction={{ xs: "column", md: "row" }} spacing={2}>
          <TextField
            label="Password"
            type="password"
            placeholder="Min 10 characters, Aa1!"
            value={registerForm.password}
            onChange={(e) => handleRegisterChange("password", e.target.value)}
            required
            fullWidth
            helperText={strongPasswordHint}
          />
          <TextField
            label="Confirm password"
            type="password"
            placeholder="Repeat password"
            value={registerForm.confirmPassword}
            onChange={(e) =>
              handleRegisterChange("confirmPassword", e.target.value)
            }
            required
            fullWidth
          />
        </Stack>

        {registerForm.role === "student" ? (
          <Stack spacing={2}>
            <TextField
              label="Student ID"
              placeholder="e.g. A123456"
              value={registerForm.studentId}
              onChange={(e) => handleRegisterChange("studentId", e.target.value)}
              required
              fullWidth
            />
            <FormControl fullWidth required>
              <InputLabel id="school-select-label">University</InputLabel>
              <Select
                labelId="school-select-label"
                label="University"
                value={registerForm.school}
                onChange={(e) => {
                  const value = e.target.value;
                  setRegisterForm((prev) => ({
                    ...prev,
                    school: value,
                    customSchool: value === OTHER_OPTION ? prev.customSchool : "",
                    faculty: "",
                    customFaculty: "",
                  }));
                }}
              >
                <MenuItem value="">
                  <em>Select your university</em>
                </MenuItem>
                {SCHOOL_OPTIONS.map((option) => (
                  <MenuItem key={option.value} value={option.value}>
                    {option.label}
                  </MenuItem>
                ))}
                <MenuItem value={OTHER_OPTION}>Other university</MenuItem>
              </Select>
            </FormControl>
            {registerForm.school === OTHER_OPTION && (
              <TextField
                label="University name"
                placeholder="Enter your university"
                value={registerForm.customSchool}
                onChange={(e) =>
                  handleRegisterChange("customSchool", e.target.value)
                }
                required
                fullWidth
              />
            )}
            {renderFacultyControls()}
          </Stack>
        ) : (
          <Stack spacing={2}>
            <TextField
              label="Organization / team"
              placeholder="Aalto Ventures Program"
              value={registerForm.organizationName}
              onChange={(e) =>
                handleRegisterChange("organizationName", e.target.value)
              }
              required
              fullWidth
            />
            <TextField
              label="Role / title"
              placeholder="Program Coordinator"
              value={registerForm.contactTitle}
              onChange={(e) =>
                handleRegisterChange("contactTitle", e.target.value)
              }
              required
              fullWidth
            />
          </Stack>
        )}

        <Stack
          direction={{ xs: "column", md: "row" }}
          spacing={2}
          justifyContent="space-between"
        >
          <Button
            variant="outlined"
            color="secondary"
            onClick={() => setRegisterStage("choice")}
            sx={{ flexBasis: { md: "40%" }, textTransform: "none" }}
          >
            Change registration type
          </Button>
          <Button
            type="submit"
            variant="contained"
            size="large"
            disabled={loading}
            sx={{
              flexGrow: 1,
              background: "linear-gradient(135deg, #007bff 0%, #7f5dff 100%)",
              "&:hover": {
                background: "linear-gradient(135deg, #0066cc 0%, #6b4dd9 100%)",
              },
            }}
            endIcon={
              loading ? <CircularProgress size={18} color="inherit" /> : undefined
            }
          >
            {loading ? "Creating..." : "Create my space"}
          </Button>
        </Stack>
        <Typography variant="body2" color="text.secondary" textAlign="center">
          Already have access?{" "}
          <Link
            to="/auth/login"
            style={{
              color: "#5de0ff",
              textDecoration: "none",
              fontWeight: 600,
            }}
          >
            Sign in instead
          </Link>
        </Typography>
      </Stack>
    </Box>
  );

  const renderRegisterContent = () => {
    if (registerSuccess) {
      return (
        <Stack spacing={3} alignItems="center">
          <Chip
            label="Verification email sent"
            color="primary"
            variant="outlined"
          />
          <Typography variant="h5" fontWeight={600} textAlign="center">
            Check your inbox
          </Typography>
          <Typography variant="body1" color="text.secondary" textAlign="center">
            We sent a verification link to <strong>{registeredEmail}</strong>.
            Please verify within 24 hours to activate your workspace.
          </Typography>
          <Button
            variant="contained"
            size="large"
            component={Link}
            to="/auth/login"
            sx={{
              background: "linear-gradient(135deg, #007bff 0%, #7f5dff 100%)",
              "&:hover": {
                background: "linear-gradient(135deg, #0066cc 0%, #6b4dd9 100%)",
              },
            }}
          >
            Back to sign in
          </Button>
        </Stack>
      );
    }

    if (registerStage === "choice") {
      return renderRoleChoice();
    }

    return renderRegisterForm();
  };

  const headerTitle = registerSuccess
    ? "Verify your inbox"
    : registerStage === "choice"
    ? "How do you want to join?"
    : registerForm.role === "organization"
    ? "Tell us about your collective"
    : "Tell us about your campus life";

  const headerSubtitle = registerSuccess
    ? `We just emailed ${registeredEmail || "you"} with the final verification step.`
    : registerStage === "choice"
    ? "Select whether you are joining as a student maker or as an organization that publishes opportunities."
    : registerForm.role === "organization"
    ? "Share the essentials so we can unlock partner tooling for you."
    : "We partner with students across Finland—let us know where you study.";

  return (
    <Box
      sx={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background: "radial-gradient(ellipse at top left, #101820, #050505)",
        padding: 2,
      }}
    >
      <Container maxWidth="md">
        <Paper
          elevation={0}
          sx={{
            padding: { xs: 3, md: 5 },
            background: "rgba(7, 12, 30, 0.96)",
            borderRadius: 3,
            border: "1px solid rgba(93, 224, 255, 0.25)",
            position: "relative",
            overflow: "hidden",
            "&::before": {
              content: '""',
              position: "absolute",
              inset: 0,
              pointerEvents: "none",
              background:
                "radial-gradient(circle at 20% 20%, rgba(93,224,255,0.15), transparent 45%)",
            },
          }}
        >
          <Stack spacing={3} sx={{ position: "relative" }}>
            <Stack spacing={1}>
              <Chip
                label={registerSuccess ? "Join the network" : "Continue building"}
                color="primary"
                variant="outlined"
                sx={{ width: "fit-content" }}
              />
              <Typography variant="h4" fontWeight={700}>
                {headerTitle}
              </Typography>
              <Typography variant="body1" color="text.secondary">
                {headerSubtitle}
              </Typography>
            </Stack>

            {error && (
              <Alert severity="error" variant="outlined">
                {error}
              </Alert>
            )}

            {renderRegisterContent()}

            <Button
              component={Link}
              to="/"
              variant="text"
              size="small"
              sx={{
                alignSelf: "center",
                textTransform: "none",
                color: "text.secondary",
              }}
            >
              Back to home
            </Button>
          </Stack>
        </Paper>
      </Container>
    </Box>
  );
}

