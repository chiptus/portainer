angular.module('portainer.app').controller('EdgeStacksDatatableController', [
  '$scope',
  '$controller',
  'DatatableService',
  function ($scope, $controller, DatatableService) {
    angular.extend(this, $controller('GenericDatatableController', { $scope: $scope }));

    this.filters = {
      state: {
        open: false,
        enabled: false,
        showActiveStacks: true,
        showUnactiveStacks: true,
      },
    };

    this.columnVisibility = {
      state: {
        open: false,
      },
      columns: {
        targetVersion: {
          label: 'Target Version',
          display: false,
        },
      },
    };

    this.onColumnVisibilityChange = onColumnVisibilityChange.bind(this);
    function onColumnVisibilityChange(columns) {
      this.columnVisibility.columns = columns;
      DatatableService.setColumnVisibilitySettings(this.tableKey, this.columnVisibility);
    }

    this.applyFilters = applyFilters.bind(this);
    function applyFilters(stack) {
      const { showActiveStacks, showUnactiveStacks } = this.filters.state;
      if (stack.Orphaned) {
        return stack.OrphanedRunning || this.settings.allOrphanedStacks;
      } else {
        return (stack.Status === 1 && showActiveStacks) || (stack.Status === 2 && showUnactiveStacks) || stack.External || !stack.Status;
      }
    }

    this.onFilterChange = onFilterChange.bind(this);
    function onFilterChange() {
      const { showActiveStacks, showUnactiveStacks } = this.filters.state;
      this.filters.state.enabled = !showActiveStacks || !showUnactiveStacks;
      DatatableService.setDataTableFilters(this.tableKey, this.filters);
    }

    this.$onInit = function () {
      this.setDefaults();
      this.prepareTableFromDataset();

      this.state.orderBy = this.orderBy;
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

      var storedSettings = DatatableService.getDataTableSettings(this.tableKey);
      if (storedSettings !== null) {
        this.settings = storedSettings;
        this.settings.open = false;
      }
      this.onSettingsRepeaterChange();

      var storedColumnVisibility = DatatableService.getColumnVisibilitySettings(this.tableKey);
      if (storedColumnVisibility !== null) {
        this.columnVisibility = storedColumnVisibility;
      }
    };
  },
]);
