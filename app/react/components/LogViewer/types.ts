import { CSSProperties, RefObject } from 'react';
import { UseQueryResult } from 'react-query';

import { SearchStatusInterface } from '@@/LogViewer/LogController/SearchStatus/SearchStatus';

export type TailType = number | '';

export type LogViewerRefType = RefObject<HTMLDivElement>;

export interface GetLogsParamsInterface {
  stdout?: boolean;
  stderr?: boolean;
  timestamps?: boolean;
  since?: number;
  sinceSeconds?: number;
  tail?: TailType;
  container?: string;
}

export interface RawLogsSection {
  sectionName?: string;
  sectionNameColor?: string;
  logs: string;
}

export type RawLogsInterface = string | RawLogsSection[];

export type GetLogsFnType = (
  getLogsParams: GetLogsParamsInterface
) => Promise<RawLogsInterface>;

export interface LogSpanInterface {
  text: string;
  style: CSSProperties;
}

export interface LogInterface {
  line: string;
  lowerLine: string;
  lineNumber?: number;
  numOfKeywords?: number;
  firstKeywordIndex?: number;
  spans: LogSpanInterface[];
}

export interface KeywordIndexInterface {
  lineNumber: number;
  indexInLine: number;
}

export interface ProcessedLogsInterface {
  logs: LogInterface[];
  keywordIndexes: KeywordIndexInterface[];
  totalKeywords: number;
}

export interface ControllerStatesInterface {
  keyword: string;
  setKeyword: (input: string) => void;
  filter: boolean;
  setFilter: (value: boolean) => void;
  autoRefresh: boolean;
  setAutoRefresh: (value: boolean) => void;
  since: number;
  setSince: (value: number) => void;
  tail: TailType;
  setTail: (value: TailType) => void;
  showTimestamp: boolean;
  setShowTimestamp: (value: boolean) => void;
  wrapLine: boolean;
  setWrapLine: (value: boolean) => void;
  showLineNumbers: boolean;
  setShowLineNumbers: (value: boolean) => void;
}

export interface LogViewerContextInterface {
  hideFetch?: boolean;
  hideLines?: boolean;
  hideTimestamp?: boolean;
  controllerStates: ControllerStatesInterface;
  logViewerRef: LogViewerRefType;
  logs: ProcessedLogsInterface;
  visibleStartIndex: number;
  setVisibleStartIndex: (value: number) => void;
  logsQuery: UseQueryResult;
  searchStatus: SearchStatusInterface;
  resourceName: string;
}

export interface LevelSpansInterface {
  [key: string]: LogSpanInterface;
}

export type JSONStackTrace = {
  func: string;
  line: string;
  source: string;
}[];

export type JSONLog = {
  [k: string]: unknown;
  time: number;
  level: string;
  caller: string;
  message: string;
  stack_trace?: JSONStackTrace;
};

export type Pair = [string, string];
