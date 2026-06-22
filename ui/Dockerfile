# Build stage
FROM ghcr.io/cirruslabs/flutter:stable AS builder

WORKDIR /app
# Copy only UI source (context is repo root — see docker-compose.yaml)
COPY ui/pubspec.yaml ui/pubspec.lock ./
RUN flutter pub get

COPY ui/ .
# Empty API_BASE_URL → requests go to same origin (proxied by nginx to backend)
RUN flutter build web --release --dart-define=API_BASE_URL=

# Runtime stage — nginx serves static Flutter web + proxies API calls
FROM nginx:1.27-alpine
COPY --from=builder /app/build/web /usr/share/nginx/html
COPY ui/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
