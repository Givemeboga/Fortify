// Central backend address. Local dev falls back to localhost:8500;
// Docker/cloud bakes the public backend URL in at build time:
//
//   VITE_API_URL=https://api.example.com npm run build
//
// (docker-compose.yml wires this through automatically via its own
// VITE_API_URL variable.)
export const API_BASE = (
  import.meta.env.VITE_API_URL || "http://localhost:8500"
).replace(/\/$/, "")

// api("/scans") → "<base>/scans"
export const api = (path) => `${API_BASE}${path.startsWith("/") ? path : `/${path}`}`
