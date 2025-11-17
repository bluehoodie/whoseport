package docker

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/bluehoodie/whoseport/internal/model"
)

// Retriever retrieves detailed Docker container information.
type Retriever interface {
	// GetContainerInfo retrieves detailed information about a container
	GetContainerInfo(containerID string, processInfo *model.ProcessInfo) (*ContainerInfo, error)
}

// DefaultRetriever implements container information retrieval.
type DefaultRetriever struct{}

// NewRetriever creates a new Docker container retriever.
func NewRetriever() Retriever {
	return &DefaultRetriever{}
}

// GetContainerInfo retrieves comprehensive container information.
func (r *DefaultRetriever) GetContainerInfo(containerID string, processInfo *model.ProcessInfo) (*ContainerInfo, error) {
	info := &ContainerInfo{
		ID:        containerID,
		ShortID:   containerID,
		ProcessID: processInfo.ID,
		ProcessCmd: processInfo.Command,
	}

	// Truncate to short ID if it's a full ID
	if len(containerID) > 12 {
		info.ShortID = containerID[:12]
	}

	// Get detailed inspect data
	if err := r.populateInspectData(info); err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	// Get real-time stats (non-blocking)
	// We don't fail if stats aren't available
	_ = r.populateStatsData(info)

	return info, nil
}

// populateInspectData uses docker inspect to get container details.
func (r *DefaultRetriever) populateInspectData(info *ContainerInfo) error {
	cmd := exec.Command("docker", "inspect", info.ID)
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	// Parse JSON output
	var inspectData []map[string]interface{}
	if err := json.Unmarshal(output, &inspectData); err != nil {
		return err
	}

	if len(inspectData) == 0 {
		return fmt.Errorf("no inspect data returned")
	}

	data := inspectData[0]

	// Extract basic information
	if name, ok := data["Name"].(string); ok {
		info.Name = strings.TrimPrefix(name, "/")
	}

	// Config section
	if config, ok := data["Config"].(map[string]interface{}); ok {
		if image, ok := config["Image"].(string); ok {
			info.Image = image
		}
		if cmd, ok := config["Cmd"].([]interface{}); ok {
			cmdParts := make([]string, 0, len(cmd))
			for _, part := range cmd {
				if s, ok := part.(string); ok {
					cmdParts = append(cmdParts, s)
				}
			}
			info.Command = strings.Join(cmdParts, " ")
		}
		if labels, ok := config["Labels"].(map[string]interface{}); ok {
			info.Labels = make(map[string]string)
			for k, v := range labels {
				if s, ok := v.(string); ok {
					info.Labels[k] = s
				}
			}
		}
	}

	// State section
	if state, ok := data["State"].(map[string]interface{}); ok {
		if status, ok := state["Status"].(string); ok {
			info.State = status
			info.Status = status
		}
		if startedAt, ok := state["StartedAt"].(string); ok {
			info.CreatedAt = startedAt
			// Calculate running duration
			if t, err := time.Parse(time.RFC3339Nano, startedAt); err == nil {
				duration := time.Since(t)
				info.RunningFor = formatDuration(duration)
			}
		}
	}

	// Image ID
	if imageID, ok := data["Image"].(string); ok {
		info.ImageID = imageID
		if len(imageID) > 19 && strings.HasPrefix(imageID, "sha256:") {
			info.ImageID = imageID[7:19] // Short hash
		}
	}

	// Host config for restart policy
	if hostConfig, ok := data["HostConfig"].(map[string]interface{}); ok {
		if restartPolicy, ok := hostConfig["RestartPolicy"].(map[string]interface{}); ok {
			if name, ok := restartPolicy["Name"].(string); ok {
				info.RestartPolicy = name
			}
		}
	}

	// Network settings
	if netSettings, ok := data["NetworkSettings"].(map[string]interface{}); ok {
		// Port mappings
		if ports, ok := netSettings["Ports"].(map[string]interface{}); ok {
			info.Ports = r.parsePortMappings(ports)
			info.PortString = r.formatPortString(info.Ports)
		}

		// Networks
		if networks, ok := netSettings["Networks"].(map[string]interface{}); ok {
			info.Networks = make([]string, 0, len(networks))
			for netName, netData := range networks {
				info.Networks = append(info.Networks, netName)

				// Get IP from first network
				if info.IPAddress == "" {
					if netMap, ok := netData.(map[string]interface{}); ok {
						if ip, ok := netMap["IPAddress"].(string); ok {
							info.IPAddress = ip
						}
						if gw, ok := netMap["Gateway"].(string); ok {
							info.Gateway = gw
						}
						if mac, ok := netMap["MacAddress"].(string); ok {
							info.MacAddress = mac
						}
					}
				}
			}
		}
	}

	// Mounts
	if mounts, ok := data["Mounts"].([]interface{}); ok {
		info.Mounts = make([]Mount, 0, len(mounts))
		for _, mountData := range mounts {
			if m, ok := mountData.(map[string]interface{}); ok {
				mount := Mount{}
				if t, ok := m["Type"].(string); ok {
					mount.Type = t
				}
				if s, ok := m["Source"].(string); ok {
					mount.Source = s
				}
				if d, ok := m["Destination"].(string); ok {
					mount.Destination = d
				}
				if rw, ok := m["RW"].(bool); ok {
					if rw {
						mount.Mode = "rw"
					} else {
						mount.Mode = "ro"
					}
				}
				info.Mounts = append(info.Mounts, mount)
			}
		}
	}

	// Platform
	if platform, ok := data["Platform"].(string); ok {
		info.Platform = platform
	}

	return nil
}

