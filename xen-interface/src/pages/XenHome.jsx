import React, { useState, useEffect } from 'react';
import { Link as RouterLink } from 'react-router-dom';
import xenService from '../services/xenService';
import websocketResponseHandlerService from '../services/websocketResponseHandlerService';
import {
  Box,
  Container,
  Flex,
  Heading,
  Text,
  Button,
  VStack,
  HStack,
  SimpleGrid,
  Card,
  CardHeader,
  CardBody,
  Divider,
  FormControl,
  FormLabel,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  NumberIncrementStepper,
  NumberDecrementStepper,
  useColorModeValue,
  Icon,
  Badge,
  useToast,
  Spinner,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalCloseButton,
  useDisclosure,
} from '@chakra-ui/react';
import { InfoIcon, TimeIcon, StarIcon, CheckCircleIcon, WarningIcon } from '@chakra-ui/icons';
import { connectWithProvider, walletStore } from '../utils/walletUtils';
import { ethers } from 'ethers';

// XENScollMint合约ABI（仅包含我们需要的方法）
const xenScollMintABI = [
  {
    "inputs": [
      {
        "internalType": "uint256",
        "name": "_start",
        "type": "uint256"
      },
      {
        "internalType": "uint256",
        "name": "_count",
        "type": "uint256"
      },
      {
        "internalType": "address",
        "name": "target",
        "type": "address"
      },
      {
        "internalType": "bytes",
        "name": "data",
        "type": "bytes"
      }
    ],
    "name": "increaseAndExecute",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  }
];

// 合约地址
const XEN_CONTRACT_ADDRESS = '0xbA8a29De182E415f355b1B65DCC60eD98Ac427cF';

