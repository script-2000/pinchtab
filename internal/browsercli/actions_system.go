package browsercli

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func Health(client *http.Client, base, token string) {
	DoGet(client, base, token, "/health", nil)
}

func Instances(client *http.Client, base, token string) {
	body := DoGetRaw(client, base, token, "/instances", nil)

	// Parse and format as JSON
	var instances []map[string]any
	if err := json.Unmarshal(body, &instances); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse instances: %v\n", err)
		os.Exit(1)
	}

	// Transform to cleaner output format
	output := make([]map[string]any, len(instances))
	for i, inst := range instances {
		id, _ := inst["id"].(string)
		port, _ := inst["port"].(string)
		headless, _ := inst["headless"].(bool)
		status, _ := inst["status"].(string)

		mode := "headless"
		if !headless {
			mode = "headed"
		}

		output[i] = map[string]any{
			"id":     id,
			"port":   port,
			"mode":   mode,
			"status": status,
		}
	}

	// Output as JSON
	data, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(data))
}

// --- profiles ---

func Profiles(client *http.Client, base, token string) {
	result := DoGet(client, base, token, "/profiles", nil)

	// Display profiles in a friendly format
	if profiles, ok := result["profiles"].([]interface{}); ok && len(profiles) > 0 {
		fmt.Println()
		for _, prof := range profiles {
			if m, ok := prof.(map[string]any); ok {
				name, _ := m["name"].(string)

				fmt.Printf("👤 %s\n", name)
			}
		}
		fmt.Println()
	} else {
		fmt.Println("No profiles available")
	}
}

// --- internal helpers ---

// getInstances fetches the list of running instances
func getInstances(client *http.Client, base, token string) []map[string]any {
	resp, err := http.NewRequest("GET", base+"/instances", nil)
	if err != nil {
		return nil
	}
	if token != "" {
		resp.Header.Set("Authorization", "Bearer "+token)
	}

	result, err := client.Do(resp)
	if err != nil || result.StatusCode >= 400 {
		return nil
	}
	defer func() { _ = result.Body.Close() }()

	var data map[string]any
	if err := json.NewDecoder(result.Body).Decode(&data); err != nil {
		log.Printf("warning: error decoding instances response: %v", err)
	}

	if instances, ok := data["instances"].([]interface{}); ok {
		converted := make([]map[string]any, len(instances))
		for i, inst := range instances {
			if m, ok := inst.(map[string]any); ok {
				converted[i] = m
			}
		}
		return converted
	}
	return nil
}

// launchInstance launches a default instance
func launchInstance(client *http.Client, base, token string, profile string) {
	body := map[string]any{"profile": profile}
	DoPost(client, base, token, "/instances/launch", body)
}
