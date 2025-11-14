# Resources 目录

这个目录包含应用资源文件。

## 文件说明

### icon.png
应用图标文件（512x512 像素）

**注意**: 当前使用 Fyne 内置图标作为临时方案。如需自定义图标，请替换此文件为：
- 分辨率: 512x512 或更高
- 格式: PNG
- 内容: 建议使用 Claude 或代理相关的图标设计

### claude-proxy
代理服务二进制文件（运行时内嵌）

**注意**: 此文件在构建时自动生成，不需要手动创建。

构建命令会从 `../proxy/build/claude-proxy` 复制此文件。

## 生成图标

你可以使用以下工具创建 macOS 应用图标：

1. **在线工具**:
   - https://appicon.co (推荐)
   - https://makeappicon.com

2. **命令行工具**:
   ```bash
   # 从 PNG 生成 .icns
   brew install imagemagick
   magick convert icon.png -resize 512x512 icon.icns
   ```

3. **macOS 预览**:
   - 打开 PNG 文件
   - 文件 → 导出 → 格式选择 ICNS
