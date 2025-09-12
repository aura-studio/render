// Package render 提供命令行渲染工具，支持Jinja2模板渲染，参数可来自文件、环境变量。
package render

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	jinja2 "github.com/kluctl/go-jinja2"
)

// Execute 解析命令行参数，读取模板和渲染参数，执行Jinja2渲染并输出结果。
func Execute() {
	templatePath := flag.String("t", "", "模板文件路径，未指定则使用标准输入")
	flag.String("d", "", "参数文件路径（JSON），未指定则用环境变量")
	flag.String("o", "", "输出文件路径，未指定则输出到标准输出")
	flag.Parse()

	// 读取模板内容（支持-t参数或标准输入）
	var tpl string
	if *templatePath == "" {
		tplBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "从标准输入读取模板失败: %v\n", err)
			os.Exit(1)
		}
		tpl = string(tplBytes)
	} else {
		tplBytes, err := os.ReadFile(*templatePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取模板文件失败: %v\n", err)
			os.Exit(1)
		}
		tpl = string(tplBytes)
	}

	// 参数处理（支持-d参数、JINJA2_VARS、RENDER_前缀环境变量）
	var data map[string]interface{}
	dFlag := flag.Lookup("d")
	dataPath := ""
	if dFlag != nil {
		dataPath = dFlag.Value.String()
	}
	if dataPath != "" {
		dataBytes, err := os.ReadFile(dataPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取参数文件失败: %v\n", err)
			os.Exit(1)
		}
		if err := json.Unmarshal(dataBytes, &data); err != nil {
			fmt.Fprintf(os.Stderr, "解析参数文件失败: %v\n", err)
			os.Exit(1)
		}
	} else if env := os.Getenv("JINJA2_VARS"); env != "" {
		if err := json.Unmarshal([]byte(env), &data); err != nil {
			fmt.Fprintf(os.Stderr, "解析JINJA2_VARS失败: %v\n", err)
			os.Exit(1)
		}
	} else {
		data = make(map[string]interface{})
		for _, env := range os.Environ() {
			if !strings.HasPrefix(env, "RENDER_") {
				continue
			}
			parts := strings.SplitN(env, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := parts[0][7:]
			val := parts[1]
			var v interface{}
			if err := json.Unmarshal([]byte(val), &v); err == nil {
				data[key] = v
			} else {
				data[key] = val
			}
		}
	}

	// 渲染（构造WithGlobal参数并调用go-jinja2）
	var opts []jinja2.Jinja2Opt
	for k, v := range data {
		opts = append(opts, jinja2.WithGlobal(k, v))
	}

	j2, err := jinja2.NewJinja2("render", 1, opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化Jinja2失败: %v\n", err)
		os.Exit(1)
	}
	defer j2.Close()

	result, err := j2.RenderString(tpl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "渲染模板失败: %v\n", err)
		os.Exit(1)
	}

	// 输出（支持-o参数或标准输出）
	oFlag := flag.Lookup("o")
	outputPath := ""
	if oFlag != nil {
		outputPath = oFlag.Value.String()
	}
	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(result), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "写入输出文件失败: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Print(result)
	}
}
