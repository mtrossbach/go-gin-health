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

type HealthCheckMeta struct {
	Fatal bool
	Identifier string
	DisplayName string
}

type HealthCheck interface {
	GetMeta() HealthCheckMeta
	ExecuteCheck(map[string][]string) (Status, *string)
}

type HealthCheckManager struct {
	displayName string
	healthChecks []HealthCheck
}

func NewHealthCheckManager(displayName string) *HealthCheckManager {
	return &HealthCheckManager{
		displayName: displayName,
	}
}

func (h *HealthCheckManager) Register(check HealthCheck) {
	h.healthChecks = append(h.healthChecks, check)
}

func (h *HealthCheckManager)SetupWithGinAndPrefix(prefix string, e *gin.Engine) {
	e.GET(prefix + "/healthz", h.healthz)
	e.GET(prefix + "/healthz/", h.healthz)
}

func (h *HealthCheckManager)SetupWithGin(e *gin.Engine) {
	e.GET("/healthz", h.healthz)
	e.GET("/healthz/", h.healthz)
}

func (h *HealthCheckManager)healthz(c *gin.Context) {
	var worst Status = UP

	response := make(map[string]interface{})

	for _, item := range h.healthChecks {
		meta := item.GetMeta()
		result, message := item.ExecuteCheck(c.Request.URL.Query())

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