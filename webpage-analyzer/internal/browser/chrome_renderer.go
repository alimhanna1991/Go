package browser

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"
)

// ChromeRenderer renders DOM content via headless Chrome.
type ChromeRenderer struct {
	commandPath string
	timeout     time.Duration
}

// NewChromeRenderer creates a renderer if Chrome is available on the host.
func NewChromeRenderer() *ChromeRenderer {
	commandPath, err := exec.LookPath("google-chrome")
	if err != nil {
		return nil
	}

	return &ChromeRenderer{
		commandPath: commandPath,
		timeout:     15 * time.Second,
	}
}

// RenderHTML returns the post-JavaScript DOM for a page.
func (r *ChromeRenderer) RenderHTML(url string) (string, error) {
	if r == nil || r.commandPath == "" {
		return "", errors.New("chrome renderer unavailable")
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		r.commandPath,
		"--headless",
		"--disable-gpu",
		"--no-sandbox",
		"--no-first-run",
		"--disable-dev-shm-usage",
		"--window-size=1440,2200",
		"--virtual-time-budget=12000",
		"--dump-dom",
		url,
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return "", errors.New(strings.TrimSpace(stderr.String()))
		}
		return "", err
	}

	return stdout.String(), nil
}
