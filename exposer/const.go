package exposer

type EdgeFlowType string

const (
	EdgeFlowTypeExpose EdgeFlowType = "expose"
	EdgeFlowTypeAccess EdgeFlowType = "access"
)

const (
	EdgeFlowTypeHeaderKey  = "X-Edge-Flow-Type"
	EdgeDeviceIDHeaderKey  = "X-Edge-Device-ID"
	EdgeServiceIDHeaderKey = "X-Edge-Service-ID"
)
