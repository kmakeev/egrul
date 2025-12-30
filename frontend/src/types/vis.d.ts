// Типы для vis.js библиотек
declare module 'vis-network' {
  export interface Node {
    id: string | number;
    label?: string;
    title?: string;
    color?: {
      background?: string;
      border?: string;
      highlight?: {
        background?: string;
        border?: string;
      };
    };
    font?: {
      color?: string;
      size?: number;
      face?: string;
    };
    size?: number;
    shape?: string;
    margin?: number;
    borderWidth?: number;
    chosen?: {
      node?: (values: Record<string, unknown>) => void;
    };
  }

  export interface Edge {
    id: string | number;
    from: string | number;
    to: string | number;
    label?: string;
    color?: {
      color?: string;
      highlight?: string;
      hover?: string;
    };
    font?: {
      color?: string;
      size?: number;
      face?: string;
      strokeWidth?: number;
      strokeColor?: string;
    };
    width?: number;
    arrows?: {
      to?: {
        enabled?: boolean;
        scaleFactor?: number;
      };
      from?: {
        enabled?: boolean;
        scaleFactor?: number;
      };
    };
    smooth?: {
      enabled?: boolean;
      type?: string;
      roundness?: number;
    };
    chosen?: {
      edge?: (values: Record<string, unknown>) => void;
    };
  }

  export interface NetworkOptions {
    nodes?: {
      borderWidth?: number;
      shadow?: {
        enabled?: boolean;
        color?: string;
        size?: number;
        x?: number;
        y?: number;
      };
    };
    edges?: {
      shadow?: {
        enabled?: boolean;
        color?: string;
        size?: number;
        x?: number;
        y?: number;
      };
      smooth?: {
        enabled?: boolean;
        type?: string;
        roundness?: number;
      };
    };
    physics?: {
      enabled?: boolean;
      stabilization?: {
        enabled?: boolean;
        iterations?: number;
        updateInterval?: number;
      };
      barnesHut?: {
        gravitationalConstant?: number;
        centralGravity?: number;
        springLength?: number;
        springConstant?: number;
        damping?: number;
        avoidOverlap?: number;
      };
    };
    interaction?: {
      hover?: boolean;
      hoverConnectedEdges?: boolean;
      selectConnectedEdges?: boolean;
      tooltipDelay?: number;
      zoomView?: boolean;
      dragView?: boolean;
    };
    layout?: {
      improvedLayout?: boolean;
      clusterThreshold?: number;
    };
  }

  export interface NetworkData {
    nodes: DataSet;
    edges: DataSet;
  }

  export class Network {
    constructor(container: HTMLElement, data: NetworkData, options?: NetworkOptions);
    
    on(event: string, callback: (params: Record<string, unknown>) => void): void;
    once(event: string, callback: () => void): void;
    off(event: string, callback: (params: Record<string, unknown>) => void): void;
    
    destroy(): void;
    setData(data: NetworkData): void;
    
    getScale(): number;
    moveTo(options: {
      position?: { x: number; y: number };
      scale?: number;
      offset?: { x: number; y: number };
      animation?: {
        duration: number;
        easingFunction: string;
      };
    }): void;
    
    fit(options?: {
      nodes?: string[];
      animation?: {
        duration: number;
        easingFunction: string;
      };
    }): void;
    
    body: {
      data: {
        nodes: DataSet;
        edges: DataSet;
      };
    };
  }
}

declare module 'vis-data' {
  export class DataSet<T = Record<string, unknown>> {
    constructor(data?: T[]);
    
    add(data: T | T[]): void;
    update(data: T | T[]): void;
    remove(id: string | number | string[] | number[]): void;
    clear(): void;
    
    get(options?: Record<string, unknown>): T[];
    getIds(): (string | number)[];
    
    on(event: string, callback: (event: string, properties: Record<string, unknown>, senderId?: string) => void): void;
    off(event: string, callback: (event: string, properties: Record<string, unknown>, senderId?: string) => void): void;
  }
}