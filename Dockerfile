FROM golang:1.24.2

WORKDIR /app

# airをインストール
RUN go install github.com/air-verse/air@latest

COPY . .
RUN go mod download

CMD ["air"]
