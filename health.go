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
	Description string
}

type HealthCheck interface {
	GetMeta() HealthCheckMeta
	ExecuteCheck(map[string][]string) Status
}

type HealthzMetaResponse struct {
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Identifier string `json:"identifier"`
}

var healthChecks []HealthCheck

func RegisterHealthCheck(check HealthCheck) {
	healthChecks = append(healthChecks, check)
}

func Setup(e *gin.Engine) {
	e.GET("/healthz", healthz)
	e.GET("/healthz/meta", healthzMeta)
}

func healthz(c *gin.Context) {
	var worst Status = UP

	var response map[string]interface{}

	for _, item := range healthChecks {
		meta := item.GetMeta()
		result := item.ExecuteCheck(c.Request.URL.Query())

		if meta.Fatal && result > worst {
			worst = result
		}

		response[meta.Identifier] = map[string]string {
			"status": statusToString(result),
		}
	}

	response["status"] = statusToString(worst)

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

func healthzMeta(c *gin.Context) {

	var result []HealthzMetaResponse

	for _, item := range healthChecks {
		meta := item.GetMeta()
		result = append(result, HealthzMetaResponse{
			Description: meta.Description,
			DisplayName: meta.DisplayName,
			Identifier: meta.Identifier,
		})
	}

	c.JSON(200, result)
}