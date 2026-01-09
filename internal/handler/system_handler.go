package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	systemmetrics "github.com/metalpoch/local-synapse/internal/infrastructure/system_metrics"
)

type systemHandler struct {
}

func NewSystemHandler() *systemHandler {
	return &systemHandler{}
}

func (hdlr *systemHandler) Stats(c echo.Context) error {
	metrics, err := systemmetrics.GetSystemMetrics()
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err})
	}

	return c.JSON(http.StatusOK, metrics)
}
