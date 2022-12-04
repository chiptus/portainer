import { LogSpanInterface } from '@@/LogViewer/types';

export function LogSpan({ style, text }: LogSpanInterface, index: number) {
  return (
    <span key={index} style={style}>
      {text}
    </span>
  );
}
