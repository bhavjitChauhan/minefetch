//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type port struct {
	os, arch string
}

var ports = [...]port{
	{"darwin", "arm64"},
	{"linux", "amd64"},
	{"linux", "arm64"},
	{"windows", "amd64"},
}

func main() {
	start := time.Now()
	cmd := exec.Command("git", "describe", "--tags", "--dirty", "--always")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		log.Fatalln(err)
	}
	version := strings.TrimPrefix(strings.TrimSpace(buf.String()), "v")
	if version == "" {
		log.Fatalln("empty version string")
	}
	err = os.RemoveAll("bin")
	if err != nil {
		log.Fatalln(err)
	}
	err = os.Mkdir("bin", 0755)
	if err != nil {
		log.Fatalln(err)
	}
	var wg sync.WaitGroup
	ch := make(chan struct{}, len(ports))
	wg.Add(len(ports))
	for _, p := range ports {
		go func(p port) {
			start := time.Now()
			name := fmt.Sprintf("minefetch_%s_%s", p.os, p.arch)
			if p.os == "windows" {
				name += ".exe"
			}
			file := filepath.Join("bin", name)
			log.Printf("%s/%s: building...", p.os, p.arch)
			cmd := exec.Command("go", "build", "-o", file, "-ldflags", "-s -X main.version="+version)
			cmd.Env = append(os.Environ(), "GOOS="+p.os, "GOARCH="+p.arch)
			err := cmd.Run()
			if err != nil {
				ch <- struct{}{}
				log.Printf("%s/%s: %v", p.os, p.arch, err)
			} else {
				log.Printf("%s/%s: done (%v)", p.os, p.arch, time.Since(start).Round(time.Millisecond))
			}
			wg.Done()
		}(p)
	}
	wg.Wait()
	close(ch)
	if len(ch) > 0 {
		log.Fatalln("failed")
	}
	log.Printf("done (%v)", time.Since(start).Round(time.Millisecond))
}
