# Talos Demo App - GitOps CI/CD Pipeline

Simple Go web application demonstrating full GitOps workflow with Harbor + ArgoCD + BGP LoadBalancer.

## Architecture

```
Developer → GitHub → (GitHub Actions on FRR) → Harbor Registry → ArgoCD → Kubernetes → BGP VIP
```

## Repository Structure

```
/app/                       # Application source code
  ├── main.go              # Go web server
  └── Dockerfile           # Multi-stage build
  
/deploy/                   # Kubernetes manifests (watched by ArgoCD)
  └── deployment.yaml      # Deployment + Service + Namespace

/.github/workflows/        # CI/CD automation
  └── ci-cd.yaml          # Build → Push → Update manifest
```

## Workflow

1. **Developer pushes code** to main branch (app/* changes)
2. **GitHub Actions** (self-hosted runner on FRR):
   - Builds Docker image
   - Tags with commit SHA (e.g., `abc1234`)
   - Pushes to Harbor: `harbor.int-talos-poc.pocketcove.net/library/demo-app:abc1234`
   - Updates `deploy/deployment.yaml` with new image tag
   - Commits change back to repo
3. **ArgoCD** detects manifest change
   - Auto-syncs from Git
   - Deploys to Kubernetes
4. **Cilium** assigns LoadBalancer VIP (10.0.1.199)
5. **BGP** advertises /32 route to FRR ToR
6. **Service accessible** at http://10.0.1.199

## Prerequisites

### GitHub Repository Setup

1. Create repo on GitHub: `talos-demo-app`
2. Push this code
3. Add secrets:
   - `HARBOR_USERNAME`: admin (temporary, will replace with robot account)
   - `HARBOR_PASSWORD`: Harbor12345

### Self-Hosted Runner Setup (FRR)

Runner is on FRR (10.0.1.50) - **TEMPORARY** per expert guidance.

**Access FRR:**
```bash
ssh -i ~/.ssh/pocketcove-bastion-20251217.pem ec2-user@18.235.42.220
ssh -i ~/.ssh/pocketcove-bastion-20251217.pem ubuntu@10.0.1.50
```

**Install runner:**
```bash
# On FRR
mkdir -p ~/actions-runner && cd ~/actions-runner

# Download runner (check GitHub for latest version)
curl -o actions-runner-linux-x64-2.321.0.tar.gz -L \
  https://github.com/actions/runner/releases/download/v2.321.0/actions-runner-linux-x64-2.321.0.tar.gz

# Extract
tar xzf actions-runner-linux-x64-2.321.0.tar.gz

# Configure (you'll need token from GitHub repo settings)
./config.sh --url https://github.com/YOUR_USERNAME/talos-demo-app --token YOUR_TOKEN

# Install as service
sudo ./svc.sh install
sudo ./svc.sh start
```

### ArgoCD Application

Created via kubectl from bastion (see deployment instructions).

## Deployment Instructions

### 1. Create GitHub Repository

```bash
# From local Mac
cd ~/Workspace/talos-demo-app
gh repo create talos-demo-app --public --source=. --remote=origin --push
# Or create manually at github.com/new
```

### 2. Push Code

```bash
git add .
git commit -m "Initial commit: Demo app with GitOps workflow"
git branch -M main
git remote add origin https://github.com/YOUR_USERNAME/talos-demo-app.git
git push -u origin main
```

### 3. Setup Self-Hosted Runner

Follow "Self-Hosted Runner Setup" section above.

### 4. Add GitHub Secrets

In GitHub repo: Settings → Secrets and variables → Actions → New repository secret

- `HARBOR_USERNAME`: `admin`
- `HARBOR_PASSWORD`: `Harbor12345`

### 5. Deploy ArgoCD Application

```bash
# From bastion
cat > demo-app-argocd.yaml << 'EOF'
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: demo-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/YOUR_USERNAME/talos-demo-app.git
    targetRevision: HEAD
    path: deploy
  destination:
    server: https://kubernetes.default.svc
    namespace: demo-app
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
EOF

kubectl apply -f demo-app-argocd.yaml
```

### 6. Verify Deployment

```bash
# Check ArgoCD app
kubectl get application -n argocd demo-app

# Check pods
kubectl get pods -n demo-app

# Check service and VIP
kubectl get svc -n demo-app

# Test from FRR
curl http://10.0.1.199
```

## Testing the Full Loop

1. Make a change to `app/main.go`
2. Commit and push to main
3. Watch GitHub Actions build
4. See Harbor receive new image
5. See ArgoCD detect and sync
6. See pods rolling update
7. Test service at http://10.0.1.199

## Load Testing

From FRR or external client:
```bash
# Install hey
wget https://hey-release.s3.us-east-2.amazonaws.com/hey_linux_amd64
chmod +x hey_linux_amd64

# Run load test
./hey_linux_amd64 -z 60s -c 10 http://10.0.1.199
```

## Notes

- **Runner Location:** FRR (10.0.1.50) - TEMPORARY
  - Expert guidance: Migrate to dedicated client VM or WARP-enrolled Mac later
  - Document: Do not use ToR/router for permanent CI
  
- **LoadBalancer VIP:** 10.0.1.199 (from pool 10.0.1.192-207)
  - Requires `bgp: blue` label (enforced by Kyverno)
  - Advertised via Cilium BGP to FRR (ASN 65000)

- **Image Tagging:** Commit SHA (Option A - proper GitOps)
  - Not `:latest` (prevents ArgoCD confusion)
  - Manifest update commits trigger ArgoCD sync

- **Harbor Authentication:** Using admin temporarily
  - TODO: Create robot account (cicd-2 task)
  - Separate push/pull credentials
