FROM scratch

# ---
# ---
# ---

COPY go-http-meta-proxy /

# ---
# ---
# ---

EXPOSE 8080

# ---
# ---
# ---

ENTRYPOINT ["/go-http-meta-proxy"]

# ---
