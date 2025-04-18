import React, { useState, useEffect } from 'react';
import { walletStore, connectWithProvider, formatAddress, disconnectWallet } from '../utils/walletUtils';
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalCloseButton,
  Flex,
  Text,
  Box,
  Divider,
  Link,
  Image,
  VStack,
  Spinner,
  Button,
} from '@chakra-ui/react';

// 钱包图标数据
const walletIcons = {
  phantom: "data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMTA4IiBoZWlnaHQ9IjEwOCIgdmlld0JveD0iMCAwIDEwOCAxMDgiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+CjxyZWN0IHdpZHRoPSIxMDgiIGhlaWdodD0iMTA4IiByeD0iMjYiIGZpbGw9IiNBQjlGRjIiLz4KPHBhdGggZmlsbC1ydWxlPSJldmVub2RkIiBjbGlwLXJ1bGU9ImV2ZW5vZGQiIGQ9Ik00Ni41MjY3IDY5LjkyMjlDNDIuMDA1NCA3Ni44NTA5IDM0LjQyOTIgODUuNjE4MiAyNC4zNDggODUuNjE4MkMxOS41ODI0IDg1LjYxODIgMTUgODMuNjU2MyAxNSA3NS4xMzQyQzE1IDUzLjQzMDUgNDQuNjMyNiAxOS44MzI3IDcyLjEyNjggMTkuODMyN0M4Ny43NjggMTkuODMyNyA5NCAzMC42ODQ2IDk0IDQzLjAwNzlDOTQgNTguODI1OCA4My43MzU1IDc2LjkxMjIgNzMuNTMyMSA3Ni45MTIyQzcwLjI5MzkgNzYuOTEyMiA2OC43MDUzIDc1LjEzNDIgNjguNzA1MyA3Mi4zMTRDNjguNzA1MyA3MS41NzgzIDY4LjgyNzUgNzAuNzgxMiA2OS4wNzE5IDY5LjkyMjlDNjUuNTg5MyA3NS44Njk5IDU4Ljg2ODUgODEuMzg3OCA1Mi41NzU0IDgxLjM4NzhDNDcuOTkzIDgxLjM4NzggNDUuNjcxMyA3OC41MDYzIDQ1LjY3MTMgNzQuNDU5OEM0NS42NzEzIDcyLjk4ODQgNDUuOTc2OCA3MS40NTU2IDQ2LjUyNjcgNjkuOTIyOVpNODMuNjc2MSA0Mi41Nzk0QzgzLjY3NjEgNDYuMTcwNCA4MS41NTc1IDQ3Ljk2NTggNzkuMTg3NSA0Ny45NjU4Qzc2Ljc4MTYgNDcuOTY1OCA3NC42OTg5IDQ2LjE3MDQgNzQuNjk4OSA0Mi41Nzk0Qzc0LjY5ODkgMzguOTg4NSA3Ni43ODE2IDM3LjE5MzEgNzkuMTg3NSAzNy4xOTMxQzgxLjU1NzUgMzcuMTkzMSA4My42NzYxIDM4Ljk4ODUgODMuNjc2MSA0Mi41Nzk0Wk03MC4yMTAzIDQyLjU3OTVDNzAuMjEwMyA0Ni4xNzA0IDY4LjA5MTYgNDcuOTY1OCA2NS43MjE2IDQ3Ljk2NThDNjMuMzE1NyA0Ny45NjU4IDYxLjIzMyA0Ni4xNzA0IDYxLjIzMyA0Mi41Nzk1QzYxLjIzMyAzOC45ODg1IDYzLjMxNTcgMzcuMTkzMSA2NS43MjE2IDM3LjE5MzFDNjguMDkxNiAzNy4xOTMxIDcwLjIxMDMgMzguOTg4NSA3MC4yMTAzIDQyLjU3OTVaIiBmaWxsPSIjRkZGREY4Ii8+Cjwvc3ZnPgo=",
  metamask: "data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMzMiIGhlaWdodD0iMzAiIHZpZXdCb3g9IjAgMCAzMyAzMCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTMxLjE0NTcgMC4wMDAxMjIwN0wxOC42NTcyIDkuNzAwMTJMMjAuOTcxNSA0LjM5NDEyTDMxLjE0NTcgMC4wMDAxMjIwN1oiIGZpbGw9IiNFMTc3MjYiLz4KPHBhdGggZD0iTTEuODUzNTIgMC4wMDAxMjIwN0wxNC4yNDM1IDkuNzk0MTJMMTIuMDI3NSA0LjM5NDEyTDEuODUzNTIgMC4wMDAxMjIwN1pNMjYuNTk5NSAyMS40ODgxTDIzLjIyMzUgMjYuNDI0MUwzMC4zMzU1IDI4LjMxMjFMMzIuMzc1NSAyMS41ODIxTDI2LjU5OTUgMjEuNDg4MVpNMC42MzU1MiAyMS41ODIxTDIuNjc1NTIgMjguMzEyMUw5Ljc4NzUyIDI2LjQyNDFMNi40MTE1MiAyMS40ODgxTDAuNjM1NTIgMjEuNTgyMVoiIGZpbGw9IiNFMjc2MjUiLz4KPHBhdGggZD0iTTkuNDExNTIgMTMuMTI4MUw3LjQyMzUyIDE2LjAwMDFMMTQuNDgzNSAxNi4zMDAxTDE0LjI0MzUgOC41NzYxMkw5LjQxMTUyIDEzLjEyODFaTTIzLjU4NzUgMTMuMTI4MUwxOC42NTc1IDguNDgyMTJMMTguNjU3NSAxNi4zMDAxTDI1LjU3NTUgMTYuMDAwMUwyMy41ODc1IDEzLjEyODFaTTkuNzg3NTIgMjYuNDI0MUwxNC4wNTU1IDI0LjQzNjFMMTAuMzY3NSAyMS42MzYxTDkuNzg3NTIgMjYuNDI0MVpNMTguOTQzNSAyNC40MzYxTDIzLjIyMzUgMjYuNDI0MUwyMi42MzE1IDIxLjYzNjFMMTguOTQzNSAyNC40MzYxWiIgZmlsbD0iI0UyNzYyNSIvPgo8cGF0aCBkPSJNMjMuMjIzNSAyNi40MjQxTDE4Ljk0MzUgMjQuNDM2MUwxOS4yODM1IDI3LjE0ODFMMTkuMjM1NSAyOC4yMTgxTDIzLjIyMzUgMjYuNDI0MVpNOS43ODc1MiAyNi40MjQxTDEzLjc3NTUgMjguMjE4MUwxMy43Mzk1IDI3LjE0ODFMMTQuMDU1NSAyNC40MzYxTDkuNzg3NTIgMjYuNDI0MVoiIGZpbGw9IiNEMUQxRDEiLz4KPHBhdGggZD0iTTEzLjg3MTUgMTkuODI0MUwxMC4zMTk1IDE4Ljc5MDFMMTIuNzc1NSAxNy42ODQxTDEzLjg3MTUgMTkuODI0MVpNMTkuMTI3NSAxOS44MjQxTDIwLjIyMzUgMTcuNjg0MUwyMi42OTE1IDE4Ljc5MDFMMTkuMTI3NSAxOS44MjQxWiIgZmlsbD0iIzIzMzQ0NyIvPgo8cGF0aCBkPSJNOS43ODc1MiAyNi40MjQxTDEwLjM5MTUgMjEuNDg4MUw2LjQxMTUyIDIxLjU4MjFMOS43ODc1MiAyNi40MjQxWk0yMi42MDc1IDIxLjQ4ODFMMjMuMjIzNSAyNi40MjQxTDI2LjU5OTUgMjEuNTgyMUwyMi42MDc1IDIxLjQ4ODFaTTI1LjU3NTUgMTYuMDAwMUwxOC42NTc1IDE2LjMwMDFMMTkuMTM5NSAxOS44MjQxTDIwLjIzNTUgMTcuNjg0MUwyMi43MDM1IDE4Ljc5MDFMMjUuNTc1NSAxNi4wMDAxWk0xMC4zMTk1IDE4Ljc5MDFMMTIuNzc1NSAxNy42ODQxTDEzLjg3MTUgMTkuODI0MUwxNC4zNjM1IDE2LjMwMDFMNy40MjM1MiAxNi4wMDAxTDEwLjMxOTUgMTguNzkwMVoiIGZpbGw9IiNDRDZDMjMiLz4KPHBhdGggZD0iTTcuNDIzNTIgMTYuMDAwMUwxMC4zNjc1IDIxLjYzNjFMMTAuMzE5NSAxOC43OTAxTDcuNDIzNTIgMTYuMDAwMVpNMjIuNzAzNSAxOC43OTAxTDIyLjYzMTUgMjEuNjM2MUwyNS41NzU1IDE2LjAwMDFMMjIuNzAzNSAxOC43OTAxWk0xNC4zNjM1IDE2LjMwMDFMMTMuODcxNSAxOS44MjQxTDE0LjQ4MzUgMjMuOTI0MUwxNC42MTE1IDE4LjUwMjFMMTQuMzYzNSAxNi4zMDAxWk0xOC42NTc1IDE2LjMwMDFMMTguNDE5NSAxOC40OTAxTDE4LjUxMTUgMjMuOTI0MUwxOS4xMzk1IDE5LjgyNDFMMTguNjU3NSAxNi4zMDAxWiIgZmlsbD0iI0UyNzUyNSIvPgo8cGF0aCBkPSJNMTkuMTM5NSAxOS44MjQxTDE4LjUxMTUgMjMuOTI0MUwxOC45NDM1IDI0LjQzNjFMMjIuNjMxNSAyMS42MzYxTDIyLjcwMzUgMTguNzkwMUwxOS4xMzk1IDE5LjgyNDFaTTEwLjMxOTUgMTguNzkwMUwxMC4zNjc1IDIxLjYzNjFMMTQuMDU1NSAyNC40MzYxTDE0LjQ4MzUgMjMuOTI0MUwxMy44NzE1IDE5LjgyNDFMMTAuMzE5NSAxOC43OTAxWiIgZmlsbD0iI0YwNUIyQiIvPgo8cGF0aCBkPSJNMTkuMjM1NSAyOC4yMTgxTDE5LjI4MzUgMjcuMTQ4MUwxOC45NjM1IDI2Ljg3NDFIMTQuMDM1NUwxMy43Mzk1IDI3LjE0ODFMMTMuNzc1NSAyOC4yMTgxTDkuNzg3NTIgMjYuNDI0MUwxMS4xNjc1IDI3LjU0NDFMMTMuOTk5NSAyOS41MDAxSDE4Ljk5OTVMMjEuODQzNSAyNy41NDQxTDIzLjIyMzUgMjYuNDI0MUwxOS4yMzU1IDI4LjIxODFaIiBmaWxsPSIjQzBBQzlEIi8+CjxwYXRoIGQ9Ik0xOC45NDM1IDI0LjQzNjFMMTguNTExNSAyMy45MjQxSDE0LjQ4MzVMMTQuMDU1NSAyNC40MzYxTDEzLjczOTUgMjcuMTQ4MUwxNC4wMzU1IDI2Ljg3NDFIMTguOTYzNUwxOS4yODM1IDI3LjE0ODFMMTguOTQzNSAyNC40MzYxWiIgZmlsbD0iIzE2MTYxNiIvPgo8cGF0aCBkPSJNMzEuNjI3NSA5Ljk5NjEyTDMyLjY3NTUgNS4xMTYxMkwzMS4xNDU3IDAuMDAwMTIyMDdMMTguOTQzNSA5LjEyMDEyTDIzLjU4NzUgMTMuMTI4MUwzMC4xOTU1IDE0Ljk3MDFMMzEuNjI3NSAxMy4zMzYxTDMxLjAwMzUgMTIuODcyMUwzMi4wMDM1IDExLjk0NDFMMzEuMjM5NSAxMS4zMjgxTDMyLjIzOTUgMTAuNTY0MUwzMS42Mjc1IDkuOTk2MTJaTTAuMzIzNTIgNS4xMTYxMkwxLjM3MTUyIDkuOTk2MTJMMC43NDc1MiAxMC41NjQxTDEuNzQ3NTIgMTEuMzI4MUwwLjk5NTUyIDExLjk0NDFMMi4wMDc1MiAxMi44NzIxTDEuMzcxNTIgMTMuMzM2MUwyLjgwMzUyIDE0Ljk3MDFMOS40MTE1MiAxMy4xMjgxTDE0LjA1NTUgOS4xMjAxMkwxLjg1MzUyIDAuMDAwMTIyMDdMMC4zMjM1MiA1LjExNjEyWiIgZmlsbD0iIzc2M0UxQSIvPgo8cGF0aCBkPSJNMzAuMTk1NSAxNC45NzAxTDIzLjU4NzUgMTMuMTI4MUwyNS41NzU1IDE2LjAwMDFMMjIuNjMxNSAyMS42MzYxTDI2LjU5OTUgMjEuNTgyMUgzMi4zNzU1TDMwLjE5NTUgMTQuOTcwMVpNOS40MTE1MiAxMy4xMjgxTDIuODAzNTIgMTQuOTcwMUwwLjYzNTUyIDIxLjU4MjFINi40MTE1MkwxMC4zNjc1IDIxLjYzNjFMNy40MjM1MiAxNi4wMDAxTDkuNDExNTIgMTMuMTI4MVpNMTguNjU3NSAxNi4zMDAxTDE4Ljk0MzUgOS4xMjAxMkwyMC45ODM1IDQuMzk0MTJIMTIuMDI3NUwxNC4wNTU1IDkuMTIwMTJMMTQuMzYzNSAxNi4zMDAxTDE0LjQ4MzUgMTguNTAyMUwxNC40ODM1IDIzLjkyNDFIMTguNTExNUwxOC41MjM1IDE4LjUwMjFMMTguNjU3NSAxNi4zMDAxWiIgZmlsbD0iI0YwNUIyQiIvPgo8L3N2Zz4K",
  okx: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADAAAAAwCAYAAABXAvmHAAAACXBIWXMAAAsTAAALEwEAmpwYAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAJDSURBVHgB7Zq9jtpAEMfHlhEgQLiioXEkoAGECwoKxMcTRHmC5E3IoyRPkPAEkI7unJYmTgEFTYwA8a3NTKScLnCHN6c9r1e3P2llWQy7M/s1Gv1twCP0ej37dDq9x+Zut1t3t9vZjDEHIiSRSPg4ZpDL5fxkMvn1cDh8m0wmfugfO53OoFQq/crn8wxfY9EymQyrVCqMfHvScZx1p9ls3pFxXBy/bKlUipGPrVbLuQqAfsCliq3zl0H84zwtjQrOw4Mt1W63P5LvBm2d+Xz+YzqdgkqUy+WgWCy+Mc/nc282m4FqLBYL+3g8fjDxenq72WxANZbLJeA13zDX67UDioL5ybXwafMYu64Ltn3bdDweQ5R97fd7GyhBQMipx4POeEDHIu2LfDdBIGGz+hJ9CQ1ABjoA2egAZPM6AgiCAEQhsi/C4jHyPA/6/f5NG3Ks2+3CYDC4aTccDrn6ojG54MnEvG00GoVmWLIRNZ7wTCwDHYBsdACy0QHIhiuRETxlICWpMMhGZHmqS8qH6JLyGegAZKMDkI0uKf8X4SWlaZo+Pp1bRrwlJU8ZKLIvUjKh0WiQ3sRUbNVq9c5Ebew7KEo2m/1p4jJ4qAmDaqDQBzj5XyiAT4VCQezJigAU+IDU+z8vJFnGWeC+bKQV/5VZ71FV6L7PA3gg3tXrdQ+DgLhC+75Wq3no69P3MC0NFQpx2lL04Ql9gHK1bRDjsSBIvScBnDTk1WrlGIZBorIDEYJj+rhdgnQ67VmWRe0zlplXl81vcyEt0rSoYDUAAAAASUVORK5CYII="
};

