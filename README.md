# ShowTa云盘

用于快速搭建个人网盘和企业网盘, 『开箱即用』。 支持本地存储、阿里云盘等类型的存储挂载，随时随地查看和分享视频、图片、文件等内容。

基于 Go + Vue3 研发的网盘系统。

+ 支持预览视频、图片、文本、音频等
+ 不同用户访问权限控制
+ 可以设置文件夹访问密码
+ 支持用户通知、公告显示
+ 通过WebDAV协议, 在PC端、手机端、电视端播放ShowTa云盘上的电影、视频、音频资源。
+ 支持把ShowTa云盘映射到本地, 当作本地磁盘使用。

## 支持的平台
+ Windows
+ Linux
+ MacOS
+ 群晖

## 快速安装启动
[查看使用文档](https://www.overlink.top/intro/install/package.html)

### 默认账户
系统首次启动时会自动创建以下默认账户：

- **管理员账户**:
  - 用户名: `admin`
  - 密码: `123456`

- **访客账户**:
  - 用户名: `guest`
  - 密码: (无密码，默认只读访问)

> **安全提醒**: 首次登录后请立即修改默认密码，以确保系统安全。

### 数据库配置

系统支持 SQLite（默认）和 MySQL 数据库：

#### SQLite 配置（默认）
```ini
[database]
type = sqlite
dbname = runtime/data/nano.db
```

#### MySQL 配置
```ini
[database]
type = mysql
user = your_username
password = your_password
host = localhost
port = 3306
dbname = showta
# TLS options (optional)
tls = false
tls_skip_verify = false
tls_ca_file =
tls_cert_file =
tls_key_file =
```

> **注意**: 使用 MySQL 时，请确保数据库已创建，并且用户具有相应的权限。

## WebDAV 配置说明

ShowTa云盘内置完整的 WebDAV 服务器实现，支持通过 WebDAV 协议访问和管理云端文件。

### 连接配置
- **服务器地址**: `http://[服务器IP]:[端口]/dav`
- **默认端口**: 8888
- **示例**: `http://localhost:8888/dav`

### 认证方式
- 使用应用内的用户名和密码进行基本认证(Basic Auth)
- 支持所有已创建的用户账户
- 权限控制与Web界面保持一致

### 支持的功能
- 文件上传、下载、删除
- 目录创建、删除
- 文件复制、移动
- 属性查询(PROPFIND)
- 资源锁定(LOCK/UNLOCK)
- 支持所有WebDAV客户端

### 使用示例

#### Windows 资源管理器映射网络驱动器
1. 打开"此电脑"
2. 点击"映射网络驱动器"
3. 地址栏输入: `http://localhost:8888/dav`
4. 输入用户名和密码

#### macOS Finder 连接
1. 打开Finder
2. 菜单栏选择"前往" -> "连接服务器"
3. 服务器地址输入: `http://localhost:8888/dav`
4. 点击"连接"并输入凭证

#### Linux 命令行挂载
```bash
# 安装 davfs2
sudo apt-get install davfs2

# 挂载 WebDAV
sudo mount -t davfs http://localhost:8888/dav /mnt/webdav
```

#### 第三方客户端
支持所有标准 WebDAV 客户端，包括：
- Cyberduck (Mac/Windows)
- WinSCP (Windows)
- FileZilla (跨平台)
- 各种移动设备文件管理器

### 注意事项
- WebDAV 使用与Web界面相同的用户认证系统
- 支持所有已挂载的存储后端(本地存储、阿里云盘、百度网盘等)
- 文件操作权限受用户角色限制
- 大文件传输建议在网络稳定的环境下进行

## 构建说明
本项目使用 Makefile 进行标准化构建：

```bash
# 构建当前平台版本
make build

# 构建所有平台版本
make build-all

# 开发模式
make dev

# 清理构建产物
make clean

# 构建 Docker 镜像
make docker
```

支持的平台：
- Linux (AMD64, ARM64)
- Windows (AMD64, ARM64)
- macOS (AMD64, ARM64)

## 在线演示
[打开演示地址](http://demo.overlink.top:8888/)


## 功能展示
#### 文件列表
![文件列表](https://www.overlink.top/md/list.png)

#### 视频预览
![视频预览](https://www.overlink.top/md/video.png)

#### 图片预览
![图片预览](https://www.overlink.top/md/img.png)

#### 文本预览
![文本预览](https://www.overlink.top/md/txt.png)

#### 音频预览
![音频预览](https://www.overlink.top/md/mp3.png)

#### 登录
<img src="https://www.overlink.top/md/login.png" width="358">

#### 挂载存储(阿里云盘为例)
<img src="https://www.overlink.top/md/mount.png" width="685">

#### 设置文件夹加密&公告
<img src="https://www.overlink.top/md/folder.png" width="550">