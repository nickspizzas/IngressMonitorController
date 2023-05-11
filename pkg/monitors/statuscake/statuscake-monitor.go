package statuscake

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	gocache "github.com/patrickmn/go-cache"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	statuscake "github.com/StatusCakeDev/statuscake-go"
	"github.com/StatusCakeDev/statuscake-go/credentials"
	endpointmonitorv1alpha1 "github.com/stakater/IngressMonitorController/v2/api/v1alpha1"
	"github.com/stakater/IngressMonitorController/v2/pkg/config"
	"github.com/stakater/IngressMonitorController/v2/pkg/models"
)

var cache = gocache.New(5*time.Minute, 5*time.Minute)
var log = logf.Log.WithName("statuscake-monitor")
var statusCodes = []string{
	"204", // No content
	"205", // Reset content
	"206", // Partial content
	"303", // See other
	"305", // Use proxy
	// https://en.wikipedia.org/wiki/List_of_HTTP_status_codes#4xx_Client_errors
	// https://support.cloudflare.com/hc/en-us/articles/115003014512/
	"400",
	"401",
	"402",
	"403",
	"404",
	"405",
	"406",
	"407",
	"408",
	"409",
	"410",
	"411",
	"412",
	"413",
	"414",
	"415",
	"416",
	"417",
	"418",
	"421",
	"422",
	"423",
	"424",
	"425",
	"426",
	"428",
	"429",
	"431",
	"444",
	"451",
	"499",
	// https://support.cloudflare.com/hc/en-us/articles/115003011431/
	"500",
	"501",
	"502",
	"503",
	"504",
	"505",
	"506",
	"507",
	"508",
	"509",
	"510",
	"511",
	"520",
	"521",
	"522",
	"523",
	"524",
	"525",
	"526",
	"527",
	"530",
	"598",
	"599",
}

// StatusCakeMonitorService is the service structure for StatusCake
type StatusCakeMonitorService struct {
	apiKey   string
	url      string
	username string
	cgroup   string
	client   *http.Client
}

func (service *StatusCakeMonitorService) Equal(oldMonitor models.Monitor, newMonitor models.Monitor) bool {

	old := service.generateUptimeTest(oldMonitor, service.cgroup)
	new := service.generateUptimeTest(newMonitor, service.cgroup)

	if !(reflect.DeepEqual(old, new)) {
		log.Info(fmt.Sprintf("Changes deteced for monitor %s", oldMonitor.Name))
		return false
	}
	return true
}

func (service *StatusCakeMonitorService) generateUptimeTest(m models.Monitor, cgroup string) statuscake.UptimeTest {

	// Retrieve provider configuration
	providerConfig, _ := m.Config.(*endpointmonitorv1alpha1.StatusCakeConfig)

	uptimeTest := statuscake.UptimeTest{}
	uptimeTest.Name = m.Name

	uptimeTest.TestType = statuscake.UptimeTestTypeHTTP
	if providerConfig != nil && len(providerConfig.TestType) > 0 {
		uptimeTest.TestType = statuscake.UptimeTestType(providerConfig.TestType)
	}

	unEscapedURL, _ := url.QueryUnescape(m.URL)
	uptimeTest.WebsiteURL = unEscapedURL

	uptimeTest.CheckRate = statuscake.UptimeTestCheckRateFiveMinutes
	if providerConfig != nil && providerConfig.CheckRate > 0 {
		uptimeTest.CheckRate = statuscake.UptimeTestCheckRate(providerConfig.CheckRate)
	}

	if providerConfig != nil && len(providerConfig.ContactGroup) > 0 {
		uptimeTest.ContactGroups = convertStringToArray(providerConfig.ContactGroup)
	} else {
		if cgroup != "" {
			uptimeTest.ContactGroups = convertStringToArray(cgroup)
		}
	}

	if providerConfig != nil && len(providerConfig.TestTags) > 0 {
		uptimeTest.Tags = convertStringToArray(providerConfig.TestTags)
	}

	uptimeTest.StatusCodes = statusCodes
	if providerConfig != nil && len(providerConfig.StatusCodes) > 0 {
		uptimeTest.StatusCodes = convertStringToArray(providerConfig.StatusCodes)
	}

	if providerConfig != nil {
		uptimeTest.Paused = providerConfig.Paused
		uptimeTest.FollowRedirects = providerConfig.FollowRedirect
		uptimeTest.EnableSSLAlert = providerConfig.EnableSSLAlert
	}

	if providerConfig != nil && providerConfig.TriggerRate > 0 {
		uptimeTest.TriggerRate = int32(providerConfig.TriggerRate)
	}

	if providerConfig != nil && providerConfig.Confirmation > 0 {
		uptimeTest.Confirmation = int32(providerConfig.Confirmation)
	}

	if providerConfig != nil {
		uptimeTest.FindString = &providerConfig.FindString
	}

	return uptimeTest
}

