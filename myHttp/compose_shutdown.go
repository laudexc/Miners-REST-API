package myHttp

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const dockerSocketPath = "/var/run/docker.sock"

type dockerContainerInspect struct {
	Config struct {
		Labels map[string]string `json:"Labels"`
	} `json:"Config"`
}

type dockerContainerListItem struct {
	ID     string            `json:"Id"`
	State  string            `json:"State"`
	Labels map[string]string `json:"Labels"`
}

func stopComposeSupportContainers(ctx context.Context) error {
	if _, err := os.Stat(dockerSocketPath); err != nil {
		return err
	}

	client := dockerSocketClient()
	project, currentService, err := currentComposeIdentity(ctx, client)
	if err != nil {
		return err
	}
	if project == "" {
		return errors.New("compose project label is not available")
	}

	containers, err := composeContainers(ctx, client, project)
	if err != nil {
		return err
	}

	for _, container := range containers {
		service := container.Labels["com.docker.compose.service"]
		if service == "" || service == currentService {
			continue
		}
		if service != "prometheus" && service != "grafana" {
			continue
		}
		if container.State != "running" {
			continue
		}
		if err := stopDockerContainer(ctx, client, container.ID); err != nil {
			return err
		}
	}

	return nil
}

func dockerSocketClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				var dialer net.Dialer
				return dialer.DialContext(ctx, "unix", dockerSocketPath)
			},
		},
	}
}

func currentComposeIdentity(ctx context.Context, client *http.Client) (string, string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://docker/containers/"+strings.TrimSpace(hostname)+"/json", nil)
	if err != nil {
		return "", "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", errors.New("failed to inspect current container")
	}

	var inspected dockerContainerInspect
	if err := json.NewDecoder(resp.Body).Decode(&inspected); err != nil {
		return "", "", err
	}

	labels := inspected.Config.Labels
	return labels["com.docker.compose.project"], labels["com.docker.compose.service"], nil
}

func composeContainers(ctx context.Context, client *http.Client, project string) ([]dockerContainerListItem, error) {
	filter := map[string][]string{
		"label": {"com.docker.compose.project=" + project},
	}
	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}

	endpoint := "http://docker/containers/json?all=true&filters=" + url.QueryEscape(string(filterJSON))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New("failed to list compose containers")
	}

	var containers []dockerContainerListItem
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, err
	}

	return containers, nil
}

func stopDockerContainer(ctx context.Context, client *http.Client, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://docker/containers/"+id+"/stop?t=10", nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("failed to stop compose support container")
	}

	return nil
}
