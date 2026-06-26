package output

import (
	"encoding/json"
	"os"
	"os/exec"
	"testing"
)

func TestExitAlwaysReturnsZeroForZabbix(t *testing.T) {
	states := []int{StateOK, StateWarning, StateCritical, StateUnknown}

	for _, state := range states {
		state := state
		t.Run(StateName(state), func(t *testing.T) {
			t.Parallel()

			if os.Getenv("CHECK_DOMAIN_EXIT_TEST") == "1" {
				Exit(state, 10, "2026-12-31", "test message")
				return
			}

			cmd := exec.Command(os.Args[0], "-test.run=TestExitAlwaysReturnsZeroForZabbix/"+StateName(state))
			cmd.Env = append(os.Environ(), "CHECK_DOMAIN_EXIT_TEST=1")
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("expected exit code 0, got error: %v, output: %s", err, output)
			}

			var result Result
			if err := json.Unmarshal(output, &result); err != nil {
				t.Fatalf("invalid JSON output: %v", err)
			}
			if result.State != StateName(state) {
				t.Fatalf("state = %q, want %q", result.State, StateName(state))
			}
		})
	}
}