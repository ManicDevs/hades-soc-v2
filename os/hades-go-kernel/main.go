package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var dmesgBuf []string
var bootTime time.Time
var supervisor *Supervisor

func main() {
	bootTime = time.Now()
	bootSequence()

	// Initialize supervisor and optionally auto-start services
	supervisor = NewSupervisor()
	if os.Getenv("HADES_AUTO_START_SERVICES") != "0" {
		// Attempt to find hades-server in repo tree and start it
		if path, err := findBinary("hades-server"); err == nil {
			svc := &Service{
				Name:         "hades",
				Path:         path,
				Args:         []string{},
				AutoRestart:  true,
				RestartDelay: 5 * time.Second,
			}
			supervisor.AddService(svc)
			if err := supervisor.StartService("hades"); err != nil {
				appendDmesg("failed to auto-start hades: " + err.Error())
			} else {
				appendDmesg("auto-started hades: " + path)
			}
		} else {
			appendDmesg("hades-server not found for auto-start: " + err.Error())
		}
	}

	startShell()
}

// Service supervision (very small supervisor)

type Service struct {
	Name         string
	Path         string
	Args         []string
	Cmd          *exec.Cmd
	AutoRestart  bool
	RestartDelay time.Duration
	RestartCount int
	StartedAt    time.Time
	lock         sync.Mutex
}

type Supervisor struct {
	services map[string]*Service
	mu       sync.Mutex
}

func NewSupervisor() *Supervisor {
	return &Supervisor{services: make(map[string]*Service)}
}

func (s *Supervisor) AddService(svc *Service) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services[svc.Name] = svc
}

func (s *Supervisor) StartService(name string) error {
	s.mu.Lock()
	svc, ok := s.services[name]
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("service not found: %s", name)
	}

	svc.lock.Lock()
	defer svc.lock.Unlock()

	if svc.Cmd != nil && svc.Cmd.Process != nil {
		return fmt.Errorf("service %s already running (pid=%d)", name, svc.Cmd.Process.Pid)
	}

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, svc.Path, svc.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", name, err)
	}

	svc.Cmd = cmd
	svc.StartedAt = time.Now()
	go s.monitorService(svc)
	return nil
}

func (s *Supervisor) monitorService(svc *Service) {
	// Wait for the process to exit and optionally restart
	err := svc.Cmd.Wait()
	exitTime := time.Now()

	svc.lock.Lock()
	svc.Cmd = nil
	svc.RestartCount++
	svc.lock.Unlock()

	note := fmt.Sprintf("service %s exited at %s (err=%v)", svc.Name, exitTime.Format(time.RFC3339), err)
	appendDmesg(note)

	if svc.AutoRestart {
		// Simple backoff strategy
		delay := svc.RestartDelay
		if svc.RestartCount > 3 {
			delay = delay * 2
		}
		appendDmesg(fmt.Sprintf("restarting service %s in %s", svc.Name, delay))
		time.Sleep(delay)
		if err := s.StartService(svc.Name); err != nil {
			appendDmesg(fmt.Sprintf("failed to restart %s: %v", svc.Name, err))
		}
	}
}

func (s *Supervisor) StopService(name string) error {
	s.mu.Lock()
	svc, ok := s.services[name]
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("service not found: %s", name)
	}

	svc.lock.Lock()
	defer svc.lock.Unlock()

	if svc.Cmd == nil || svc.Cmd.Process == nil {
		return fmt.Errorf("service %s not running", name)
	}

	if err := svc.Cmd.Process.Signal(os.Interrupt); err != nil {
		_ = svc.Cmd.Process.Kill()
	}
	go func() { _ = svc.Cmd.Wait(); svc.lock.Lock(); svc.Cmd = nil; svc.lock.Unlock() }()
	return nil
}

func (s *Supervisor) Status(name string) string {
	s.mu.Lock()
	svc, ok := s.services[name]
	s.mu.Unlock()
	if !ok {
		return "not-found"
	}
	svc.lock.Lock()
	defer svc.lock.Unlock()
	if svc.Cmd != nil && svc.Cmd.Process != nil {
		pid := svc.Cmd.Process.Pid
		upt := time.Since(svc.StartedAt)
		return fmt.Sprintf("running pid=%d uptime=%s", pid, upt.String())
	}
	return "stopped"
}

// findBinary searches upward from common start points for an executable named 'name'
func findBinary(name string) (string, error) {
	starts := []string{}
	if wd, err := os.Getwd(); err == nil {
		starts = append(starts, wd)
	}
	if exe, err := os.Executable(); err == nil {
		starts = append(starts, filepath.Dir(exe))
	}
	for _, start := range starts {
		dir := start
		for i := 0; i < 8; i++ {
			cand := filepath.Join(dir, name)
			if info, err := os.Stat(cand); err == nil && !info.IsDir() {
				if info.Mode()&0111 != 0 {
					abs, _ := filepath.Abs(cand)
					return abs, nil
				}
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	if p, err := exec.LookPath(name); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("executable %s not found", name)
}

func bootSequence() {
	banner := []string{
		"HADES V2 — Pure Go Kernel",
		"Initializing Go runtime subsystems...",
		"Probing CPU…",
		"Initializing memory manager...",
		"Bringing up virtual network interfaces...",
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
		case "service":
			if len(args) == 0 {
				fmt.Println("usage: service <start|stop|status> <name>")
				break
			}
			action := args[0]
			svcName := ""
			if len(args) > 1 {
				svcName = args[1]
			}
			switch action {
			case "start":
				if svcName == "" {
					fmt.Println("specify service name")
					break
				}
				if err := supervisor.StartService(svcName); err != nil {
					fmt.Println("error:", err)
				} else {
					fmt.Println("started", svcName)
				}
			case "stop":
				if svcName == "" {
					fmt.Println("specify service name")
					break
				}
				if err := supervisor.StopService(svcName); err != nil {
					fmt.Println("error:", err)
				} else {
					fmt.Println("stopped", svcName)
				}
			case "status":
				if svcName == "" {
					fmt.Println("services:")
					for n := range supervisor.services {
						fmt.Println(" -", n, ":", supervisor.Status(n))
					}
					break
				}
				fmt.Println(svcName, ":", supervisor.Status(svcName))
			default:
				fmt.Println("unknown service action", action)
			}
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
	fmt.Println("  service     - manage services: service <start|stop|status> <name>")
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
