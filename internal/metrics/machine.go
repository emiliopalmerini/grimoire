package metrics

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var machineID string

func GetMachineID() string {
	if machineID != "" {
		return machineID
	}

	id, err := getMachineID()
	if err != nil {
		id = generateFallbackID()
	}
	machineID = id
	return machineID
}

func getMachineID() (string, error) {
	switch runtime.GOOS {
	case "linux":
		data, err := os.ReadFile("/etc/machine-id")
		if err == nil {
			return hashID(strings.TrimSpace(string(data))), nil
		}
		data, err = os.ReadFile("/var/lib/dbus/machine-id")
		if err == nil {
			return hashID(strings.TrimSpace(string(data))), nil
		}
		return "", err

	case "darwin":
		cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
		output, err := cmd.Output()
		if err != nil {
			return "", err
		}
		for _, line := range strings.Split(string(output), "\n") {
			if strings.Contains(line, "IOPlatformUUID") {
				parts := strings.Split(line, "=")
				if len(parts) == 2 {
					uuid := strings.Trim(strings.TrimSpace(parts[1]), "\"")
					return hashID(uuid), nil
				}
			}
		}
		return "", os.ErrNotExist

	default:
		return "", os.ErrNotExist
	}
}

func generateFallbackID() string {
	hostname, _ := os.Hostname()
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	return hashID(user + "@" + hostname)
}

func hashID(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:8])
}
