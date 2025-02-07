package leaderElection

import "testing"

func TestClusterMembersAdd(t *testing.T) {
	cm := newClusterMembers()
	cm.addMember("member1")
	cm.addMember("member0")

	result := cm.getMembers()
	if result[0] != "member0" || result[1] != "member1" {
		t.Fatalf("Expected [member0 member1], got %v", result)
	}
}
