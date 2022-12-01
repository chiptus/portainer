import { JSONLog, LogInterface, Pair } from '@@/LogViewer/types';
import { newBlankLog, newLogService } from '@@/LogViewer/helpers/commons';
import { JSONColors } from '@@/LogViewer/helpers/consts';
import { formatStackTraceLog } from '@@/LogViewer/helpers/prettifyLogs/helpers/stackTrace';

function formatMainLog(
  newLogs: LogInterface[],
  timestamp: string,
  json: JSONLog
) {
  const log = newBlankLog();
  const logService = newLogService(log);

  const {
    level,
    caller,
    message,
    time,
    stack_trace: stackTrace,
    ...restJson
  } = json;

  if (timestamp) {
    logService.pushSpan(timestamp);
  }

  logService.pushTimeSpan(time);
  logService.pushLevelSpan(level);
  logService.pushCallerSpan(caller);
  logService.pushSpan(message, JSONColors.Magenta);
  logService.pushPairs(Object.entries(restJson) as Pair[]);

  newLogs.push(log);
}

export function prettifyJSONLog(
  newLogs: LogInterface[],
  line: string,
  timestamp: string
) {
  let json: JSONLog;

  try {
    json = JSON.parse(line);
  } catch {
    return false;
  }

  formatMainLog(newLogs, timestamp, json);
  formatStackTraceLog(newLogs, timestamp, json.stack_trace);
  return true;
}
