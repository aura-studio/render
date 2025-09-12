#!/bin/sh
set -e

echo "== 多参数文件+JINJA2_VARS+RENDER_混合测试 =="

# a.json: {"var1": "foo", "var2": 1, "var3": [1,2]}
# b.json: {"var2": "bar", "var3": [4,5]}
# JINJA2_VARS: {"var4": {"subvar1": "envsub"}}
# RENDER_var1: baz

export JINJA2_VARS='{"var4":{"subvar1":"envsub"}}'
export RENDER_var1=baz
output=$(go run ../main.go -t test.tpl -d a.json -d b.json)
unset JINJA2_VARS RENDER_var1

expected="hello, baz!
var2: bar
var3: 4,5
var4.subvar1: envsub"

if [ "$output" = "$expected" ]; then
  echo "混合参数合并测试通过"
else
  echo "混合参数合并测试失败"
  echo "输出: $output"
  exit 1
fi
