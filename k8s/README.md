# Kubernetes Deployment Guide

## Prerequisites

1. Kubernetes cluster (v1.24+)
2. kubectl configured
3. NGINX Ingress Controller installed
4. cert-manager installed (for TLS)

## Quick Start

### 1. Create Secrets

```bash
# Copy template and fill in actual values
cp k8s/base/secret.yaml.template k8s/base/secret.yaml
cp k8s/base/postgres-secret.yaml.template k8s/base/postgres-secret.yaml

# Edit the files with your actual credentials
vim k8s/base/secret.yaml
vim k8s/base/postgres-secret.yaml
```

### 2. Deploy to Development

```bash
kubectl apply -k k8s/overlays/dev
```

### 3. Deploy to Staging

```bash
kubectl apply -k k8s/overlays/staging
```

### 4. Deploy to Production

```bash
kubectl apply -k k8s/overlays/production
```

## Manual Deployment Steps

### Create Namespace

```bash
kubectl apply -f k8s/base/namespace.yaml
```

### Create Secrets

```bash
kubectl create secret generic studio-secrets \
  --from-literal=STUDIO_DATABASE_DSN="postgres://user:pass@postgres:5432/studio_db" \
  --from-literal=STUDIO_JWT_SECRET="your-secret-key" \
  -n studio

kubectl create secret generic postgres-secret \
  --from-literal=password="your-postgres-password" \
  -n studio
```

### Deploy Database

```bash
kubectl apply -f k8s/base/postgres-statefulset.yaml
kubectl apply -f k8s/base/postgres-service.yaml
kubectl apply -f k8s/base/redis-deployment.yaml
kubectl apply -f k8s/base/redis-service.yaml
```

### Deploy Application

```bash
kubectl apply -f k8s/base/configmap.yaml
kubectl apply -f k8s/base/backend-deployment.yaml
kubectl apply -f k8s/base/backend-service.yaml
kubectl apply -f k8s/base/frontend-deployment.yaml
kubectl apply -f k8s/base/frontend-service.yaml
kubectl apply -f k8s/base/ingress.yaml
kubectl apply -f k8s/base/hpa.yaml
```

## Verify Deployment

```bash
# Check all pods
kubectl get pods -n studio

# Check services
kubectl get svc -n studio

# Check ingress
kubectl get ingress -n studio

# View logs
kubectl logs -f deployment/studio-backend -n studio
kubectl logs -f deployment/studio-frontend -n studio

# Check health
kubectl exec -n studio deployment/studio-backend -- wget -q -O- http://localhost:8080/health
```

## Scaling

```bash
# Manual scaling
kubectl scale deployment studio-backend --replicas=5 -n studio

# HPA will automatically scale based on CPU/Memory
kubectl get hpa -n studio
```

## Rollback

```bash
# View rollout history
kubectl rollout history deployment/studio-backend -n studio

# Rollback to previous version
kubectl rollout undo deployment/studio-backend -n studio

# Rollback to specific revision
kubectl rollout undo deployment/studio-backend --to-revision=2 -n studio
```

## Database Migrations

Migrations run automatically on pod startup. To run manually:

```bash
# Get backend pod name
POD=$(kubectl get pod -n studio -l app=studio-backend -o jsonpath='{.items[0].metadata.name}')

# Check migration status
kubectl exec -n studio $POD -- /server --migrate-status

# Run migrations manually (if needed)
kubectl exec -n studio $POD -- /server --migrate-up
```

## Troubleshooting

### Pod not starting

```bash
kubectl describe pod <pod-name> -n studio
kubectl logs <pod-name> -n studio
```

### Database connection issues

```bash
# Test database connectivity
kubectl exec -n studio deployment/studio-backend -- nc -zv postgres 5432

# Check database logs
kubectl logs -n studio statefulset/postgres
```

### Ingress not working

```bash
# Check ingress controller
kubectl get pods -n ingress-nginx

# Check ingress events
kubectl describe ingress studio-ingress -n studio
```

## CI/CD Integration

The deployment workflow (`.github/workflows/deploy.yml`) automatically deploys:

- **dev**: On push to feature branches
- **staging**: On push to main branch
- **production**: On tag push (v*)

### Manual deployment via GitHub Actions

1. Go to Actions tab
2. Select "Deploy to Kubernetes"
3. Click "Run workflow"
4. Choose environment (dev/staging/production)

## Monitoring

```bash
# Watch pod status
kubectl get pods -n studio -w

# View resource usage
kubectl top pods -n studio
kubectl top nodes

# Check HPA status
kubectl get hpa -n studio -w
```

## Cleanup

```bash
# Delete specific environment
kubectl delete -k k8s/overlays/dev

# Delete everything
kubectl delete namespace studio
```