// convertValuesToString changes multiple values returned by same key to string for validation purposes
func convertUrlValuesToString(vals url.Values, key string) string {
	var valuesArray []string
	for k, v := range vals {
		if k == key {
			valuesArray = append(valuesArray, v...)
		}
	}
	return strings.Join(valuesArray, ",")
}

// convertStringToArray function is used to convert string to []string
func convertStringToArray(stringValues string) []string {
	stringArray := strings.Split(stringValues, ",")
	return stringArray
}

// Setup function is used to initialise the StatusCake service
func (service *StatusCakeMonitorService) Setup(p config.Provider) {
	service.apiKey = p.ApiKey
	service.url = p.ApiURL
	service.username = p.Username
	service.cgroup = p.AlertContacts
	service.client = &http.Client{}
}

// GetByName function will Get a monitor by it's name
func (service *StatusCakeMonitorService) GetByName(name string) (*models.Monitor, error) {
	monitors := service.GetAll()
	if len(monitors) != 0 {
		for _, monitor := range monitors {
			if monitor.Name == name {
				return &monitor, nil
			}
		}
	}
	errorString := "GetByName Request failed for name: " + name
	return nil, errors.New(errorString)

}

// GetByID function will Get a monitor by it's ID
func (service *StatusCakeMonitorService) GetById(id string) (*models.Monitor, error) {

	bearer := credentials.NewBearerWithStaticToken(service.apiKey)
	client := statuscake.NewClient(statuscake.WithRequestCredentials(bearer))

	resp, err := client.GetUptimeTest(context.Background(), id).Execute()

	if err != nil {
		log.Error(nil, "Getting monitor with id "+id+" failed: "+fmt.Sprintf("%+v", statuscake.Errors(err)))
		return nil, err
	}

	return StatusCakeMonitorMonitorToBaseMonitorMapper(resp.Data), nil
}

// GetAll function will fetch all monitors
func (service *StatusCakeMonitorService) GetAll() []models.Monitor {
	var statusCakeMonitorData []statuscake.UptimeTestOverview

	cached, found := cache.Get("uptime-checks")
	if found {
		return StatusCakeMonitorMonitorsToBaseMonitorsMapper(cached.([]statuscake.UptimeTestOverview))
	}

	page := 1
	for {
		res := service.fetchMonitors(page)
		statusCakeMonitorData = append(statusCakeMonitorData, res.Data...)
		if page >= int(res.Metadata.PageCount) {
			break
		}
		page += 1
	}

	cache.Set("uptime-checks", statusCakeMonitorData, gocache.DefaultExpiration)

	return StatusCakeMonitorMonitorsToBaseMonitorsMapper(statusCakeMonitorData)
}

