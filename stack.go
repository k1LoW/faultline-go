package gobrake

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// getBacktrace returns the stacktrace associated with e. If e is an
// error from the errors package its stacktrace is extracted, otherwise
// the current stacktrace is collected end returned.
func getBacktrace(e interface{}, depth int) []StackFrame {
	if err, ok := e.(stackTracer); ok {
		return backtraceFromErrorWithStackTrace(err)
	}

	frames := make([]StackFrame, 0)
	for i := depth; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		packageName, funcName := packageFuncName(pc)
		if stackFilter(packageName, funcName, file, line) {
			frames = frames[:0]
			continue
		}
		frames = append(frames, StackFrame{
			File: file,
			Line: line,
			Func: funcName,
		})
	}

	return frames
}

func packageFuncName(pc uintptr) (string, string) {
	f := runtime.FuncForPC(pc)
	if f == nil {
		return "", ""
	}

	packageName := ""
	funcName := f.Name()

	if ind := strings.LastIndex(funcName, "/"); ind > 0 {
		packageName += funcName[:ind+1]
		funcName = funcName[ind+1:]
	}
	if ind := strings.Index(funcName, "."); ind > 0 {
		packageName += funcName[:ind]
		funcName = funcName[ind+1:]
	}

	return packageName, funcName
}

func stackFilter(packageName, funcName string, file string, line int) bool {
	return packageName == "runtime" && funcName == "panic"
}

// stackTraces returns the stackTrace of an error.
// It is part of the errors package public interface.
type stackTracer interface {
	StackTrace() errors.StackTrace
}

// backtraceFromErrorWithStackTrace extracts the stacktrace from e.
func backtraceFromErrorWithStackTrace(e stackTracer) []StackFrame {
	const sep = "\n\t"

	stackTrace := e.StackTrace()
	frames := make([]StackFrame, len(stackTrace))
	for i, f := range stackTrace {
		line, _ := strconv.ParseInt(fmt.Sprintf("%d", f), 10, 64)
		file := fmt.Sprintf("%+s", f)
		if ind := strings.Index(file, sep); ind != -1 {
			file = file[ind+len(sep):]
		}
		frames[i] = StackFrame{
			File: file,
			Line: int(line),
			Func: fmt.Sprintf("%n", f),
		}
	}
	return frames
}

// getTypeName returns the type name of e.
func getTypeName(e interface{}) string {
	if err, ok := e.(error); ok {
		e = errors.Cause(err)
	}
	return fmt.Sprintf("%T", e)
}
