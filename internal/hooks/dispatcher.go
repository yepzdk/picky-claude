package hooks

import "fmt"

// Hook is a function that handles a specific hook event.
// It receives the parsed input and is responsible for writing output
// (via WriteOutput, BlockWithError, or ExitOK).
type Hook func(input *Input) error

// registry maps hook names to their implementations.
var registry = map[string]Hook{}

// Register adds a hook implementation to the registry.
func Register(name string, h Hook) {
	registry[name] = h
}

// Dispatch reads hook input from stdin, looks up the named hook, and runs it.
// Returns an error if the hook is not found or fails.
func Dispatch(name string) error {
	h, ok := registry[name]
	if !ok {
		return fmt.Errorf("unknown hook: %s", name)
	}

	input, err := ReadInput()
	if err != nil {
		return fmt.Errorf("reading hook input: %w", err)
	}

	return h(input)
}

// RegisteredHooks returns the names of all registered hooks.
func RegisteredHooks() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
