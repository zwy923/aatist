FROM golang:1.25-alpine

WORKDIR /app

# 拷贝 go.mod 和 go.sum 先安装依赖（缓存优化）
COPY go.mod go.sum ./
RUN go mod download

# 拷贝所有代码
COPY . .

# 编译为静态二进制
RUN go build -o event_logger

# 暴露端口
EXPOSE 8081

# 启动服务
CMD ["./event_logger"]
