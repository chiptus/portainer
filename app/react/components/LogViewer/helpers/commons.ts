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
        date = format(new Date(time * 1000), 'Y/MM/dd hh:mmaa');
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
      if (!index) {
        pushSpan('|', JSONColors.Magenta);
      }

      pushSpan(`${key}=`, JSONColors.Blue, false);
      pushSpan(value, key === 'error' ? JSONColors.Red : JSONColors.Magenta);
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
