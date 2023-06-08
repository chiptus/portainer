import { CSSProperties } from 'react';
import { parse } from 'ansicolor';
import stripAnsi from 'strip-ansi';
import toStyle from 'css-to-style';

import { LogInterface } from '@@/LogViewer/types';
import { newBlankLog } from '@@/LogViewer/helpers/commons';

// Format ANSI logs string to log array of spans
export function formatANSILogs(stripedLogs: string) {
  const { spans } = parse(stripedLogs);

  const logs: LogInterface[] = [];

  let log = newBlankLog();

  function pushLog() {
    logs.push(log);
    log = newBlankLog();
  }

  function pushSpan(rawText: string, style: CSSProperties) {
    const text = stripAnsi(rawText);
    log.line += text;
    log.spans.push({ text, style });
  }

  function isLogEmpty() {
    return log.spans.length === 1 && log.spans[0].text === '';
  }

  for (let i = 0; i < spans.length; i += 1) {
    const span = spans[i];
    const style = (toStyle(span.css) as CSSProperties) || {};

    const texts = span.text.split('\n');
    for (let j = 0; j < texts.length; j += 1) {
      if (j) {
        pushLog();
      }

      const text = texts[j];
      pushSpan(text, style);
    }
  }

  if (!isLogEmpty()) pushLog();

  return logs;
}
