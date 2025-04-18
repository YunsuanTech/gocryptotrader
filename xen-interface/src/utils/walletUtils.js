// 钱包工具函数

// 注意：在JavaScript中，我们不需要显式声明全局事件类型
// 这里我们直接使用CustomEvent，浏览器会自动处理

// 存储检测到的钱包提供者
let providers = [];

// 格式化地址显示（显示前5位和后4位）
export const formatAddress = (addr) => {
  if (!addr) return '';
  const upperAfterLastTwo = addr.slice(0, 2) + addr.slice(2);
  return `${upperAfterLastTwo.substring(0, 5)}...${upperAfterLastTwo.substring(39)}`;
};

// 钱包存储服务
export const walletStore = {
  // 获取所有检测到的钱包提供者
  value: () => providers,
  
  // 订阅钱包提供者变化
  subscribe: (callback) => {
    function onAnnouncement(event) {
      // 检查是否已存在相同UUID的提供者
      if (providers.map((p) => p.info.uuid).includes(event.detail.info.uuid))
        return;
      
      // 添加新的提供者
      providers = [...providers, event.detail];
      callback();
    }

    // 监听钱包公告事件
    window.addEventListener("eip6963:announceProvider", onAnnouncement);

    // 请求钱包提供者
    window.dispatchEvent(new Event("eip6963:requestProvider"));

    // 返回清理函数
    return () =>
      window.removeEventListener("eip6963:announceProvider", onAnnouncement);
  },
};

// 连接到选定的钱包提供者
export const connectWithProvider = async (providerWithInfo) => {
  try {
    // 请求用户账户
    const accounts = await providerWithInfo.provider.request({
      method: "eth_requestAccounts"
    });
    
    // 返回连接结果
    return {
      success: true,
      address: accounts?.[0] || '',
      wallet: providerWithInfo
    };
  } catch (error) {
    console.error("连接钱包失败:", error);
    return {
      success: false,
      error: error.message || "连接失败"
    };
  }
};

// 断开钱包连接
export const disconnectWallet = () => {
  // 这里只是简单地返回断开连接的状态
  // 实际应用中可能需要更多的清理工作
  return {
    success: true
  };
};