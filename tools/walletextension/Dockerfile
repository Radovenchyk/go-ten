# Build stages:
# 1. system: prepares the OS with required binaries
# 2. get-dependencies: downloads the Go modules
# 3. build-wallet: copies source code and builds binaries using compiler cache
# 4. final: creates a lightweight image with only the executables

FROM golang:1.22.1-alpine3.19 as system

# Install build tools and dependencies
RUN apk add build-base bash git
ENV CGO_ENABLED=1
ARG TESTNET_TYPE

# Stage to download Go dependencies
FROM system as get-dependencies
RUN mkdir -p /home/ten/go-ten

# Set up workspace and download dependencies
WORKDIR /home/ten/go-ten
COPY go.mod .
COPY go.sum .
RUN go mod download

# Stage to build the wallet extension executable
FROM get-dependencies as build-wallet
COPY . /home/ten/go-ten

# Build the wallet extension binary
WORKDIR /home/ten/go-ten/tools/walletextension/main
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -o ../bin/wallet_extension_linux

# Final lightweight image with just the necessary files to run the wallet extension
FROM alpine:3.18

# Copy executable and required files to the final image
COPY --from=build-wallet /home/ten/go-ten/tools/walletextension/bin /home/ten/go-ten/tools/walletextension/bin
COPY --from=build-wallet /home/ten/go-ten/tools/walletextension/storage/database /home/ten/go-ten/tools/walletextension/storage/database

# Set entrypoint to directly run the wallet extension binary
ENTRYPOINT ["/home/ten/go-ten/tools/walletextension/bin/wallet_extension_linux"]
