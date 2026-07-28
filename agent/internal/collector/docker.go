package collector

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"
)

// ContainerInfo represents a Docker container.
type ContainerInfo struct {
	Name   string `json:"name"`
	Image  string `json:"image"`
	Status string `json:"status"`
	State  string `json:"state"`
}

const dockerSocket = "/var/run/docker.sock"

// CollectDocker lists running Docker containers via the Docker Engine API.
// Returns an empty slice (not an error) if Docker is unavailable.
func CollectDocker() []ContainerInfo {
	if runtime.GOOS != "linux" {
		return nil
	}

	if _, err := os.Stat(dockerSocket); os.IsNotExist(err) {
		return nil
	}

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.DialTimeout("unix", dockerSocket, 5*time.Second)
			},
		},
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get("http://localhost/v1.41/containers/json?all=true")
	if err != nil {
		log.Printf("[docker] failed to query API: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[docker] API returned status %d", resp.StatusCode)
		return nil
	}

	var raw []struct {
		Names  []string `json:"Names"`
		Image  string   `json:"Image"`
		State  string   `json:"State"`
		Status string   `json:"Status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		log.Printf("[docker] failed to decode response: %v", err)
		return nil
	}

	containers := make([]ContainerInfo, 0, len(raw))
	for _, c := range raw {
		name := ""
		if len(c.Names) > 0 {
			name = c.Names[0]
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
		}
		containers = append(containers, ContainerInfo{
			Name:   name,
			Image:  c.Image,
			Status: c.Status,
			State:  c.State,
		})
	}

	return containers
}
