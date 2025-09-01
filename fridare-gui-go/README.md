# Fridare GUI

基于 Golang + Fyne 的跨平台 Frida 魔改工具，包含图形界面版本和命令行工具。

## 功能特性

### �️ 多工具集成
- **GUI版本** (fridare-gui.exe): 完整图形界面操作
- **DEB创建工具** (fridare-create.exe): 命令行DEB包创建
- **DEB修改工具** (fridare-patch.exe): 命令行DEB包修改
- 三套工具互相配合，灵活适应不同使用场景

### �📥 下载模块
- 自动获取 Frida 最新版本列表
- 支持多平台下载（Android、iOS、Windows、Linux、macOS）
- 支持多种文件类型（frida-server、frida-gadget、frida-tools）
- 带进度显示的下载功能
- 网络代理支持

### 🔧 二进制魔改
- 自动检测文件类型（frida-server、frida-agent.dylib）
- 预定义修补模式和自定义十六进制修补
- Magic值替换绕过frida检测
- 名称和端口修改
- 文件备份功能

### 📦 iOS DEB 打包
- **DEB包创建**: 从原始frida文件创建全新DEB包
- **DEB包修改**: 修改现有DEB包的内容  
- **标准权限**: 符合Debian官方包管理规范
- **Root/Rootless**: 自动适配iOS越狱环境
- 可配置包名、版本和依赖关系
- 支持Root/Rootless结构
- 自动权限设置和文件校验

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

## 快速开始

### 安装方式

#### 方式一：下载预编译版本 (推荐)
1. 访问 [Releases 页面](../../releases) 下载最新版本
2. 解压到任意目录
3. 双击运行 `fridare-gui.exe` 开始使用

#### 方式二：从源码构建
```bash
# 克隆项目
git clone https://github.com/your-repo/fridare.git
cd fridare/fridare-gui-go

# 安装依赖
go mod tidy

# 构建所有工具
make build

# 运行GUI版本
./build/fridare-gui.exe
```

## 项目结构

```
fridare-gui-go/
├── cmd/                    # 命令行工具集合
│   ├── gui/               # GUI版本主程序
│   │   └── main.go
│   ├── create/            # DEB包创建CLI工具
│   │   └── main.go
│   └── patch/             # DEB包修改CLI工具
│       └── main.go
├── internal/               # 内部包
│   ├── config/            # 配置管理
│   │   └── config.go
│   ├── core/              # 核心功能
│   │   ├── frida.go       # Frida 相关操作
│   │   ├── patcher.go     # 二进制修补
│   │   ├── debpackager.go # DEB包处理
│   │   └── hexreplace.go  # 十六进制替换
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

#### 构建所有版本 (推荐)
```bash
# 构建GUI + 两个CLI工具
make build

# 查看构建结果
ls build/
# fridare-gui.exe     - GUI版本 
# fridare-create.exe  - DEB包创建工具
# fridare-patch.exe   - DEB包修改工具
```

#### 单独构建
```bash
# 仅构建GUI版本
make build-gui

# 仅构建DEB创建工具
make build-create

# 仅构建DEB修改工具  
make build-patch
```

#### 运行应用程序
```bash
# 运行GUI版本
make run
# 或者
./build/fridare-gui.exe

# 运行CLI工具
./build/fridare-create.exe --help
./build/fridare-patch.exe --help
```

#### 跨平台构建
```bash
# 构建所有平台版本
make build-all

# 使用 fyne package 打包GUI版本
make package
```

## 工具使用说明

### 1. GUI版本 (fridare-gui.exe)
提供完整的图形界面，包含所有功能模块：
- 📥 下载模块
- 🔧 魔改模块  
- 📦 iOS魔改+打包模块
- 🆕 创建DEB包模块
- 🛠️ frida-tools魔改模块

### 2. DEB包创建工具 (fridare-create.exe)
从原始frida文件创建全新的DEB包：

```bash
# 基本用法
./fridare-create.exe -server frida-server -magic agent -output agent.deb

# 包含agent库的完整包
./fridare-create.exe -server frida-server -agent frida-agent.dylib -magic agent -rootless -output agent.deb

# 从现有DEB包提取agent并创建新包
./fridare-create.exe -server frida-server -extract-deb frida_17.2.17.deb -magic agent -output agent.deb
```

### 3. DEB包修改工具 (fridare-patch.exe)
修改现有DEB包的内容：

```bash
# 修改现有DEB包
./fridare-patch.exe input.deb output.deb magic 27042
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
