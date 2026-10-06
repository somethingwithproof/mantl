<!-- SPDX-License-Identifier: Apache-2.0 -->
# Static Website Example

Production-ready static website hosting on Kubernetes demonstrating:

- Nginx-based static file serving
- CDN integration (CloudFront example)
- Aggressive caching strategies
- Compression (gzip/brotli)
- Security headers (CSP, X-Frame-Options, etc.)
- Auto-scaling for traffic spikes
- Zero-downtime deployments
- Multi-region CDN distribution

## Architecture

```
┌─────────────┐
│   Users     │
└──────┬──────┘
       │
       v
┌─────────────────────┐
│   CloudFront CDN    │
│  - Edge Caching     │
│  - HTTPS/TLS        │
│  - GZIP/Brotli      │
└──────┬──────────────┘
       │ Cache Miss
       v
┌─────────────────────┐
│  Kubernetes Ingress │
│  - Rate Limiting    │
│  - WAF Rules        │
└──────┬──────────────┘
       │
       v
┌─────────────────────┐
│  Nginx Pods (3+)    │
│  - Static Files     │
│  - Cache Headers    │
│  - Security Headers │
└─────────────────────┘
```

## Quick Start

### Local Development

```bash
# Serve locally with Python
cd src
python -m http.server 8000

# Or use nginx locally
nginx -c nginx.conf -p .
```

### Build Container

```bash
docker build -t mantl-website:1.0.0 .

# Run locally
docker run -p 8080:8080 mantl-website:1.0.0

# Test
curl http://localhost:8080/
```

### Deploy to Kubernetes

```bash
# Update image registry
sed -i 's|REGISTRY|your-registry.example.com|g' k8s/deployment.yaml

# Deploy
kubectl apply -f k8s/deployment.yaml

# Verify
kubectl get pods -n static-sites
kubectl get ingress -n static-sites

# Access via ingress
curl https://mantl.example.com/
```

## CDN Integration

### AWS CloudFront Setup

```bash
# 1. Create ACM certificate (must be in us-east-1 for CloudFront)
aws acm request-certificate \
  --domain-name mantl.example.com \
  --subject-alternative-names www.mantl.example.com \
  --validation-method DNS \
  --region us-east-1

# 2. Validate certificate via DNS
# (Add CNAME records provided by ACM)

# 3. Create CloudFront distribution
aws cloudfront create-distribution \
  --distribution-config file://k8s/cdn-cloudfront.yaml

# 4. Update Route53 to point to CloudFront
aws route53 change-resource-record-sets \
  --hosted-zone-id ZONE_ID \
  --change-batch file://route53-changes.json

# 5. Wait for distribution to deploy (15-20 minutes)
aws cloudfront get-distribution --id DISTRIBUTION_ID

# 6. Test CDN
curl -I https://mantl.example.com/
# Look for: X-Cache: Hit from cloudfront
```

### Alternative: Cloudflare

```bash
# 1. Add site to Cloudflare
# 2. Update nameservers to Cloudflare
# 3. Configure Page Rules:
#   - Cache Level: Standard
#   - Browser Cache TTL: 4 hours
#   - Edge Cache TTL: 1 year (for static assets)

# 4. Enable features:
#   - Auto Minify (JS, CSS, HTML)
#   - Brotli compression
#   - HTTP/3 (QUIC)
#   - Always Use HTTPS
#   - Automatic HTTPS Rewrites

# 5. Test
curl -I https://mantl.example.com/
# Look for: CF-Cache-Status: HIT
```

## Caching Strategy

### Asset Types and Cache Duration

| Asset Type | Cache Duration | CloudFront | Browser |
|------------|---------------|------------|---------|
| HTML | 5 minutes | 5m | 5m |
| CSS/JS (versioned) | 1 year | 1y | 1y |
| Images | 1 year | 1y | 1y |
| Fonts | 1 year | 1y | 1y |
| API responses | No cache | 0 | 0 |

### Cache Invalidation

```bash
# Invalidate CloudFront cache after deployment
aws cloudfront create-invalidation \
  --distribution-id DISTRIBUTION_ID \
  --paths "/*"

# Invalidate specific paths
aws cloudfront create-invalidation \
  --distribution-id DISTRIBUTION_ID \
  --paths "/index.html" "/styles.css"

# Check invalidation status
aws cloudfront get-invalidation \
  --distribution-id DISTRIBUTION_ID \
  --id INVALIDATION_ID
```

### Versioned Assets

For CSS/JS files, use content hashing:

```bash
# Build with webpack/vite (automatic hash in filename)
# styles.abc123.css instead of styles.css

# This allows aggressive caching without invalidation
```

## Security

### Content Security Policy

The nginx config includes CSP headers:

```
Content-Security-Policy:
  default-src 'self';
  script-src 'self' 'unsafe-inline';
  style-src 'self' 'unsafe-inline';
  img-src 'self' data: https:;
```

Customize for your needs:

```nginx
add_header Content-Security-Policy "
  default-src 'self';
  script-src 'self' https://trusted-cdn.com;
  style-src 'self' https://fonts.googleapis.com;
  font-src 'self' https://fonts.gstatic.com;
  img-src 'self' data: https:;
  connect-src 'self' https://api.example.com;
" always;
```

### Security Headers Validation

```bash
# Test security headers
curl -I https://mantl.example.com/

# Should include:
# X-Frame-Options: SAMEORIGIN
# X-Content-Type-Options: nosniff
# X-XSS-Protection: 1; mode=block
# Content-Security-Policy: ...
# Referrer-Policy: strict-origin-when-cross-origin

# Test with securityheaders.com
curl "https://securityheaders.com/?q=mantl.example.com"
```

