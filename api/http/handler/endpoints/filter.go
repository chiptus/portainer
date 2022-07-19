package endpoints

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"golang.org/x/exp/slices"
)

type EnvironmentsQuery struct {
	search              string
	types               []portaineree.EndpointType
	tagIds              []portaineree.TagID
	endpointIds         []portaineree.EndpointID
	tagsPartialMatch    bool
	groupIds            []portaineree.EndpointGroupID
	status              []portaineree.EndpointStatus
	edgeDevice          *bool
	edgeDeviceUntrusted bool
	provisioned         bool
	name                string
}

func parseQuery(r *http.Request) (EnvironmentsQuery, error) {
	search, _ := request.RetrieveQueryParameter(r, "search", true)
	if search != "" {
		search = strings.ToLower(search)
	}

	status, err := getNumberArrayQueryParameter[portaineree.EndpointStatus](r, "status")
	if err != nil {
		return EnvironmentsQuery{}, err
	}

	groupIDs, err := getNumberArrayQueryParameter[portaineree.EndpointGroupID](r, "groupIds")
	if err != nil {
		return EnvironmentsQuery{}, err
	}

	provisioned, _ := request.RetrieveBooleanQueryParameter(r, "provisioned", true)

	endpointTypes, err := getNumberArrayQueryParameter[portaineree.EndpointType](r, "types")
	if err != nil {
		return EnvironmentsQuery{}, err
	}

	tagIDs, err := getNumberArrayQueryParameter[portaineree.TagID](r, "tagIds")
	if err != nil {
		return EnvironmentsQuery{}, err
	}

	tagsPartialMatch, _ := request.RetrieveBooleanQueryParameter(r, "tagsPartialMatch", true)

	endpointIDs, err := getNumberArrayQueryParameter[portaineree.EndpointID](r, "endpointIds")
	if err != nil {
		return EnvironmentsQuery{}, err
	}

	name, _ := request.RetrieveQueryParameter(r, "name", true)

	edgeDeviceParam, _ := request.RetrieveQueryParameter(r, "edgeDevice", true)

	var edgeDevice *bool
	if edgeDeviceParam != "" {
		edgeDevice = BoolAddr(edgeDeviceParam == "true")
	}

	edgeDeviceUntrusted, _ := request.RetrieveBooleanQueryParameter(r, "edgeDeviceUntrusted", true)

	return EnvironmentsQuery{
		search:              search,
		types:               endpointTypes,
		tagIds:              tagIDs,
		endpointIds:         endpointIDs,
		tagsPartialMatch:    tagsPartialMatch,
		groupIds:            groupIDs,
		status:              status,
		edgeDevice:          edgeDevice,
		edgeDeviceUntrusted: edgeDeviceUntrusted,
		provisioned:         provisioned,
		name:                name,
	}, nil
}

func (handler *Handler) filterEndpointsByQuery(filteredEndpoints []portaineree.Endpoint, query EnvironmentsQuery, groups []portaineree.EndpointGroup, settings *portaineree.Settings) ([]portaineree.Endpoint, int, error) {
	totalAvailableEndpoints := len(filteredEndpoints)

	if len(query.endpointIds) > 0 {
		filteredEndpoints = filteredEndpointsByIds(filteredEndpoints, query.endpointIds)
	}

	if len(query.groupIds) > 0 {
		filteredEndpoints = filterEndpointsByGroupIDs(filteredEndpoints, query.groupIds)
	}

	if query.name != "" {
		filteredEndpoints = filterEndpointsByName(filteredEndpoints, query.name)
	}

	if query.edgeDevice != nil {
		filteredEndpoints = filterEndpointsByEdgeDevice(filteredEndpoints, *query.edgeDevice, query.edgeDeviceUntrusted)
	} else {
		// If the edgeDevice parameter is not set, we need to filter out the untrusted edge devices
		filteredEndpoints = filter(filteredEndpoints, func(endpoint portaineree.Endpoint) bool {
			return !endpoint.IsEdgeDevice || endpoint.UserTrusted
		})
	}

	if len(query.status) > 0 {
		filteredEndpoints = filterEndpointsByStatuses(filteredEndpoints, query.status, settings)
	}

	if query.search != "" {
		tags, err := handler.dataStore.Tag().Tags()
		if err != nil {
			return nil, 0, errors.WithMessage(err, "Unable to retrieve tags from the database")
		}

		tagsMap := make(map[portaineree.TagID]string)
		for _, tag := range tags {
			tagsMap[tag.ID] = tag.Name
		}

		filteredEndpoints = filterEndpointsBySearchCriteria(filteredEndpoints, groups, tagsMap, query.search)
	}

	if len(query.types) > 0 {
		filteredEndpoints = filterEndpointsByTypes(filteredEndpoints, query.types)
	}

	if len(query.tagIds) > 0 {
		filteredEndpoints = filteredEndpointsByTags(filteredEndpoints, query.tagIds, groups, query.tagsPartialMatch)
	}

	if query.provisioned {
		filteredEndpoints = filteredNonProvisionedEndpoints(filteredEndpoints)
	}

	return filteredEndpoints, totalAvailableEndpoints, nil
}

