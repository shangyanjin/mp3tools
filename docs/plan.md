# MP3 管理工具开发计划

## 项目概述

开发一个 MP3 文件管理工具，用于批量处理音频文件的标签和元数据，解决编码问题，支持自动化批量操作。

## 核心功能

### 1. 文件夹扫描
- 递归扫描指定目录及其子目录
- 支持多种音频格式（MP3、FLAC、M4A 等）
- 生成文件列表和统计信息

### 2. 标签读取
- 读取音频文件的元数据标签（ID3、Vorbis 等）
- 显示标题、艺术家、专辑、年份等信息
- 检测标签编码问题

### 3. 编码修复
- 自动检测标签编码（UTF-8、GBK、GB2312 等）
- 识别乱码和编码错误
- 批量转换到统一编码格式

### 4. 批量元数据更新
- 批量修改标签信息
- 支持规则化重命名
- 自动填充缺失的元数据
- 完全自动化的标签处理流程
- 自动化专辑：从目录名自动推导专辑信息
- 自动化标题：支持补零格式（01 标题、02 标题，而不是 1 标题、2 标题）

### 5. 进度显示
- 类似 wget/axel 风格的输出显示
- 支持多线程处理时同时显示各线程进度
- 实时显示处理状态、进度百分比、速度等信息
- 清晰的终端输出格式

## 模块架构

### internal/scanner
- 递归目录遍历
- 音频文件识别和过滤
- 文件列表生成

### internal/tagger
- 标签解析（只读）
- 支持多格式音频标签（ID3、Vorbis 等）
- 元数据提取

### internal/writer
- ID3v2.4 标签写入（UTF-8）
- 使用 github.com/bogem/id3v2
- 支持原地更新和输出到新文件

### internal/encoder
- 自动编码检测
- 编码转换（GBK、GB2312 -> UTF-8）
- 乱码识别

### internal/processor
- 批量处理协调
- 多线程并发处理
- Worker pool 模式

### internal/display
- 实时进度显示
- 终端输出格式化
- 统计信息汇总

## 技术要点

- 递归目录遍历
- 自动编码检测和转换
- 多格式音频标签库支持
- 批量处理性能优化
- 多线程并发处理
- 实时进度显示和终端输出格式化
- 错误处理和日志记录
- 职责分离：tagger 只读，writer 只写

## 命令设计

```bash
mp3tools scan <path>              # 扫描并显示标签
mp3tools fix <path>               # 修复编码
mp3tools tag <path>               # 自动填充标签
mp3tools test <path>              # 测试结果显示
mp3tools <command> <path> -f      # 强制覆盖已存在的标签
mp3tools <command> <path> -n 5    # 指定线程数量（默认 5）
mp3tools <command> <path> -u      # 更新原MP3文件（覆盖）
mp3tools <command> <path> -o <dir>       # 输出到指定目录，保持目录结构（默认 output）
```

## 命令示范结果

### help 命令示例输出（默认）

```
$ mp3tools

MP3 Tools - Audio file metadata management utility

Usage:
  mp3tools <command> [path] [options]

Commands:
  scan <path>    Scan directory and display audio file tags
  fix <path>     Fix encoding issues in audio file tags
  tag <path>     Auto-fill missing metadata tags
  test <path>    Display test results

Options:
  -f, --force    Force overwrite existing tags
  -n, --threads  Number of worker threads (default: 5)
  -u, --update   Update original MP3 files (overwrite)
  -o, --outdir   Output directory, preserve directory structure (default: output)

Examples:
  mp3tools scan ./music
  mp3tools test ./music
  mp3tools tag ./music -f
  mp3tools fix ./music -n 8 -f
  mp3tools fix ./music -u -n 5
  mp3tools tag ./music -o ./custom

For more information about a command, use:
  mp3tools <command> --help
```

### test 命令示例输出

```
$ mp3tools test ./music

Scanning directory: ./music
Found 15 audio files

[1/15] Processing: 01 歌曲名.mp3 | Current tags: Title="1 歌曲名", Album="", Artist="" | Encoding: GBK -> UTF-8 | Auto-derived: Album="music" (from directory name) | Auto-formatted: Title="01 歌曲名" (zero-padded format) | Updated: ✓

---

Statistics:
  Total files: 15
  Successfully processed: 15
  Failed: 0
  Encoding fixed: 12
  Tags updated: 15
  Auto-derived albums: 8
  Auto-formatted titles: 10



