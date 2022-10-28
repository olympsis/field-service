FROM golang:1.19.1-bullseye
WORKDIR /app
COPY ./ ./
RUN go build -o /docker
RUN go mod download
ENV PORT=7002
ENV NIKOLA=1943
ENV DATABASE=mongodb://192.168.1.211:27017
ENV DB_NAME=olympsis
ENV USER_COL=fields
ENV KEY=SESHAT

EXPOSE 7002

CMD ["/docker"]
