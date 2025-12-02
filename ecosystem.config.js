module.exports = {
    apps: [
        {
            // 进程名称（必填，唯一标识）
            name: "md-to-pdf-service",
            // 启动命令（Golang 二进制文件路径）
            script: "./md-to-pdf",
            // 环境变量（可选，Golang 可通过 os.Getenv() 读取）
            env: {
                GO_ENV: "production",
            },
            // 日志配置（推荐配置，便于问题排查）
            log_date_format: "YYYY-MM-DD HH:mm:ss",  // 日志时间格式
            out_file: "./logs/out.log",              // 标准输出日志（stdout）
            error_file: "./logs/error.log",          // 错误日志（stderr）
            merge_logs: true,                        // 合并日志（避免多进程日志拆分）
            log_rotate_size: "10M",                  // 日志轮转大小（10MB 自动分割）
            log_rotate_backups: 10,                  // 保留日志备份数（10 个）
            // 进程管理配置
            instances: "1",  // 启动实例数（max 表示根据 CPU 核心数自动分配）
            autorestart: true,     // 进程崩溃后自动重启（生产环境必备）
            watch: false,          // 禁用文件监听（Golang 是编译型语言，无需热重载）
            max_memory_restart: "1G",  // 内存占用超过 1G 自动重启（避免内存泄漏）
            restart_delay: 3000,    // 重启延迟 3 秒（避免频繁重启）
        },
    ],
};