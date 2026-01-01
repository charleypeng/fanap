# Fanap - CPU温度控制风扇程序
FROM alpine:3.19

# 安装必要的工具
RUN apk add --no-cache ca-certificates

# 创建运行用户（非root）
RUN addgroup -g 1000 fanap && \
    adduser -D -u 1000 -G fanap -s /sbin/nologin -c "Fanap Service" fanap

# 创建工作目录
WORKDIR /app

# 复制二进制文件
COPY fanap-linux-amd64 /app/fanap

# 创建日志目录
RUN mkdir -p /var/log/fanap && \
    chown -R fanap:fanap /var/log/fanap

# 切换到非root用户
USER fanap

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD pgrep fanap || exit 1

# 设置环境变量
ENV FANAP_INTERVAL=5s
ENV FANAP_LOW_TEMP=40.0
ENV FANAP_HIGH_TEMP=75.0
ENV FANAP_MIN_PWM=50
ENV FANAP_MAX_PWM=255
ENV FANAP_SENSOR=auto
ENV FANAP_PWM=auto
ENV FANAP_VERBOSE=false

# 运行程序
ENTRYPOINT ["/app/fanap"]
CMD ["-verbose"]
