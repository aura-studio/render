package render

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	pongo2 "github.com/flosch/pongo2/v6"
)

// Execute 作为程序唯一入口，处理参数和退出码
func Execute() {
	tplFile := flag.String("f", "", "template file (.tmpl/.template)")
	outFile := flag.String("o", "", "output file (optional)")
	checkOnly := flag.Bool("check-only", false, "only validate required variables; no render")
	emitDefaults := flag.Bool("emit-defaults", false, "print shell exports for defaults and exit")
	flag.Parse()

	if err := runAll(*tplFile, *outFile, *checkOnly, *emitDefaults); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func runAll(tplFile, outFile string, checkOnly, emitDefaults bool) error {
	if tplFile == "" {
		return fmt.Errorf("-f required")
	}
	if ext := filepath.Ext(tplFile); ext != ".tmpl" && ext != ".template" {
		return fmt.Errorf("unsupported suffix: %s", ext)
	}
	if st, err := os.Stat(tplFile); err != nil {
		return err
	} else if st.IsDir() {
		return fmt.Errorf("template is directory")
	}

	b, err := os.ReadFile(tplFile)
	if err != nil {
		return err
	}
	cleaned, required, defaults, err := parseDirectives(string(b))
	if err != nil {
		return err
	}

	if emitDefaults {
		keys := make([]string, 0, len(defaults))
		for k := range defaults {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("export %s=%q\n", k, defaults[k])
		}
		return nil
	}

	ctx := envContext()
	for k, v := range defaults {
		if cur, ok := ctx[k]; !ok || cur == "" {
			ctx[k] = v
		}
	}

	if len(required) > 0 {
		missing := []string{}
		for _, rv := range required {
			if val, ok := ctx[rv]; !ok || strings.TrimSpace(fmt.Sprint(val)) == "" {
				missing = append(missing, rv)
			}
		}
		if len(missing) > 0 {
			return fmt.Errorf("missing required variables: %s", strings.Join(missing, ", "))
		}
	}

	if checkOnly {
		return nil
	}

	tpl, err := pongo2.FromString(cleaned)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	out, err := tpl.Execute(ctx)
	if err != nil {
		return fmt.Errorf("execute: %w", err)
	}

	target := outFile
	if target == "" {
		target = strings.TrimSuffix(tplFile, filepath.Ext(tplFile))
	}
	if err = os.WriteFile(target, []byte(out), 0644); err != nil {
		return err
	}
	log.Printf("rendered %s => %s", tplFile, target)
	return nil
}

// parseDirectives extracts custom directives and returns cleaned template string.
// Directives:
//
//	{% required VAR1 VAR2 %}
//	{% default VAR=value %}
//
// Variables with defaults are not treated as required even if also listed.
func parseDirectives(s string) (clean string, required []string, defaults map[string]string, err error) {
	reReq := regexp.MustCompile(`(?m)\{\%\s*required\s+([A-Za-z0-9_\s]+?)\s*\%\}`)
	reDef := regexp.MustCompile(`(?m)\{\%\s*default\s+([A-Za-z0-9_]+)\s*=\s*(.*?)\s*\%\}`)
	reqSet := map[string]struct{}{}
	defaults = map[string]string{}

	s = reReq.ReplaceAllStringFunc(s, func(m string) string {
		sub := reReq.FindStringSubmatch(m)
		if len(sub) == 2 {
			for _, f := range strings.Fields(sub[1]) {
				if f != "" {
					reqSet[f] = struct{}{}
				}
			}
		}
		return ""
	})
	s = reDef.ReplaceAllStringFunc(s, func(m string) string {
		sub := reDef.FindStringSubmatch(m)
		if len(sub) == 3 {
			key := sub[1]
			raw := strings.TrimSpace(sub[2])
			val := strings.Trim(raw, " \"'")
			if strings.HasPrefix(raw, "`") && strings.HasSuffix(raw, "`") && len(raw) >= 2 {
				cmdStr := strings.Trim(raw, "`")
				if out, e := runShell(cmdStr); e == nil {
					val = out
				} else {
					err = fmt.Errorf("default command failed for %s: %w", key, e)
				}
			}
			defaults[key] = val
		}
		return ""
	})
	for k := range reqSet {
		if _, has := defaults[k]; !has {
			required = append(required, k)
		}
	}
	sort.Strings(required)
	clean = s
	return
}

func envContext() pongo2.Context {
	ctx := pongo2.Context{}
	for _, e := range os.Environ() {
		if i := strings.IndexByte(e, '='); i > 0 {
			ctx[e[:i]] = e[i+1:]
		}
	}
	return ctx
}

// runShell executes command via sh/bash/cmd depending on platform.
func runShell(cmdStr string) (string, error) {
	shell := "sh"
	args := []string{"-c", cmdStr}
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("bash"); err == nil {
			shell = "bash"
			args = []string{"-c", cmdStr}
		} else {
			shell = "cmd"
			args = []string{"/C", cmdStr}
		}
	}
	c := exec.Command(shell, args...)
	out, err := c.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
