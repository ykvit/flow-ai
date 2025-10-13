export interface Model {
  modified_at: string;
  name: string;
  size: number;
}

export interface PullStatus {
  name: string;
  status: string;
  digest?: string;
  total?: number;
  completed?: number;
}

export interface ModelDetails {
  modelfile: string;
  parameters: string;
  template: string;
  details: {
    format: string;
    family: string;
    families: string[] | null;
    parameter_size: string;
    quantization_level: string;
  };
}

export interface DeleteModelPayload {
  name: string;
}

export interface PullModelPayload {
  name: string;
  stream?: boolean;
}

export interface ShowModelPayload {
  name: string;
}