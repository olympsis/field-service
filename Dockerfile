FROM golang:1.19.1-bullseye
WORKDIR /app
COPY ./ ./
RUN go build -o /docker
RUN go mod download
ENV PORT=7002
ENV NIKOLA=1943
ENV DB_ADDR=192.168.1.123
ENV DB_USR=service
ENV DB_PASS=b2x5bXBzaXMgbWljcm8tc2VydmljZXMgMjAyMg==
ENV DB_NAME=olympsis
ENV DB_COL=fields
ENV KEY=SZkp78avQkxGyjRakxb5Ob08zqjguNRA

EXPOSE 7002

CMD ["/docker"]
