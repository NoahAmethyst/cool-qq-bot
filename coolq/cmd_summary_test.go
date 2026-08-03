package coolq

import "testing"

func TestEconomicSummaryCommandRegisteredForGroupAndPrivate(t *testing.T) {
	const command = "经济总结"

	if _, ok := cmdInfo[command]; !ok {
		t.Fatalf("cmdInfo missing %q", command)
	}
	if !stringSliceContains(groupCmdList, command) {
		t.Fatalf("groupCmdList missing %q", command)
	}
	if _, ok := groupCmdHandlers[command]; !ok {
		t.Fatalf("groupCmdHandlers missing %q", command)
	}
	if !stringSliceContains(privateCmdList, command) {
		t.Fatalf("privateCmdList missing %q", command)
	}
	if _, ok := privateCmdHandlers[command]; !ok {
		t.Fatalf("privateCmdHandlers missing %q", command)
	}
}

func stringSliceContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
