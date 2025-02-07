package internal

import "errors"

type subscriptionManager struct {
	subscriptions map[string]chan []byte
}

func newSubscriptionManager() subscriptionManager {
	return subscriptionManager{make(map[string]chan []byte)}
}

func (sm subscriptionManager) add(subscription string, c chan []byte) {
	sm.subscriptions[subscription] = c
}

func (sm subscriptionManager) remove(subscription string) {
	delete(sm.subscriptions, subscription)
}

func (sm subscriptionManager) get(subscription string) (chan []byte, error) {
	c, ok := sm.subscriptions[subscription]
	if !ok {
		return nil, errors.New("subscription not found")
	}
	return c, nil
}
