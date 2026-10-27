package config

import "strconv"

func setWorkCapacityConfigValue(config *WorkConfig, key, value string) bool {
	var field *int

	switch key {
	case "max_active_workers":
		field = &config.MaxActiveWorkers
	case "ready_queue_capacity":
		field = &config.ReadyQueueCapacity
	default:
		return false
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		parsed = -1
	}

	*field = parsed

	return true
}
