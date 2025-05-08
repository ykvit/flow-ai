export interface Message { 
    sender: 'user' | 'assistant';
    text: string;
    id?: string;
  }
  

  export interface Chat { 
      id: string;
      title: string;
      messages: Message[];
      createdAt: Date;
      lastModified: Date;
      model: string;
  }
  
  export interface OllamaTagModel { 
      name: string;
      model: string;
      modified_at: string;
      size: number;
      digest: string;
      details: { /* ... */ };
  }
  
  export interface OllamaMessage {
      role: string;
      content: string;
}
  export interface OllamaChatResponse {
      message: any;
}
  export interface OllamaTagsResponse { models: OllamaTagModel[]; }
  export interface BackendChatsListItem {
      model: any;
      createdAt: string | number | Date;
      id: any;
      title: any;
      lastModified: string | number | Date;
}
  export interface BackendChatsResponse { chats: BackendChatsListItem[]; }
  export interface BackendChatResponse { chat: Chat; }