## Performance Optimization

### Enable HTTP/3

CloudFront supports HTTP/3 automatically when enabled.

For nginx (if not using CDN):

```nginx
listen 443 quic reuseport;
listen 443 ssl;
http2 on;
http3 on;

add_header Alt-Svc 'h3=":443"; ma=86400';
```

### Preload Critical Assets

Add to HTML `<head>`:

```html
<link rel="preload" href="/styles.css" as="style">
<link rel="preload" href="/app.js" as="script">
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="dns-prefetch" href="https://api.example.com">
```

### Service Worker for Offline Support

```javascript
// sw.js
self.addEventListener('install', event => {
    event.waitUntil(
        caches.open('mantl-v1').then(cache => {
            return cache.addAll([
                '/',
                '/styles.css',
                '/app.js',
                '/favicon.svg'
            ]);
        })
    );
});

self.addEventListener('fetch', event => {
    event.respondWith(
        caches.match(event.request).then(response => {
            return response || fetch(event.request);
        })
    );
});
```

Register in HTML:

```html
<script>
if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/sw.js');
}
</script>
```

## Deployment

### Blue-Green Deployment

```bash
# Deploy new version as "green"
kubectl apply -f k8s/deployment-green.yaml

# Test green version
kubectl port-forward -n static-sites deploy/mantl-website-green 8080:8080

# Switch ingress to green
kubectl patch ingress mantl-website -n static-sites --type=merge -p '
spec:
  rules:
  - host: mantl.example.com
    http:
      paths:
      - backend:
          service:
            name: mantl-website-green
            port:
              number: 80
'

# Invalidate CDN cache
aws cloudfront create-invalidation \
  --distribution-id DISTRIBUTION_ID \
  --paths "/*"

# If successful, delete blue
kubectl delete deployment mantl-website-blue -n static-sites
```

### CI/CD Pipeline

```yaml
# Example GitHub Actions workflow
name: Deploy Static Website

on:
  push:
    branches: [main]
    paths:
      - 'examples/static-website/**'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Build Docker image
      run: |
        cd examples/static-website
        docker build -t ${{ secrets.REGISTRY }}/mantl-website:${{ github.sha }} .

    - name: Push to registry
      run: |
        echo "${{ secrets.REGISTRY_PASSWORD }}" | docker login ${{ secrets.REGISTRY }} -u ${{ secrets.REGISTRY_USERNAME }} --password-stdin
        docker push ${{ secrets.REGISTRY }}/mantl-website:${{ github.sha }}

    - name: Update Kubernetes deployment
      run: |
        kubectl set image deployment/mantl-website \
          nginx=${{ secrets.REGISTRY }}/mantl-website:${{ github.sha }} \
          -n static-sites

    - name: Invalidate CDN cache
      run: |
        aws cloudfront create-invalidation \
          --distribution-id ${{ secrets.CLOUDFRONT_DISTRIBUTION_ID }} \
          --paths "/*"
```

## Monitoring

### Nginx Metrics

```bash
# Enable nginx-prometheus-exporter sidecar
kubectl apply -f k8s/nginx-exporter.yaml

# Metrics available:
# - nginx_http_requests_total
# - nginx_http_request_duration_seconds
# - nginx_connections_active
```

### CDN Metrics (CloudFront)

```bash
# View CloudFront metrics in CloudWatch
aws cloudwatch get-metric-statistics \
  --namespace AWS/CloudFront \
  --metric-name Requests \
  --dimensions Name=DistributionId,Value=DISTRIBUTION_ID \
  --start-time 2024-03-15T00:00:00Z \
  --end-time 2024-03-15T23:59:59Z \
  --period 3600 \
  --statistics Sum

# Key metrics:
# - Requests (total requests)
# - BytesDownloaded (data transfer)
# - 4xxErrorRate (client errors)
# - 5xxErrorRate (server errors)
# - CacheHitRate (% served from cache)
```

## Troubleshooting

### CDN Not Caching

```bash
# Check cache headers from origin
curl -I https://mantl.example.com/ -H "Host: origin.example.com"

# Check CloudFront cache status
curl -I https://mantl.example.com/
# Look for: X-Cache: Hit from cloudfront

# Common issues:
# 1. Cache-Control: no-cache set by origin
# 2. Vary: * header preventing caching
# 3. Set-Cookie headers preventing caching
```

### Stale Content After Deployment

```bash
# Create CDN invalidation
aws cloudfront create-invalidation \
  --distribution-id DISTRIBUTION_ID \
  --paths "/*"

# For faster updates, use versioned asset names:
# styles.abc123.css instead of styles.css
```

## Cost Optimization

### CloudFront Cost Reduction

```bash
# 1. Use cheaper price class (exclude expensive regions)
# PriceClass_100: US, Canada, Europe
# PriceClass_200: + Asia, Middle East, Africa
# PriceClass_All: + South America, Australia

# 2. Enable compression (reduces data transfer)
# 3. Optimize images (WebP format, proper sizing)
# 4. Set appropriate cache TTLs (reduce origin requests)
# 5. Use Origin Shield (reduce origin load)
```

## Additional Resources

- [CloudFront Best Practices](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/ConfiguringCaching.html)
- [Nginx Performance Tuning](https://www.nginx.com/blog/tuning-nginx/)
- [Web Performance Optimization](https://web.dev/fast/)
