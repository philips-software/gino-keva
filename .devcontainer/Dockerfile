ARG HADOLINT_VERSION=2.4.1
FROM hadolint/hadolint:${HADOLINT_VERSION}-alpine as hadolint

FROM mcr.microsoft.com/vscode/devcontainers/go:1.18

RUN go install github.com/ahmetb/govvv@v0.3.0 && go install github.com/go-delve/delve/cmd/dlv@master

COPY --from=hadolint /bin/hadolint /bin/
