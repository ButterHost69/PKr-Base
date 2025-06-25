package utils

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/ButterHost69/PKr-Base/logger"
)

func ClearScreen() {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func RandomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[r.Intn(len(letters))]
	}
	return string(s)
}

func PrintProgressBar(progress int, total int, barLength int) {
	percent := float64(progress) / float64(total)
	hashes := int(percent * float64(barLength))
	spaces := barLength - hashes

	fmt.Printf("\r[%s%s] %.2f%%",
		strings.Repeat("#", hashes),
		strings.Repeat(" ", spaces),
		percent*100)
}

// 192.168.65.1 & 192.168.137.1 are used as VM-Bridge, so ignoring those IP's
func isPrivateIPv4(ip net.IP) bool {
	return ip[0] == 10 ||
		(ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31) ||
		(ip[0] == 192 && ip[1] == 168 && ip[2] != 65 && ip[3] != 1) ||
		(ip[0] == 192 && ip[1] == 168 && ip[2] != 137 && ip[3] != 1)
}

func ReturnListOfPrivateIPs() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logger.USER_LOGGER.Println("Error while Retrieving Address of all Network Interfaces:", err)
		logger.USER_LOGGER.Println("Source: ReturnListOfPrivateIPs()")
		return nil, err
	}

	private_ips := []string{}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ip4 := ipNet.IP.To4(); ip4 != nil {
				// Check if it's a private IP
				if isPrivateIPv4(ip4) {
					private_ips = append(private_ips, ip4.String())
				}
			}
		}
	}
	return private_ips, nil
}
