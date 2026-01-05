# Local Kubernetes Deployment Guide

Since you don't have an active Azure subscription, you can deploy the entire application locally using **Docker Desktop** (with Kubernetes enabled) or **Minikube**.

## Prerequisites

1.  **Docker Desktop**: Installed and running.
2.  **Kubernetes**: Enabled in Docker Desktop settings (Settings -> Kubernetes -> Enable Kubernetes).
3.  **kubectl**: Installed (usually comes with Docker Desktop).

## 1. Build Images Locally

Since we are running locally, we don't need to push to a registry (ACR). We just need the images to exist in your local Docker cache.

**Important:** If using **Minikube**, run this first: `eval $(minikube docker-env)` (Linux/Mac) or `minikube -p minikube docker-env | Invoke-Expression` (PowerShell).

Run these commands in the root of your project:

```bash
# 1. Build Backend Services
docker build -t food-delivery/auth-service:latest --build-arg SERVICE_NAME=auth-service -f Dockerfile.backend .
docker build -t food-delivery/order-service:latest --build-arg SERVICE_NAME=order-service -f Dockerfile.backend .
docker build -t food-delivery/restaurant-service:latest --build-arg SERVICE_NAME=restaurant-service -f Dockerfile.backend .
docker build -t food-delivery/delivery-service:latest --build-arg SERVICE_NAME=delivery-service -f Dockerfile.backend .
docker build -t food-delivery/payment-service:latest --build-arg SERVICE_NAME=payment-service -f Dockerfile.backend .
docker build -t food-delivery/tracking-service:latest --build-arg SERVICE_NAME=tracking-service -f Dockerfile.backend .

# 2. Build Frontend
docker build -t food-delivery/frontend:latest -f frontend/food-delivery-web/Dockerfile frontend/food-delivery-web

### Important for Kind Users
If you are using **Kind**, you must load the images into the cluster nodes:
```bash
kind load docker-image food-delivery/auth-service:latest --name food-delivery
kind load docker-image food-delivery/order-service:latest --name food-delivery
kind load docker-image food-delivery/restaurant-service:latest --name food-delivery
kind load docker-image food-delivery/delivery-service:latest --name food-delivery
kind load docker-image food-delivery/payment-service:latest --name food-delivery
kind load docker-image food-delivery/tracking-service:latest --name food-delivery
kind load docker-image food-delivery/frontend:latest --name food-delivery
```
```

## 2. Prepare Manifests

Ensure your Kubernetes manifests use the local image names and `imagePullPolicy: IfNotPresent` (which they already do by default in the files I created).

**Verify:** Check `k8s/services/auth-service.yaml` (and others) to ensure `image` is `food-delivery/auth-service:latest`. If you modified them for Azure earlier, revert them back:

**Windows (PowerShell) - Revert to local images:**
```powershell
Get-ChildItem k8s/services/*.yaml | ForEach-Object {
    (Get-Content $_.FullName) -replace 'image: .*/food-delivery/', 'image: food-delivery/' | Set-Content $_.FullName
}
```

## 3. Deploy to Local Kubernetes

```bash
# 1. Apply Infrastructure (DBs, Kafka, etc.)
kubectl apply -f k8s/infrastructure/

# Wait for them to be ready
kubectl wait --for=condition=ready pod -l app=postgres --timeout=120s
kubectl wait --for=condition=ready pod -l app=redis --timeout=120s

# 2. Apply Services
kubectl apply -f k8s/services/

# 3. Apply Ingress (Optional locally, we will use port-forwarding instead)
kubectl apply -f k8s/ingress.yaml
```

## 4. Access the Application

The easiest way to access services locally without configuring complex Ingress Controllers is using `port-forward`.

### Access Frontend
Open a new terminal:
```bash
kubectl port-forward svc/frontend 8080:80
```
Now open your browser to: **http://localhost:8080**

### Access Kong API Gateway (Backend)
Open another terminal:
```bash
kubectl port-forward svc/kong 8000:8000
```
Now your frontend can talk to `http://localhost:8000` (ensure your frontend config points here).

### Access Jaeger (Tracing)
```bash
kubectl port-forward svc/jaeger 16686:16686
```
Open browser to: **http://localhost:16686**

## 5. Cleanup

To stop everything:
```bash
kubectl delete -f k8s/services/
kubectl delete -f k8s/infrastructure/
```
