package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

var dmesgBuf []string
var bootTime time.Time

func main() {
	bootTime = time.Now()
	bootSequence()
	startShell()
}

func bootSequence() {
	banner := []string{
		"HADES V2 — Pure Go Kernel",
		"Initializing Go runtime subsystems...",
		"Probing CPU…",
		"Initializing memory manager…",
		"Bringing up virtual network interfaces…",
		"Starting init (PID 1) — hades shell",
	}

	for _, l := range banner {
		println(l)
		appendDmesg(l)
		time.Sleep(150 * time.Millisecond)
	}

	println()
	println("Boot complete — Welcome to HADES. Type 'help' for commands.")
	appendDmesg("Boot complete")
}

func appendDmesg(msg string) {
	timestamp := time.Now().Format(time.RFC3339)
	dmesgBuf = append(dmesgBuf, fmt.Sprintf("%s %s", timestamp, msg))
}

func startShell() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("hades> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println()
				return
			}
			fmt.Fprintln(os.Stderr, "error reading input:", err)
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		cmd := fields[0]
		args := fields[1:]
		switch cmd {
		case "help":
			doHelp()
		case "uname":
			doUname()
		case "ps":
			doPs()
		case "meminfo":
			doMeminfo()
		case "cpuinfo":
			doCPUInfo()
		case "uptime":
			doUptime()
		case "dmesg":
			doDmesg(args)
		case "ls":
			doLs(args)
		case "cat":
			doCat(args)
		case "ifconfig":
			doIfconfig()
		case "clear":
			doClear()
		case "shutdown", "poweroff":
			fmt.Println("Shutting down... Goodbye.")
			appendDmesg("Shutdown requested via shell")
			os.Exit(0)
		case "exit":
			return
		default:
			fmt.Println("unknown command:", cmd)
		}
	}
}

func doHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  help        - show this help")
	fmt.Println("  uname       - print OS and architecture")
	fmt.Println("  ps          - list processes (best-effort)")
	fmt.Println("  meminfo     - memory usage information")
	fmt.Println("  cpuinfo     - CPU information")
	fmt.Println("  uptime      - system uptime (since this process started)")
	fmt.Println("  dmesg       - show kernel messages")
	fmt.Println("  ls [path]   - list directory")
	fmt.Println("  cat <file>  - show file contents")
	fmt.Println("  ifconfig    - list network interfaces")
	fmt.Println("  clear       - clear the screen")
	fmt.Println("  shutdown    - power off (exit)")
	fmt.Println("  exit        - exit shell")
}

func doUname() {
	// Try to read /proc/version for kernel info; fallback to Go runtime
	if data, err := ioutil.ReadFile("/proc/version"); err == nil {
		fmt.Printf("%s %s\n", strings.TrimSpace(string(data)), runtime.GOARCH)
		return
	}
	fmt.Printf("%s/%s (Go runtime)\n", runtime.GOOS, runtime.GOARCH)
}

func doPs() {
	// Attempt to read /proc to list PIDs; otherwise show current process
	if ents, err := ioutil.ReadDir("/proc"); err == nil {
		pids := []int{}
		for _, e := range ents {
			if !e.IsDir() {
				continue
			}
			if pid, err := strconv.Atoi(e.Name()); err == nil {
				pids = append(pids, pid)
			}
		}
		sort.Ints(pids)
		fmt.Printf("PID\tCMD\n")
		for _, pid := range pids {
			cmdlinePath := filepath.Join("/proc", strconv.Itoa(pid), "cmdline")
			if data, err := ioutil.ReadFile(cmdlinePath); err == nil {
				cmdline := strings.ReplaceAll(string(data), "\x00", " ")
				if cmdline == "" {
					// try reading comm
					commPath := filepath.Join("/proc", strconv.Itoa(pid), "comm")
					if c, err := ioutil.ReadFile(commPath); err == nil {
						cmdline = strings.TrimSpace(string(c))
					}
				}
				fmt.Printf("%d\t%s\n", pid, cmdline)
			}
		}
		return
	}
	// Fallback
	fmt.Printf("PID\tCMD\n")
	fmt.Printf("%d\t%s\n", os.Getpid(), os.Args[0])
}

func doMeminfo() {
	// Prefer /proc/meminfo
	if data, err := ioutil.ReadFile("/proc/meminfo"); err == nil {
		fmt.Print(string(data))
		return
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB\n", bToMb(m.Alloc))
	fmt.Printf("TotalAlloc = %v MiB\n", bToMb(m.TotalAlloc))
	fmt.Printf("Sys = %v MiB\n", bToMb(m.Sys))
	fmt.Printf("NumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 { return b / 1024 / 1024 }

func doCPUInfo() {
	// Try /proc/cpuinfo
	if data, err := ioutil.ReadFile("/proc/cpuinfo"); err == nil {
		fmt.Print(string(data))
		return
	}
	fmt.Printf("CPUs: %d\n", runtime.NumCPU())
}

func doUptime() {
	if data, err := ioutil.ReadFile("/proc/uptime"); err == nil {
		parts := strings.Fields(string(data))
		if len(parts) > 0 {
			if f, err := strconv.ParseFloat(parts[0], 64); err == nil {
				d := time.Duration(f) * time.Second
				fmt.Printf("Uptime: %s\n", d.String())
				return
			}
		}
	}
	// fallback to process start time
	d := time.Since(bootTime)
	fmt.Printf("Uptime (simulated): %s\n", d.String())
}

func doDmesg(args []string) {
	if len(dmesgBuf) == 0 {
		fmt.Println("(no kernel messages)")
		return
	}
	for _, l := range dmesgBuf {
		fmt.Println(l)
	}
}

func doLs(args []string) {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}
	ents, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println("ls: error:", err)
		return
	}
	for _, e := range ents {
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		fmt.Println(name)
	}
}

func doCat(args []string) {
	if len(args) == 0 {
		fmt.Println("cat: missing file")
		return
	}
	for _, p := range args {
		data, err := ioutil.ReadFile(p)
		if err != nil {
			fmt.Printf("cat: %s: %v\n", p, err)
			continue
		}
		fmt.Print(string(data))
	}
}

func doIfconfig() {
	ifs, err := net.Interfaces()
	if err != nil {
		fmt.Println("ifconfig: error listing interfaces:", err)
		return
	}
	for _, it := range ifs {
		addrs, _ := it.Addrs()
		fmt.Printf("%s: flags=%s\n", it.Name, it.Flags.String())
		for _, a := range addrs {
			fmt.Printf("    %s\n", a.String())
		}
	}
}

func doClear() {
	// ANSI clear screen
	fmt.Print("\033[H\033[2J")
}
