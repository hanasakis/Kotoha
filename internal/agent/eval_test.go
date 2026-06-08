package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type evalScenario struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	UserMessage    string            `json:"user_message"`
	ExpectedTool   string            `json:"expected_tool"`
	ExpectedToolArgs map[string]interface{} `json:"expected_tool_args"`
	ExpectedNoTool bool              `json:"expected_no_tool"`
	Evaluation     struct {
		IntentMatch string `json:"intent_match"`
		Language    string `json:"language"`
		Difficulty  string `json:"difficulty"`
		FollowUp    *struct {
			UserMessage string `json:"user_message"`
			ExpectedTool string `json:"expected_tool"`
		} `json:"follow_up"`
		Notes string `json:"notes"`
	} `json:"evaluation"`
}

type evalFile struct {
	Scenarios []evalScenario `json:"scenarios"`
}

func loadEvalScenarios(t *testing.T) []evalScenario {
	t.Helper()

	// Try multiple paths: relative from project root and from test working dir
	candidates := []string{
		"evaluation/agent_evals.json",
		"../../evaluation/agent_evals.json",
		filepath.Join("..", "..", "evaluation", "agent_evals.json"),
	}

	var path string
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			path = c
			break
		}
	}
	if path == "" {
		t.Skip("agent_evals.json not found in any expected location")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read eval file: %v", err)
	}

	var ef evalFile
	if err := json.Unmarshal(data, &ef); err != nil {
		t.Fatalf("failed to parse eval file: %v", err)
	}
	return ef.Scenarios
}

func TestEvalAgent_ToolSelection(t *testing.T) {
	scenarios := loadEvalScenarios(t)
	if len(scenarios) == 0 {
		t.Fatal("no scenarios loaded")
	}

	results := make(map[string]string) // id -> "pass" or failure reason
	passed := 0
	failed := 0

	for _, sc := range scenarios {
		t.Run(sc.ID+"_"+sc.Name, func(t *testing.T) {
			results[sc.ID] = "pass"

			// Chat detection scenarios
			if sc.ExpectedNoTool {
				tool, _, ok := classifyIntent(sc.UserMessage)
				if ok {
					t.Errorf("[%s] %s: expected no tool (chat), got tool=%q",
						sc.ID, sc.Name, tool)
					results[sc.ID] = "wrong tool: " + tool
				}
				return
			}

			// Tool detection scenarios
			tool, args, ok := classifyIntent(sc.UserMessage)
			if !ok {
				t.Errorf("[%s] %s: expected tool %q, got chat",
					sc.ID, sc.Name, sc.ExpectedTool)
				results[sc.ID] = "got chat"
				return
			}

			if tool != sc.ExpectedTool {
				t.Errorf("[%s] %s: tool mismatch — got %q, want %q",
					sc.ID, sc.Name, tool, sc.ExpectedTool)
				results[sc.ID] = "wrong tool: " + tool
				return
			}

			// Check tool args if specified — verify key exists and value is non-empty
			if sc.ExpectedToolArgs != nil && len(sc.ExpectedToolArgs) > 0 {
				var gotArgs map[string]interface{}
				if err := json.Unmarshal([]byte(args), &gotArgs); err != nil {
					t.Errorf("[%s] %s: failed to parse args %q: %v", sc.ID, sc.Name, args, err)
					results[sc.ID] = "args parse error"
					return
				}
				for k := range sc.ExpectedToolArgs {
					gotVal, ok := gotArgs[k]
					if !ok {
						t.Errorf("[%s] %s: args missing key %q, got %v",
							sc.ID, sc.Name, k, gotArgs)
						results[sc.ID] = "missing arg " + k
						return
					}
					gotStr := formatVal(gotVal)
					if gotStr == "" || gotStr == "null" || gotStr == `""` {
						t.Errorf("[%s] %s: arg %q is empty", sc.ID, sc.Name, k)
						results[sc.ID] = "empty arg " + k
						return
					}
				}
			}
		})
	}

	// Summary
	for _, v := range results {
		if v == "pass" {
			passed++
		} else {
			failed++
		}
	}

	t.Logf("=== Evaluation Summary ===")
	t.Logf("Total: %d | Passed: %d | Failed: %d",
		len(scenarios), passed, failed)
	if passed+failed > 0 {
		t.Logf("Tool Selection Accuracy: %d/%d (%.1f%%)",
			passed, passed+failed, float64(passed)/float64(passed+failed)*100)
	}
}

func formatVal(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
