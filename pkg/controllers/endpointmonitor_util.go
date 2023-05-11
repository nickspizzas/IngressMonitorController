package controllers

import (
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
	"github.com/stakater/IngressMonitorController/v2/pkg/monitors"
)

func getMonitorById(monitorService monitors.MonitorServiceProxy, uptimeId string) *models.Monitor {

	monitor, _ := monitorService.GetById(uptimeId)
	// Monitor Exists
	if monitor != nil {
		return monitor
	}
	return nil
}

func findMonitorByName(monitorService monitors.MonitorServiceProxy, monitorName string) *models.Monitor {

	monitor, _ := monitorService.GetByName(monitorName)
	// Monitor Exists
	if monitor != nil {
		return monitor
	}
	return nil
}
