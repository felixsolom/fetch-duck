FROM gcr.io/distroless/static-debian11

WORKDIR /app

COPY fetch-duck .

EXPOSE 8080

USER 65532:65532

ENTRYPOINT ["/app/fetch-duck"]