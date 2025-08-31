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
)

type port struct {
	os, arch string
}

var ports = [...]port{
	{"linux", "amd64"},
	{"linux", "arm64"},
	{"darwin", "arm64"},
	{"windows", "amd64"},
}

func main() {
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
	err = os.MkdirAll("bin", os.ModeDir)
	if err != nil {
		log.Fatalln(err)
	}
	var wg sync.WaitGroup
	for _, p := range ports {
		go func(p port) {
			file := filepath.Join("bin", fmt.Sprintf("minefetch_%s_%s_%s", version, p.os, p.arch))
			if p.os == "windows" {
				file += ".exe"
			}
			log.Printf("Building %s...", file)
			cmd := exec.Command("go", "build", "-o="+file, "-ldflags=-s")
			cmd.Env = append(os.Environ(), "GOOS="+p.os, "GOARCH="+p.arch)
			err := cmd.Run()
			if err != nil {
				log.Fatalln(err)
			}
			wg.Done()
		}(p)
	}
	wg.Add(len(ports))
	wg.Wait()
}
