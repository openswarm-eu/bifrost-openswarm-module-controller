package leaderElection

import (
	"slices"
	"strings"
	"sync"
)

type clusterMembers struct {
	members []string
	mu      sync.Mutex
}

func newClusterMembers() *clusterMembers {
	return &clusterMembers{members: make([]string, 0)}
}

func (m *clusterMembers) addMember(id string) {
	defer m.mu.Unlock()
	m.mu.Lock()

	m.members = append(m.members, id)

	slices.SortFunc(m.members, func(a string, b string) int {
		return strings.Compare(a, b)
	})
}

func (m *clusterMembers) removeMember(id string) {
	defer m.mu.Unlock()
	m.mu.Lock()

	for i, v := range m.members {
		if v == id {
			m.members = append(m.members[:i], m.members[i+1:]...)
			return
		}
	}
}

func (m *clusterMembers) getMembers() []string {
	return m.members
}
