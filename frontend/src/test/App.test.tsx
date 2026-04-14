import { render, screen } from '@testing-library/react';
import { CssVarsProvider } from '@mui/material/styles';
import theme from '../theme/theme';
import App from '../App';

// Mock the stores to prevent API calls during testing
vi.mock('../stores/chats', () => ({
  useChatStore: vi.fn((selector) => {
    const state = {
      chats: [],
      currentChat: null,
      isLoading: false,
      isStreaming: false,
      streamingContent: '',
      error: null,
      fetchChats: vi.fn(),
      fetchChatById: vi.fn(),
      createMessage: vi.fn(),
      deleteChat: vi.fn(),
      updateChatTitle: vi.fn(),
      setCurrentChat: vi.fn(),
      clearCurrentChat: vi.fn(),
      createNewChat: vi.fn(),
    };
    return selector ? selector(state) : state;
  }),
}));

vi.mock('../stores/settings', () => ({
  useSettingsStore: vi.fn((selector) => {
    const state = {
      settings: { main_model: '', support_model: '', system_prompt: '' },
      isLoading: false,
      error: null,
      isSuccess: false,
      fetchSettings: vi.fn(),
      updateSettings: vi.fn(),
      resetSuccess: vi.fn(),
    };
    return selector ? selector(state) : state;
  }),
}));

vi.mock('../stores/models', () => ({
  useModelsStore: vi.fn((selector) => {
    const state = {
      models: [],
      currentModelDetails: null,
      pullStatus: null,
      isLoading: false,
      error: null,
      fetchModels: vi.fn(),
      deleteModel: vi.fn(),
      showModelInfo: vi.fn(),
      pullModel: vi.fn(),
    };
    return selector ? selector(state) : state;
  }),
}));

describe('App', () => {
  it('renders the new chat button', () => {
    render(
      <CssVarsProvider theme={theme}>
        <App />
      </CssVarsProvider>
    );
    expect(screen.getByText('New Chat')).toBeInTheDocument();
  });

  it('renders the message input', () => {
    render(
      <CssVarsProvider theme={theme}>
        <App />
      </CssVarsProvider>
    );
    expect(screen.getByPlaceholderText('Type a message...')).toBeInTheDocument();
  });
});