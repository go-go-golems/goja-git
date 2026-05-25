package main

import (
	"fmt"
	"os"

	"github.com/dop251/goja"
	"github.com/go-go-golems/goja-git/pkg/gitjs"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <script.js>\n", os.Args[0])
		os.Exit(1)
	}

	scriptPath := os.Args[1]
	scriptBytes, err := os.ReadFile(scriptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading script: %v\n", err)
		os.Exit(1)
	}

	rt := goja.New()

	// Install the git module
	gitjs.InstallGit(rt)

	// Add a simple console.log implementation
	console := rt.NewObject()
	console.Set("log", func(call goja.FunctionCall) goja.Value {
		args := make([]interface{}, len(call.Arguments))
		for i, arg := range call.Arguments {
			args[i] = arg.Export()
		}
		fmt.Println(args...)
		return goja.Undefined()
	})
	rt.Set("console", console)

	// Add JSON support
	rt.Set("JSON", rt.NewObject())
	rt.RunString(`
		JSON.stringify = function(obj, replacer, space) {
			return __internal_stringify(obj, space || 0);
		};
		JSON.parse = function(str) {
			return __internal_parse(str);
		};
	`)

	rt.Set("__internal_stringify", func(call goja.FunctionCall) goja.Value {
		obj := call.Argument(0).Export()
		indent := 0
		if len(call.Arguments) > 1 {
			if num := call.Argument(1).ToInteger(); num > 0 {
				indent = int(num)
			}
		}

		result := stringify(obj, indent, 0)
		return rt.ToValue(result)
	})

	// Run the script
	_, err = rt.RunString(string(scriptBytes))
	if err != nil {
		if ex, ok := err.(*goja.Exception); ok {
			fmt.Fprintf(os.Stderr, "JavaScript error: %v\n", ex.String())
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}

func stringify(obj interface{}, indent, depth int) string {
	switch v := obj.(type) {
	case nil:
		return "null"
	case bool:
		if v {
			return "true"
		}
		return "false"
	case string:
		return fmt.Sprintf("%q", v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprintf("%v", v)
	case map[string]interface{}:
		if len(v) == 0 {
			return "{}"
		}
		if indent == 0 {
			result := "{"
			first := true
			for k, val := range v {
				if !first {
					result += ","
				}
				first = false
				result += fmt.Sprintf("%q:%s", k, stringify(val, indent, depth+1))
			}
			result += "}"
			return result
		} else {
			result := "{\n"
			first := true
			for k, val := range v {
				if !first {
					result += ",\n"
				}
				first = false
				for i := 0; i < (depth+1)*indent; i++ {
					result += " "
				}
				result += fmt.Sprintf("%q: %s", k, stringify(val, indent, depth+1))
			}
			result += "\n"
			for i := 0; i < depth*indent; i++ {
				result += " "
			}
			result += "}"
			return result
		}
	case []interface{}:
		if len(v) == 0 {
			return "[]"
		}
		if indent == 0 {
			result := "["
			for i, val := range v {
				if i > 0 {
					result += ","
				}
				result += stringify(val, indent, depth+1)
			}
			result += "]"
			return result
		} else {
			result := "[\n"
			for i, val := range v {
				if i > 0 {
					result += ",\n"
				}
				for j := 0; j < (depth+1)*indent; j++ {
					result += " "
				}
				result += stringify(val, indent, depth+1)
			}
			result += "\n"
			for i := 0; i < depth*indent; i++ {
				result += " "
			}
			result += "]"
			return result
		}
	default:
		return fmt.Sprintf("%v", v)
	}
}
