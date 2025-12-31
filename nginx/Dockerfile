# Build frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /app

# Build argument for API URL
ARG NEXT_PUBLIC_API_URL=http://localhost:8080
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL

# Copy package files
COPY frontend/package*.json ./
RUN npm ci

# Copy frontend source
COPY frontend/ ./

# Build Next.js app
RUN npm run build

# Production stage with Nginx
FROM nginx:alpine

# Copy built static files
COPY --from=frontend-builder /app/out /usr/share/nginx/html

# Copy nginx config
COPY nginx/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]

