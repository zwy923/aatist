import React, { useMemo, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  FormControl,
  MenuItem,
  Select,
  TextField,
  Typography,
} from "@mui/material";
import { useAuth } from "../features/auth/hooks/useAuth";
import "./AuthRegister.css";

const SCHOOL_OPTIONS = [
  "School of Arts, Design and Architecture",
  "School of Business",
  "School of Chemical Engineering",
  "School of Electrical Engineering",
  "School of Engineering",
  "School of Science",
];

const DEPARTMENT_OPTIONS = [
  "Design",
  "Media",
  "Architecture",
  "Film",
  "Visual Communication",
  "Other",
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
  department: "",
  program: "",
  yearOfEnrollment: "",
  emailLocalPart: "",
  password: "",
  confirmPassword: "",
  agreed: false,
});

function Register() {
  const navigate = useNavigate();
  const { register, loading } = useAuth();
  const [mode, setMode] = useState("client");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [clientForm, setClientForm] = useState(createClientForm());
  const [studentForm, setStudentForm] = useState(createStudentForm());

  const titleText = useMemo(
    () =>
      mode === "client"
        ? "Join Aatist as a Client."
        : "Join Aatist as a verified Aalto student.",
    [mode]
  );

  const updateClient = (field, value) => {
    setClientForm((prev) => ({ ...prev, [field]: value }));
  };
  const updateStudent = (field, value) => {
    setStudentForm((prev) => ({ ...prev, [field]: value }));
  };

  const validatePassword = (password) => password.length >= 8;

  const submitClient = async (event) => {
    event.preventDefault();
    setError("");
    setSuccess("");

    if (!clientForm.agreed) {
      setError("Please agree to the Terms & Privacy Policy.");
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
    setTimeout(() => navigate(result.autoLogin ? "/dashboard" : "/auth/login/client"), 700);
  };

  const submitStudent = async (event) => {
    event.preventDefault();
    setError("");
    setSuccess("");

    const email = `${studentForm.emailLocalPart.trim()}@aalto.fi`;
    if (!studentForm.agreed) {
      setError("Please confirm you are an Aalto student and agree to the policy.");
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
        school: studentForm.school,
        faculty: studentForm.department,
        major: studentForm.program.trim(),
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
    setTimeout(() => navigate(result.autoLogin ? "/dashboard" : "/auth/login/student"), 700);
  };

  return (
    <main className={`register-auth-page ${mode === "client" ? "register-client-mode" : ""}`}>
      <header className="register-header">
        <Link to="/" className="register-brand" aria-label="Aatist Home">
          <span className="register-brand-icon">A</span>
          <span className="register-brand-text">atist</span>
        </Link>
        <nav className="register-nav" aria-label="Primary">
          <Link to="/talents" className="register-nav-link">Hire Talent</Link>
          <Link to="/opportunities" className="register-nav-link">Opportunities</Link>
          <Link to="/about" className="register-nav-link">About</Link>
        </nav>
      </header>

      <section className={`register-hero ${mode === "student" ? "student-bg" : "client-bg"}`}>
        <article className="register-card">
          <div className="register-card-logo">A atist.</div>
          <h1>{titleText}</h1>
          <p>It only takes 1 minute :)</p>

          {mode === "client" && (
            <div className="register-mode-toggle" role="tablist" aria-label="Register role">
              <button
                type="button"
                className="active"
                onClick={() => setMode("client")}
              >
                Client
              </button>
              <button
                type="button"
                onClick={() => setMode("student")}
              >
                Student
              </button>
            </div>
          )}

          {error && <Alert severity="error">{error}</Alert>}
          {success && <Alert severity="success">{success}</Alert>}

          {mode === "client" ? (
            <Box component="form" className="register-form" onSubmit={submitClient}>
              <TextField
                placeholder="Full Name"
                value={clientForm.name}
                onChange={(e) => updateClient("name", e.target.value)}
                required
                fullWidth
              />
              <TextField
                placeholder="Company"
                value={clientForm.company}
                onChange={(e) => updateClient("company", e.target.value)}
                required
                fullWidth
              />
              <TextField
                placeholder="Role"
                value={clientForm.title}
                onChange={(e) => updateClient("title", e.target.value)}
                required
                fullWidth
              />
              <TextField
                type="email"
                placeholder="Email"
                value={clientForm.email}
                onChange={(e) => updateClient("email", e.target.value)}
                required
                fullWidth
              />
              <TextField
                type="password"
                placeholder="Password (6+ characters)"
                value={clientForm.password}
                onChange={(e) => updateClient("password", e.target.value)}
                required
                fullWidth
              />
              <TextField
                type="password"
                placeholder="Confirm Password"
                value={clientForm.confirmPassword}
                onChange={(e) => updateClient("confirmPassword", e.target.value)}
                required
                fullWidth
              />

              <label className="register-checkbox">
                <input
                  type="checkbox"
                  checked={clientForm.agreed}
                  onChange={(e) => updateClient("agreed", e.target.checked)}
                />
                <span>
                  I agree to the <Link to="/terms" className="register-terms-link">Terms & Privacy Policy</Link>.
                </span>
              </label>

              <div className="register-submit-wrap">
                <Button type="submit" className="register-submit" disabled={loading}>
                  {loading ? <CircularProgress size={20} color="inherit" /> : "Create Account"}
                </Button>
              </div>
            </Box>
          ) : (
            <Box component="form" className="register-form" onSubmit={submitStudent}>
              <TextField
                placeholder="Full Name"
                value={studentForm.name}
                onChange={(e) => updateStudent("name", e.target.value)}
                required
                fullWidth
              />
              <TextField
                placeholder="Preferred Name (Optional)"
                value={studentForm.preferredName}
                onChange={(e) => updateStudent("preferredName", e.target.value)}
                fullWidth
              />
              <FormControl fullWidth required>
                <Select
                  displayEmpty
                  value={studentForm.school}
                  onChange={(e) => updateStudent("school", e.target.value)}
                >
                  <MenuItem value="" disabled>
                    School (Drop-Down)
                  </MenuItem>
                  {SCHOOL_OPTIONS.map((item) => (
                    <MenuItem key={item} value={item}>
                      {item}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
              <FormControl fullWidth required>
                <Select
                  displayEmpty
                  value={studentForm.department}
                  onChange={(e) => updateStudent("department", e.target.value)}
                >
                  <MenuItem value="" disabled>
                    Department (Drop-Down)
                  </MenuItem>
                  {DEPARTMENT_OPTIONS.map((item) => (
                    <MenuItem key={item} value={item}>
                      {item}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
              <TextField
                placeholder="Program"
                value={studentForm.program}
                onChange={(e) => updateStudent("program", e.target.value)}
                required
                fullWidth
              />
              <TextField
                placeholder="Year of enrollment"
                value={studentForm.yearOfEnrollment}
                onChange={(e) => updateStudent("yearOfEnrollment", e.target.value)}
                required
                fullWidth
              />
              <div className="aalto-email-input">
                <TextField
                  placeholder="Aalto Email Address"
                  value={studentForm.emailLocalPart}
                  onChange={(e) => updateStudent("emailLocalPart", e.target.value)}
                  required
                  fullWidth
                />
                <span>@aalto.fi</span>
              </div>
              <TextField
                type="password"
                placeholder="Password (6+ characters)"
                value={studentForm.password}
                onChange={(e) => updateStudent("password", e.target.value)}
                required
                fullWidth
              />
              <TextField
                type="password"
                placeholder="Confirm Password"
                value={studentForm.confirmPassword}
                onChange={(e) => updateStudent("confirmPassword", e.target.value)}
                required
                fullWidth
              />

              <label className="register-checkbox">
                <input
                  type="checkbox"
                  checked={studentForm.agreed}
                  onChange={(e) => updateStudent("agreed", e.target.checked)}
                />
                <span>
                  I confirm that I am an Aalto student and agree to the{" "}
                  <Link to="/terms" className="register-terms-link">Terms & Privacy Policy</Link>.
                </span>
              </label>

              <div className="register-submit-wrap">
                <Button type="submit" className="register-submit" disabled={loading}>
                  {loading ? <CircularProgress size={20} color="inherit" /> : "Create Account"}
                </Button>
              </div>
            </Box>
          )}

          <p className="register-signin-link">
            Already have an account? <Link to="/auth/login">Sign in</Link>
            {mode === "student" && (
              <> · <button type="button" className="register-mode-link" onClick={() => setMode("client")}>Register as client</button></>
            )}
            {mode === "client" && (
              <> · <button type="button" className="register-mode-link" onClick={() => setMode("student")}>Register as student</button></>
            )}
          </p>
        </article>
      </section>
    </main>
  );
}

export default Register;

