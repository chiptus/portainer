import { JSONStackTrace, LogInterface, Pair } from '@@/LogViewer/types';
import { newBlankLog, newLogService } from '@@/LogViewer/helpers/commons';
import { JSONColors } from '@@/LogViewer/helpers/consts';

// extractStackTraceLog returns the stack trace JSON object in pairs if
// there is any, and remove the stack_trace item from pairs.
// Otherwise, extractStackTraceLog returns null
export function extractStackTraceLog(timestamp: string, pairs: Pair[]) {
  const stackTraceIndex = pairs.findIndex((p) => p[0] === 'stack_trace');

  if (stackTraceIndex !== -1) {
    try {
      const stackTraceJson = JSON.parse(pairs[stackTraceIndex][1]);
      pairs.splice(stackTraceIndex);
      return stackTraceJson;
    } catch {
      return null;
    }
  }

  return null;
}

export function formatStackTraceLog(
  newLogs: LogInterface[],
  timestamp: string,
  stackTraceJson: JSONStackTrace | undefined
) {
  if (stackTraceJson) {
    stackTraceJson.forEach(({ func, line: lineNumber, source }) => {
      const log = newBlankLog();
      const logService = newLogService(log);

      if (timestamp) {
        logService.pushSpan(timestamp);
      }

      for (let i = 0; i < 4; i += 1) {
        logService.pushSpaceSpan();
      }

      logService.pushSpan('at', JSONColors.Grey);
      logService.pushSpan(func, JSONColors.Red);
      logService.pushSpan(`(${source}:${lineNumber})`, JSONColors.Grey);

      newLogs.push(log);
    });
  }
}
