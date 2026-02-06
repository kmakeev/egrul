/**
 * SSE Client для получения real-time уведомлений от Notification Hub
 */

export interface NotificationEvent {
  id: string;
  type: string;
  entity_type: string;
  entity_id: string;
  entity_name: string;
  change_type: string;
  field_name: string;
  old_value: string;
  new_value: string;
  is_significant: boolean;
  timestamp: string;
  region_code: string;
}

type EventHandler = (event: NotificationEvent) => void;
type ConnectionHandler = () => void;
type ErrorHandler = (error: Event) => void;

interface SSEClientOptions {
  token: string;
  onMessage?: EventHandler;
  onConnected?: ConnectionHandler;
  onDisconnected?: ConnectionHandler;
  onError?: ErrorHandler;
  reconnectDelay?: number; // Initial delay in ms
  maxReconnectDelay?: number; // Max delay in ms
  heartbeatTimeout?: number; // Timeout in ms
}

export class SSEClient {
  private eventSource: EventSource | null = null;
  private token: string;
  private reconnectAttempts = 0;
  private reconnectTimer: NodeJS.Timeout | null = null;
  private heartbeatTimer: NodeJS.Timeout | null = null;
  private lastEventId: string | null = null;
  private isIntentionallyClosed = false;

  private readonly baseUrl: string;
  private readonly reconnectDelay: number;
  private readonly maxReconnectDelay: number;
  private readonly heartbeatTimeout: number;

  private onMessage?: EventHandler;
  private onConnected?: ConnectionHandler;
  private onDisconnected?: ConnectionHandler;
  private onError?: ErrorHandler;

  constructor(options: SSEClientOptions) {
    this.token = options.token;
    this.onMessage = options.onMessage;
    this.onConnected = options.onConnected;
    this.onDisconnected = options.onDisconnected;
    this.onError = options.onError;

    // Для SSE используем базовый URL напрямую (без /api/v1)
    this.baseUrl = 'http://localhost:8080';
    this.reconnectDelay = options.reconnectDelay || 1000;
    this.maxReconnectDelay = options.maxReconnectDelay || 30000;
    this.heartbeatTimeout = options.heartbeatTimeout || 60000;
  }

  /**
   * Подключиться к SSE stream
   */
  public connect(): void {
    if (this.eventSource) {
      console.warn('[SSEClient] Already connected');
      return;
    }

    this.isIntentionallyClosed = false;

    try {
      // Создаем URL с токеном в query параметре
      const url = new URL(`${this.baseUrl}/notifications/stream`);
      url.searchParams.set('token', this.token);
      if (this.lastEventId) {
        url.searchParams.set('lastEventId', this.lastEventId);
      }

      this.eventSource = new EventSource(url.toString());

      // Обработчик успешного подключения
      this.eventSource.addEventListener('open', () => {
        this.reconnectAttempts = 0;
        this.resetHeartbeatTimer();
        this.onConnected?.();
      });

      // Обработчик сообщений (основной канал)
      this.eventSource.addEventListener('message', (event: MessageEvent) => {
        this.handleMessage(event);
      });

      // Обработчик событий change_detected
      this.eventSource.addEventListener('change_detected', (event: MessageEvent) => {
        this.handleMessage(event);
      });

      // Обработчик события connected
      this.eventSource.addEventListener('connected', (_event: MessageEvent) => {
        this.resetHeartbeatTimer();
      });

      // Обработчик ошибок
      this.eventSource.addEventListener('error', (_error: Event) => {
        console.error('[SSEClient] Connection error');

        // EventSource автоматически переподключается, но мы можем обработать это
        if (this.eventSource?.readyState === EventSource.CLOSED) {
          this.handleDisconnect();
        }

        this.onError?.(_error);
      });

    } catch (error) {
      console.error('[SSEClient] Failed to create EventSource:', error);
      this.scheduleReconnect();
    }
  }

  /**
   * Отключиться от SSE stream
   */
  public disconnect(): void {
    this.isIntentionallyClosed = true;

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.heartbeatTimer) {
      clearTimeout(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }

    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }

    this.onDisconnected?.();
  }

  /**
   * Обновить токен (переподключиться с новым токеном)
   */
  public updateToken(newToken: string): void {
    if (this.token === newToken) {
      return;
    }

    this.token = newToken;

    // Переподключиться с новым токеном
    const wasConnected = this.eventSource !== null;
    this.disconnect();
    if (wasConnected) {
      this.connect();
    }
  }

  /**
   * Проверить, подключен ли клиент
   */
  public isConnected(): boolean {
    return this.eventSource?.readyState === EventSource.OPEN;
  }

  // Приватные методы

  private handleMessage(event: MessageEvent): void {
    this.resetHeartbeatTimer();

    // Сохранить lastEventId для восстановления соединения
    if (event.lastEventId) {
      this.lastEventId = event.lastEventId;
    }

    try {
      const data = JSON.parse(event.data) as NotificationEvent;

      // Игнорировать heartbeat и служебные события
      if (data.type === 'connected' || data.type === 'heartbeat') {
        return;
      }

      this.onMessage?.(data);
    } catch (error) {
      console.error('[SSEClient] Failed to parse message:', error, event.data);
    }
  }

  private handleDisconnect(): void {
    this.eventSource?.close();
    this.eventSource = null;

    if (this.heartbeatTimer) {
      clearTimeout(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }

    this.onDisconnected?.();

    // Переподключиться, если не закрыто намеренно
    if (!this.isIntentionallyClosed) {
      this.scheduleReconnect();
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer || this.isIntentionallyClosed) {
      return;
    }

    // Exponential backoff
    const delay = Math.min(
      this.reconnectDelay * Math.pow(2, this.reconnectAttempts),
      this.maxReconnectDelay
    );

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.reconnectAttempts++;
      this.connect();
    }, delay);
  }

  private resetHeartbeatTimer(): void {
    if (this.heartbeatTimer) {
      clearTimeout(this.heartbeatTimer);
    }

    this.heartbeatTimer = setTimeout(() => {
      console.warn('[SSEClient] Heartbeat timeout - connection may be stale');
      // Переподключиться при отсутствии heartbeat
      this.handleDisconnect();
    }, this.heartbeatTimeout);
  }
}
