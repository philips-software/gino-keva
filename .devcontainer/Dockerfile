ARG HADOLINT_VERSION=2.4.1
FROM hadolint/hadolint:${HADOLINT_VERSION}-alpine as hadolint

FROM mcr.microsoft.com/vscode/devcontainers/go:1.17

# Install hadolint
COPY --from=hadolint /bin/hadolint /bin/