// populateStatsData uses docker stats to get real-time resource usage.
func (r *DefaultRetriever) populateStatsData(info *ContainerInfo) error {
	// Use --no-stream to get a single snapshot
	cmd := exec.Command("docker", "stats", "--no-stream", "--format",
		"{{.CPUPerc}}|{{.MemUsage}}|{{.MemPerc}}|{{.NetIO}}|{{.BlockIO}}|{{.PIDs}}",
		info.ID)

	output, err := cmd.Output()
	if err != nil {
		return err
	}

	line := strings.TrimSpace(string(output))
	parts := strings.Split(line, "|")

	if len(parts) >= 6 {
		info.CPUPercent = strings.TrimSpace(parts[0])
		info.MemUsage = strings.TrimSpace(parts[1])
		info.MemPercent = strings.TrimSpace(parts[2])
		info.NetIO = strings.TrimSpace(parts[3])
		info.BlockIO = strings.TrimSpace(parts[4])
		info.PIDs = strings.TrimSpace(parts[5])
	}

	return nil
}

// parsePortMappings converts Docker port map to PortMapping structs.
func (r *DefaultRetriever) parsePortMappings(ports map[string]interface{}) []PortMapping {
	mappings := make([]PortMapping, 0)

	for containerPort, bindings := range ports {
		if bindings == nil {
			continue
		}

		bindingsList, ok := bindings.([]interface{})
		if !ok {
			continue
		}

		for _, binding := range bindingsList {
			bindingMap, ok := binding.(map[string]interface{})
			if !ok {
				continue
			}

			mapping := PortMapping{}

			// Parse container port (format: "8080/tcp")
			portParts := strings.Split(containerPort, "/")
			if len(portParts) >= 1 {
				mapping.ContainerPort = portParts[0]
			}
			if len(portParts) >= 2 {
				mapping.Protocol = portParts[1]
			}

			// Parse host binding
			if hostIP, ok := bindingMap["HostIp"].(string); ok {
				mapping.HostIP = hostIP
			}
			if hostPort, ok := bindingMap["HostPort"].(string); ok {
				mapping.HostPort = hostPort
			}

			mappings = append(mappings, mapping)
		}
	}

	return mappings
}

// formatPortString creates a human-readable port mapping string.
func (r *DefaultRetriever) formatPortString(ports []PortMapping) string {
	if len(ports) == 0 {
		return ""
	}

	parts := make([]string, 0, len(ports))
	for _, p := range ports {
		hostPart := p.HostPort
		if p.HostIP != "" && p.HostIP != "0.0.0.0" {
			hostPart = p.HostIP + ":" + p.HostPort
		}

		part := fmt.Sprintf("%s->%s/%s", hostPart, p.ContainerPort, p.Protocol)
		parts = append(parts, part)
	}

	return strings.Join(parts, ", ")
}

// formatDuration formats a duration in human-readable form.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		if minutes > 0 {
			return fmt.Sprintf("%d hours %d minutes", hours, minutes)
		}
		return fmt.Sprintf("%d hours", hours)
	}

	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	if hours > 0 {
		return fmt.Sprintf("%d days %d hours", days, hours)
	}
	return fmt.Sprintf("%d days", days)
}
