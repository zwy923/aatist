import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import { CssBaseline, ThemeProvider, createTheme } from "@mui/material";
import App from "./App.jsx";
import Verify from "./pages/Verify.jsx";
import { UserProvider } from "./store/userStore.jsx";
import "./styles/global.css";

const theme = createTheme({
  palette: {
    mode: "dark",
    primary: { main: "#5de0ff" },
    secondary: { main: "#ffb877" },
    background: {
      default: "#030617",
      paper: "rgba(7, 12, 30, 0.96)",
    },
    text: {
      primary: "#f5fbff",
      secondary: "rgba(245, 251, 255, 0.65)",
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
          borderRadius: 24,
          border: "1px solid rgba(93, 224, 255, 0.25)",
          backgroundImage:
            "radial-gradient(circle at top, rgba(27,37,78,0.95), rgba(7,10,28,0.95))",
        },
      },
    },
  },
});

ReactDOM.createRoot(document.getElementById("root")).render(
  <React.StrictMode>
    <ThemeProvider theme={theme}>
      <CssBaseline />
    <UserProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<App />} />
          <Route path="/verify" element={<Verify />} />
        </Routes>
      </BrowserRouter>
    </UserProvider>
    </ThemeProvider>
  </React.StrictMode>
);
