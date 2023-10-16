package edgeconfigs

import (
	"net/http"
	"strconv"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer/pkg/libhttp/request"
)

type SearchFieldGetters = []func(portaineree.EdgeConfig) string

type SearchQueryParams struct {
	search string
}

type ConfigCategoryQueryParams struct {
	configCategory portaineree.EdgeConfigCategory
}

func searchFn(configs []portaineree.EdgeConfig, params SearchQueryParams, getters SearchFieldGetters) []portaineree.EdgeConfig {
	search := params.search

	if search == "" {
		return configs
	}

	results := []portaineree.EdgeConfig{}

	for confIdx := range configs {
		config := configs[confIdx]
		for getIdx := range getters {
			getter := getters[getIdx]
			if strings.Contains(getter(config), search) {
				results = append(results, config)
				break
			}
		}
	}

	return results
}

type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

type SortQueryParams struct {
	sort  string
	order SortOrder
}

func sortFn(configs []portaineree.EdgeConfig, params SortQueryParams) []portaineree.EdgeConfig {
	return configs
}

type PaginationQueryParams struct {
	start int
	limit int
}

func paginateFn(configs []portaineree.EdgeConfig, params PaginationQueryParams) []portaineree.EdgeConfig {
	start := params.start
	limit := params.limit

	if limit == 0 {
		return configs
	}

	max := len(configs)

	if start < 0 {
		start = 0
	}

	if start > max {
		start = max
	}

	end := start + limit
	if end > max {
		end = max
	}

	return configs[start:end]
}

type QueryParams struct {
	ConfigCategoryQueryParams
	SearchQueryParams
	SortQueryParams
	PaginationQueryParams
}

func extractListModifiersQueryParams(r *http.Request) QueryParams {
	search, _ := request.RetrieveQueryParameter(r, "search", true)
	sortField, _ := request.RetrieveQueryParameter(r, "sort", true)
	sortOrder, _ := request.RetrieveQueryParameter(r, "order", true)
	start, _ := request.RetrieveNumericQueryParameter(r, "start", true)
	limit, _ := request.RetrieveNumericQueryParameter(r, "limit", true)
	configCategory, _ := request.RetrieveQueryParameter(r, "category", true)

	if configCategory == "" {
		configCategory = string(portaineree.EdgeConfigCategoryConfig)
	}

	return QueryParams{
		ConfigCategoryQueryParams{
			configCategory: portaineree.EdgeConfigCategory(configCategory),
		},
		SearchQueryParams{
			search: search,
		},
		SortQueryParams{
			sort:  sortField,
			order: SortOrder(sortOrder),
		},
		PaginationQueryParams{
			start: start,
			limit: limit,
		},
	}
}

type FilterResult struct {
	configs        []portaineree.EdgeConfig
	totalCount     int
	totalAvailable int
}

func searchOrderAndPaginate(configs []portaineree.EdgeConfig, params QueryParams, searchConfig SearchFieldGetters) FilterResult {
	totalAvailable := len(configs)

	configs = searchFn(configs, params.SearchQueryParams, searchConfig)
	configs = sortFn(configs, params.SortQueryParams)

	totalCount := len(configs)
	configs = paginateFn(configs, params.PaginationQueryParams)

	return FilterResult{
		configs:        configs,
		totalCount:     totalCount,
		totalAvailable: totalAvailable,
	}
}

func filterConfigCategory(configs []portaineree.EdgeConfig, params QueryParams) []portaineree.EdgeConfig {
	configCategory := params.configCategory

	if configCategory == "" {
		return configs
	}

	results := []portaineree.EdgeConfig{}

	for confIdx := range configs {
		config := configs[confIdx]
		if configCategory == config.Category {
			results = append(results, config)
		}
	}

	return results
}

func applyFilterResultsHeaders(w *http.ResponseWriter, result FilterResult) {
	(*w).Header().Set("X-Total-Count", strconv.Itoa(result.totalCount))
	(*w).Header().Set("X-Total-Available", strconv.Itoa(result.totalAvailable))
}
