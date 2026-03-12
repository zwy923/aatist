import React from "react";
import { Navigate } from "react-router-dom";

export default function ClientLogin() {
  return <Navigate to="/auth/login?mode=client" replace />;
}
