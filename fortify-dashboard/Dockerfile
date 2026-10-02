# --- Stage 1: build the static bundle ---
FROM node:20-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
# Public backend URL baked into the bundle (compose passes this through;
# defaults to local dev). The browser calls this address, so in the cloud it
# must be the backend's *public* URL, not an internal hostname.
ARG VITE_API_URL=http://localhost:8500
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

# --- Stage 2: serve the built files ---
FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
# nginx's base image already serves /usr/share/nginx/html by default — no CMD needed
