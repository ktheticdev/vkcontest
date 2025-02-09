package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/go-ping/ping"
)

type Status struct {
	IP            string    `json:"ip"`
	PingTime      int64     `json:"ping_time"`
	LastSuccessAt time.Time `json:"last_success_at"`
}

func getContainerIPs(ctx context.Context, cli *client.Client) ([]string, error) {
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, container := range containers {
		for _, network := range container.NetworkSettings.Networks {
			if network.IPAddress != "" {
				ips = append(ips, network.IPAddress)
			}
		}
	}
	return ips, nil
}

func pingIP(ip string) (int64, bool) {
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		log.Printf("Ошибка создания пингера для %s: %v", ip, err)
		return 0, false
	}
	pinger.Count = 3
	pinger.Timeout = 3 * time.Second
	if err := pinger.Run(); err != nil {
		log.Printf("Ошибка пинга %s: %v", ip, err)
		return 0, false
	}
	stats := pinger.Statistics()
	return int64(stats.AvgRtt / time.Millisecond), true
}

func sendStatus(backendURL string, status Status) error {
	jsonData, err := json.Marshal(status)
	if err != nil {
		return err
	}
	resp, err := http.Post(backendURL+"/statuses", "application/json", io.NopCloser(bytes.NewReader(jsonData)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func main() {
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		log.Fatal("BACKEND_URL не установлен")
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Ошибка создания Docker клиента: %v", err)
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		ctx := context.Background()
		ips, err := getContainerIPs(ctx, cli)
		if err != nil {
			log.Printf("Ошибка получения IP контейнеров: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		for _, ip := range ips {
			pingTime, ok := pingIP(ip)
			if !ok {
				log.Printf("Пинг для %s не удался", ip)
				continue
			}
			status := Status{
				IP:            ip,
				PingTime:      pingTime,
				LastSuccessAt: time.Now(),
			}
			if err := sendStatus(backendURL, status); err != nil {
				log.Printf("Ошибка отправки данных для %s: %v", ip, err)
			} else {
				log.Printf("Данные успешно отправлены для %s", ip)
			}
		}
		<-ticker.C
	}
}
