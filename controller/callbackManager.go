package controller

// callbackManager stores callback functions identified by unique string IDs
type callbackManager struct {
	callbacks map[string]func([]byte)
}

// newCallbackManager creates a new callback manager
func newCallbackManager() *callbackManager {
	return &callbackManager{
		callbacks: make(map[string]func([]byte)),
	}
}

// addCallback adds a callback function with the given ID
// If a callback with the ID already exists, it will be overwritten
func (cm *callbackManager) addCallback(id string, callback func([]byte)) {
	cm.callbacks[id] = callback
}

// getCallback retrieves and removes a callback by its ID
// Returns the callback function and a boolean indicating if it was found
func (cm *callbackManager) getCallback(id string) (func([]byte), bool) {
	callback, exists := cm.callbacks[id]
	if exists {
		delete(cm.callbacks, id)
	}
	return callback, exists
}
