import { LevelSpansInterface } from '@@/LogViewer/types';

export const JSONColors = {
  Grey: 'var(--text-log-viewer-color-json-grey)',
  Magenta: 'var(--text-log-viewer-color-json-magenta)',
  Yellow: 'var(--text-log-viewer-color-json-yellow)',
  Green: 'var(--text-log-viewer-color-json-green)',
  Red: 'var(--text-log-viewer-color-json-red)',
  Blue: 'var(--text-log-viewer-color-json-blue)',
};

const DebugSpan = {
  style: {
    fontWeight: 'bold',
    color: JSONColors.Grey,
  },
  text: 'DBG',
};

const InfoSpan = {
  style: {
    fontWeight: 'bold',
    color: JSONColors.Green,
  },
  text: 'INF',
};

const WarnSpan = {
  style: {
    fontWeight: 'bold',
    color: JSONColors.Yellow,
  },
  text: 'WRN',
};

const ErrorSpan = {
  style: {
    fontWeight: 'bold',
    color: JSONColors.Red,
  },
  text: 'ERR',
};

export const LevelSpans: LevelSpansInterface = {
  debug: DebugSpan,
  DBG: DebugSpan,
  info: InfoSpan,
  INF: InfoSpan,
  warn: WarnSpan,
  WRN: WarnSpan,
  error: ErrorSpan,
  ERR: ErrorSpan,
};

export const SpaceSpan = { style: {}, text: ' ' };

export const TIMESTAMP_LENGTH = 30;

const MAC_OR_LIN = navigator.userAgent.indexOf('Mac') > -1 ? 'mac' : 'lin';
const BROWSER_OS_PLATFORM =
  navigator.userAgent.indexOf('Windows') > -1 ? 'win' : MAC_OR_LIN;
export const NEW_LINE_BREAKER = BROWSER_OS_PLATFORM === 'win' ? '\r\n' : '\n';