func filterEndpointsByGroupIDs(endpoints []portaineree.Endpoint, endpointGroupIDs []portaineree.EndpointGroupID) []portaineree.Endpoint {
	filteredEndpoints := make([]portaineree.Endpoint, 0)

	for _, endpoint := range endpoints {
		if slices.Contains(endpointGroupIDs, endpoint.GroupID) {
			filteredEndpoints = append(filteredEndpoints, endpoint)
		}
	}

	return filteredEndpoints
}

func filterEndpointsBySearchCriteria(endpoints []portaineree.Endpoint, endpointGroups []portaineree.EndpointGroup, tagsMap map[portaineree.TagID]string, searchCriteria string) []portaineree.Endpoint {
	filteredEndpoints := make([]portaineree.Endpoint, 0)

	for _, endpoint := range endpoints {
		endpointTags := convertTagIDsToTags(tagsMap, endpoint.TagIDs)
		if endpointMatchSearchCriteria(&endpoint, endpointTags, searchCriteria) {
			filteredEndpoints = append(filteredEndpoints, endpoint)
			continue
		}

		if endpointGroupMatchSearchCriteria(&endpoint, endpointGroups, tagsMap, searchCriteria) {
			filteredEndpoints = append(filteredEndpoints, endpoint)
		}
	}

	return filteredEndpoints
}

func filterEndpointsByStatuses(endpoints []portaineree.Endpoint, statuses []portaineree.EndpointStatus, settings *portaineree.Settings) []portaineree.Endpoint {
	filteredEndpoints := make([]portaineree.Endpoint, 0)

	for _, endpoint := range endpoints {
		status := endpoint.Status
		if endpointutils.IsEdgeEndpoint(&endpoint) {
			isCheckValid := false
			edgeCheckinInterval := endpoint.EdgeCheckinInterval
			if endpoint.EdgeCheckinInterval == 0 {
				edgeCheckinInterval = settings.EdgeAgentCheckinInterval
			}

			if edgeCheckinInterval != 0 && endpoint.LastCheckInDate != 0 {
				isCheckValid = time.Now().Unix()-endpoint.LastCheckInDate <= int64(edgeCheckinInterval*EdgeDeviceIntervalMultiplier+EdgeDeviceIntervalAdd)
			}

			status = portaineree.EndpointStatusDown // Offline
			if isCheckValid {
				status = portaineree.EndpointStatusUp // Online
			}
		}

		if slices.Contains(statuses, status) {
			filteredEndpoints = append(filteredEndpoints, endpoint)
		}
	}

	return filteredEndpoints
}

func endpointMatchSearchCriteria(endpoint *portaineree.Endpoint, tags []string, searchCriteria string) bool {
	if strings.Contains(strings.ToLower(endpoint.Name), searchCriteria) {
		return true
	}

	if strings.Contains(strings.ToLower(endpoint.URL), searchCriteria) {
		return true
	}

	if endpoint.Status == portaineree.EndpointStatusUp && searchCriteria == "up" {
		return true
	} else if endpoint.Status == portaineree.EndpointStatusDown && searchCriteria == "down" {
		return true
	}
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), searchCriteria) {
			return true
		}
	}

	return false
}

func endpointGroupMatchSearchCriteria(endpoint *portaineree.Endpoint, endpointGroups []portaineree.EndpointGroup, tagsMap map[portaineree.TagID]string, searchCriteria string) bool {
	for _, group := range endpointGroups {
		if group.ID == endpoint.GroupID {
			if strings.Contains(strings.ToLower(group.Name), searchCriteria) {
				return true
			}
			tags := convertTagIDsToTags(tagsMap, group.TagIDs)
			for _, tag := range tags {
				if strings.Contains(strings.ToLower(tag), searchCriteria) {
					return true
				}
			}
		}
	}

	return false
}

func filterEndpointsByTypes(endpoints []portaineree.Endpoint, endpointTypes []portaineree.EndpointType) []portaineree.Endpoint {
	filteredEndpoints := make([]portaineree.Endpoint, 0)

	typeSet := map[portaineree.EndpointType]bool{}
	for _, endpointType := range endpointTypes {
		typeSet[portaineree.EndpointType(endpointType)] = true
	}

	for _, endpoint := range endpoints {
		if typeSet[endpoint.Type] {
			filteredEndpoints = append(filteredEndpoints, endpoint)
		}
	}
	return filteredEndpoints
}

func filterEndpointsByEdgeDevice(endpoints []portaineree.Endpoint, edgeDevice bool, untrusted bool) []portaineree.Endpoint {
	filteredEndpoints := make([]portaineree.Endpoint, 0)

	for _, endpoint := range endpoints {
		if shouldReturnEdgeDevice(endpoint, edgeDevice, untrusted) {
			filteredEndpoints = append(filteredEndpoints, endpoint)
		}
	}
	return filteredEndpoints
}

