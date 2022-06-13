angular.module('portainer.app').controller('EndpointsDatatableController', [
  '$scope',
  '$controller',
  '$interval',
  'DatatableService',
  'PaginationService',
  'Notifications',
  function ($scope, $controller, $interval, DatatableService, PaginationService) {
    angular.extend(this, $controller('GenericDatatableController', { $scope: $scope }));

    this.state = Object.assign(this.state, {
      allowSelection: this.allowSelection,
      orderBy: this.orderBy,
      loading: true,
      filteredDataSet: [],
      totalFilteredDataset: 0,
      pageNumber: 1,
      interval: undefined,
    });

    this.paginationChanged = function () {
      this.state.loading = true;
      this.state.filteredDataSet = [];
      const start = (this.state.pageNumber - 1) * this.state.paginatedItemLimit + 1;
      this.retrievePage(start, this.state.paginatedItemLimit, this.state.textFilter)
        .then((data) => {
          this.state.filteredDataSet = data.endpoints;
          this.state.totalFilteredDataSet = data.totalCount;
        })
        .finally(() => {
          this.state.loading = false;
        });
    };

    this.onPageChange = function (newPageNumber) {
      this.state.pageNumber = newPageNumber;
      this.paginationChanged();
    };

    this.setReferrer = function () {
      window.localStorage.setItem('wizardReferrer', 'environments');
    };

    /**
     * Overridden
     */
    this.onTextFilterChange = function () {
      var filterValue = this.state.textFilter;
      DatatableService.setDataTableTextFilters(this.tableKey, filterValue);
      this.paginationChanged();
    };

    /**
     * Overridden
     */
    this.changePaginationLimit = function () {
      PaginationService.setPaginationLimit(this.tableKey, this.state.paginatedItemLimit);
      this.paginationChanged();
    };

    this.startPolling = function (interval) {
      this.state.interval = $interval(this.paginationChanged.bind(this), interval);
    };

    this.$onDestroy = function () {
      $interval.cancel(this.state.interval);
      this.state.interval = undefined;
    };

    /**
     * Do not allow selecting provisioning items
     */
    this.allowSelection = function (item) {
      if (item.Status === 3) {
        return false;
      }
      return true;
    };

    /**
     * Overridden
     */
    this.$onInit = function () {
      this.setDefaults();
      this.prepareTableFromDataset();

      var storedOrder = DatatableService.getDataTableOrder(this.tableKey);
      if (storedOrder !== null) {
        this.state.reverseOrder = storedOrder.reverse;
        this.state.orderBy = storedOrder.orderBy;
      }

      var textFilter = DatatableService.getDataTableTextFilters(this.tableKey);
      if (textFilter !== null) {
        this.state.textFilter = textFilter;
        this.onTextFilterChange();
      }

      var storedFilters = DatatableService.getDataTableFilters(this.tableKey);
      if (storedFilters !== null) {
        this.filters = storedFilters;
      }
      if (this.filters && this.filters.state) {
        this.filters.state.open = false;
      }

      this.paginationChanged();
      this.startPolling(30000);
    };
  },
]);
