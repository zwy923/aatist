import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import { CssBaseline, ThemeProvider, createTheme } from "@mui/material";
import Landing from "./pages/Landing.jsx";
import Login from "./pages/Login.jsx";
import ClientLogin from "./pages/ClientLogin.jsx";
import StudentLogin from "./pages/StudentLogin.jsx";
import AuthCallback from "./pages/AuthCallback.jsx";
import Register from "./pages/Register.jsx";
import Dashboard from "./pages/Dashboard.jsx";
import Profile from "./pages/Profile.jsx";
import Verify from "./pages/Verify.jsx";
import MessagesPage from "./pages/Messages.jsx";
import OpportunitiesPage from "./pages/Opportunities.jsx";
import OpportunityDetailPage from "./pages/OpportunityDetail.jsx";
import TalentsPage from "./pages/Talents.jsx";
import PublicProfilePage from "./pages/PublicProfile.jsx";
import AboutPage from "./pages/About.jsx";
import TermsPage from "./pages/Terms.jsx";
import AuthSessionGate from "./shared/components/auth/AuthSessionGate.jsx";
import { ChatProvider } from "./features/messages/ChatProvider.jsx";

import "./styles/global.css";

const theme = createTheme({
  palette: {
    mode: "light",
    primary: { main: "#1976d2" },
    secondary: { main: "#7b1fa2" },
    background: {
      default: "#f5f7fb",
      paper: "#ffffff",
    },
    text: {
      primary: "#1a1a1a",
      secondary: "#666666",
    },
  },
  shape: { borderRadius: 16 },
  typography: {
    fontFamily: '"Inter", "Space Grotesk", "Roboto", "Helvetica", sans-serif',
  },
  components: {
    MuiDialog: {
      styleOverrides: {
        paper: {
          borderRadius: 16,
          border: "1px solid #e5e7eb",
          backgroundImage: "none",
        },
      },
    },
  },
});

ReactDOM.createRoot(document.getElementById("root")).render(
  <React.StrictMode>
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <BrowserRouter
        future={{
          v7_startTransition: true,
          v7_relativeSplatPath: true,
        }}
      >
        <AuthSessionGate>
          <ChatProvider>
          <Routes>
            <Route path="/" element={<Landing />} />
            <Route path="/auth/login" element={<Login />} />
            <Route path="/auth/login/client" element={<ClientLogin />} />
            <Route path="/auth/login/student" element={<StudentLogin />} />
            <Route path="/auth/callback" element={<AuthCallback />} />
            <Route path="/auth/register" element={<Register />} />
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/profile" element={<Profile />} />
            <Route path="/verify" element={<Verify />} />
            <Route path="/messages" element={<MessagesPage />} />
            <Route path="/opportunities" element={<OpportunitiesPage />} />
            <Route path="/opportunities/:id" element={<OpportunityDetailPage />} />
            <Route path="/talents" element={<TalentsPage />} />
            <Route path="/about" element={<AboutPage />} />
            <Route path="/terms" element={<TermsPage />} />
            <Route path="/users/:id" element={<PublicProfilePage />} />
          </Routes>
          </ChatProvider>
        </AuthSessionGate>
      </BrowserRouter>
    </ThemeProvider>
  </React.StrictMode>
);
