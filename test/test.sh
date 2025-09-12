#!/bin/sh
set -e

echo "== 测试参数文件渲染 =="
output=$(go run ../main.go -t test.tpl -d test.json)
expected="hello, world!
var2: 1
var3: 1,2,3
var4.subvar1: subworld"
if [ "$output" = "$expected" ]; then
  echo "参数文件渲染通过"
else
  echo "参数文件渲染失败"
  echo "输出: $output"
  exit 1
fi

echo "== 测试JINJA2_VARS环境变量渲染 =="
export JINJA2_VARS='{"var1":"env","var2":2,"var3":[4,5],"var4":{"subvar1":"envsub"}}'
output=$(go run ../main.go -t test.tpl)
unset JINJA2_VARS
expected="hello, env!
var2: 2
var3: 4,5
var4.subvar1: envsub"
if [ "$output" = "$expected" ]; then
  echo "JINJA2_VARS渲染通过"
else
  echo "JINJA2_VARS渲染失败"
  echo "输出: $output"
  exit 1
fi

echo "== 测试RENDER_环境变量渲染 =="
export RENDER_var1=foo RENDER_var2=3 RENDER_var3='[6,7]' RENDER_var4='{"subvar1":"bar"}'
output=$(go run ../main.go -t test.tpl)
unset RENDER_var1 RENDER_var2 RENDER_var3 RENDER_var4
# 由于RENDER_变量全部为字符串，var3和var4.subvar1的渲染会有差异
expected="hello, foo!
var2: 3
var3: 6,7
var4.subvar1: bar"
if [ "$output" = "$expected" ]; then
  echo "RENDER_环境变量渲染通过"
else
  echo "RENDER_环境变量渲染失败"
  echo "输出: $output"
  exit 1
fi

echo "所有测试通过"
