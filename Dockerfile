# ---- build stage ----
FROM golang:1.22-alpine AS build
WORKDIR /src

# No third-party dependencies, so there is nothing to download; copying the
# module file first still lets Docker cache this layer.
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/groupie-tracker .

# ---- run stage ----
# distroless/static gives a tiny image with no shell and a non-root user.
FROM gcr.io/distroless/static-debian12
COPY --from=build /bin/groupie-tracker /groupie-tracker

# The server reads PORT (defaulting to 8080); hosts inject their own value.
ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/groupie-tracker"]
