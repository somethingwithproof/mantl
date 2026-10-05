# Multi-stage build for static website
# Official nginx:1.31-alpine image, pinned to its multi-platform digest.
FROM nginx@sha256:df221db836e1754089190208cee7eeda94f233197056426eda74a43ab1abeac2

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf && \
    rm /etc/nginx/nginx.conf

# Copy custom nginx config
COPY nginx.conf /etc/nginx/nginx.conf

# Copy static files
COPY src/ /usr/share/nginx/html/

# Create required directories with proper permissions
RUN mkdir -p /var/cache/nginx /var/log/nginx /tmp && \
    chown -R nginx:nginx /var/cache/nginx /var/log/nginx /tmp /usr/share/nginx/html && \
    chmod -R 755 /usr/share/nginx/html

# Run as non-root user
USER nginx

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]
