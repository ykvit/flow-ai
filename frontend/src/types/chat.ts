export interface Chat {
  created_at: string;
  id: string;
  model: string;
  title: string;
  updated_at: string;
  user_id: string;
}

export interface Message {
  content: string;
  id: string;
  metadata: object;
  model: string;
  parent_id: string;
  role: 'user' | 'assistant';
  timestamp: string;
}

export interface ChatWithMessages extends Chat {
  messages: Message[];
}

export interface CreateMessagePayload {
  chat_id: string;
  content: string;
  model: string;
  options?: {
    repeat_penalty?: number;
    seed?: number;
    system?: string;
    temperature?: number;
    top_k?: number;
    top_p?: number;
  };
  support_model?: string;
  system_prompt?: string;
}

export interface UpdateTitlePayload {
  title: string;
}