package endpoints

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/portainer/libhttp/request"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/internal/utils"
)

const (
	EdgeDeviceFilterAll       = "all"
	EdgeDeviceFilterTrusted   = "trusted"
	EdgeDeviceFilterUntrusted = "untrusted"
	EdgeDeviceFilterNone      = "none"
)

const (
	EdgeDeviceIntervalMultiplier = 2
	EdgeDeviceIntervalAdd        = 20
)

// @id EndpointList
// @summary List environments(endpoints)
// @description List all environments(endpoints) based on the current user authorizations. Will
// @description return all environments(endpoints) if using an administrator or team leader account otherwise it will
// @description only return authorized environments(endpoints).
// @description **Access policy**: restricted
// @tags endpoints
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param start query int false "Start searching from"
// @param search query string false "Search query"
// @param groupId query int false "List environments(endpoints) of this group"
// @param limit query int false "Limit results to this value"
// @param types query []int false "List environments(endpoints) of this type"
// @param tagIds query []int false "search environments(endpoints) with these tags (depends on tagsPartialMatch)"
// @param tagsPartialMatch query bool false "If true, will return environment(endpoint) which has one of tagIds, if false (or missing) will return only environments(endpoints) that has all the tags"
// @param endpointIds query []int false "will return only these environments(endpoints)"
// @param edgeDeviceFilter query string false "will return only these edge environments, none will return only regular edge environments" Enum("all", "trusted", "untrusted", "none")
// @param name query string false "will return only environments(endpoints) with this name"
// @success 200 {array} portaineree.Endpoint "Endpoints"
// @failure 500 "Server error"
// @router /endpoints [get]
func (handler *Handler) endpointList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	start, _ := request.RetrieveNumericQueryParameter(r, "start", true)
	if start != 0 {
		start--
	}

	search, _ := request.RetrieveQueryParameter(r, "search", true)
	if search != "" {
		search = strings.ToLower(search)
	}

	groupID, _ := request.RetrieveNumericQueryParameter(r, "groupId", true)
	limit, _ := request.RetrieveNumericQueryParameter(r, "limit", true)
	sortField, _ := request.RetrieveQueryParameter(r, "sort", true)
	sortOrder, _ := request.RetrieveQueryParameter(r, "order", true)

	var statuses []int
	request.RetrieveJSONQueryParameter(r, "status", &statuses, true)

	var groupIDs []int
	request.RetrieveJSONQueryParameter(r, "groupIds", &groupIDs, true)

	provisioned, _ := request.RetrieveBooleanQueryParameter(r, "provisioned", true)

	var endpointTypes []int
	request.RetrieveJSONQueryParameter(r, "types", &endpointTypes, true)

	var tagIDs []portaineree.TagID
	request.RetrieveJSONQueryParameter(r, "tagIds", &tagIDs, true)

	tagsPartialMatch, _ := request.RetrieveBooleanQueryParameter(r, "tagsPartialMatch", true)

	var endpointIDs []portaineree.EndpointID
	request.RetrieveJSONQueryParameter(r, "endpointIds", &endpointIDs, true)

	endpointGroups, err := handler.dataStore.EndpointGroup().EndpointGroups()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve environment groups from the database", err}
	}

	endpoints, err := handler.dataStore.Endpoint().Endpoints()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve environments from the database", err}
	}

	settings, err := handler.dataStore.Settings().Settings()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve settings from the database", err}
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve info from request context", err}
	}

	filteredEndpoints := security.FilterEndpoints(endpoints, endpointGroups, securityContext)
	totalAvailableEndpoints := len(filteredEndpoints)

	if groupID != 0 {
		filteredEndpoints = filterEndpointsByGroupIDs(filteredEndpoints, []int{groupID})
	}

	if endpointIDs != nil {
		filteredEndpoints = filteredEndpointsByIds(filteredEndpoints, endpointIDs)
	}

	if len(groupIDs) > 0 {
		filteredEndpoints = filterEndpointsByGroupIDs(filteredEndpoints, groupIDs)
	}

	name, _ := request.RetrieveQueryParameter(r, "name", true)
	if name != "" {
		filteredEndpoints = filterEndpointsByName(filteredEndpoints, name)
	}

	edgeDeviceFilter, _ := request.RetrieveQueryParameter(r, "edgeDeviceFilter", false)
	if edgeDeviceFilter != "" {
		filteredEndpoints = filterEndpointsByEdgeDevice(filteredEndpoints, edgeDeviceFilter)
	}

	if len(statuses) > 0 {
		filteredEndpoints = filterEndpointsByStatuses(filteredEndpoints, statuses, settings)
	}

	if search != "" {
		tags, err := handler.dataStore.Tag().Tags()
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve tags from the database", err}
		}
		tagsMap := make(map[portaineree.TagID]string)
		for _, tag := range tags {
			tagsMap[tag.ID] = tag.Name
		}
		filteredEndpoints = filterEndpointsBySearchCriteria(filteredEndpoints, endpointGroups, tagsMap, search)
	}

	if endpointTypes != nil {
		filteredEndpoints = filterEndpointsByTypes(filteredEndpoints, endpointTypes)
	}

	if tagIDs != nil {
		filteredEndpoints = filteredEndpointsByTags(filteredEndpoints, tagIDs, endpointGroups, tagsPartialMatch)
	}

	if provisioned {
		filteredEndpoints = filteredNonProvisionedEndpoints(filteredEndpoints)
	}

	// Sort endpoints by field
	sortEndpointsByField(filteredEndpoints, endpointGroups, sortField, sortOrder == "desc")

	filteredEndpointCount := len(filteredEndpoints)

	paginatedEndpoints := paginateEndpoints(filteredEndpoints, start, limit)

	for idx := range paginatedEndpoints {
		hideFields(&paginatedEndpoints[idx])
		paginatedEndpoints[idx].ComposeSyntaxMaxVersion = handler.ComposeStackManager.ComposeSyntaxMaxVersion()
		if paginatedEndpoints[idx].EdgeCheckinInterval == 0 {
			paginatedEndpoints[idx].EdgeCheckinInterval = settings.EdgeAgentCheckinInterval
		}
		paginatedEndpoints[idx].QueryDate = time.Now().Unix()
	}

	w.Header().Set("X-Total-Count", strconv.Itoa(filteredEndpointCount))
	w.Header().Set("X-Total-Available", strconv.Itoa(totalAvailableEndpoints))
	return response.JSON(w, paginatedEndpoints)
}

