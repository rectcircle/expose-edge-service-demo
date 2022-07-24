package helper

func RouteKey(serviceID, deviceID string) string {
	return "exposer-route-table:" + serviceID + ":" + deviceID
}
