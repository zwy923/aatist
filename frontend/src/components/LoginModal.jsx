import React, { useEffect, useMemo, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Card,
  CardActionArea,
  Chip,
  CircularProgress,
  Collapse,
  Dialog,
  DialogContent,
  FormControl,
  Grid,
  IconButton,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  Tab,
  Tabs,
  TextField,
  Typography,
} from "@mui/material";
import CloseRoundedIcon from "@mui/icons-material/CloseRounded";
import { authAPI } from "../services/api";
import { useUser } from "../store/userStore.jsx";

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
    description:
      "",
    badge: "Create & learn",
  },
  organization: {
    title: "Organization / partner",
    accent: "#ffb877",
    description:
      "",
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

export default function LoginModal({ onClose, mode = "login" }) {
  const [activeTab, setActiveTab] = useState(mode === "register" ? "register" : "login");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [registerSuccess, setRegisterSuccess] = useState(false);
  const [registeredEmail, setRegisteredEmail] = useState("");
  const [registerStage, setRegisterStage] = useState("choice");
  const [registerForm, setRegisterForm] = useState(createRegisterForm());

  const { login } = useUser();

  const isLogin = activeTab === "login";

  useEffect(() => {
    const nextTab = mode === "register" ? "register" : "login";
    setActiveTab(nextTab);
    setError("");
    setRegisterSuccess(false);
    if (nextTab === "register") {
      setRegisterStage("choice");
      setRegisterForm(createRegisterForm());
    }
  }, [mode]);

  const handleTabChange = (_, value) => {
    setActiveTab(value);
    setError("");
    setRegisterSuccess(false);
    if (value === "register") {
      setRegisterStage("choice");
      setRegisterForm(createRegisterForm());
    }
  };

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

  const selectedSchool = useMemo(
    () => SCHOOL_OPTIONS.find((option) => option.value === registerForm.school),
    [registerForm.school]
  );

  const handleLogin = async (e) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    const form = e.currentTarget;

    try {
      const response = await authAPI.login(
        form.email.value,
        form.password.value
      );
      login(response.user, response.access_token);
      onClose();
    } catch (err) {
      setError(err.response?.data?.detail || "Sign-in failed, please try again.");
    } finally {
      setLoading(false);
    }
  };

  const handleRegister = async (e) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    if (!registerForm.role) {
      setError("Please choose how you’d like to use Aatist.");
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

    const cleanProfile = Object.fromEntries(
      Object.entries(profile).filter(([, value]) => Boolean(value))
    );

    try {
      await authAPI.register({
        name: registerForm.name,
        email: registerForm.email,
        password: registerForm.password,
        role: registerForm.role,
        ...(Object.keys(cleanProfile).length ? { profile: cleanProfile } : {}),
      });
      setRegisterSuccess(true);
      setRegisteredEmail(registerForm.email);
      setRegisterForm(createRegisterForm());
      setRegisterStage("choice");
    } catch (err) {
      setError(err.response?.data?.detail || "Registration failed, try again.");
      setRegisterSuccess(false);
    } finally {
      setLoading(false);
    }
  };

  const emailPlaceholder =
    registerForm.role === "organization"
      ? "you@studio.com"
      : "firstname.lastname@studentmail.com";

  const headerTitle = isLogin
    ? "Sign in to Aatist"
    : registerSuccess
    ? "Verify your inbox"
    : registerStage === "choice"
    ? "How do you want to join?"
    : registerForm.role === "organization"
    ? "Tell us about your collective"
    : "Tell us about your campus life";

  const headerSubtitle = isLogin
    ? "Access your studio, review briefs, and keep shipping."
    : registerSuccess
    ? `We just emailed ${registeredEmail || "you"} with the final verification step.`
    : registerStage === "choice"
    ? "Select whether you are joining as a student maker or as an organization that publishes opportunities."
    : registerForm.role === "organization"
    ? "Share the essentials so we can unlock partner tooling for you."
    : "We partner with students across Finland—let us know where you study.";

  const renderLoginForm = () => (
    <Box component="form" onSubmit={handleLogin}>
      <Stack spacing={3}>
        <TextField
          name="email"
          type="email"
          label="Email"
          placeholder="you@aatist.fi"
          fullWidth
          required
        />
        <TextField
          name="password"
          type="password"
          label="Password"
          placeholder="Enter your password"
          fullWidth
          required
        />
        <Button
          type="submit"
          variant="contained"
          size="large"
          disabled={loading}
          endIcon={
            loading ? <CircularProgress size={18} color="inherit" /> : undefined
          }
        >
          {loading ? "Signing in..." : "Sign in"}
        </Button>
        <Typography variant="body2" color="text.secondary" textAlign="center">
          New here?{" "}
          <Button
            variant="text"
            size="small"
            onClick={() => handleTabChange(null, "register")}
            sx={{ textTransform: "none", fontWeight: 600 }}
          >
            Create an account
          </Button>
        </Typography>
      </Stack>
    </Box>
  );

  const renderRoleChoice = () => (
    <Stack spacing={3}>
      <Typography variant="body1" color="text.secondary">
        Choose the path that matches how you collaborate with the campus
        ecosystem.
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
        onClick={() => handleTabChange(null, "login")}
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
            sx={{ flexGrow: 1 }}
            endIcon={
              loading ? <CircularProgress size={18} color="inherit" /> : undefined
            }
          >
            {loading ? "Creating..." : "Create my space"}
          </Button>
        </Stack>
        <Typography variant="body2" color="text.secondary" textAlign="center">
          Already have access?{" "}
          <Button
            variant="text"
            size="small"
            onClick={() => handleTabChange(null, "login")}
            sx={{ textTransform: "none", fontWeight: 600 }}
          >
            Sign in instead
          </Button>
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
            onClick={() => {
              setRegisterSuccess(false);
              setActiveTab("login");
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

  return (
    <Dialog
      open
      maxWidth="md"
      fullWidth
      keepMounted
      onClose={onClose}
      BackdropProps={{
        sx: {
          backdropFilter: "blur(14px)",
          backgroundColor: "rgba(3, 5, 15, 0.85)",
        },
      }}
    >
      <DialogContent
        sx={{
          position: "relative",
          overflow: "hidden",
          px: { xs: 3, md: 6 },
          py: { xs: 3, md: 5 },
        }}
      >
        <Box
          sx={{
            position: "absolute",
            inset: 0,
            pointerEvents: "none",
            background:
              "radial-gradient(circle at 20% 20%, rgba(93,224,255,0.15), transparent 45%)",
          }}
        />
        <IconButton
          onClick={onClose}
          sx={{
            position: "absolute",
            top: 12,
            right: 12,
            zIndex: 1,
          }}
        >
          <CloseRoundedIcon />
        </IconButton>
        <Stack spacing={3} sx={{ position: "relative" }}>
          <Stack spacing={1}>
            <Chip
              label={isLogin ? "Continue building" : "Join the network"}
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

          <Tabs
            value={activeTab}
            onChange={handleTabChange}
            textColor="primary"
            indicatorColor="primary"
            sx={{
              width: "fit-content",
              "& .MuiTab-root": { textTransform: "none", fontWeight: 600 },
            }}
          >
            <Tab label="Sign in" value="login" />
            <Tab label="Join Aatist" value="register" />
          </Tabs>

          <Collapse in={Boolean(error)}>
            <Alert severity="error" variant="outlined">
              {error}
            </Alert>
          </Collapse>

          {isLogin ? renderLoginForm() : renderRegisterContent()}
        </Stack>
      </DialogContent>
    </Dialog>
  );
}
