# .devcontainer/Dockerfile
FROM golang:1.24.2-bookworm

# Add user arguments, including one for the Docker GID
ARG USERNAME=vscode
ARG USER_UID=1001
ARG USER_GID=1001
# Argument to accept the host's Docker GID
ARG DOCKER_GID

# Install system dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    gnupg \
    git \
    postgresql-client \
    && rm -rf /var/lib/apt/lists/*

# Add Docker's official GPG key and set up the repository
RUN install -m 0755 -d /etc/apt/keyrings
RUN curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc
RUN chmod a+r /etc/apt/keyrings/docker.asc
RUN echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  tee /etc/apt/sources.list.d/docker.list > /dev/null
RUN apt-get update

# Install Docker CLI
RUN apt-get install -y docker-ce-cli

# Install Tailwind CSS CLI
RUN arch=$(uname -m) && \
    if [ "$arch" = "x86_64" ]; then arch="x64"; fi && \
    if [ "$arch" = "aarch64" ]; then arch="arm64"; fi && \
    curl -sLO "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-${arch}" && \
    chmod +x "tailwindcss-linux-${arch}" && \
    mv "tailwindcss-linux-${arch}" /usr/local/bin/tailwindcss

RUN if getent group docker > /dev/null; then \
        groupmod -g $DOCKER_GID docker; \
    else \
        groupadd -g $DOCKER_GID docker; \
    fi

# Create the user and add them to the docker group
RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --shell /bin/bash --uid $USER_UID --gid $USER_GID -m $USERNAME \
    && apt-get install -y sudo \
    && usermod -aG docker $USERNAME \
    && echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME

# Switch to the non-root user
USER $USERNAME

# Set up workspace and Go environment
WORKDIR /workspace
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH

# Install Go tools
RUN go install golang.org/x/tools/gopls@latest
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go install honnef.co/go/tools/cmd/staticcheck@latest