func (service *StatusCakeMonitorService) fetchMonitors(page int) *statuscake.UptimeTests {

	bearer := credentials.NewBearerWithStaticToken(service.apiKey)
	client := statuscake.NewClient(statuscake.WithRequestCredentials(bearer))

	resp, err := client.ListUptimeTests(context.Background()).Execute()

	if err != nil {
		log.Error(nil, "Getting all monitors failed: "+fmt.Sprintf("%+v", statuscake.Errors(err)))
		return nil
	}

	return &resp
}

// Add will create a new Monitor
func (service *StatusCakeMonitorService) Add(m models.Monitor) (*string, error) {
	defer cache.Flush()

	uptimeTest := service.generateUptimeTest(m, service.cgroup)

	bearer := credentials.NewBearerWithStaticToken(service.apiKey)
	client := statuscake.NewClient(statuscake.WithRequestCredentials(bearer))

	resp, err := client.CreateUptimeTest(context.Background()).
		Name(uptimeTest.Name).
		TestType(uptimeTest.TestType).
		WebsiteURL(uptimeTest.WebsiteURL).
		CheckRate(uptimeTest.CheckRate).
		ContactGroups(uptimeTest.ContactGroups).
		Tags(uptimeTest.Tags).
		StatusCodes(uptimeTest.StatusCodes).
		Paused(uptimeTest.Paused).
		FollowRedirects(uptimeTest.FollowRedirects).
		EnableSSLAlert(uptimeTest.EnableSSLAlert).
		TriggerRate(uptimeTest.TriggerRate).
		Confirmation(uptimeTest.Confirmation).
		FindString(*uptimeTest.FindString).
		Execute()

	if err != nil {
		log.Error(nil, "Adding monitor "+m.Name+"failed: "+fmt.Sprintf("%+v", statuscake.Errors(err)))
		return nil, err
	}

	id := resp.Data.NewID
	log.Info(fmt.Sprintf("Added monitor %s with id %s", m.Name, id))

	return &id, nil
}

// Update will update an existing Monitor
func (service *StatusCakeMonitorService) Update(m models.Monitor) error {
	defer cache.Flush()

	bearer := credentials.NewBearerWithStaticToken(service.apiKey)
	client := statuscake.NewClient(statuscake.WithRequestCredentials(bearer))

	uptimeTest := service.generateUptimeTest(m, service.cgroup)

	err := client.UpdateUptimeTest(context.Background(), m.ID).
		Name(uptimeTest.Name).
		WebsiteURL(uptimeTest.WebsiteURL).
		CheckRate(uptimeTest.CheckRate).
		ContactGroups(uptimeTest.ContactGroups).
		Tags(uptimeTest.Tags).
		StatusCodes(uptimeTest.StatusCodes).
		Paused(uptimeTest.Paused).
		FollowRedirects(uptimeTest.FollowRedirects).
		EnableSSLAlert(uptimeTest.EnableSSLAlert).
		TriggerRate(uptimeTest.TriggerRate).
		Confirmation(uptimeTest.Confirmation).
		FindString(*uptimeTest.FindString).
		Execute()

	if err != nil {
		log.Error(nil, "Updating monitor "+m.Name+"failed: "+fmt.Sprintf("%+v", statuscake.Errors(err)))
		return err
	}

	log.Info(fmt.Sprintf("Updated monitor %s with id %s", m.Name, m.ID))

	return nil
}

// Remove will delete an existing Monitor
func (service *StatusCakeMonitorService) Remove(m models.Monitor) {
	defer cache.Flush()

	bearer := credentials.NewBearerWithStaticToken(service.apiKey)
	client := statuscake.NewClient(statuscake.WithRequestCredentials(bearer))

	err := client.DeleteUptimeTest(context.Background(), m.ID).Execute()

	if err != nil {
		log.Error(nil, "Deleting monitor "+m.Name+"failed: "+fmt.Sprintf("%+v", statuscake.Errors(err)))
		return
	}

	log.Info(fmt.Sprintf("Deleted monitor %swith id %s", m.Name, m.ID))

	return
}
