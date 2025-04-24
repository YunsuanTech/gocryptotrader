import { Subject, Observable, interval } from 'rxjs';
import { map, share } from 'rxjs/operators';
import websocketService, { WebSocketMessage } from './websocketService';

// WebSocket连接URL
const WEBSOCKET_URL = 'ws://localhost:9051/ws';

// WebSocket响应处理服务
export class WebsocketResponseHandlerService {
  public messages: Subject<WebSocketMessage>;
  public shared: Observable<WebSocketMessage>;
  public isConnected = false;
  private connectionCheckInterval: any = null;

  constructor() {
    // 初始化 messages Subject
    this.messages = new Subject<WebSocketMessage>();

    // 连接 WebSocket 服务
    websocketService
      .connect(WEBSOCKET_URL)
      .pipe(
        map((response: MessageEvent) => {

          const data = response.data as string;

          // 按 }{ 分割多个 JSON 对象
          const jsonStrings = data.split('}{').map((msg, index) => {
            // 为非第一个对象添加 { 前缀
            if (index > 0) msg = '{' + msg;
            // 为非最后一个对象添加 } 后缀
            if (index < data.split('}{').length - 1) msg = msg + '}';
            return msg;
          });

          return jsonStrings;
        })
      )
      .subscribe((jsonStrings: string[]) => {
        // 处理每个分割后的 JSON 字符串
        jsonStrings.forEach(jsonStr => {
          try {
            const websocketResponseMessage = JSON.parse(jsonStr);
            const event = websocketResponseMessage.Event || websocketResponseMessage.event;
            const data = websocketResponseMessage.Data || websocketResponseMessage.data;
            const exchange = websocketResponseMessage.exchange;
            const assetType = websocketResponseMessage.assetType;
            const error = websocketResponseMessage.error;

            const responseMessage: WebSocketMessage = {
              event,
              data,
              exchange,
              assetType,
              error
            };

            console.log('WebSocket消息接收:', responseMessage);
            this.messages.next(responseMessage);
          } catch (e) {
            console.error('JSON parse failed:', e, 'Data:', jsonStr);
            this.messages.next({
                event: 'error', error: 'Invalid JSON received',
                data: undefined
            });
          }
        });
      });

    // 共享观察者，允许多个订阅者
    this.shared = this.messages.pipe(share());

    // 启动连接状态检查
    this.startConnectionCheck();
  }

  /**
   * 开始连接状态检查
   * 每2秒检查一次WebSocket连接状态，并在状态变化时通知
   */
  private startConnectionCheck(): void {
    try {
      this.connectionCheckInterval = interval(10000).subscribe(() => {
        const previousState = this.isConnected;
        this.isConnected = websocketService.isConnected;
        if (previousState !== this.isConnected) {
          console.log(`WebSocket连接状态变更: ${this.isConnected ? '已连接' : '未连接'}`);
          // 当连接状态变化时，发送通知
          this.messages.next({
            event: 'connection_status_change',
            data: this.isConnected,
            error: this.isConnected ? undefined : '连接已断开'
          });
        }
      });
    } catch (error) {
      console.error('启动连接状态检查失败:', error);
    }
  }

  /**
   * 发送认证请求
   * 向服务器发送用户认证信息
   * @param username 用户名
   * @param passwordHash 密码哈希值
   */
  public authenticate(username: string, passwordHash: string): void {
    if (!this.isConnected) {
      console.error('WebSocket未连接，无法发送认证请求');
      return;
    }
    console.log('发送认证请求:', username);
    try {
      websocketService.sendMessage('auth', { Username: username, Password: passwordHash });
    } catch (error) {
      console.error('发送认证请求失败:', error);
      this.messages.next({
        event: 'error',
        error: '发送认证请求失败',
        data: undefined
      });
    }
  }


  /**
   * 关闭 WebSocket 连接
   * 清理资源并关闭与服务器的连接
   */
  public close(): void {
    console.log('关闭WebSocket连接');
    if (this.connectionCheckInterval) {
      this.connectionCheckInterval.unsubscribe();
      this.connectionCheckInterval = null;
    }
    try {
      websocketService.close();
      this.isConnected = false;
    } catch (error) {
      console.error('关闭WebSocket连接失败:', error);
      this.messages.next({
        event: 'error',
        error: '关闭WebSocket连接失败',
        data: undefined
      });
    }
  }
}

// 导出单例实例
const websocketResponseHandlerService = new WebsocketResponseHandlerService();
export default websocketResponseHandlerService;