const XenHome = () => {
  // 使用Chakra UI的颜色模式
  const bgColor = useColorModeValue('white', 'gray.800');
  const secondaryBgColor = useColorModeValue('gray.50', 'gray.700');
  const borderColor = useColorModeValue('gray.200', 'gray.700');
  const textColor = useColorModeValue('gray.600', 'gray.400');
  const accentColor = 'hsl(142, 76%, 36%)';
  const accentColorLight = 'rgba(24, 201, 100, 0.1)';
  
  // 状态管理
  const [lockDays, setLockDays] = useState(30);
  const [slotCount, setSlotCount] = useState(1);
  const [walletAddress, setWalletAddress] = useState('');
  const [walletInfo, setWalletInfo] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [txHash, setTxHash] = useState('');
  const [error, setError] = useState('');
  const [estimatedReward, setEstimatedReward] = useState(2.5);
  const [estimatedDailyRate, setEstimatedDailyRate] = useState(0.42);
  const [expiryDate, setExpiryDate] = useState('');
  const [xenRecords, setXenRecords] = useState([]);
  
  // Toast通知
  const toast = useToast();
  
  // 交易结果模态框
  const { isOpen, onOpen, onClose } = useDisclosure();
  
  // 监听WebSocket消息
  useEffect(() => {
    const subscription = websocketResponseHandlerService.shared.subscribe(message => {
      if (message.event === 'getxensbychainname' || message.event === 'getxensbystatusandchain') {
        // 处理获取Xen记录的响应
        if (message.data && Array.isArray(message.data)) {
          setXenRecords(message.data);
        }
      } else if (message.event === 'addxen') {
        // 处理添加Xen记录的响应
        toast({
          title: '添加Xen记录',
          description: message.error ? `添加失败: ${message.error}` : '添加成功',
          status: message.error ? 'error' : 'success',
          duration: 5000,
          isClosable: true,
          position: 'top',
        });
        
        // 刷新记录列表
        if (!message.error) {
          xenService.getXensByChainName('base');
        }
      }
    });
    
    // 初始加载Xen记录
    if (websocketResponseHandlerService.isConnected) {
      xenService.getXensByChainName('base');
    }
    
    return () => subscription.unsubscribe();
  }, []);
  
  // 监听钱包连接事件
  useEffect(() => {
    const handleWalletConnected = (event) => {
      const { address, wallet } = event.detail;
      setWalletAddress(address);
      setWalletInfo(wallet);
    };

    // 监听钱包断开连接事件
    const handleWalletDisconnected = () => {
      setWalletAddress('');
      setWalletInfo(null);
    };

    // 添加事件监听器
    window.addEventListener('wallet-connected', handleWalletConnected);
    window.addEventListener('wallet-disconnected', handleWalletDisconnected);
    
    // 清理函数
    return () => {
      window.removeEventListener('wallet-connected', handleWalletConnected);
      window.removeEventListener('wallet-disconnected', handleWalletDisconnected);
    };
  }, []);
  
  // 计算到期日期
  useEffect(() => {
    const date = new Date();
    date.setDate(date.getDate() + lockDays);
    setExpiryDate(date.toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' }));
    
    // 简单的奖励估算逻辑（实际应用中可能需要更复杂的计算）
    setEstimatedReward((lockDays / 30 * 1.5 + 1).toFixed(1));
    setEstimatedDailyRate(((lockDays / 30 * 0.2) + 0.22).toFixed(2));
  }, [lockDays]);
  

  // 处理创建槽位
  const handleCreateSlot = async () => {
    // 使用xenService添加Xen记录
    const addXenToService = () => {
      try {
        // 创建Xen记录请求对象
        const xenRequest = {
          slot: Date.now(), // 使用时间戳作为临时槽位ID
          chainName: 'base',
          count: slotCount,
          days: lockDays,
          ranking: 0, // 初始排名
          amp: 0,     // 初始放大器值
          eaa: 0,     // 初始EAA值
          status: 'pending',
          expectedReward: parseFloat(estimatedReward)
        };
        
        // 调用xenService添加记录
        xenService.addXen(xenRequest);
      } catch (error) {
        console.error('添加Xen记录失败:', error);
        setError(`添加Xen记录失败: ${error.message}`);
        toast({
          title: '添加Xen记录失败',
          description: error.message,
          status: 'error',
          duration: 5000,
          isClosable: true,
          position: 'top',
        });
      }
    };
    
    // 重置状态
    setError('');
    setTxHash('');
    
    // 检查钱包连接
    if (!walletAddress) {
      setError('请先连接钱包');
      toast({
        title: '未连接钱包',
        description: '请先连接钱包后再创建槽位',
        status: 'warning',
        duration: 5000,
        isClosable: true,
        position: 'top',
      });
      return;
    }
    
    try {
      setIsLoading(true);
      
      // 获取钱包提供者
      const provider = walletInfo.provider;
      
      // 检查当前网络是否为Base网络
      const chainId = await provider.request({ method: 'eth_chainId' });
      if (chainId !== '0x2105') { // Base网络的chainId为0x2105
        setError('请切换到Base网络');
        toast({
          title: '网络错误',
          description: '请将钱包切换到Base网络后再创建槽位',
          status: 'warning',
          duration: 5000,
          isClosable: true,
          position: 'top',
        });
        setIsLoading(false);
        return;
      }
      
      // 创建ethers提供者和签名者
      const ethersProvider = new ethers.BrowserProvider(provider);
      const signer = await ethersProvider.getSigner();
      
      // 创建合约实例
      const xenContract = new ethers.Contract(XEN_CONTRACT_ADDRESS, xenScollMintABI, signer);
      
      // 构造increaseAndExecute方法的参数
      const startValue = 0; // 起始值，通常从0开始
      const countValue = 100; // 固定为100，根据合约要求
      
      // 这里需要根据实际情况设置目标合约地址和调用数据
      // 例如，如果要与XEN代币合约交互，这里应该是XEN代币合约地址
      const targetAddress = '0xffcbF84650cE02DaFE96926B37a0ac5E34932fa5'; // Sepolia XEN合约地址
      
      // 构造调用数据，使用固定的方法ID前缀，后面跟着锁定天数的十六进制表示
      // 0x9ff054df是mint方法的选择器，后面跟着32字节的参数（锁定天数）
      // 使用ethers.js的正确方式构造参数
      const abiCoder = new ethers.AbiCoder();
      const encodedParams = abiCoder.encode(['uint256'], [lockDays]);
      const data = '0x9ff054df' + encodedParams.slice(2); // 去掉encodedParams前面的0x
      console.log('Data:', data);
      // 估算gas费用
      const gasEstimate = await xenContract.increaseAndExecute.estimateGas(
        startValue,
        countValue,
        targetAddress,
        data
      );
      
      // 发送交易
      const tx = await xenContract.increaseAndExecute(
        startValue,
        countValue,
        targetAddress,
        data,
        {
          gasLimit: Math.floor(Number(gasEstimate) * 1.2), // 增加20%的gas限制以确保交易成功
        }
      );
      
      // 保存交易哈希
      setTxHash(tx.hash);
      
      // 显示成功消息
      toast({
        title: '交易已发送',
        description: '您的槽位创建交易已提交到区块链',
        status: 'success',
        duration: 5000,
        isClosable: true,
        position: 'top',
      });
      
      // 打开交易结果模态框
      onOpen();
      
      // 等待交易确认
      await tx.wait();
      
      toast({
        title: '交易已确认',
        description: '您的槽位已成功创建',
        status: 'success',
        duration: 5000,
        isClosable: true,
        position: 'top',
      });
      
    } catch (err) {
      console.error('创建槽位失败:', err);
      setError(err.message || '交易失败，请重试');
      
      toast({
        title: '交易失败',
        description: err.message || '创建槽位失败，请重试',
        status: 'error',
        duration: 5000,
        isClosable: true,
        position: 'top',
      });
    } finally {
      setIsLoading(false);
    }
  };
  
  return (
    <Container maxW="1008px" py={8}>
      <VStack spacing={8} align="stretch">
        {/* 欢迎区域 */}
        <Box
          bg={bgColor}
          p={6}
          borderRadius="xl"
          boxShadow="sm"
          borderWidth="1px"
          borderColor={borderColor}
          position="relative"
          overflow="hidden"
        >
          <Box
            position="absolute"
            top="0"
            right="0"
            width="200px"
            height="200px"
            bg={accentColorLight}
            borderRadius="full"
            transform="translate(30%, -30%)"
            zIndex="0"
          />
          
          <VStack spacing={4} align="start" position="relative" zIndex="1">
            <Heading as="h1" size="xl">XEN 合约</Heading>
            <Text fontSize="lg" color={textColor}>
              XEN是一个去中心化的加密货币项目，基于自我保管原则，没有预挖、没有创始人分配、没有管理密钥，完全透明。
            </Text>
            <HStack spacing={4}>
              <Button
                as={RouterLink}
                to="/xen-slots"
                colorScheme="green"
                bg={accentColor}
                size="lg"
                _hover={{ bg: `${accentColor}/90`, transform: 'translateY(-2px)' }}
                boxShadow="md"
              >
                查看我的槽位
              </Button>
              <Button
                variant="outline"
                colorScheme="green"
                size="lg"
                _hover={{ bg: accentColorLight, transform: 'translateY(-2px)' }}
              >
                了解更多
              </Button>
            </HStack>
          </VStack>
        </Box>
        
        {/* 统计数据 */}
       
        
        {/* 创建新槽位 */}
        <Box
          bg={bgColor}
          p={6}
          borderRadius="xl"
          boxShadow="sm"
          borderWidth="1px"
          borderColor={borderColor}
        >
          <VStack spacing={6} align="stretch">
            <Heading as="h2" size="lg">创建新槽位</Heading>
            <Text color={textColor}>
              通过创建槽位，您可以参与XEN网络并获得奖励。选择锁定期和数量来自定义您的收益潜力。
              本操作将在Base网络上调用XENScollMint合约的increaseAndExecute方法，创建100个代理合约。
            </Text>
            
            <SimpleGrid columns={{ base: 1, md: 2 }} spacing={6}>
              <FormControl>
                <FormLabel>锁定期 (天)</FormLabel>
                <NumberInput 
                  value={lockDays} 
                  onChange={(valueString) => setLockDays(parseInt(valueString))} 
                  min={1} 
                  max={365}
                >
                  <NumberInputField />
                  <NumberInputStepper>
                    <NumberIncrementStepper />
                    <NumberDecrementStepper />
                  </NumberInputStepper>
                </NumberInput>
                <Text fontSize="sm" color={textColor} mt={2}>
                  <Icon as={InfoIcon} mr={1} />
                  锁定期越长，奖励倍数越高
                </Text>
              </FormControl>
              
              <FormControl>
                <FormLabel>槽位数量</FormLabel>
                <NumberInput 
                  value={100} 
                  isReadOnly={true}
                  min={100} 
                  max={100}
                >
                  <NumberInputField />
                  <NumberInputStepper>
                    <NumberIncrementStepper isDisabled={true} />
                    <NumberDecrementStepper isDisabled={true} />
                  </NumberInputStepper>
                </NumberInput>
                <Text fontSize="sm" color={textColor} mt={2}>
                  <Icon as={InfoIcon} mr={1} />
                  XENScollMint合约要求固定创建100个代理合约
                </Text>
              </FormControl>
            </SimpleGrid>
            
            <Divider my={2} />
            
            <Box bg={secondaryBgColor} p={4} borderRadius="md">
              <VStack spacing={3} align="stretch">
                <Flex justify="space-between">
                  <Text>预计收益倍数:</Text>
                  <Text fontWeight="bold">{estimatedReward}x</Text>
                </Flex>
                <Flex justify="space-between">
                  <Text>预计每日收益率:</Text>
                  <Text fontWeight="bold">{estimatedDailyRate}%</Text>
                </Flex>
                <Flex justify="space-between">
                  <Text>到期日期:</Text>
                  <Text fontWeight="bold">{expiryDate}</Text>
                </Flex>
              </VStack>
            </Box>
            
            {error && (
              <Alert status="error" borderRadius="md" mb={4}>
                <AlertIcon />
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}
            
            <Button
              colorScheme="green"
              bg={accentColor}
              size="lg"
              _hover={{ bg: `${accentColor}/90`, transform: 'translateY(-2px)' }}
              boxShadow="md"
              onClick={handleCreateSlot}
              isLoading={isLoading}
              loadingText="处理中..."
              disabled={isLoading}
            >
              在Base网络创建槽位
            </Button>
          </VStack>
        </Box>
        
 
        
        {/* 交易结果模态框 */}
        <Modal isOpen={isOpen} onClose={onClose} isCentered>
          <ModalOverlay backdropFilter="blur(6px)" />
          <ModalContent borderRadius="xl" p={4}>
            <ModalHeader>交易状态</ModalHeader>
            <ModalCloseButton />
            <ModalBody pb={6}>
              {txHash ? (
                <VStack spacing={4} align="stretch">
                  <Alert status="success" borderRadius="md">
                    <AlertIcon />
                    <Box>
                      <AlertTitle>交易已提交</AlertTitle>
                      <AlertDescription>
                        您的槽位创建交易已成功提交到区块链
                      </AlertDescription>
                    </Box>
                  </Alert>
                  
                  <Box>
                    <Text fontWeight="bold" mb={2}>交易详情:</Text>
                    <Text fontSize="sm" wordBreak="break-all">
                      交易哈希: {txHash}
                    </Text>
                    <Text fontSize="sm" mt={2}>
                      锁定期: {lockDays} 天
                    </Text>
                    <Text fontSize="sm">
                      槽位数量: {slotCount}
                    </Text>
                    <Text fontSize="sm">
                      到期日期: {expiryDate}
                    </Text>
                  </Box>
                  
                  <Button as="a" colorScheme="blue" variant="outline" size="sm" 
                    href={`https://etherscan.io/tx/${txHash}`} target="_blank" rel="noopener noreferrer">
                    在区块浏览器中查看
                  </Button>
                </VStack>
              ) : error ? (
                <Alert status="error" borderRadius="md">
                  <AlertIcon />
                  <Box>
                    <AlertTitle>交易失败</AlertTitle>
                    <AlertDescription>{error}</AlertDescription>
                  </Box>
                </Alert>
              ) : (
                <VStack spacing={4}>
                  <Spinner size="xl" color={accentColor} />
                  <Text>处理中...</Text>
                </VStack>
              )}
            </ModalBody>
          </ModalContent>
        </Modal>
      </VStack>
    </Container>
  );
};

export default XenHome;