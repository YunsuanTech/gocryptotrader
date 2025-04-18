import React, { useState, useEffect } from 'react';
import { Link as RouterLink, useLocation } from 'react-router-dom';
import {
  Box,
  Flex,
  HStack,
  Button,
  Text,
  Container,
  useBreakpointValue, // 用于响应式文本切换
  Image,
} from '@chakra-ui/react';
import WalletModal from './WalletModal';
import { formatAddress } from '../utils/walletUtils';

// 导航项组件
const NavLink = ({ children, to = '/', isActive }) => (
  <RouterLink to={to}>
    <Text
      px={{ base: '12px', md: '16px' }}
      py={{ base: '8px', md: '13px' }}
      rounded={'xl'}
      _hover={{
        textDecoration: 'none',
        bg: 'rgba(24, 201, 100, 0.08)',
        color: '#18c964',
        transform: 'translateY(-1px)',
      }}
      fontWeight={isActive ? 'bold' : 'medium'}
      fontSize={{ base: '14px', md: '16px' }}
      transition='all 0.2s'
      letterSpacing='0.3px'
      color={isActive ? '#18c964' : 'gray.600'}
      bg={isActive ? 'rgba(24, 201, 100, 0.1)' : 'transparent'}
      border='none'
    >
      {children}
    </Text>
  </RouterLink>
);

// 主导航组件
const Navbar = () => {
  const location = useLocation();
  const [isWalletModalOpen, setIsWalletModalOpen] = useState(false);
  const [walletAddress, setWalletAddress] = useState('');
  const [walletInfo, setWalletInfo] = useState(null);

  // 动态切换按钮文本
  const buttonText = useBreakpointValue({ base: 'Connect', md: 'Connect Wallet' });
  
  // 监听钱包连接事件
  useEffect(() => {
    const handleWalletConnected = (event) => {
      const { address, wallet } = event.detail;
      console.log('钱包已连接:', address); // 添加日志以便调试
      setWalletAddress(address);
      console.log('钱包信息:', wallet); // 添加日志以便调试
      setWalletInfo(wallet);
    };

    // 监听钱包断开连接事件
    const handleWalletDisconnected = () => {
      console.log('钱包已断开连接'); // 添加日志以便调试
      setWalletAddress('');
      setWalletInfo(null);
    };

    // 监听钱包连接和断开连接事件
    window.addEventListener('wallet-connected', handleWalletConnected);
    window.addEventListener('wallet-disconnected', handleWalletDisconnected);
    
    // 清理函数
    return () => {
      window.removeEventListener('wallet-connected', handleWalletConnected);
      window.removeEventListener('wallet-disconnected', handleWalletDisconnected);
    };
  }, []);
  
  // 处理钱包连接按钮点击
  const handleWalletConnect = () => {
    if (walletAddress) {
      // 如果已连接，显示钱包菜单（这里简化为重新打开模态框）
      setIsWalletModalOpen(true);
    } else {
      // 未连接，打开连接模态框
      setIsWalletModalOpen(true);
    }
  };
  return (
    <>
      {/* 顶部导航栏 */}
      <Box
        as='header'
        position='sticky'
        top='0'
        zIndex='20'
        height='105px'
        py='25px'
        backdropFilter='blur(6px)'
      >
        <Container maxW='1008px' px={0}>
          <Flex
            h='65px'
            alignItems={'center'}
            justifyContent={'space-between'}
            px='8px'
            rounded='12px'
          >
            {/* 左侧品牌名称 */}
            <Box fontWeight='bold'>
              <RouterLink to='/'>
                <Text fontSize={{ base: 'xl', md: '2xl' }} fontWeight='bold' color='brand.500'>
                  XEN Trade
                </Text>
              </RouterLink>
            </Box>

            {/* 中间导航链接（桌面端） - 左对齐 */}
            <HStack as={'nav'} spacing={4} display={{ base: 'none', md: 'flex' }} justifyContent={'flex-start'} flex={1} ml={4}>
              <NavLink  to='/' isActive={location.pathname === '/'}>XEN</NavLink>
              <NavLink to='/xen-slots' isActive={location.pathname === '/xen-slots'}>SLOT</NavLink>
            </HStack>

        {/* 右侧按钮 */}
        <Button
            variant={walletAddress ? 'outline' : 'solid'}
            size={{ base: 'sm', md: 'md' }}
            height={{ base: '40px', md: '40px' }} // 统一高度为40px
            minWidth={{ base: walletAddress ? '120px' : '82.73px', md: '120px' }} // 设置最小宽度确保内容不会挤压
            maxWidth={{ base: '160px', md: '180px' }} // 设置最大宽度避免按钮过宽
            px={{ base: '12px', md: '12px' }} // 水平内边距
            py={{ base: '8px', md: '8px' }} // 垂直内边距
            rounded='xl'
            bg={walletAddress ? 'transparent' : 'hsl(142, 76%, 36%)'} // 连接后透明背景，未连接时绿色背景
            border={walletAddress ? '1px solid' : 'none'}
            borderColor={walletAddress ? 'gray.200' : 'transparent'}
            _hover={{ 
              bg: walletAddress ? 'gray.50' : 'hsl(142, 76%, 36%)/90', 
              transform: 'translateY(-1px)' 
            }}
            color={walletAddress ? 'gray.700' : 'white'} // 连接后灰色文字，未连接时白色文字
            boxShadow={walletAddress ? 'sm' : 'md'} // 阴影
            fontSize={{ base: 'sm', md: 'md' }} // 字体大小
            onClick={handleWalletConnect}
            display="inline-flex"
            alignItems="center"
            justifyContent="center"
            gap={2}
            data-wallet-status={walletAddress ? 'connected' : 'disconnected'} // 添加数据属性用于调试
            overflow="hidden" // 防止内容溢出
            textOverflow="ellipsis" // 文本溢出时显示省略号
            whiteSpace="nowrap" // 防止文本换行
          >
            {walletAddress ? (
              <>
                <Image 
                  src={walletInfo.info.icon} 
                  alt="Wallet Icon" 
                  boxSize="20px"
                  objectFit="contain"
                  borderRadius="4px"
                  mr="1px"
                />
                {formatAddress(walletAddress)}
              </>
            ) : (
              buttonText
            )}
          </Button>
          </Flex>
        </Container>
      </Box>

      {/* 底部导航栏（移动端） */}
      <Box
        position='fixed'
        bottom='0'
        left='0'
        right='0'
        zIndex='20'
        bg='white'
        borderTop='1px solid hsl(142, 76%, 36%, 0.2)' // 顶部边框为浅绿色
        roundedTop='12px'
        shadow='sm'
        display={{ base: 'flex', md: 'none' }}
        justifyContent='space-around'
        py='2'
      >
        <NavLink to='/' isActive={location.pathname === '/'}>XEN</NavLink>
        <NavLink to='/xen-slots' isActive={location.pathname === '/xen-slots'}>SLOT</NavLink>
      </Box>

      {/* 钱包连接模态框 */}
      <WalletModal isOpen={isWalletModalOpen} onClose={() => setIsWalletModalOpen(false)} />
    </>
  );
};

export default Navbar;