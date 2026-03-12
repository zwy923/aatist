import React from "react";
import { Navigate } from "react-router-dom";

export default function StudentLogin() {
  return <Navigate to="/auth/login?mode=student" replace />;
}
