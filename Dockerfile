FROM golang:1.19.1-bullseye
WORKDIR /app
COPY ./ ./
RUN go build -o /docker
RUN go mod download
ENV PORT=7002
ENV NIKOLA=1943
ENV DB_URL=mongodb://service:b2x5bXBzaXMgbWljcm8tc2VydmljZXMgMjAyMg==@master-db.olympsis.internal/olympsis
ENV DB_NAME=olympsis
ENV DB_COL=fields
ENV KEY=SESHAT

EXPOSE 7002

CMD ["/docker"]
