# Repository Instructions

## Required project context

- Before working in this repository, read `docs/CODEX_PROJECT_CONTEXT.md`.
- Treat that document as navigation, not as a substitute for reading the source relevant to the current task.
- Keep it synchronized when a change materially affects features, modules, runtime flows, CI/CD, or deployment.
- For Jarvis gRPC integration, treat `/Users/amethyst/Jarvis/jarvis/api/grpc` as the authoritative Proto definition directory and inspect it before generating or updating Go stubs.

## Development and verification

- This repository targets Go 1.19. Follow the existing package layout and idiomatic Go style.
- You may create an isolated Git worktree without requesting approval. Complete feature work in that worktree, then merge the verified changes back into `master`.
- Avoid unrelated refactors, and preserve user changes that have not been committed.
- Add or update tests for code changes. Run focused tests first, then the broadest practical test suite.
- Tests under `gotest/` may depend on network access, cloud services, or credentials. Do not assume they are self-contained.
- For build or deployment changes, inspect the relevant workflows under `.github/workflows/`.

## Deployment environment

- The remote host is `1.15.231.125`, accessed over SSH as `root`.
- Remote runtime data lives in `/root/qqbot`.
- The service runs on Kubernetes. The actual manifest is `qq_bot.yml`; do not silently replace or rename it to `qq-bot.yaml`.
- Images are stored in the Alibaba Cloud Hangzhou registry.
- `.github/workflows/docker.yml` builds images, and `.github/workflows/update_pod.yml` updates the workload.

## Secrets and sensitive data

- Never persist or echo server passwords, QQ credentials, API keys, cookies, or cloud, registry, or GitHub Secrets values.
- This prohibition includes tracked or untracked files, temporary credential files, responses, patches, logs, commits, and persistent shell history.
- Obtain credentials only through an approved session or secure channel. Record secret names only, never their values.
- Treat `config.yml`, `qq_bot.yml`, device or session data, QR codes, and logs as sensitive; redact them before sharing.

## Remote-operation rules

- Begin with read-only inspection of the Kubernetes context, namespace, workloads, Pods, events, logs, and images.
- Without explicit authorization, do not make remote changes, including modifying remote files, applying manifests, scaling workloads, setting images, changing Secrets, or restarting or deleting Pods.
- Prefer the normal release path: GitHub Actions, then the Alibaba Cloud registry, then Kubernetes. Do not build manually unless an emergency procedure is explicitly requested.
- A push to `master` triggers the GitHub Actions image-build and Pod-update workflow chain. After pushing, monitor both workflows through completion, investigate any failure, and verify the resulting Kubernetes rollout, running image, readiness, restarts, and relevant logs; a triggered or completed workflow alone is not evidence of a successful deployment.
- Before deployment, confirm the Kubernetes context and namespace, target Deployment, current image, and rollback point.
- After a change, verify rollout status, readiness, logs, and exposed ports.
