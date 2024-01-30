FROM golang:1.21.5 as builder
# 작업 디렉토리를 설정합니다.
WORKDIR /app/NameServer_Finder

# 소스 코드를 복사합니다.
COPY . .

# Go 어플리케이션을 빌드합니다.
RUN go build -o NameServer_Finder src/main.go

# 최종 이미지를 가져옵니다.
FROM alpine:latest

RUN apk add --no-cache libc6-compat
# 작업 디렉토리를 설정합니다.
WORKDIR /app/NameServer_Finder

# 빌드한 실행 파일을 복사합니다.
COPY --from=builder /app/NameServer_Finder .


# 실행할 명령을 지정합니다.
ENTRYPOINT ["./NameServer_Finder"]