# Build stage
FROM node:20-alpine as builder

WORKDIR /app

# Copy package files only
COPY package.json package-lock.json* ./

# Clean install with additional rollup dependency for Alpine
RUN npm cache clean --force && \
    npm install && \
    npm install @rollup/rollup-linux-x64-musl --save-dev

# Copy source code (excluding node_modules via .dockerignore)
COPY . .

# Build the application
RUN npm run build

# Production stage
FROM nginx:alpine

# Copy custom nginx config
COPY nginx.conf /etc/nginx/nginx.conf

# Copy built app from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Expose port 80
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"] 