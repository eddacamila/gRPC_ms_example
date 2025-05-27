# 🔨 Etapa de construcción
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# ⚙️ Compilar binario estático para Linux
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .

# 🧱 Etapa para obtener certificados raíz
FROM debian:bullseye-slim AS certs
RUN apt-get update && \
    apt-get install --yes --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# 📦 Etapa final súper minimalista
FROM scratch

WORKDIR /root/

# Copiar el binario compilado
COPY --from=builder /app/server .
COPY --from=builder /app/transport /root/transport

# Copiar certificados
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Variable de entorno para TLS
ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt

EXPOSE 50051

ENTRYPOINT ["./server"]
