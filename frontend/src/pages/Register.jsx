import React, { useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  FormControl,
  FormControlLabel,
  ListSubheader,
  MenuItem,
  Select,
  TextField,
  Typography,
} from "@mui/material";
import { useAuth } from "../features/auth/hooks/useAuth";
import { getGoogleOAuthStartURL } from "../shared/utils/googleOAuth";
import "./Landing.css";
import "./AuthRegister.css";
import "./AuthLogin.css";
import {
  AALTO_PROGRAMMES,
  AALTO_REGISTRATION_SCHOOL_OPTIONS,
  buildRegistrationAcademicFields,
  getProgrammeByName,
  programmeMatchesSchoolFilter,
} from "../constants/aaltoProgrammes";
import { aaltoOutlinedSelectSx, aaltoSelectMenuProps } from "../shared/styles/aaltoSelectSx";

const CLIENT_LETTERS = [
  { char: "C", x: 20, y: 80, r: -20 },
  { char: "L", x: 100, y: 60, r: 22 },
  { char: "I", x: 220, y: 20, r: -38 },
  { char: "E", x: 340, y: 90, r: -5 },
  { char: "N", x: 440, y: -30, r: 28 },
  { char: "T", x: 540, y: 70, r: -10 },
];

const STUDENT_LETTERS = [
  { char: "A", x: 20, y: 70, r: 10 },
  { char: "A", x: 120, y: 40, r: -32 },
  { char: "T", x: 240, y: 10, r: 24 },
  { char: "I", x: 360, y: 50, r: -3 },
  { char: "S", x: 460, y: 80, r: -18 },
  { char: "T", x: 560, y: 30, r: 5, scale: 1.15 },
];

const createClientForm = () => ({
  name: "",
  company: "",
  title: "",
  email: "",
  password: "",
  confirmPassword: "",
  agreed: false,
});

const createStudentForm = () => ({
  name: "",
  preferredName: "",
  school: "",
  program: "",
  yearOfEnrollment: "",
  emailLocalPart: "",
  password: "",
  confirmPassword: "",
  agreed: false,
});

