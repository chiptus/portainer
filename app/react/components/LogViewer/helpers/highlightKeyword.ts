import {LogInterface, LogSpanInterface} from "@@/LogViewer/types";

const BEGIN = 'B';
const MID = 'M';
const END = 'E';

function getMaskKeyword(keyword: string) {
  const len = keyword.length;

  const begin = len >= 2 ? BEGIN : '';
  const mid = MID.repeat(len >= 2 ? len - 2 : len - 1);

  return `${begin}${mid}${END}`
}

const KeywordStyle1 = {
  color: 'black',
  backgroundColor: '#f0ab01',
}

const KeywordStyle2 = {
  color: 'black',
  backgroundColor: '#f9e0a2',
}

export function highlightKeyword(log: LogInterface, keyword: string, focusedKeywordIndexInLine: number) {
  if (!keyword.length) {
    return log.spans;
  }

  let keywordCount = 0;

  const maskKeyword = getMaskKeyword(keyword);

  let maskLine = log.line.toLowerCase().replaceAll(keyword.toLowerCase(), maskKeyword);

  const newSpans: LogSpanInterface[] = [];
  let newSpan: LogSpanInterface = {
    text: '',
    style: {},
  };

  function pushNewSpan() {
    if (newSpan.text) {
      newSpans.push(newSpan);
    }
  }

  function newKeywordSpan(char: string) {
    pushNewSpan();

    newSpan = {
      text: char,
      style: keywordCount === focusedKeywordIndexInLine ? KeywordStyle1 : KeywordStyle2,
    }

    keywordCount += 1;
  }

  function pushKeywordChar(char: string) {
    newSpan.text += char;
  }

  function endKeywordSpan(char: string) {
    if (keyword.length === 1) {
      // no BEGIN in maskLine
      newKeywordSpan(char);
    } else {
      pushKeywordChar(char);
    }
    pushNewSpan();
    newSpan = {text: '', style:{}};
  }

  function pushNormalChar(char: string, span: LogSpanInterface) {
    const newText = newSpan.text + char;
    newSpan = {...span, text: newText}
  }

  for (let i = 0; i < log.spans.length; i += 1) {
    const span = log.spans[i];
    const text = String(span.text);

    for (let j = 0; j < text.length; j += 1) {
      const char = text.charAt(j);
      const maskChar = maskLine.charAt(0)
      maskLine = maskLine.substring(1);

      if (maskChar === BEGIN) {
        newKeywordSpan(char);
      } else if (maskChar === MID) {
        pushKeywordChar(char);
      } else if (maskChar === END) {
        endKeywordSpan(char);
      } else {
        pushNormalChar(char, span);
      }
    }

    if (newSpan.style !== KeywordStyle1 && newSpan.style !== KeywordStyle2) {
      pushNewSpan();
      newSpan = {text: '', style:{}};
    }
  }

  return newSpans;
}
