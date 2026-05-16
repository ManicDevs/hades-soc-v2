package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) <= 1 {
		usage()
		return
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	switch cmd {
	case "ls":
		doLs(args)
	case "cat":
		doCat(args)
	case "hostname":
		doHostname()
	case "echo":
		doEcho(args)
	case "sleep":
		doSleep(args)
	case "whoami":
		doWhoami()
	case "ps":
		doPs()
	case "ifconfig":
		doIfconfig()
	case "uptime":
		doUptime()
	default:
		fmt.Fprintf(os.Stderr, "gobox: unknown applet '%s'\n", cmd)
		usage()
	}
}

func usage() {
	fmt.Println("gobox - minimal Go multi-call utility")
	fmt.Println("Usage: gobox <ls|cat|hostname|echo|sleep|whoami|ps|ifconfig|uptime> ...")
}

func doLs(args []string) {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}
	ents, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ls: %v\n", err)
		os.Exit(2)
	}
	names := []string{}
	for _, e := range ents {
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		names = append(names, name)
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Println(n)
	}
}

func doCat(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "cat: missing file")
		os.Exit(2)
	}
	for _, p := range args {
		b, err := ioutil.ReadFile(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cat: %s: %v\n", p, err)
			continue
		}
		os.Stdout.Write(b)
	}
}

func doHostname() {
	h, err := os.Hostname()
	if err != nil {
		fmt.Println("unknown")
	} else {
		fmt.Println(h)
	}
}

func doEcho(args []string) { fmt.Println(strings.Join(args, " ")) }

func doSleep(args []string) {
	if len(args) == 0 {
		time.Sleep(1 * time.Second)
		return
	}
	d, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, "sleep: invalid")
		os.Exit(2)
	}
	time.Sleep(time.Duration(d) * time.Second)
}

func doWhoami() {
	u := os.Getenv("USER")
	if u == "" {
		u = "root"
	}
	fmt.Println(u)
}

func doPs() {
	ents, err := ioutil.ReadDir("/proc")
	if err != nil {
		fmt.Println("ps: /proc unavailable")
		return
	}
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
	for _, pid := range pids {
		cmdlinePath := filepath.Join("/proc", strconv.Itoa(pid), "cmdline")
		if b, err := ioutil.ReadFile(cmdlinePath); err == nil {
			cmd := strings.ReplaceAll(string(b), "\x00", " ")
			if cmd == "" {
				if c, err := ioutil.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "comm")); err == nil {
					cmd = strings.TrimSpace(string(c))
				}
			}
			fmt.Printf("%d %s\n", pid, cmd)
		}
	}
}

func doIfconfig() {
	ifs, err := net.Interfaces()
	if err != nil {
		fmt.Println("ifconfig: error listing interfaces")
		return
	}
	for _, it := range ifs {
		addrs, _ := it.Addrs()
		fmt.Printf("%s:\n", it.Name)
		for _, a := range addrs {
			fmt.Printf("  %s\n", a.String())
		}
	}
}

func doUptime() {
	if b, err := ioutil.ReadFile("/proc/uptime"); err == nil {
		parts := strings.Fields(string(b))
		if len(parts) > 0 {
			fmt.Println(parts[0])
			return
		}
	}
	fmt.Println("0.00")
}