// 钱包连接模态框组件
const WalletModal = ({ isOpen, onClose }) => {
  // 状态管理
  const [providers, setProviders] = useState([]);
  const [selectedWallet, setSelectedWallet] = useState(null);
  const [userAccount, setUserAccount] = useState('');
  const [isConnecting, setIsConnecting] = useState(false);
  const [error, setError] = useState('');
  
  // 监听钱包提供者变化
  useEffect(() => {
    if (!isOpen) return;
    
    // 清除之前的状态
    setError('');
    
    // 订阅钱包提供者变化
    const unsubscribe = walletStore.subscribe(() => {
      setProviders(walletStore.value());
    });
    
    return () => {
      if (unsubscribe) unsubscribe();
    };
  }, [isOpen]);
  
  // 处理钱包连接
  const handleWalletConnect = async (providerWithInfo) => {
    try {
      setIsConnecting(true);
      setError('');
      
      // 连接到选定的钱包提供者
      const result = await connectWithProvider(providerWithInfo);
      
      if (result.success) {
        setSelectedWallet(result.wallet);
        setUserAccount(result.address);
        // 触发钱包连接事件，传递地址和钱包信息
        window.dispatchEvent(new CustomEvent('wallet-connected', {
          detail: {
            address: result.address,
            wallet: result.wallet
          }
        }));
        // 连接成功后延迟关闭模态框，让用户看到连接成功的状态
        setTimeout(() => {
          onClose();
        }, 1000);
      } else {
        setError(result.error || '连接失败');
      }
    } catch (error) {
      console.error('连接钱包失败:', error);
      setError(error.message || '连接失败');
    } finally {
      setIsConnecting(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} isCentered size="md">
      <ModalOverlay backdropFilter="blur(6px)" bg="whiteAlpha.300" />
      <ModalContent
        borderRadius="xl"
        maxW="450px"
        fontFamily="'Plus Jakarta Sans', sans-serif"
        position="relative"
        zIndex={100}
      >
        <ModalCloseButton 
          position="absolute" 
          top="4" 
          right="4" 
          _hover={{ opacity: 0.8 }}
          transition="opacity"
          zIndex={102}
          aria-label="Close"
        />
        <ModalHeader p={4} pb={0}>
          <Text fontSize="16px" fontWeight="bold" color="zinc.900">
            {userAccount ? '钱包已连接' : 'Wallet Connection'}
          </Text>
          <Text fontSize="14px" fontWeight="bold" color="rgba(2, 2, 30, 0.64)">
            {userAccount ? '您的钱包已成功连接' : 'Connect your wallet to use XEN Trade'}
          </Text>
        </ModalHeader>

        <Box h="23px" />

        <ModalBody p={4} pt={0}>
          {isConnecting && (
            <Flex justifyContent="center" alignItems="center" direction="column" py={6}>
              <Spinner size="xl" color="purple.500" mb={4} />
              <Text>正在连接钱包，请在钱包中确认...</Text>
            </Flex>
          )}
          
          {error && (
            <Box bg="red.50" p={3} borderRadius="md" mb={4}>
              <Text color="red.500">{error}</Text>
            </Box>
          )}
          
          {userAccount ? (
            <Flex 
              direction="column" 
              alignItems="center" 
              justifyContent="center" 
              py={4}
              bg="gray.50"
              borderRadius="xl"
            >
              <Image 
                src={selectedWallet?.info?.icon} 
                alt={selectedWallet?.info?.name} 
                w="48px" 
                h="48px" 
                mb={3} 
              />
              <Text fontWeight="bold" mb={1}>{selectedWallet?.info?.name}</Text>
              <Flex alignItems="center" mb={3}>
                <Text fontSize="sm" color="gray.600">{formatAddress(userAccount)}</Text>
                <Box 
                  as="button" 
                  ml={2} 
                  p={1} 
                  borderRadius="md" 
                  _hover={{ bg: 'gray.200' }}
                  onClick={() => {
                    navigator.clipboard.writeText(userAccount);
                    // 可以添加一个提示，表示已复制
                    alert('地址已复制到剪贴板');
                  }}
                >
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                    <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                  </svg>
                </Box>
              </Flex>
              <Button
                size="sm"
                colorScheme="red"
                variant="outline"
                onClick={() => {
                  // 调用断开连接函数
                  const result = disconnectWallet();
                  if (result.success) {
                    setUserAccount('');
                    setSelectedWallet(null);
                    // 触发钱包断开连接事件
                    window.dispatchEvent(new CustomEvent('wallet-disconnected'));
                    onClose();
                  }
                }}
              >
                断开连接
              </Button>
            </Flex>
          ) : (
            <VStack spacing={2} align="stretch">
            {/* Phantom 钱包 */}
            <Flex 
              cursor="pointer" 
              borderBottom="1px solid" 
              borderColor="gray.200"
              py={2}
              onClick={() => {
                const phantomProvider = providers.find(p => p.info.name === 'Phantom');
                if (phantomProvider) {
                  handleWalletConnect(phantomProvider);
                } else {
                  setError('未检测到Phantom钱包，请确保已安装Phantom扩展');
                }
              }}
            >
              <Flex 
                w="full" 
                alignItems="center" 
                justifyContent="space-between" 
                gap={2} 
                rounded="12px" 
                px={3} 
                py={4}
                _hover={{ bg: 'gray.50' }}
              >
                <Flex alignItems="center" gap={3}>
                  <Image src={walletIcons.phantom} alt="Phantom icon" w="24px" h="24px" />
                  <Text fontSize="base" fontWeight="medium" lineHeight="20px">
                    Phantom
                  </Text>
                </Flex>
                <Box>
                  <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-5 w-5 text-gray-400">
                    <path d="m9 18 6-6-6-6"></path>
                  </svg>
                </Box>
              </Flex>
            </Flex>

            {/* MetaMask 钱包 */}
            <Flex 
              cursor="pointer" 
              borderBottom="1px solid" 
              borderColor="gray.200"
              py={2}
              onClick={() => {
                const metaMaskProvider = providers.find(p => p.info.name === 'MetaMask');
                if (metaMaskProvider) {
                  handleWalletConnect(metaMaskProvider);
                } else {
                  setError('未检测到MetaMask钱包，请确保已安装MetaMask扩展');
                }
              }}
            >
              <Flex 
                w="full" 
                alignItems="center" 
                justifyContent="space-between" 
                gap={2} 
                rounded="12px" 
                px={3} 
                py={4}
                _hover={{ bg: 'gray.50' }}
              >
                <Flex alignItems="center" gap={3}>
                  <Image src={walletIcons.metamask} alt="MetaMask icon" w="24px" h="24px" />
                  <Text fontSize="base" fontWeight="medium" lineHeight="20px">
                    MetaMask
                  </Text>
                </Flex>
                <Box>
                  <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-5 w-5 text-gray-400">
                    <path d="m9 18 6-6-6-6"></path>
                  </svg>
                </Box>
              </Flex>
            </Flex>

            {/* OKX 钱包 */}
            <Flex 
              cursor="pointer" 
              borderBottom="1px solid" 
              borderColor="gray.200"
              py={2}
              onClick={() => {
                const okxProvider = providers.find(p => p.info.name === 'OKX Wallet');
                if (okxProvider) {
                  handleWalletConnect(okxProvider);
                } else {
                  setError('未检测到OKX钱包，请确保已安装OKX Wallet扩展');
                }
              }}
            >
              <Flex 
                w="full" 
                alignItems="center" 
                justifyContent="space-between" 
                gap={2} 
                rounded="12px" 
                px={3} 
                py={4}
                _hover={{ bg: 'gray.50' }}
              >
                <Flex alignItems="center" gap={3}>
                  <Image src={walletIcons.okx} alt="OKX Wallet icon" w="24px" h="24px" />
                  <Text fontSize="base" fontWeight="medium" lineHeight="20px">
                    OKX Wallet
                  </Text>
                </Flex>
                <Box>
                  <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-5 w-5 text-gray-400">
                    <path d="m9 18 6-6-6-6"></path>
                  </svg>
                </Box>
              </Flex>
            </Flex>
          </VStack>
          )}

          <Box h="10px" />

          <Box textAlign="center">
            <Text fontSize="11px" fontWeight="semibold" color="rgba(2, 2, 30, 0.64)">
              XEN Trade is in BETA. By continuing to use XEN Trade, you agree to comply with and be bound by the &nbsp;
              <Link href="#" target="_blank" color="#6B40E0" textDecoration="underline">
                Terms of Service
              </Link>
              &nbsp;and&nbsp;
              <Link href="#" target="_blank" color="#6B40E0" textDecoration="underline">
                Privacy Policy
              </Link>
              , which govern your access to and use of the platform.
            </Text>
          </Box>
        </ModalBody>
      </ModalContent>
    </Modal>
  );
};

export default WalletModal;