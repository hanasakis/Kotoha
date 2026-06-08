import apiClient from './client';

export interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
}

export interface ChatRequest {
  message: string;
  history?: ChatMessage[];
}

export interface ChatResponse {
  reply: string;
}

export async function sendChat(request: ChatRequest): Promise<ChatResponse> {
  const { data } = await apiClient.post('/agent/chat', request);
  return data.data;
}
