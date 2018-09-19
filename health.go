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
	ExecuteCheck(map[string][]string) (Status, *string)
}

type HealthzMetaComponent struct {
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Identifier string `json:"identifier"`
}

type HealthzMetaResponse struct {
	DisplayName string `json:"displayName"`
	Identifier string `json:"identifier"`
	Components []HealthzMetaComponent `json:"components"`
}

var healthChecks []HealthCheck
var healthDisplayName string
var healthIdentifier string

func RegisterHealthCheck(check HealthCheck) {
	healthChecks = append(healthChecks, check)
}

func Setup(identifier string, displayName string, e *gin.Engine) {
	healthIdentifier = identifier
	healthDisplayName = displayName
	e.GET("/healthz", healthz)
	e.GET("/healthz/meta", healthzMeta)
}

func healthz(c *gin.Context) {
	var worst Status = UP

	var response map[string]interface{}

	for _, item := range healthChecks {
		meta := item.GetMeta()
		result, message := item.ExecuteCheck(c.Request.URL.Query())

		if meta.Fatal && result > worst {
			worst = result
		}

		if message != nil && len(*message) > 0 {
			response[meta.Identifier] = map[string]string {
				"status": statusToString(result),
				"message": *message,
			}
		} else {
			response[meta.Identifier] = map[string]string {
				"status": statusToString(result),
			}
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

	var result []HealthzMetaComponent

	for _, item := range healthChecks {
		meta := item.GetMeta()
		result = append(result, HealthzMetaComponent{
			Description: meta.Description,
			DisplayName: meta.DisplayName,
			Identifier: meta.Identifier,
		})
	}

	c.JSON(200, HealthzMetaResponse{
		Identifier: healthIdentifier,
		DisplayName: healthDisplayName,
		Components: result,
	})
}