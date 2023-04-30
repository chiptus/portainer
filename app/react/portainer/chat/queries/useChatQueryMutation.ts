import { useMutation } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { mutationOptions, withError } from '@/react-tools/react-query';

import { ChatQueryPayload, ChatQueryResponse, ChatResponse } from '../types';
import { chatQueryResponseToChatResponse } from '../chat.converter';

export function useChatQueryMutation() {
  return useMutation(
    sendChatQuery,
    mutationOptions(withError('Unable to send chat query'))
  );
}

async function sendChatQuery(query: ChatQueryPayload): Promise<ChatResponse> {
  try {
    const { data } = await axios.post<ChatQueryResponse>(buildUrl(), query);
    return chatQueryResponseToChatResponse(data);
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}

const BASE_URL = '/chat';
function buildUrl() {
  return BASE_URL;
}
