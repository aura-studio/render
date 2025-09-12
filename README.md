# Render 工具

基于 Go 语言和 github.com/kluctl/go-jinja2 实现的 Jinja2 模板渲染命令行工具。

## 特性
- 支持 Jinja2 语法模板渲染
- 模板文件可通过 `-t` 指定，未指定时从标准输入读取
- 渲染参数可通过 `-d` 指定 JSON 文件，未指定时优先读取环境变量 `JINJA2_VARS`，否则自动收集所有 `RENDER_` 前缀的环境变量
- 渲染结果可通过 `-o` 指定输出文件，未指定时输出到标准输出

## 使用方法

### 1. 基本用法

```bash
# 指定模板文件和参数文件，输出到标准输出
go run main.go -t template.tpl -d vars.json

# 指定模板文件，参数通过环境变量传递
go run main.go -t template.tpl

# 从标准输入读取模板内容
echo 'hello, {{ var1 }}' | go run main.go -d vars.json
```

### 2. 参数传递优先级
1. `-d` 指定参数文件（JSON 格式）
2. 环境变量 `JINJA2_VARS`（内容为 JSON 字符串）
3. 所有以 `RENDER_` 开头的环境变量（全部作为字符串类型）

#### 示例：JINJA2_VARS
```bash
export JINJA2_VARS='{"var1":"world","var2":1}'
go run main.go -t template.tpl
```

#### 示例：RENDER_ 前缀环境变量
```bash
export RENDER_var1=foo RENDER_var2=bar
go run main.go -t template.tpl
```

### 3. 输出到文件

```bash
go run main.go -t template.tpl -d vars.json -o result.txt
```

## 参数说明
- `-t` 模板文件路径，未指定时从标准输入读取
- `-d` 参数文件路径（JSON），未指定时用环境变量
- `-o` 输出文件路径，未指定时输出到标准输出

## 模板与参数示例

**模板文件 template.tpl**
```
hello, {{ var1 }}!
var2: {{ var2 }}
var3: {{ var3|join(",") }}
var4.subvar1: {{ var4.subvar1 }}
```

**参数文件 vars.json**
```
{
    "var1": "world",
    "var2": 1,
    "var3": [1,2,3],
    "var4": {"subvar1": "subworld"}
}
```

## 依赖
- Go 1.18+
- github.com/kluctl/go-jinja2

## License
Apache-2.0
