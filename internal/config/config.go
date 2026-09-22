package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	BotToken       string
	ChatID         string
	CPUThreshold   float64
	RAMThreshold   float64
	DiskThreshold  float64
	Disks          []string
	DiskRootPrefix string
	Interval       time.Duration
}

func Load(path string) (*Config, error) {
	loadDotenv(path)

	cfg := &Config{
		BotToken:       os.Getenv("TELEGRAM_BOT_TOKEN"),
		ChatID:         os.Getenv("TELEGRAM_CHAT_ID"),
		CPUThreshold:   getFloat("CPU_THRESHOLD", 80),
		RAMThreshold:   getFloat("RAM_THRESHOLD", 80),
		DiskThreshold:  getFloat("DISK_THRESHOLD", 80),
		Disks:          getList("DISKS", []string{"sda"}),
		DiskRootPrefix: os.Getenv("DISK_ROOT_PREFIX"),
		Interval:       time.Duration(getInt("CHECK_INTERVAL_SECONDS", 60)) * time.Second,
	}

	if cfg.BotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if cfg.ChatID == "" {
		return nil, fmt.Errorf("TELEGRAM_CHAT_ID is required")
	}

	return cfg, nil
}

func loadDotenv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, val)
		}
	}
}

func getFloat(key string, def float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}

func getInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func getList(key string, def []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return def
	}
	return out
}
