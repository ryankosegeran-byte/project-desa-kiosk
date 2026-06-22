// Same-origin in production (SPA served by the Go server on serv00).
// In dev, Vite proxies /api -> http://localhost:3000 (see vite.config.ts).
// Override with VITE_API_URL if the API lives on a different host.
const BASE_URL = import.meta.env.VITE_API_URL ?? "";

// Exposed for callers that need a raw fetch (e.g. binary/PDF responses).
export const API_BASE = BASE_URL;

export async function request(path: string, options: RequestInit = {}) {
  if (typeof window === "undefined") {
    return null;
  }
  const token = localStorage.getItem("token");
  
  const headers = new Headers(options.headers);
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  if (!(options.body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers,
  });

  if (response.status === 401 && path !== "/api/auth/login") {
    localStorage.clear();
    window.location.href = "/";
    throw new Error("Sesi telah berakhir, silakan login kembali");
  }

  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.error || "Terjadi kesalahan pada server");
  }

  return data;
}

// authFetch is a raw fetch (returns the Response) for binary responses like PDF
// or DOCX download, but applies the same auth + 401 handling as request():
// on an expired/invalid token it logs out and redirects to login.
export async function authFetch(path: string, options: RequestInit = {}): Promise<Response> {
  const token = localStorage.getItem("token");
  const headers = new Headers(options.headers);
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  const response = await fetch(`${BASE_URL}${path}`, { ...options, headers });
  if (response.status === 401) {
    localStorage.clear();
    window.location.href = "/";
    throw new Error("Sesi telah berakhir, silakan login kembali");
  }
  return response;
}

export function getUser() {
  if (typeof window === "undefined") {
    return null;
  }
  const userStr = localStorage.getItem("user");
  if (!userStr) return null;
  try {
    return JSON.parse(userStr);
  } catch (e) {
    return null;
  }
}

