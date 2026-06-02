# Multi-stage Dockerfile for Angular Admin Dashboard
# Optimized for Choreo deployment

# Stage 1: Build the Angular application
FROM node:20 AS build

WORKDIR /app

# Copy package files and install dependencies
COPY package*.json ./
RUN npm ci --legacy-peer-deps

# Copy source code
COPY . .

# Build the Angular app for production
RUN npm run build

# Stage 2: Serve with nginx
FROM nginxinc/nginx-unprivileged:alpine

# Set working directory
WORKDIR /usr/share/nginx/html

# Switch to root for setup operations
USER root

# Remove default nginx static content
RUN rm -rf /usr/share/nginx/html/*

# Create non-root user for Choreo (UID 10014 is required by Choreo)
RUN addgroup -g 10014 choreo && \
    adduser -D -u 10014 -G choreo choreouser

# Copy built Angular app from build stage
COPY --from=build /app/dist/my-angular-app/browser /usr/share/nginx/html

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Set permissions for the choreo user
RUN chown -R 10014:10014 /usr/share/nginx/html && \
    chown -R 10014:10014 /var/cache/nginx && \
    chown -R 10014:10014 /var/log/nginx && \
    chown -R 10014:10014 /etc/nginx/conf.d && \
    touch /var/run/nginx.pid && \
    chown -R 10014:10014 /var/run/nginx.pid

# Switch to non-root user
USER 10014

# Expose port 8080 (Choreo standard)
EXPOSE 8080

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