function Register() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { register, loading } = useAuth();
  const [step, setStep] = useState(() => {
    const mode = searchParams.get("mode");
    if (mode === "client" || mode === "student") return mode;
    return "select";
  });

  useEffect(() => {
    const mode = searchParams.get("mode");
    if (mode === "client" || mode === "student") setStep(mode);
  }, [searchParams]);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [clientForm, setClientForm] = useState(createClientForm());
  const [studentForm, setStudentForm] = useState(createStudentForm());

  const updateClient = (field, value) => {
    setClientForm((prev) => ({ ...prev, [field]: value }));
  };
  const updateStudent = (field, value) => {
    setStudentForm((prev) => {
      const next = { ...prev, [field]: value };
      if (field === "school") next.program = "";
      return next;
    });
  };

  const programmesBySchool = useMemo(() => {
    const filtered = AALTO_PROGRAMMES.filter((p) => programmeMatchesSchoolFilter(p.school, studentForm.school));
    const map = new Map();
    for (const p of filtered) {
      if (!map.has(p.school)) map.set(p.school, []);
      map.get(p.school).push(p);
    }
    const entries = [...map.entries()].map(([schoolName, progs]) => [
      schoolName,
      [...progs].sort((a, b) => a.name.localeCompare(b.name)),
    ]);
    entries.sort(([a], [b]) => a.localeCompare(b));
    return entries;
  }, [studentForm.school]);

  useEffect(() => {
    setStudentForm((prev) => {
      const visible = AALTO_PROGRAMMES.filter((p) => programmeMatchesSchoolFilter(p.school, prev.school));
      if (prev.program && !visible.some((p) => p.name === prev.program)) {
        return { ...prev, program: "" };
      }
      return prev;
    });
  }, [studentForm.school]);

  const validatePassword = (password) => password.length >= 8;

  const submitClient = async (event) => {
    event.preventDefault();
    setError("");
    setSuccess("");

    if (!clientForm.agreed) {
      setError("Please agree to the Terms & Privacy Policy.");
      return;
    }
    if (!clientForm.name.trim()) {
      setError("Name is required.");
      return;
    }
    if (!validatePassword(clientForm.password)) {
      setError("Password must be at least 8 characters.");
      return;
    }
    if (clientForm.password !== clientForm.confirmPassword) {
      setError("Passwords do not match.");
      return;
    }
    if (!clientForm.company.trim()) {
      setError("Company name is required.");
      return;
    }

    const payload = {
      name: clientForm.name.trim(),
      email: clientForm.email.trim(),
      password: clientForm.password,
      role: "organization",
      profile: {
        organizationName: clientForm.company.trim(),
        contactTitle: clientForm.title.trim(),
      },
    };

    const result = await register(payload);
    if (!result.success) {
      setError(result.error || "Registration failed.");
      return;
    }

    setSuccess(result.autoLogin ? "Account created successfully." : "Account created successfully. Please sign in.");
    setClientForm(createClientForm());
    setTimeout(() => navigate(result.autoLogin ? "/talents" : "/auth/login/client"), 700);
  };

  const submitStudent = async (event) => {
    event.preventDefault();
    setError("");
    setSuccess("");

    const email = `${studentForm.emailLocalPart.trim()}@aalto.fi`;
    if (!studentForm.school) {
      setError("Please select your school.");
      return;
    }
    if (!studentForm.agreed) {
      setError("Please confirm you are an Aalto student and agree to the policy.");
      return;
    }
    if (!studentForm.name.trim()) {
      setError("Name is required.");
      return;
    }
    if (!studentForm.program.trim()) {
      setError("Please select your degree programme.");
      return;
    }
    const selectedProgramme = getProgrammeByName(studentForm.program);
    if (!selectedProgramme) {
      setError("Please select a valid degree programme.");
      return;
    }
    const academic = buildRegistrationAcademicFields(selectedProgramme);
    if (!academic) {
      setError("Could not save academic information. Please try again.");
      return;
    }
    if (!studentForm.emailLocalPart.trim()) {
      setError("Aalto email address is required.");
      return;
    }
    if (!validatePassword(studentForm.password)) {
      setError("Password must be at least 8 characters.");
      return;
    }
    if (studentForm.password !== studentForm.confirmPassword) {
      setError("Passwords do not match.");
      return;
    }

    const payload = {
      name: studentForm.name.trim(),
      email,
      password: studentForm.password,
      role: "student",
      profile: {
        school: academic.school,
        faculty: academic.faculty,
        major: academic.major,
        studentId: studentForm.yearOfEnrollment.trim(),
      },
    };

    const result = await register(payload);
    if (!result.success) {
      setError(result.error || "Registration failed.");
      return;
    }

    setSuccess(result.autoLogin ? "Student account created successfully." : "Student account created successfully. Please sign in.");
    setStudentForm(createStudentForm());
    setTimeout(() => navigate(result.autoLogin ? "/talents" : "/auth/login/student"), 700);
  };

  // Step 1: Join selection
  if (step === "select") {
    return (
      <main className="register-auth-page register-select-step">
        <header className="landing-header">
          <Link to="/" className="brand" aria-label="Aatist Home">
            <span className="brand-icon">A</span>
            <span className="brand-text">atist</span>
          </Link>
          <nav className="landing-nav" aria-label="Primary">
            <Link to="/" className="nav-link active">Home</Link>
            <Link to="/talents" className="nav-link">Hire Talent</Link>
            <Link to="/opportunities" className="nav-link">Opportunities</Link>
          </nav>
          <div className="nav-actions">
            <Link to="/auth/register" className="nav-btn nav-btn-signup">Sign up</Link>
            <Link to="/auth/login" className="nav-btn nav-btn-login">Log in</Link>
          </div>
        </header>

        <section className="register-hero register-hero-split">
          <div className="register-split-left" />
          <div className="register-split-right" />
          <div className="register-hero-word register-hero-word-client" aria-hidden="true">
            {CLIENT_LETTERS.map((l, i) => (
              <span key={i} className="register-hero-letter" style={{ transform: `translate(${l.x}px, ${l.y}px) rotate(${l.r}deg)` }}>{l.char}</span>
            ))}
          </div>
          <div className="register-hero-word register-hero-word-student" aria-hidden="true">
            {STUDENT_LETTERS.map((l, i) => (
              <span key={i} className={`register-hero-letter ${l.scale ? "register-hero-letter-large" : ""}`} style={{ transform: `translate(${l.x}px, ${l.y}px) rotate(${l.r}deg)` }}>{l.char}</span>
            ))}
          </div>

          <article className="register-card register-select-card">
            <Link to="/" className="register-back-link">← Back</Link>
            <h1>Join Aatist !</h1>
            <p className="register-select-sub">Select how you&apos;ll use the platform</p>

            <div className="register-role-tiles">
              <button type="button" className="register-tile register-tile-client" onClick={() => setStep("client")}>
                <span className="tile-letter-big">C</span>
                <span className="tile-letters-stack">LIENT</span>
              </button>
              <button type="button" className="register-tile register-tile-artist" onClick={() => setStep("student")}>
                <span className="tile-letter-big">A</span>
                <span className="tile-letters-stack">ATIST</span>
              </button>
            </div>

            <p className="register-footer-link">
              Already have an account? <Link to="/auth/login">Log in</Link>
            </p>
          </article>
        </section>
      </main>
    );
  }

  // Step 2: Client registration form
  if (step === "client") {
    return (
      <main className="register-auth-page register-client-form-step">
        <header className="landing-header">
          <Link to="/" className="brand" aria-label="Aatist Home">
            <span className="brand-icon">A</span>
            <span className="brand-text">atist</span>
          </Link>
          <nav className="landing-nav" aria-label="Primary">
            <Link to="/" className="nav-link active">Home</Link>
            <Link to="/talents" className="nav-link">Hire Talent</Link>
            <Link to="/opportunities" className="nav-link">Opportunities</Link>
          </nav>
          <div className="nav-actions">
            <Link to="/auth/register" className="nav-btn nav-btn-signup">Sign up</Link>
            <Link to="/auth/login" className="nav-btn nav-btn-login">Log in</Link>
          </div>
        </header>

        <section className="register-hero register-hero-client">
          <div className="register-hero-word register-hero-word-client" aria-hidden="true">
            {CLIENT_LETTERS.map((l, i) => (
              <span key={i} className="register-hero-letter" style={{ transform: `translate(${l.x}px, ${l.y}px) rotate(${l.r}deg)` }}>{l.char}</span>
            ))}
          </div>

          <article className="register-card register-form-card">
            <button
              type="button"
              className="register-back-link"
              onClick={() => {
                setStep("select");
                setSearchParams({});
              }}
            >
              ← Back
            </button>
            <h1>Join Aatist as a <span className="register-title-accent">Client</span></h1>
            <p className="register-sub-text">It only takes 1 minute : )</p>

            {error && <Alert severity="error" sx={{ mt: 2 }}>{error}</Alert>}
            {success && <Alert severity="success" sx={{ mt: 2 }}>{success}</Alert>}

            <div className="login-divider">
              <span>Or continue with</span>
            </div>
            <div className="social-row">
              <button
                type="button"
                className="social-btn"
                aria-label="Google"
                onClick={() => {
                  window.location.href = getGoogleOAuthStartURL();
                }}
              >
                <span className="social-google">G</span>
              </button>
              <button type="button" className="social-btn" aria-label="Apple">
                <svg className="apple-logo" viewBox="0 0 24 24" fill="currentColor" width="20" height="20" aria-hidden>
                  <path d="M17.05 20.28c-.98.95-2.05.8-3.08.35-1.09-.46-2.09-.48-3.24 0-1.44.62-2.2.44-3.06-.35C2.79 15.25 3.51 7.59 9.05 7.31c1.35.07 2.29.74 3.08.8 1.18-.24 2.31-.93 3.57-.84 1.51.12 2.65.72 3.4 1.8-3.12 1.87-2.38 5.98.48 7.13-.57 1.5-1.31 2.99-2.54 4.09l.01-.01zM12.03 7.25c-.15-2.23 1.66-4.07 3.74-4.42.42 2.43-2.43 4.66-3.74 4.42z" />
                </svg>
              </button>
            </div>
            <Box sx={{ display: "flex", alignItems: "center", gap: 1, my: 2, color: "text.secondary" }}>
              <Box sx={{ flex: 1, height: "1px", bgcolor: "divider" }} />
              <Typography variant="caption">Or sign up with email</Typography>
              <Box sx={{ flex: 1, height: "1px", bgcolor: "divider" }} />
            </Box>

            <Box component="form" className="register-form register-client-form" onSubmit={submitClient}>
              <div className="register-section">
                <span className="register-section-title">Basic Information</span>
                <div className="register-section-line" />
              </div>
              <div className="register-form-row">
                <div className="register-field">
                  <label>Name<span className="required">*</span></label>
                  <TextField
                    placeholder="Your name"
                    value={clientForm.name}
                    onChange={(e) => updateClient("name", e.target.value)}
                    required
                    fullWidth
                    size="small"
                  />
                </div>
                <div className="register-field">
                  <label>Company Name</label>
                  <TextField
                    placeholder="Your company"
                    value={clientForm.company}
                    onChange={(e) => updateClient("company", e.target.value)}
                    fullWidth
                    size="small"
                  />
                </div>
                <div className="register-field">
                  <label>Role / Title</label>
                  <TextField
                    placeholder="Your role"
                    value={clientForm.title}
                    onChange={(e) => updateClient("title", e.target.value)}
                    fullWidth
                    size="small"
                  />
                </div>
              </div>
              <div className="register-form-row register-account-row">
                <div className="register-field">
                  <label>Email Address<span className="required">*</span></label>
                  <TextField
                    type="email"
                    placeholder="you@company.com"
                    value={clientForm.email}
                    onChange={(e) => updateClient("email", e.target.value)}
                    required
                    fullWidth
                    size="small"
                  />
                </div>
                <div className="register-field">
                  <label>Password<span className="required">*</span></label>
                  <TextField
                    type="password"
                    placeholder="(8+ characters)"
                    value={clientForm.password}
                    onChange={(e) => updateClient("password", e.target.value)}
                    required
                    fullWidth
                    size="small"
                  />
                </div>
                <div className="register-field">
                  <label>Confirm Password<span className="required">*</span></label>
                  <TextField
                    type="password"
                    placeholder="Confirm password"
                    value={clientForm.confirmPassword}
                    onChange={(e) => updateClient("confirmPassword", e.target.value)}
                    required
                    fullWidth
                    size="small"
                  />
                </div>
              </div>

              <FormControlLabel
                control={
                  <input
                    type="checkbox"
                    checked={clientForm.agreed}
                    onChange={(e) => updateClient("agreed", e.target.checked)}
                    className="register-checkbox-input"
                  />
                }
                label={
                  <span>
                    I agree to <Link to="/terms" className="register-terms-link">Terms & Privacy Policy</Link>
                  </span>
                }
                className="register-checkbox-label"
              />

              <div className="register-submit-wrap">
                <Button type="submit" className="register-submit register-submit-teal" disabled={loading}>
                  {loading ? <CircularProgress size={20} color="inherit" /> : "Create Account"}
                </Button>
              </div>
            </Box>
          </article>
        </section>
      </main>
    );
  }

  // Step 2: Student registration form
  return (
    <main className="register-auth-page register-student-form-step">
      <header className="landing-header">
        <Link to="/" className="brand" aria-label="Aatist Home">
          <span className="brand-icon">A</span>
          <span className="brand-text">atist</span>
        </Link>
        <nav className="landing-nav" aria-label="Primary">
          <Link to="/" className="nav-link active">Home</Link>
          <Link to="/talents" className="nav-link">Hire Talent</Link>
          <Link to="/opportunities" className="nav-link">Opportunities</Link>
        </nav>
        <div className="nav-actions">
          <Link to="/auth/register" className="nav-btn nav-btn-signup">Sign up</Link>
          <Link to="/auth/login" className="nav-btn nav-btn-login">Log in</Link>
        </div>
      </header>

      <section className="register-hero register-hero-student">
        <div className="register-hero-word register-hero-word-student" aria-hidden="true">
          {STUDENT_LETTERS.map((l, i) => (
            <span key={i} className={`register-hero-letter ${l.scale ? "register-hero-letter-large" : ""}`} style={{ transform: `translate(${l.x}px, ${l.y}px) rotate(${l.r}deg)` }}>{l.char}</span>
          ))}
        </div>

        <article className="register-card register-form-card register-student-card">
          <button
            type="button"
            className="register-back-link"
            onClick={() => {
              setStep("select");
              setSearchParams({});
            }}
          >
            ← Back
          </button>
          <h1>Join Aatist as a verified Aalto <span className="register-title-accent register-title-accent-blue">Student</span></h1>
          <p className="register-sub-text register-sub-text-blue">It only takes 1 minute : )</p>

          {error && <Alert severity="error" sx={{ mt: 2 }}>{error}</Alert>}
          {success && <Alert severity="success" sx={{ mt: 2 }}>{success}</Alert>}

          <Box component="form" className="register-form register-student-form" onSubmit={submitStudent}>
            <div className="register-section">
              <span className="register-section-title">Basic Information</span>
              <div className="register-section-line" />
            </div>
            <div className="register-form-row register-form-row-2">
              <div className="register-field">
                <label>Name<span className="required">*</span></label>
                <TextField
                  placeholder="Firstname Lastname"
                  value={studentForm.name}
                  onChange={(e) => updateStudent("name", e.target.value)}
                  required
                  fullWidth
                  size="small"
                />
              </div>
              <div className="register-field">
                <label>Preferred Name</label>
                <TextField
                  placeholder="e.g. Emma"
                  value={studentForm.preferredName}
                  onChange={(e) => updateStudent("preferredName", e.target.value)}
                  fullWidth
                  size="small"
                />
              </div>
            </div>

            <div className="register-section">
              <span className="register-section-title">Academic Information</span>
              <div className="register-section-line" />
            </div>
            <div className="register-form-row register-form-row-2">
              <div className="register-field">
                <label>School<span className="required">*</span></label>
                <FormControl fullWidth required size="small" variant="outlined" sx={aaltoOutlinedSelectSx}>
                  <Select
                    displayEmpty
                    value={studentForm.school}
                    onChange={(e) => updateStudent("school", e.target.value)}
                    MenuProps={aaltoSelectMenuProps}
                    inputProps={{ "aria-label": "School" }}
                  >
                    <MenuItem value="" disabled>
                      Select school
                    </MenuItem>
                    {AALTO_REGISTRATION_SCHOOL_OPTIONS.map((s) => (
                      <MenuItem key={s} value={s}>
                        {s}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </div>
              <div className="register-field">
                <label>Year of enrollment<span className="required">*</span></label>
                <TextField
                  placeholder="e.g. 2023"
                  value={studentForm.yearOfEnrollment}
                  onChange={(e) => updateStudent("yearOfEnrollment", e.target.value)}
                  required
                  fullWidth
                  size="small"
                />
              </div>
            </div>
            <div className="register-form-row">
              <div className="register-field register-field-full">
                <label>Degree programme<span className="required">*</span></label>
                <FormControl
                  fullWidth
                  required
                  size="small"
                  variant="outlined"
                  disabled={!studentForm.school}
                  sx={aaltoOutlinedSelectSx}
                >
                  <Select
                    displayEmpty
                    value={studentForm.program}
                    onChange={(e) => updateStudent("program", e.target.value)}
                    MenuProps={aaltoSelectMenuProps}
                    inputProps={{ "aria-label": "Degree programme" }}
                    renderValue={(v) =>
                      v || (
                        <span style={{ color: "#94a3b8", fontWeight: 500 }}>
                          {studentForm.school ? "Select programme" : "Select school first"}
                        </span>
                      )
                    }
                  >
                    <MenuItem value="" disabled>
                      {studentForm.school ? "Select programme" : "Select school first"}
                    </MenuItem>
                    {programmesBySchool.map(([schoolName, progs]) => [
                      <ListSubheader key={`reg-h-${schoolName}`} disableSticky>
                        {schoolName}
                      </ListSubheader>,
                      ...progs.map((p) => (
                        <MenuItem key={`reg-${schoolName}::${p.name}`} value={p.name}>
                          {p.name}
                        </MenuItem>
                      )),
                    ])}
                  </Select>
                </FormControl>
              </div>
            </div>

            <div className="register-section">
              <span className="register-section-title">Account</span>
              <div className="register-section-line" />
            </div>
            <div className="register-form-row register-form-row-2">
              <div className="register-field register-field-full">
                <label>Aalto Email<span className="required">*</span></label>
                <div className="aalto-email-field">
                  <TextField
                    placeholder="your.name"
                    value={studentForm.emailLocalPart}
                    onChange={(e) => updateStudent("emailLocalPart", e.target.value)}
                    required
                    fullWidth
                    size="small"
                  />
                  <span className="aalto-suffix">@aalto.fi</span>
                </div>
              </div>
            </div>
            <div className="register-form-row register-form-row-2">
              <div className="register-field">
                <label>Password<span className="required">*</span></label>
                <TextField
                  type="password"
                  placeholder="(8+ characters)"
                  value={studentForm.password}
                  onChange={(e) => updateStudent("password", e.target.value)}
                  required
                  fullWidth
                  size="small"
                />
              </div>
              <div className="register-field">
                <label>Confirm Password<span className="required">*</span></label>
                <TextField
                  type="password"
                  placeholder="(8+ characters)"
                  value={studentForm.confirmPassword}
                  onChange={(e) => updateStudent("confirmPassword", e.target.value)}
                  required
                  fullWidth
                  size="small"
                />
              </div>
            </div>

            <FormControlLabel
              control={
                <input
                  type="checkbox"
                  checked={studentForm.agreed}
                  onChange={(e) => updateStudent("agreed", e.target.checked)}
                  className="register-checkbox-input"
                />
              }
              label={
                <span>
                  I agree to <Link to="/terms" className="register-terms-link">Terms & Privacy Policy</Link>
                </span>
              }
              className="register-checkbox-label"
            />

            <div className="register-submit-wrap">
              <Button type="submit" className="register-submit register-submit-blue" disabled={loading}>
                {loading ? <CircularProgress size={20} color="inherit" /> : "Create Account"}
              </Button>
            </div>
          </Box>
        </article>
      </section>
    </main>
  );
}

export default Register;
