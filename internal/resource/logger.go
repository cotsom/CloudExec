package resource

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/cotsom/CloudExec/internal/utils"
)

func getGorutineID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	s := strings.TrimPrefix(string(b), "goroutine ")
	s = s[:strings.Index(s, " ")]
	gid, _ := strconv.ParseUint(s, 10, 64)
	return gid
}

type Logger struct {
	buffers map[uint64]string
	mu      sync.Mutex
}

func (l *Logger) printGorutine(text string) {
	gid := getGorutineID()

	if gid == 1 {
		fmt.Print(text)
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.buffers[gid] += text
}

func (l *Logger) Info(text string) {
	l.printGorutine(utils.ColorizeFMT(utils.ColorBlue, fmt.Sprintf("[*] %s", text)))
}

func (l *Logger) Found(text string) {
	l.printGorutine(utils.ColorizeFMT(utils.ColorGreen, fmt.Sprintf("[+] %s", text)))
}

func (l *Logger) Error(text string) {
	l.printGorutine(utils.ColorizeFMT(utils.ColorRed, fmt.Sprintf("[-] %s", text)))
}

func (l *Logger) Fatal(text string) {
	l.printGorutine(utils.ColorizeFMT(utils.ColorYellow, fmt.Sprintf("[!!!] %s", text)))
}

func (l *Logger) Raw(text string) {
	l.printGorutine(fmt.Sprintf("%s\n", text))
}

func (l *Logger) List(text string) {
	l.printGorutine(fmt.Sprintf("-> %s\n", text))
}

func (l *Logger) DeferPrint() {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Print(l.buffers[getGorutineID()])
}

func NewLogger() *Logger {
	return &Logger{
		buffers: make(map[uint64]string),
	}
}
