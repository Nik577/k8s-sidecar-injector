K8s Sidecar Injector: Beyond Manual YAML
Built by an engineer who's tired of chasing developers to add logging agents.

Why I built this?
In a large-scale cluster, you can't trust every team to remember to include a security agent or a log collector in their Deployment. I wanted a way to enforce Platform Standards without touching a single line of the application's code. This injector acts as a "silent guardian" at the API level.

 How it's different from a "Hello World" webhook:
No Recompilation: Most tutorials hardcode the sidecar. Here, I’ve implemented a watcher that reads templates from a ConfigMap. Want to swap Fluentd for Vector? Just update the YAML and the webhook picks it up.

Real-world TLS: I didn't stop at gen-certs.sh. The project is pre-configured to work with cert-manager, which is how 99% of Big Tech companies handle webhook certificates.

Production Safety: It won't crash your API server. I’ve included Graceful Shutdown logic and a specialized Reconciler to handle high-concurrency admission requests.

 Architecture in a Nutshell
Intercept: The K8s API Server sends an AdmissionReview to this service.

Match: The service checks if the Pod needs a sidecar (based on namespace or annotations).

Patch: It generates a JSON Patch (RFC 6902) to inject the sidecar into the Pod's spec.

Deploy: The Pod starts with your main container + the auto-injected agent.

 Quick Start (The "I'm in a hurry" way)
Bash
# 1. Get the code
git clone https://github.com/Nik577/k8s-sidecar-injector.git

# 2. Deploy via Helm (The Industry Standard)
# This handles RBAC, Service, and Cert-Manager integration automatically
helm upgrade --install sidecar-injector ./deploy/helm/k8s-sidecar-injector \
  --namespace sidecar-injector --create-namespace
 Monitoring & Ops
I've baked in observability from day one:

Metrics: Check :8080/metrics for injection success/failure rates.

Logs: Structured slog (JSON) for easy searching in Grafana Loki or ELK.

Health: /healthz and /readyz endpoints for K8s Probes.

Developed by Nikita Mamonov
Feel free to open an Issue or a PR if you want to add more injection logic!
