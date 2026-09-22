package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"tg-sysmon/internal/config"
	"tg-sysmon/internal/monitor"
	"tg-sysmon/internal/telegram"
)

func main() {
	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	bot := telegram.New(cfg.BotToken)

	log.Printf("starting tg-sysmon: disks=%v interval=%s cpu>%.0f%% ram>%.0f%% disk>%.0f%%",
		cfg.Disks, cfg.Interval, cfg.CPUThreshold, cfg.RAMThreshold, cfg.DiskThreshold)

	go commandLoop(bot, cfg)
	monitorLoop(bot, cfg)
}

func monitorLoop(bot *telegram.Bot, cfg *config.Config) {
	firing := map[string]bool{}
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		stats, err := monitor.Collect(cfg.Disks, cfg.DiskRootPrefix)
		if err != nil {
			log.Printf("collect error: %v", err)
		} else {
			checkThreshold(bot, cfg, firing, "cpu", stats.CPUPercent, cfg.CPUThreshold,
				fmt.Sprintf("🔥 *CPU high*: %.1f%% (threshold %.0f%%)", stats.CPUPercent, cfg.CPUThreshold))

			checkThreshold(bot, cfg, firing, "ram", stats.RAMPercent, cfg.RAMThreshold,
				fmt.Sprintf("🧠 *RAM high*: %.1f%% (%.1fGB / %.1fGB, threshold %.0f%%)",
					stats.RAMPercent, stats.RAMUsedGB, stats.RAMTotalGB, cfg.RAMThreshold))

			for _, d := range stats.Disks {
				key := "disk:" + d.Device
				checkThreshold(bot, cfg, firing, key, d.UsedPercent, cfg.DiskThreshold,
					fmt.Sprintf("💾 *Disk high*: %s (%s) %.1f%% (threshold %.0f%%)",
						d.Device, d.Mountpoint, d.UsedPercent, cfg.DiskThreshold))
			}
		}
		<-ticker.C
	}
}

func checkThreshold(bot *telegram.Bot, cfg *config.Config, firing map[string]bool, key string, value, threshold float64, alertText string) {
	if value >= threshold {
		if !firing[key] {
			firing[key] = true
			if err := bot.SendMessage(cfg.ChatID, alertText); err != nil {
				log.Printf("send error: %v", err)
			}
		}
	} else if firing[key] {
		firing[key] = false
		msg := fmt.Sprintf("✅ *Recovered*: %s back to %.1f%%", key, value)
		if err := bot.SendMessage(cfg.ChatID, msg); err != nil {
			log.Printf("send error: %v", err)
		}
	}
}

func commandLoop(bot *telegram.Bot, cfg *config.Config) {
	offset := 0
	for {
		updates, err := bot.GetUpdates(offset)
		if err != nil {
			log.Printf("getUpdates error: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, u := range updates {
			offset = u.UpdateID + 1
			if u.Message == nil {
				continue
			}
			handleCommand(bot, cfg, u.Message.Text, fmt.Sprintf("%d", u.Message.Chat.ID))
		}
	}
}

func handleCommand(bot *telegram.Bot, cfg *config.Config, text, chatID string) {
	switch {
	case strings.HasPrefix(text, "/status"):
		stats, err := monitor.Collect(cfg.Disks, cfg.DiskRootPrefix)
		if err != nil {
			bot.SendMessage(chatID, fmt.Sprintf("error collecting stats: %v", err))
			return
		}
		bot.SendMessage(chatID, formatStatus(stats))
	case strings.HasPrefix(text, "/start"), strings.HasPrefix(text, "/help"):
		bot.SendMessage(chatID, "Commands:\n/status - current CPU, RAM, disk usage\n/help - this message")
	}
}

func formatStatus(s *monitor.Stats) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🖥 *CPU*: %.1f%%\n", s.CPUPercent))
	sb.WriteString(fmt.Sprintf("🧠 *RAM*: %.1f%% (%.1fGB / %.1fGB)\n", s.RAMPercent, s.RAMUsedGB, s.RAMTotalGB))
	for _, d := range s.Disks {
		sb.WriteString(fmt.Sprintf("💾 *%s* (%s): %.1f%%\n", d.Device, d.Mountpoint, d.UsedPercent))
	}
	if len(s.Disks) == 0 {
		sb.WriteString("💾 no matching disks found\n")
	}
	return sb.String()
}
