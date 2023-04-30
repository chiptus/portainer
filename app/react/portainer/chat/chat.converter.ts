import { ChatQueryResponse, ChatResponse } from './types';

export function chatQueryResponseToChatResponse(
  apiRes: ChatQueryResponse
): ChatResponse {
  return {
    message: apiRes.message,
    yaml: apiRes.yaml,
  };
}
