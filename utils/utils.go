package utils

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
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
