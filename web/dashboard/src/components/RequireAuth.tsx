import { type ReactNode } from "react";
import { Navigate } from "react-router-dom";

// Guards dashboard routes: redirects to login if no valid session in localStorage.
export default function RequireAuth({ children }: { children: ReactNode }) {
  const token = localStorage.getItem("token");
  const userStr = localStorage.getItem("user");

  if (!token || !userStr) {
    return <Navigate to="/" replace />;
  }

  try {
    JSON.parse(userStr);
  } catch {
    localStorage.clear();
    return <Navigate to="/" replace />;
  }

  return <>{children}</>;
}
