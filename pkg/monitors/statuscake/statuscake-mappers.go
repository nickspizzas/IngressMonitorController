package statuscake

import (
	"strings"

	statuscake "github.com/StatusCakeDev/statuscake-go"
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
)

// StatusCakeMonitorMonitorToBaseMonitorMapper function to map Statuscake structure to Monitor
func StatusCakeMonitorMonitorToBaseMonitorMapper(statuscakeData statuscake.UptimeTest) *models.Monitor {
	var m models.Monitor
	m.Name = statuscakeData.Name
	m.URL = statuscakeData.WebsiteURL
	m.ID = statuscakeData.ID

	var providerConfig endpointmonitorv1alpha1.StatusCakeConfig
	providerConfig.CheckRate = int(statuscakeData.CheckRate)
	providerConfig.Confirmation = int(statuscakeData.Confirmation)
	providerConfig.ContactGroup = strings.Join(statuscakeData.ContactGroups, ",")
	providerConfig.EnableSSLAlert = statuscakeData.EnableSSLAlert
	if statuscakeData.FindString != nil {
		providerConfig.FindString = *statuscakeData.FindString
	}
	providerConfig.FollowRedirect = statuscakeData.FollowRedirects
	providerConfig.Paused = statuscakeData.Paused
	providerConfig.TestTags = strings.Join(statuscakeData.Tags, ",")
	providerConfig.TestType = string(statuscakeData.TestType)
	providerConfig.TriggerRate = int(statuscakeData.TriggerRate)
	providerConfig.StatusCodes = strings.Join(statuscakeData.StatusCodes, ",")

	m.Config = &providerConfig

	return &m
}

// StatusCakeApiResponseDataToBaseMonitorMapper function to map Statuscake Uptime Test Response to Monitor
func StatusCakeApiResponseDataToBaseMonitorMapper(statuscakeData statuscake.UptimeTestOverview) *models.Monitor {
	var m models.Monitor
	m.Name = statuscakeData.Name
	m.URL = statuscakeData.WebsiteURL
	m.ID = statuscakeData.ID

	var providerConfig endpointmonitorv1alpha1.StatusCakeConfig
	providerConfig.CheckRate = int(statuscakeData.CheckRate)
	providerConfig.ContactGroup = strings.Join(statuscakeData.ContactGroups, ",")
	providerConfig.Paused = statuscakeData.Paused
	providerConfig.TestTags = strings.Join(statuscakeData.Tags, ",")
	providerConfig.TestType = string(statuscakeData.TestType)

	m.Config = &providerConfig

	return &m
}

// StatusCakeMonitorMonitorsToBaseMonitorsMapper function to map Statuscake structure to Monitor
func StatusCakeMonitorMonitorsToBaseMonitorsMapper(statuscakeData []statuscake.UptimeTestOverview) []models.Monitor {
	var monitors []models.Monitor
	for _, payloadData := range statuscakeData {
		monitors = append(monitors, *StatusCakeApiResponseDataToBaseMonitorMapper(payloadData))
	}
	return monitors
}
