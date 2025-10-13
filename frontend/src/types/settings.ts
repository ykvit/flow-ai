export interface Settings {
  main_model: string;
  support_model: string;
  system_prompt: string;
}

export type UpdateSettingsPayload = Settings;