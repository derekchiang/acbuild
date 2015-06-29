package main

import (
	"bufio"
	log "github.com/Sirupsen/logrus"
	"os"
	"os/exec"
)

// SupportsOverlay returns whether the system supports overlay filesystem
func supportsOverlay() bool {
	exec.Command("modprobe", "overlay").Run()

	f, err := os.Open("/proc/filesystems")
	if err != nil {
		log.Errorf("Error opening /proc/filesystems: %s", err)
		return false
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if s.Text() == "nodev\toverlay" {
			return true
		}
	}

	return false
}
