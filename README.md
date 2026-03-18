# Kubernetes Mutating Webhook: Sidecar Injector

This project implements a Kubernetes Mutating Webhook in Go that automatically injects a security sidecar (Falco) into all new Pods.

## Features
- **Admission Controller**: Listens for Pod creation requests.
- **Sidecar Injection**: Automatically adds a `security-agent` container to the Pod's `spec.containers`.
- **Filtering**: Can be configured to skip injection based on labels (e.g., `sidecar-injection: disabled`).

## Project Structure
- `main.go`: The webhook server implementation.
- `manifests/`: Kubernetes deployment manifests.
- `scripts/`: TLS certificate generation scripts.

## How to Deploy

### 1. Build the Binary
```bash
go build -o sidecar-injector main.go
```

### 2. Generate TLS Certificates
Kubernetes requires admission webhooks to run over HTTPS. Use the provided script to generate self-signed certificates and create a Kubernetes Secret:
```bash
chmod +x scripts/gen-certs.sh
./scripts/gen-certs.sh
```
*Note: This script will output a `CA_BUNDLE` string. Copy it and replace `${CA_BUNDLE}` in `manifests/webhook-config.yaml`.*

### 3. Deploy the Service
```bash
# Create the namespace
kubectl create namespace sidecar-injector

# Apply the manifests
kubectl apply -f manifests/deployment.yaml
kubectl apply -f manifests/service.yaml
kubectl apply -f manifests/webhook-config.yaml
```

### 4. Test the Injection
Create a test Pod and check its containers:
```bash
kubectl run test-pod --image=nginx
kubectl get pod test-pod -o jsonpath='{.spec.containers[*].name}'
# Output should include 'security-agent'
```

## Security Agent
By default, this injector adds `falcosecurity/falco-no-driver:latest` as a sidecar. You can customize the image and arguments in `main.go`.

---
Created for Nik577.
