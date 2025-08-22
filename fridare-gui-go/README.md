# Fridare GUI

基于 Golang + Fyne 的跨平台 Frida 魔改工具图形界面版本。

## 功能特性

### 📥 下载模块
- 自动获取 Frida 最新版本列表
- 支持多平台下载（Android、iOS、Windows、Linux、macOS）
- 支持多种文件类型（frida-server、frida-gadget、frida-tools）
- 带进度显示的下载功能
- 网络代理支持

### 🔧 二进制魔改
- 自动检测文件类型
- 预定义修补模式
- 自定义十六进制修补
- 名称和端口修改
- 文件备份功能

### 📦 iOS DEB 打包
- 自动生成 DEB 包结构
- 可配置包名和版本
- 支持多架构打包
- 自动安装到设备

### 🛠️ 系统工具
- 环境依赖检查
- Python 环境检测
- 设备连接状态
- 批量操作支持

### ⚙️ 配置管理
- 持久化配置存储
- 主题切换（明亮/暗黑/自动）
- 代理设置
- 工作目录配置

## 项目结构

```
fridare-gui-go/
├── cmd/                    # 主程序入口
│   └── main.go
├── internal/               # 内部包
│   ├── config/            # 配置管理
│   │   └── config.go
│   ├── core/              # 核心功能
│   │   ├── frida.go       # Frida 相关操作
│   │   └── patcher.go     # 二进制修补
│   ├── ui/                # 用户界面
│   │   ├── main_window.go # 主窗口
│   │   ├── download_tab.go # 下载标签页
│   │   └── tabs.go        # 其他标签页
│   └── utils/             # 工具函数
├── pkg/                   # 公共包
├── assets/                # 资源文件
├── build/                 # 构建输出
├── dist/                  # 分发包
├── go.mod                 # Go 模块文件
├── go.sum                 # 依赖校验文件
├── Makefile              # 构建脚本
└── README.md             # 项目说明
```

## 技术栈

- **Go 1.21+**: 主要编程语言
- **Fyne v2**: 跨平台 GUI 框架
- **Resty**: HTTP 客户端库
- **YAML**: 配置文件格式

## 快速开始

### 环境要求

1. Go 1.21 或更高版本
2. Git

### 安装依赖

```bash
# 克隆项目
git clone https://github.com/suifei/fridare.git
cd fridare/fridare-gui-go

# 初始化模块并安装依赖
make dev-setup
```

### 构建和运行

```bash
# 构建应用程序
./build.sh

# 运行应用程序
./build/fridare-gui.exe
```

### 手动构建

```bash
cd cmd && fyne build -o ../build/fridare-gui.exe
```
make build-all

# 使用 fyne package 打包
make package
```

## 构建脚本

项目提供了完整的 Makefile，支持以下操作：

- `make deps` - 安装依赖
- `make build` - 本地构建
- `make run` - 构建并运行
- `make build-all` - 跨平台构建
- `make package` - 使用 fyne package 打包
- `make test` - 运行测试
- `make clean` - 清理构建文件
- `make fmt` - 格式化代码
- `make vet` - 代码检查

## 配置文件

应用程序配置保存在用户配置目录下的 `fridare/config.json` 文件中：

```json
{
  "app_version": "1.0.0",
  "work_dir": "/home/user/.fridare",
  "hexreplace_path": "hexreplace/hexreplace",
  "proxy": "",
  "timeout": 30,
  "retries": 3,
  "default_port": 27042,
  "magic_name": "frida",
  "auto_confirm": false,
  "theme": "auto",
  "window_width": 1200,
  "window_height": 800,
  "debug_mode": false,
  "download_dir": "/home/user/Downloads/fridare",
  "concurrent_downloads": 3,
  "recent_versions": [],
  "recent_platforms": []
}
```

## 核心功能实现

### 1. Frida 版本管理

- 通过 GitHub API 获取 Frida 发行版信息
- 版本号智能排序
- 资源文件自动匹配
- 支持自定义版本输入

### 2. 二进制文件修补

- 基于原 hexreplace 工具的 Go 实现
- 支持预定义修补模式
- 自定义十六进制模式替换
- 文件类型自动检测

### 3. 网络下载

- 断点续传支持
- 进度实时显示
- 速度监控
- 代理服务器支持

### 4. 跨平台支持

- Windows (x64, ARM64)
- macOS (Intel, Apple Silicon)
- Linux (x64, ARM64)
- 原生外观适配

## 从 Shell 脚本迁移

本项目完整重新实现了原 `fridare.sh` 的核心功能：

### 已实现功能
- ✅ 版本列表获取和下载
- ✅ 二进制文件修补
- ✅ 配置管理
- ✅ 跨平台支持

### 待实现功能
- 🔄 iOS DEB 包生成
- 🔄 frida-tools 修补
- 🔄 环境依赖检查
- 🔄 批量操作
- 🔄 设备管理

## 开发计划

### v1.0.0 (基础版本)
- [x] 基础 UI 框架
- [x] 下载功能
- [x] 配置管理
- [ ] 二进制修补
- [ ] 基本测试

### v1.1.0 (功能完善)
- [ ] iOS DEB 打包
- [ ] frida-tools 修补
- [ ] 环境检测
- [ ] 批量操作

### v1.2.0 (高级功能)
- [ ] 插件系统
- [ ] 自动更新
- [ ] 多语言支持
- [ ] 高级配置

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

本项目采用与原 fridare 项目相同的许可证。

## 免责声明

本工具仅供学习和研究使用，请勿用于非法用途。使用本工具修改 Frida 可能违反相关软件的使用条款，用户需自行承担风险和法律责任。

## 作者

- suifei@gmail.com
- https://github.com/suifei/fridare

## 致谢

感谢原 fridare.sh 脚本的设计和实现，为本 GUI 版本提供了功能参考。
