import { LogInterface, Pair } from '@@/LogViewer/types';
import { newBlankLog, newLogService } from '@@/LogViewer/helpers/commons';
import { JSONColors } from '@@/LogViewer/helpers/consts';
import {
  extractStackTraceLog,
  formatStackTraceLog,
} from '@@/LogViewer/helpers/prettifyLogs/helpers/stackTrace';

const dateRegex = /(\d{4}\/\d{2}\/\d{2} \d{2}:\d{2}[AP]M) /; // "2022/02/01 04:30AM "
const levelRegex = /(\w{3}) /; // "INF " or "ERR "
const callerRegex = /(.+?.go:\d+) /; // "path/to/file.go:line "
const chevRegex = /> /; // "> "
const messageAndPairsRegex = /(.*)/; // include the rest of the string in a separate group

const keyRegex = /(\S+=)/g; // ""

export const ZerologRegex = concatRegex(
  dateRegex,
  levelRegex,
  callerRegex,
  chevRegex,
  messageAndPairsRegex
);

function concatRegex(...regs: RegExp[]) {
  const flags = Array.from(
    new Set(
      regs
        .map((r) => r.flags)
        .join('')
        .split('')
    )
  ).join('');
  const source = regs.map((r) => r.source).join('');
  return new RegExp(source, flags);
}

function extractPairs(messageAndPairs: string): [string, Pair[]] {
  const pairs: Pair[] = [];
  let [message, rawPairs] = messageAndPairs.split('|');

  if (!messageAndPairs.includes('|') && !rawPairs) {
    rawPairs = message;
    message = '';
  }
  message = message.trim();
  rawPairs = rawPairs.trim();

  const matches = [...rawPairs.matchAll(keyRegex)];

  matches.forEach((m, idx) => {
    const rawKey = m[0];
    const key = rawKey.slice(0, -1);
    const start = m.index || 0;
    const end = idx !== matches.length - 1 ? matches[idx + 1].index : undefined;
    const value = (
      end
        ? rawPairs.slice(start + rawKey.length, end)
        : rawPairs.slice(start + rawKey.length)
    ).trim();
    pairs.push([key, value]);
  });

  return [message, pairs];
}

function doPrettifyZeroLog(
  newLogs: LogInterface[],
  line: string,
  timestamp: string
) {
  const log = newBlankLog();
  const logService = newLogService(log);

  if (timestamp) {
    logService.pushSpan(timestamp);
  }

  const [, time, level, caller, messageAndPairs] =
    line.match(ZerologRegex) || [];
  const [message, pairs] = extractPairs(messageAndPairs);

  const stackTraceJson = extractStackTraceLog(timestamp, pairs);

  logService.pushTimeSpan(time);
  logService.pushLevelSpan(level);
  logService.pushCallerSpan(caller);
  logService.pushSpan(message, JSONColors.Magenta);
  logService.pushPairs(pairs);
  newLogs.push(log);

  if (stackTraceJson) {
    formatStackTraceLog(newLogs, timestamp, stackTraceJson);
  }
}

export function prettifyZeroLog(
  newLogs: LogInterface[],
  line: string,
  timestamp: string
): boolean {
  if (ZerologRegex.test(line)) {
    doPrettifyZeroLog(newLogs, line, timestamp);
    return true;
  }

  return false;
}