func paginateEndpoints(endpoints []portaineree.Endpoint, start, limit int) []portaineree.Endpoint {
	if limit == 0 {
		return endpoints
	}

	endpointCount := len(endpoints)

	if start > endpointCount {
		start = endpointCount
	}

	end := start + limit
	if end > endpointCount {
		end = endpointCount
	}

	return endpoints[start:end]
}

func filterEndpointsByGroupIDs(endpoints []portaineree.Endpoint, endpointGroupIDs []int) []portaineree.Endpoint {
	filteredEndpoints := make([]portaineree.Endpoint, 0)

	for _, endpoint := range endpoints {
		if utils.Contains(endpointGroupIDs, int(endpoint.GroupID)) {
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

func filterEndpointsByStatuses(endpoints []portaineree.Endpoint, statuses []int, settings *portaineree.Settings) []portaineree.Endpoint {
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

		if utils.Contains(statuses, int(status)) {
			filteredEndpoints = append(filteredEndpoints, endpoint)
		}
	}

	return filteredEndpoints
}

func sortEndpointsByField(endpoints []portaineree.Endpoint, endpointGroups []portaineree.EndpointGroup, sortField string, isSortDesc bool) {

	switch sortField {
	case "Name":
		if isSortDesc {
			sort.Stable(sort.Reverse(EndpointsByName(endpoints)))
		} else {
			sort.Stable(EndpointsByName(endpoints))
		}

	case "Group":
		endpointGroupNames := make(map[portaineree.EndpointGroupID]string, 0)
		for _, group := range endpointGroups {
			endpointGroupNames[group.ID] = group.Name
		}

		endpointsByGroup := EndpointsByGroup{
			endpointGroupNames: endpointGroupNames,
			endpoints:          endpoints,
		}

		if isSortDesc {
			sort.Stable(sort.Reverse(endpointsByGroup))
		} else {
			sort.Stable(EndpointsByGroup(endpointsByGroup))
		}

	case "Status":
		if isSortDesc {
			sort.Slice(endpoints, func(i, j int) bool {
				return endpoints[i].Status > endpoints[j].Status
			})
		} else {
			sort.Slice(endpoints, func(i, j int) bool {
				return endpoints[i].Status < endpoints[j].Status
			})
		}
	}
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

func filterEndpointsByTypes(endpoints []portaineree.Endpoint, endpointTypes []int) []portaineree.Endpoint {
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

func filterEndpointsByEdgeDevice(endpoints []portaineree.Endpoint, edgeDeviceFilter string) []portaineree.Endpoint {
	filteredEndpoints := make([]portaineree.Endpoint, 0)

	for _, endpoint := range endpoints {
		if shouldReturnEdgeDevice(endpoint, edgeDeviceFilter) {
			filteredEndpoints = append(filteredEndpoints, endpoint)
		}
	}
	return filteredEndpoints
}

func shouldReturnEdgeDevice(endpoint portaineree.Endpoint, edgeDeviceFilter string) bool {
	// none - return all endpoints that are not edge devices
	if edgeDeviceFilter == EdgeDeviceFilterNone && !endpoint.IsEdgeDevice {
		return true
	}

	if !endpointutils.IsEdgeEndpoint(&endpoint) {
		return false
	}

	switch edgeDeviceFilter {
	case EdgeDeviceFilterAll:
		return true
	case EdgeDeviceFilterTrusted:
		return endpoint.UserTrusted
	case EdgeDeviceFilterUntrusted:
		return !endpoint.UserTrusted
	}

	return false
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

func getEndpointGroup(groupID portaineree.EndpointGroupID, groups []portaineree.EndpointGroup) portaineree.EndpointGroup {
	var endpointGroup portaineree.EndpointGroup
	for _, group := range groups {
		if group.ID == groupID {
			endpointGroup = group
			break
		}
	}
	return endpointGroup
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
