FROM golang:latest

WORKDIR /app

# airをインストール
RUN go install github.com/air-verse/air@latest

COPY . .
RUN go mod download

CMD ["air"]
