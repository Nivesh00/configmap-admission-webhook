# Build application
FROM golang:1.24.5 AS build

WORKDIR /app

# Copy go files
COPY src/go.mod src/go.sum ./
RUN go mod download

# Copy modules
COPY src/module/helper.go ./module/
COPY src/module/global.go ./module/
COPY src/module/mutation.go ./module/
COPY src/module/validation.go ./module/

# Copy main file
COPY src/main.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/admission-webhook

# Now copy it into our base image
# Smaller, distroless
FROM gcr.io/distroless/static-debian12
COPY --from=build /build/admission-webhook /app/

# Expose https port
EXPOSE 443

# Run
CMD ["/app/admission-webhook"]