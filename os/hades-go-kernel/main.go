package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
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

// Lightweight pure-Go userspace kernel simulation with a tiny service supervisor
// and health monitoring. Designed for rapid development without QEMU or full
// kernel rebuilds.

var dmesgBuf []string
var bootTime time.Time
var supervisor *Supervisor

func main() {
	bootTime = time.Now()
	bootSequence()
	startShell()
}

// HealthStatus holds the last health check result for a service.
type HealthStatus struct {
	LastChecked time.Time `json:"last_checked"`
	StatusCode  int       `json:"status_code,omitempty"`
	Status      string    `json:"status"`
	Err         string    `json:"error,omitempty"`
}

// Service struct describes an external process we can supervise.
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

// Supervisor holds registered services and recent health statuses.
type Supervisor struct {
	services map[string]*Service
	health   map[string]HealthStatus
	mu       sync.Mutex
}

func NewSupervisor() *Supervisor {
	return &Supervisor{services: make(map[string]*Service), health: make(map[string]HealthStatus)}
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

func (s *Supervisor) SetHealth(name string, hs HealthStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.health[name] = hs
}

func (s *Supervisor) GetHealth(name string) (HealthStatus, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	hs, ok := s.health[name]
	return hs, ok
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

// ServiceConfig is the JSON manifest entry for a supervised service.
type ServiceConfig struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Args         []string `json:"args,omitempty"`
	AutoStart    bool     `json:"auto_start,omitempty"`
	AutoRestart  bool     `json:"auto_restart,omitempty"`
	RestartDelay int      `json:"restart_delay,omitempty"` // seconds
	HealthURL    string   `json:"health_url,omitempty"`
}

func loadServicesFromCandidates(cands []string) error {
	for _, p := range cands {
		if _, err := os.Stat(p); err == nil {
			b, err := ioutil.ReadFile(p)
			if err != nil {
				return err
			}
			var list []ServiceConfig
			if err := json.Unmarshal(b, &list); err != nil {
				return err
			}
			for _, sc := range list {
				svc := &Service{
					Name:         sc.Name,
					Path:         sc.Path,
					Args:         sc.Args,
					AutoRestart:  sc.AutoRestart,
					RestartDelay: time.Duration(sc.RestartDelay) * time.Second,
				}
				supervisor.AddService(svc)
				if sc.HealthURL != "" {
					// start health monitor for this service (does not start the service itself)
					go func(name, url string, interval time.Duration) {
						startHealthMonitor(name, url, interval)
					}(sc.Name, sc.HealthURL, 15*time.Second)
				}
				if sc.AutoStart && os.Getenv("HADES_AUTO_START_SERVICES") != "0" {
					// attempt to start service now
					if err := supervisor.StartService(sc.Name); err != nil {
						appendDmesg(fmt.Sprintf("failed to auto-start %s: %v", sc.Name, err))
					} else {
						appendDmesg(fmt.Sprintf("auto-started %s from manifest", sc.Name))
					}
				}
			}
			return nil
		}
	}
	return fmt.Errorf("no services manifest found in candidates")
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

// startHealthMonitor periodically hits an HTTP health endpoint and restarts the service if unhealthy
func startHealthMonitor(svcName, url string, interval time.Duration) {
	client := &http.Client{Timeout: 3 * time.Second}
	for {
		time.Sleep(interval)
		hs := HealthStatus{LastChecked: time.Now()}
		resp, err := client.Get(url)
		if err != nil {
			hs.Status = "FAIL"
			hs.Err = err.Error()
			supervisor.SetHealth(svcName, hs)
			appendDmesg(fmt.Sprintf("healthcheck failed for %s: %v", svcName, err))
			// Restart service
			_ = supervisor.StopService(svcName)
			time.Sleep(2 * time.Second)
			_ = supervisor.StartService(svcName)
			continue
		}
		hs.StatusCode = resp.StatusCode
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			hs.Status = "OK"
		} else {
			hs.Status = "FAIL"
			hs.Err = fmt.Sprintf("http status %d", resp.StatusCode)
		}
		_ = resp.Body.Close()
		supervisor.SetHealth(svcName, hs)
		appendDmesg(fmt.Sprintf("healthcheck %s: %s", svcName, hs.Status))
	}
}

func startShell() {
	// Attempt to load services manifest (optional)
	supervisor = NewSupervisor()
	manifestCandidates := []string{"./services.json", "./config/services.json", "./services.json"}
	_ = loadServicesFromCandidates(manifestCandidates) // best-effort

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
				fmt.Println("usage: service <start|stop|status|health|list> <name>")
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
			case "health":
				if svcName == "" {
					fmt.Println("usage: service health <name>")
					break
				}
				if hs, ok := supervisor.GetHealth(svcName); ok {
					fmt.Printf("%+v\n", hs)
				} else {
					fmt.Println("no health data for", svcName)
				}
			case "list":
				fmt.Println("services:")
				for n := range supervisor.services {
					st := supervisor.Status(n)
					hs, _ := supervisor.GetHealth(n)
					fmt.Printf(" - %s : %s (health=%s)\n", n, st, hs.Status)
				}
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
	fmt.Println("  service     - manage services: service <start|stop|status|health|list> <name>")
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
