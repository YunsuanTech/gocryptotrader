import { Subject, Observable } from 'rxjs';

// WebSocket消息接口
export interface WebSocketMessage {
  event: string;
  data: any;
  exchange?: string;
  assetType?: string;
  error?: string;
}

// WebSocket服务类
export class WebsocketService {
  private socket: WebSocket | null = null;
  private subject: Subject<MessageEvent> | null = null;
  public isConnected = false;

  // 连接到WebSocket服务器
  public connect(url: string): Subject<MessageEvent> {
    // 关闭现有的连接（如果存在）
    this.close();
    this.subject = this.createWebSocketSubject(url);
    return this.subject;
  }

  // 创建WebSocket连接
  private createWebSocketSubject(url: string): Subject<MessageEvent> {
    this.socket = new WebSocket(url);

    // 创建一个观察者
    const observable = new Observable<MessageEvent>(observer => {
      this.socket!.onmessage = observer.next.bind(observer);
      this.socket!.onerror = observer.error.bind(observer);
      this.socket!.onclose = () => {
        this.isConnected = false;
        observer.complete();
      };
      this.socket!.onopen = () => {
        this.isConnected = true;
      };

      // 清理函数
      return () => {
        if (this.socket) {
          this.socket.close();
          this.socket = null;
          this.isConnected = false;
        }
      };
    });

    // 创建一个Subject
    const subject = new Subject<MessageEvent>();
    
    // 订阅observable
    observable.subscribe(subject);
    
    return subject;
  }

  // 发送消息到WebSocket服务器
  public sendMessage(event: string, data: any = {}): void {
    if (!event || event.trim() === '') {
      console.error('WebSocket错误: 事件名称不能为空');
      return;
    }
    
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      const message = {
        Event: event,
        Data: data
      };
      this.socket.send(JSON.stringify(message));
      console.log(`发送WebSocket消息: ${event}`, data);
    } else {
      console.error('WebSocket未连接，无法发送消息');
      if (this.subject) {
        this.subject.error(new Error('WebSocket未连接，无法发送消息'));
      }
    }
  }

  // 关闭WebSocket连接
  public close(): void {
    if (this.socket) {
      this.socket.close();
      this.socket = null;
      this.isConnected = false;
    }
  }
}

// 导出单例实例
const websocketService = new WebsocketService();
export default websocketService;