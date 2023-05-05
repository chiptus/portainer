import { format } from 'date-fns';
import { takeRight } from 'lodash';
import { CSSProperties } from 'react';

import { LogInterface, LogSpanInterface, Pair } from '@@/LogViewer/types';
import { JSONColors, LevelSpans, SpaceSpan } from '@@/LogViewer/helpers/consts';

export function newBlankLog() {
  const log: LogInterface = {
    line: '',
    lowerLine: '',
    spans: [],
  };

  return log;
}

export function newLogService(log: LogInterface) {
  const newLog = log;

  function pushSpaceSpan() {
    newLog.line += SpaceSpan.text;
    newLog.spans.push(SpaceSpan);
  }

  function pushSpan(text: unknown, color = '', space = true, bold = false) {
    const textStr = String(text);

    if (textStr) {
      newLog.line += String(textStr);

      const style: CSSProperties = { color };
      if (bold) {
        style.fontWeight = 'bold';
      }
      const span: LogSpanInterface = { text: textStr, style };

      newLog.spans.push(span);
    }

    if (space) {
      pushSpaceSpan();
    }
  }

  function pushLevelSpan(level: string) {
    if (level) {
      const span = LevelSpans[level] || { text: level, style: {} };
      newLog.line += span.text;
      newLog.spans.push(span);
      pushSpaceSpan();
    }
  }

  function pushTimeSpan(time: number | string) {
    if (time) {
      let date = time;
      if (typeof time === 'number') {
        // time is a number, so it is the number of seconds OR milliseconds since Unix Epoch (1970-01-01T00:00:00.000Z)
        // we need to know if time's unit is second or millisecond
        // new Date(Date.now()*1000).getUTCFullYear() > new Date()
        // 253402214400 is the numer of seconds between Unix Epoch and 9999-12-31T00:00:00.000Z
        // if time is greater than 253402214400, then time unit cannot be second, so it is millisecond
        const timestampInMilliseconds =
          time > 253402214400 ? time : time * 1000;
        date = format(new Date(timestampInMilliseconds), 'Y/MM/dd hh:mmaa');
      }

      pushSpan(date, JSONColors.Grey);
    }
  }

  function pushCallerSpan(caller: string) {
    if (caller) {
      const trimmedCaller = takeRight(caller.split('/'), 2).join('/');
      pushSpan(trimmedCaller, JSONColors.Magenta, true, true);
      pushSpan('>', JSONColors.Blue);
    }
  }

  function pushPairs(pairs: Pair[]) {
    pairs.forEach(([key, value], index) => {
      const strValue =
        typeof value === 'string' ? value : JSON.stringify(value);

      if (!index && newLog.line) {
        pushSpan('|', JSONColors.Magenta);
      }

      pushSpan(`${key}=`, JSONColors.Blue, false);
      pushSpan(strValue, key === 'error' ? JSONColors.Red : JSONColors.Magenta);
    });
  }

  return {
    pushSpaceSpan,
    pushSpan,
    pushLevelSpan,
    pushTimeSpan,
    pushCallerSpan,
    pushPairs,
  };
}

export function startsWithTimestamp(line: string) {
  // timestamp example: 2022-09-26T21:57:30.297058516Z
  return !!line.match(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/);
}
