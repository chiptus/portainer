import { useContext } from 'react';
import { RefreshCw } from 'lucide-react';

import { Button } from '@@/buttons';
import { TableTitle } from '@@/datatables';
import { SearchBar } from '@@/datatables/SearchBar';
import {
  LogViewerContext,
  LogViewerContextInterface,
} from '@@/LogViewer/types';
import { Checkbox } from '@@/form-components/Checkbox';
import { SwitchField } from '@@/form-components/SwitchField';
import { Input, Select } from '@@/form-components/Input';
import { SinceOptions } from '@@/LogViewer/LogController/data';
import { FullScreenButton } from '@@/LogViewer/LogController/FullScreenButton/FullScreenButton';
import { CopyLogsButton } from '@@/LogViewer/LogController/CopyLogsButton/CopyLogsButton';
import { DownloadLogsButton } from '@@/LogViewer/LogController/DownloadLogsButton/DownloadLogsButton';
import { SearchStatus } from '@@/LogViewer/LogController/SearchStatus/SearchStatus';

import './LogController.css';

export function LogController() {
  const {
    searchStatus,
    controllerStates,
    hideFetch,
    hideLines,
    hideTimestamp,
    logsQuery,
  } = useContext(LogViewerContext) as LogViewerContextInterface;

  const {
    keyword,
    setKeyword,
    filter,
    setFilter,
    autoRefresh,
    setAutoRefresh,
    since,
    setSince,
    tail,
    setTail,
    showTimestamp,
    setShowTimestamp,
    wrapLine,
    setWrapLine,
    showLineNumbers,
    setShowLineNumbers,
  } = controllerStates;

  return (
    <>
      <TableTitle icon="file" label="Logs">
        <div className="tool-bar-segment">
          <SearchBar value={keyword} onChange={setKeyword}>
            {controllerStates.keyword && (
              <SearchStatus searchIndicator={searchStatus} />
            )}
          </SearchBar>

          <div className="vertical-center">
            <Checkbox
              id="filter"
              checked={filter}
              onChange={(e) => setFilter(e.target.checked)}
            />
            <span>Filter search results</span>
          </div>

          <CopyLogsButton />

          <DownloadLogsButton />
        </div>
      </TableTitle>

      <div className="tool-bar">
        <div className="tool-bar-segment">
          <div className="vertical-center">
            <SwitchField
              label=""
              checked={autoRefresh}
              onChange={setAutoRefresh}
            />
            <span>Auto refresh</span>
          </div>

          <div className="vertical-center">
            <Button
              onClick={() => logsQuery.refetch()}
              disabled={autoRefresh || logsQuery.isFetching}
              color="none"
              icon={RefreshCw}
              title="Refresh"
            />
          </div>

          {!hideFetch && (
            <div className="vertical-center">
              <span>Fetch</span>
              <Select
                value={since}
                onChange={(e) => {
                  setSince(parseInt(e.currentTarget.value, 10));
                }}
                options={SinceOptions}
              />
            </div>
          )}

          {!hideLines && (
            <div className="vertical-center">
              <span>Lines</span>
              <Input
                type="number"
                value={tail}
                min={1}
                onChange={(e) => {
                  const t = parseInt(e.target.value, 10);
                  setTail(t || '');
                }}
                className="line-number"
              />
            </div>
          )}

          {!hideTimestamp && (
            <div className="vertical-center">
              <Checkbox
                id="showTimestamp"
                checked={showTimestamp}
                onChange={(e) => setShowTimestamp(e.target.checked)}
              />
              <span>Show timestamp</span>
            </div>
          )}
        </div>

        <div className="tool-bar-segment">
          <div className="vertical-center">
            <Checkbox
              id="showLineNumber"
              checked={showLineNumbers}
              onChange={(e) => setShowLineNumbers(e.target.checked)}
            />
            <span>Show line numbers</span>
          </div>

          <div className="vertical-center">
            <Checkbox
              id="wrapLine"
              checked={wrapLine}
              onChange={(e) => setWrapLine(e.target.checked)}
            />
            <span>Wrap line</span>
          </div>

          <FullScreenButton />
        </div>
      </div>
    </>
  );
}
