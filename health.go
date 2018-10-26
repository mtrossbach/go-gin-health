package health

import "github.com/gin-gonic/gin"

const (
	UP Status = 0
	UNKNOWN Status = 1
	SLOW Status = 2
	PARTIAL Status = 3
	DOWN Status = 4
)

type Status int

const (
	Liveness ProbeType = "liveness"
	Readiness ProbeType = "readiness"
)

type ProbeType string

type HealthCheckMeta struct {
	Fatal bool
	Identifier string
	DisplayName string
}

type HealthCheck interface {
	GetMeta() HealthCheckMeta
	SupportsProbeType(ProbeType) bool
	ExecuteCheck(map[string][]string, ProbeType) (Status, *string)
}

type HealthCheckManager struct {
	displayName string
	healthChecks []HealthCheck
	shutdown bool
}

func NewHealthCheckManager(displayName string) *HealthCheckManager {
	return &HealthCheckManager{
		displayName: displayName,
		shutdown: false,
	}
}

func (h *HealthCheckManager) Register(check HealthCheck) {
	h.healthChecks = append(h.healthChecks, check)
}

func (h *HealthCheckManager)SetupWithGinAndPrefix(prefix string, e *gin.Engine) {
	e.GET(prefix + "/healthz", h.liveness)
	e.GET(prefix + "/healthz/", h.liveness)
	e.GET(prefix + "/readyz", h.readiness)
	e.GET(prefix + "/readyz/", h.readiness)
}

func (h *HealthCheckManager)SetupWithGin(e *gin.Engine) {
	e.GET("/healthz", h.liveness)
	e.GET("/healthz/", h.liveness)
	e.GET("/readyz", h.readiness)
	e.GET("/readyz/", h.readiness)
}

func (h *HealthCheckManager)Shutdown() {
	h.shutdown = true
}

func (h *HealthCheckManager)readiness(c *gin.Context) {
	if h.shutdown {
		response := make(map[string]interface{})
		response["status"] = statusToString(DOWN)
		response["_displayName"] = h.displayName
		response["_message"] = "Shutting down"
		c.JSON(503, response)
	} else {
		h.healthz(c, Readiness)
	}
}

func (h *HealthCheckManager)liveness(c *gin.Context) {
	h.healthz(c, Liveness)
}

func (h *HealthCheckManager)healthz(c *gin.Context, probeType ProbeType) {
	var worst = UP

	response := make(map[string]interface{})

	for _, item := range h.healthChecks {
		if !item.SupportsProbeType(probeType) {
			continue
		}

		meta := item.GetMeta()
		result, message := item.ExecuteCheck(c.Request.URL.Query(), probeType)

		if meta.Fatal && result > worst {
			worst = result
		}

		if message != nil && len(*message) > 0 {
			response[meta.Identifier] = map[string]string {
				"status": statusToString(result),
				"_message": *message,
				"_displayName": meta.DisplayName,
			}
		} else {
			response[meta.Identifier] = map[string]string {
				"status": statusToString(result),
				"_displayName": meta.DisplayName,
			}
		}
	}

	response["status"] = statusToString(worst)
	response["_displayName"] = h.displayName

	if worst < 2 {
		c.JSON(200, response)
	} else if worst < 3 {
		c.JSON(207, response)
	} else {
		c.JSON(503, response)
	}
}

func statusToString(status Status) string {
	if status == UP {
		return "UP"
	} else if status == SLOW {
		return "SLOW"
	} else if status == PARTIAL {
		return "PARTIAL"
	} else if status == DOWN {
		return "DOWN"
	} else {
		return "UNKNOWN"
	}
}