func shouldReturnEdgeDevice(endpoint portaineree.Endpoint, edgeDeviceParam bool, untrustedParam bool) bool {
	if !endpointutils.IsEdgeEndpoint(&endpoint) {
		return true
	}

	if !edgeDeviceParam {
		return !endpoint.IsEdgeDevice
	}

	return endpoint.IsEdgeDevice && endpoint.UserTrusted == !untrustedParam
}

func convertTagIDsToTags(tagsMap map[portaineree.TagID]string, tagIDs []portaineree.TagID) []string {
	tags := make([]string, 0)
	for _, tagID := range tagIDs {
		tags = append(tags, tagsMap[tagID])
	}
	return tags
}

func filteredNonProvisionedEndpoints(endpoints []portaineree.Endpoint) []portaineree.Endpoint {

	filteredEndpoints := make([]portaineree.Endpoint, 0)

	for _, endpoint := range endpoints {
		if endpoint.Status < portaineree.EndpointStatusProvisioning {
			filteredEndpoints = append(filteredEndpoints, endpoint)
		}
	}
	return filteredEndpoints
}

func filteredEndpointsByTags(endpoints []portaineree.Endpoint, tagIDs []portaineree.TagID, endpointGroups []portaineree.EndpointGroup, partialMatch bool) []portaineree.Endpoint {
	filteredEndpoints := make([]portaineree.Endpoint, 0)

	for _, endpoint := range endpoints {
		endpointGroup := getEndpointGroup(endpoint.GroupID, endpointGroups)
		endpointMatched := false
		if partialMatch {
			endpointMatched = endpointPartialMatchTags(endpoint, endpointGroup, tagIDs)
		} else {
			endpointMatched = endpointFullMatchTags(endpoint, endpointGroup, tagIDs)
		}

		if endpointMatched {
			filteredEndpoints = append(filteredEndpoints, endpoint)
		}
	}
	return filteredEndpoints
}

func endpointPartialMatchTags(endpoint portaineree.Endpoint, endpointGroup portaineree.EndpointGroup, tagIDs []portaineree.TagID) bool {
	tagSet := make(map[portaineree.TagID]bool)
	for _, tagID := range tagIDs {
		tagSet[tagID] = true
	}
	for _, tagID := range endpoint.TagIDs {
		if tagSet[tagID] {
			return true
		}
	}
	for _, tagID := range endpointGroup.TagIDs {
		if tagSet[tagID] {
			return true
		}
	}
	return false
}

func endpointFullMatchTags(endpoint portaineree.Endpoint, endpointGroup portaineree.EndpointGroup, tagIDs []portaineree.TagID) bool {
	missingTags := make(map[portaineree.TagID]bool)
	for _, tagID := range tagIDs {
		missingTags[tagID] = true
	}
	for _, tagID := range endpoint.TagIDs {
		if missingTags[tagID] {
			delete(missingTags, tagID)
		}
	}
	for _, tagID := range endpointGroup.TagIDs {
		if missingTags[tagID] {
			delete(missingTags, tagID)
		}
	}
	return len(missingTags) == 0
}

func filteredEndpointsByIds(endpoints []portaineree.Endpoint, ids []portaineree.EndpointID) []portaineree.Endpoint {
	filteredEndpoints := make([]portaineree.Endpoint, 0)

	idsSet := make(map[portaineree.EndpointID]bool)
	for _, id := range ids {
		idsSet[id] = true
	}

	for _, endpoint := range endpoints {
		if idsSet[endpoint.ID] {
			filteredEndpoints = append(filteredEndpoints, endpoint)
		}
	}

	return filteredEndpoints

}

func filterEndpointsByName(endpoints []portaineree.Endpoint, name string) []portaineree.Endpoint {
	if name == "" {
		return endpoints
	}

	filteredEndpoints := make([]portaineree.Endpoint, 0)

	for _, endpoint := range endpoints {
		if endpoint.Name == name {
			filteredEndpoints = append(filteredEndpoints, endpoint)
		}
	}
	return filteredEndpoints
}

func filter(endpoints []portaineree.Endpoint, predicate func(endpoint portaineree.Endpoint) bool) []portaineree.Endpoint {
	filteredEndpoints := make([]portaineree.Endpoint, 0)

	for _, endpoint := range endpoints {
		if predicate(endpoint) {
			filteredEndpoints = append(filteredEndpoints, endpoint)
		}
	}
	return filteredEndpoints
}

func getArrayQueryParameter(r *http.Request, parameter string) []string {
	list, exists := r.Form[fmt.Sprintf("%s[]", parameter)]
	if !exists {
		list = []string{}
	}

	return list
}

func getNumberArrayQueryParameter[T ~int](r *http.Request, parameter string) ([]T, error) {
	list := getArrayQueryParameter(r, parameter)
	if list == nil {
		return []T{}, nil
	}

	var result []T
	for _, item := range list {
		number, err := strconv.Atoi(item)
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to parse parameter %s", parameter)

		}

		result = append(result, T(number))
	}

	return result, nil
}
