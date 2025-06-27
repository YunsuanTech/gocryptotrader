// Solana钱包工具函数

// 格式化地址显示（显示前5位和后4位）
export const formatAddress = (addr) => {
  if (!addr) return '';
  return `${addr.substring(0, 5)}...${addr.substring(addr.length - 4)}`;
};

// 检查Phantom钱包是否已安装
export const checkPhantomWalletInstalled = () => {
  const provider = window?.phantom?.solana;
  return provider && provider.isPhantom;
};

// 钱包存储服务 - 仅支持Phantom钱包
export const walletStore = {
  // 获取Phantom钱包提供者
  value: () => {
    if (checkPhantomWalletInstalled()) {
      return [{
        provider: window.phantom.solana,
        info: {
          name: 'Phantom',
          icon: "data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMTA4IiBoZWlnaHQ9IjEwOCIgdmlld0JveD0iMCAwIDEwOCAxMDgiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+CjxyZWN0IHdpZHRoPSIxMDgiIGhlaWdodD0iMTA4IiByeD0iMjYiIGZpbGw9IiNBQjlGRjIiLz4KPHBhdGggZmlsbC1ydWxlPSJldmVub2RkIiBjbGlwLXJ1bGU9ImV2ZW5vZGQiIGQ9Ik00Ni41MjY3IDY5LjkyMjlDNDIuMDA1NCA3Ni44NTA5IDM0LjQyOTIgODUuNjE4MiAyNC4zNDggODUuNjE4MkMxOS41ODI0IDg1LjYxODIgMTUgODMuNjU2MyAxNSA3NS4xMzQyQzE1IDUzLjQzMDUgNDQuNjMyNiAxOS44MzI3IDcyLjEyNjggMTkuODMyN0M4Ny43NjggMTkuODMyNyA5NCAzMC42ODQ2IDk0IDQzLjAwNzlDOTQgNTguODI1OCA4My43MzU1IDc2LjkxMjIgNzMuNTMyMSA3Ni45MTIyQzcwLjI5MzkgNzYuOTEyMiA2OC43MDUzIDc1LjEzNDIgNjguNzA1MyA3Mi4zMTRDNjguNzA1MyA3MS41NzgzIDY4LjgyNzUgNzAuNzgxMiA2OS4wNzE5IDY5LjkyMjlDNjUuNTg5MyA3NS44Njk5IDU4Ljg2ODUgODEuMzg3OCA1Mi41NzU0IDgxLjM4NzhDNDcuOTkzIDgxLjM4NzggNDUuNjcxMyA3OC41MDYzIDQ1LjY3MTMgNzQuNDU5OEM0NS42NzEzIDcyLjk4ODQgNDUuOTc2OCA3MS40NTU2IDQ2LjUyNjcgNjkuOTIyOVpNODMuNjc2MSA0Mi41Nzk0QzgzLjY3NjEgNDYuMTcwNCA4MS41NTc1IDQ3Ljk2NTggNzkuMTg3NSA0Ny45NjU4Qzc2Ljc4MTYgNDcuOTY1OCA3NC42OTg5IDQ2LjE3MDQgNzQuNjk4OSA0Mi41Nzk0Qzc0LjY5ODkgMzguOTg4NSA3Ni43ODE2IDM3LjE5MzEgNzkuMTg3NSAzNy4xOTMxQzgxLjU1NzUgMzcuMTkzMSA4My42NzYxIDM4Ljk4ODUgODMuNjc2MSA0Mi41Nzk0Wk03MC4yMTAzIDQyLjU3OTVDNzAuMjEwMyA0Ni4xNzA0IDY4LjA5MTYgNDcuOTY1OCA2NS43MjE2IDQ3Ljk2NThDNjMuMzE1NyA0Ny45NjU4IDYxLjIzMyA0Ni4xNzA0IDYxLjIzMyA0Mi41Nzk1QzYxLjIzMyAzOC45ODg1IDYzLjMxNTcgMzcuMTkzMSA2NS43MjE2IDM3LjE5MzFDNjguMDkxNiAzNy4xOTMxIDcwLjIxMDMgMzguOTg4NSA3MC4yMTAzIDQyLjU3OTVaIiBmaWxsPSIjRkZGREY4Ii8+Cjwvc3ZnPgo=",
          uuid: 'phantom-wallet'
        }
      }];
    }
    return [];
  },
  
  // 订阅钱包提供者变化
  subscribe: (callback) => {
    // 检查Phantom钱包是否已安装
    if (checkPhantomWalletInstalled()) {
      // 立即触发回调，通知有可用的钱包
      setTimeout(callback, 0);
    }
    
    // 返回空清理函数
    return () => {};
  },
};

// 连接到Phantom钱包
export const connectWithProvider = async (providerWithInfo) => {
  try {
    if (!providerWithInfo || !providerWithInfo.provider) {
      throw new Error("未找到Phantom钱包提供者");
    }
    
    // 请求连接Phantom钱包
    const resp = await providerWithInfo.provider.connect();
    
    // 返回连接结果
    return {
      success: true,
      address: resp.publicKey.toString(),
      wallet: providerWithInfo
    };
  } catch (error) {
    console.error("连接Phantom钱包失败:", error);
    return {
      success: false,
      error: error.message || "连接失败"
    };
  }
};

// 断开钱包连接
export const disconnectWallet = async () => {
  try {
    if (checkPhantomWalletInstalled()) {
      await window.phantom.solana.disconnect();
    }
    return {
      success: true
    };
  } catch (error) {
    console.error("断开钱包连接失败:", error);
    return {
      success: false,
      error: error.message || "断开连接失败"
    };
  }
};

// 获取Solana网络上的代币余额
export const getSolanaTokens = async (walletAddress) => {
  // 这个函数将在TokensPage.jsx中实现
  return [];
};