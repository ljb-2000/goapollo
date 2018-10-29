package log

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
	"strings"
	"runtime"
	"time"
)

var logHandler = make(map[string] *os.File)
var logLock = new(sync.Mutex)

func getHandler(logPath string, level log.Level) (*os.File, error) {
	logLock.Lock()
	defer logLock.Unlock()
	//初始化当前，后天的文件句柄
	ym := time.Now().Local().Format("2006-01")
	t        := time.Now()
	day      := fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())
	dir      := fmt.Sprintf("%s/%s", logPath, ym)
	dfile    := fmt.Sprintf("%s/%s/%s-%s.log", logPath, ym, level.String(), day)
	//logsDir := &file.WPath{Dir:dir}
	if _,err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
	key := fmt.Sprintf("%s%d", day, level)
	var err error
	_, ok := logHandler[key]
	if !ok {
		//初始化当前，后天的文件句柄
		logHandler[key], err = os.OpenFile(dfile, os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_APPEND, 0755)
		if err != nil {
			return nil, err
		}
	}
	for k, v := range logHandler{
		if k != key {
			delete(logHandler, k)
			v.Close()
		}
	}
	return logHandler[key], nil
}

type ContextHook struct {
	LogPath string
}

func (hook *ContextHook) Levels() []log.Level {
	return log.AllLevels
}

func (hook *ContextHook) getCallerInfo() (string, string, int) {
	var (
		shortPath string
		funcName string
	)
	for i := 3; i < 15; i++ {
		pc, fullPath, line, ok := runtime.Caller(i)
		if !ok {
			fmt.Println("error: error during runtime.Caller")
			continue
		} else {
			lastS := strings.LastIndex(fullPath, "/")
			if lastS < 0 {
				lastS = strings.LastIndex(fullPath, "\\")
			}
			//if strings.HasPrefix(fullPath, workingDir) {
			//	shortPath = fullPath[len(workingDir):]
			//} else {
			//	shortPath = fullPath
			//}
			shortPath = fullPath[lastS+1:]
			funcName = runtime.FuncForPC(pc).Name()

			index := strings.LastIndex(funcName, ".")
			if index > 0 {
				funcName = funcName[index+1:]
			}
			if !strings.Contains(strings.ToLower(fullPath), "github.com/sirupsen/logrus") {
				return shortPath, funcName, line
				break
			}
		}
	}
	return "", "", 0
}

func (hook *ContextHook) Fire(entry *log.Entry) error {
	shortPath, funcName, callLine := hook.getCallerInfo()
	if shortPath != "" && callLine != 0 {
		entry.Message = fmt.Sprintf("[%s:%d %s] ", shortPath, callLine, funcName) + entry.Message
	}
	line, err := entry.String()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read entry, %v\n", err)
		return err
	}
	handler, err := getHandler(hook.LogPath, entry.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get log handler error, %v\n", err)
		return nil
	}
	handler.Write([]byte(line))
	return nil
}
