import React, { useState, useEffect, useCallback } from 'react';
import {
  Box,
  Container,
  Heading,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Text,
  Flex,
  Image,
  Button,
  useToast,
  Spinner,
  Badge,
  useColorModeValue,
  HStack,
  VStack,
  Tabs,
  TabList,
  Tab,
  TabPanels,
  TabPanel,
  Input,
  InputGroup,
  InputLeftElement,
  InputRightElement,
  IconButton,
  Card,
  CardHeader,
  CardBody,
  SimpleGrid,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  Divider,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  useBreakpointValue,
  Link,
  Stack,
  Tooltip,
} from '@chakra-ui/react';
import { SearchIcon, CloseIcon, RepeatIcon, InfoIcon, ExternalLinkIcon } from '@chakra-ui/icons';
import { useWalletStore } from '../stores/walletStore';
import { formatAddress } from '../utils/walletUtils';
import { Connection, PublicKey } from '@solana/web3.js';
import { TOKEN_PROGRAM_ID } from '@solana/spl-token';
import axios from 'axios';
import { debounce } from 'lodash';

// 默认代币图标
const DEFAULT_TOKEN_ICON = 'https://raw.githubusercontent.com/solana-labs/token-list/main/assets/mainnet/So11111111111111111111111111111111111111112/logo.png';

// Solana RPC节点地址
const SOLANA_RPC_ENDPOINT = 'https://api.mainnet-beta.solana.com';

// 代币元数据API地址
const TOKEN_LIST_URL = 'https://raw.githubusercontent.com/solana-labs/token-list/main/src/tokens/solana.tokenlist.json';

// 获取钱包的代币账户
const getTokenAccounts = async (walletAddress) => {
  try {
    // 创建Solana连接
    const connection = new Connection(SOLANA_RPC_ENDPOINT);
    
    // 创建钱包公钥
    const publicKey = new PublicKey(walletAddress);
    
    // 获取所有代币账户
    const tokenAccounts = await connection.getParsedTokenAccountsByOwner(
      publicKey,
      { programId: TOKEN_PROGRAM_ID }
    );
    
    // 获取代币元数据
    const tokenMetadata = await fetchTokenMetadata();
    
    // 获取常见代币价格
    const tokenPrices = await getTokenPrices();
    
    // 解析代币账户数据
    const parsedTokenAccounts = [];
    
    for (const account of tokenAccounts.value) {
      try {
        const accountInfo = account.account.data.parsed.info;
        const mint = accountInfo.mint;
        const amount = accountInfo.tokenAmount.amount;
        const decimals = accountInfo.tokenAmount.decimals;
        const uiAmount = accountInfo.tokenAmount.uiAmount;
        
        // 查找代币元数据
        const tokenInfo = tokenMetadata.find(token => token.address === mint);
        
        // 获取代币价格和计算USD价值
        let usdValue = null;
        if (tokenInfo && tokenInfo.symbol) {
          // 尝试通过symbol匹配价格
          const symbol = tokenInfo.symbol.toLowerCase();
          const priceId = Object.keys(tokenPrices).find(id => {
            return id.toLowerCase() === symbol.toLowerCase() || 
                  (tokenInfo.extensions && 
                   tokenInfo.extensions.coingeckoId && 
                   tokenInfo.extensions.coingeckoId.toLowerCase() === id.toLowerCase());
          });
          
          if (priceId && tokenPrices[priceId] && tokenPrices[priceId].usd) {
            usdValue = uiAmount * tokenPrices[priceId].usd;
          }
        }
        
        // 只添加有余额的代币或有价值的代币
        if (uiAmount > 0 || usdValue > 0) {
          parsedTokenAccounts.push({
            Mint: mint,
            Account: account.pubkey,
            Amount: amount,
            Decimals: decimals,
            UIAmount: uiAmount,
            TokenInfo: tokenInfo,
            USDValue: usdValue
          });
        }
      } catch (err) {
        console.error('解析代币账户失败:', err);
      }
    }
    
    return parsedTokenAccounts;
  } catch (error) {
    console.error('获取代币账户失败:', error);
    throw error;
  }
};

// 获取代币元数据
const fetchTokenMetadata = async () => {
  try {
    const response = await fetch(TOKEN_LIST_URL);
    const data = await response.json();
    return data.tokens || [];
  } catch (error) {
    console.error('获取代币元数据失败:', error);
    return [];
  }
};

// 获取代币价格
const getTokenPrices = async () => {
  try {
    // 获取Solana生态系统中常见代币的价格
    const response = await fetch('https://api.coingecko.com/api/v3/simple/price?ids=solana,usd-coin,tether,wrapped-bitcoin,ethereum,raydium,orca,mango-markets,marinade,serum,bonfida,step-finance,saber,mercurial-finance,meteora,drift,zeta-markets,jupiter,tensor,solend,marginfi,sanctum,kamino,mean-finance,helium,helium-mobile,helium-iot,bonk,jito,pyth-network&vs_currencies=usd');
    const data = await response.json();
    return data;
  } catch (error) {
    console.error('获取代币价格失败:', error);
    return {};
  }
};

const TokensPage = () => {
  // 状态管理
  const [tokens, setTokens] = useState([]);
  const [loading, setLoading] = useState(false);
  const [walletConnected, setWalletConnected] = useState(false);
  const [walletAddress, setWalletAddress] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [showSearch, setShowSearch] = useState(false);
  const toast = useToast();

  // 颜色模式
  const bgColor = useColorModeValue('white', 'gray.800');
  const secondaryBgColor = useColorModeValue('gray.50', 'gray.700');
  const borderColor = useColorModeValue('gray.200', 'gray.700');
  const textColor = useColorModeValue('gray.800', 'white');
  const mutedColor = useColorModeValue('gray.600', 'gray.400');
  const successColor = useColorModeValue('green.500', 'green.300');
  const accentColor = useColorModeValue('blue.500', 'blue.300');
  const cardBg = useColorModeValue('white', 'gray.800');
  const cardShadow = useColorModeValue('lg', 'dark-lg');
  
  // 响应式设计
  const isMobile = useBreakpointValue({ base: true, md: false });

  // 从钱包存储中获取钱包状态
  const { account, connected } = useWalletStore();

  // 监听钱包连接状态变化
  useEffect(() => {
    if (connected && account) {
      setWalletAddress(account);
      setWalletConnected(true);
      loadTokens(account);
    } else {
      setWalletConnected(false);
      setWalletAddress('');
      setTokens([]);
    }
  }, [connected, account]);


  // 加载代币数据
  const loadTokens = async (address) => {
    try {
      setLoading(true);
      const tokenAccounts = await getTokenAccounts(address);
      
      // 对代币进行排序，余额大的排在前面
      const sortedTokens = tokenAccounts.sort((a, b) => b.UIAmount - a.UIAmount);
      
      setTokens(sortedTokens);
    } catch (error) {
      console.error('加载代币失败:', error);
      toast({
        title: '加载失败',
        description: '无法获取代币信息，请稍后再试',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  // 刷新代币数据
  const refreshTokens = () => {
    if (walletConnected) {
      loadTokens(walletAddress);
    }
  };

  // 过滤代币
  const filteredTokens = tokens.filter(token => {
    if (!searchQuery) return true;
    const query = searchQuery.toLowerCase();
    return (
      token.Mint.toLowerCase().includes(query) ||
      (token.TokenInfo?.symbol && token.TokenInfo.symbol.toLowerCase().includes(query)) ||
      (token.TokenInfo?.name && token.TokenInfo.name.toLowerCase().includes(query))
    );
  });

  // 格式化数字显示
  const formatNumber = (num) => {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(2) + 'm';
    } else if (num >= 1000) {
      return (num / 1000).toFixed(2) + 'k';
    } else {
      return num.toFixed(2);
    }
  };

  // 格式化地址显示
  const formatAddress = (address) => {
    if (!address) return '';
    return `${address.substring(0, 4)}...${address.substring(address.length - 4)}`;
  };

  // 计算总价值
  const getTotalValue = () => {
    return filteredTokens.reduce((total, token) => {
      return total + (token.USDValue || 0);
    }, 0);
  };

  // 获取有价值的代币数量
  const getValuedTokensCount = () => {
    return filteredTokens.filter(token => token.USDValue && token.USDValue > 0).length;
  };

  return (
    <Container maxW="7xl" py={8}>
      {/* 页面标题 */}
      <VStack spacing={6} align="stretch">
        <Flex justify="space-between" align="center" flexWrap="wrap" gap={4}>
          <VStack align="start" spacing={1}>
            <Heading size="xl" color={textColor}>我的Solana代币</Heading>
            <Text color={mutedColor} fontSize="md">
              {walletConnected ? `Solana钱包地址: ${formatAddress(walletAddress)}` : '连接Phantom钱包查看您的Solana代币资产'}
            </Text>
          </VStack>
          
          <HStack spacing={3}>
            {/* 搜索框 */}
            {showSearch ? (
              <InputGroup size="md" maxW="300px">
                <InputLeftElement pointerEvents="none">
                  <SearchIcon color="gray.400" />
                </InputLeftElement>
                <Input
                  placeholder="搜索代币名称或地址"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  autoFocus
                  bg={cardBg}
                  borderColor={borderColor}
                />
                <InputRightElement>
                  <IconButton
                    icon={<CloseIcon />}
                    size="sm"
                    variant="ghost"
                    aria-label="清除搜索"
                    onClick={() => {
                      setSearchQuery('');
                      setShowSearch(false);
                    }}
                  />
                </InputRightElement>
              </InputGroup>
            ) : (
              <IconButton
                icon={<SearchIcon />}
                aria-label="搜索代币"
                variant="outline"
                colorScheme="blue"
                onClick={() => setShowSearch(true)}
              />
            )}
            
            {/* 刷新按钮 */}
            <IconButton
              icon={<RepeatIcon />}
              aria-label="刷新代币列表"
              variant="outline"
              colorScheme="blue"
              isLoading={loading}
              onClick={refreshTokens}
              isDisabled={!walletConnected}
            />
          </HStack>
        </Flex>

        {/* 统计信息卡片 */}
        {walletConnected && (
          <SimpleGrid columns={{ base: 1, md: 3 }} spacing={6}>
            <Card bg={cardBg} shadow={cardShadow} borderColor={borderColor}>
              <CardBody>
                <Stat>
                  <StatLabel color={mutedColor}>总价值</StatLabel>
                  <StatNumber color={accentColor}>
                    ${getTotalValue().toFixed(2)}
                  </StatNumber>
                  <StatHelpText color={mutedColor}>
                    <InfoIcon mr={1} />
                    基于当前市场价格
                  </StatHelpText>
                </Stat>
              </CardBody>
            </Card>
            
            <Card bg={cardBg} shadow={cardShadow} borderColor={borderColor}>
              <CardBody>
                <Stat>
                  <StatLabel color={mutedColor}>代币总数</StatLabel>
                  <StatNumber color={textColor}>
                    {filteredTokens.length}
                  </StatNumber>
                  <StatHelpText color={mutedColor}>
                    <InfoIcon mr={1} />
                    包含所有代币类型
                  </StatHelpText>
                </Stat>
              </CardBody>
            </Card>
            
            <Card bg={cardBg} shadow={cardShadow} borderColor={borderColor}>
              <CardBody>
                <Stat>
                  <StatLabel color={mutedColor}>有价值代币</StatLabel>
                  <StatNumber color={successColor}>
                    {getValuedTokensCount()}
                  </StatNumber>
                  <StatHelpText color={mutedColor}>
                    <InfoIcon mr={1} />
                    具有市场价值的代币
                  </StatHelpText>
                </Stat>
              </CardBody>
            </Card>
          </SimpleGrid>
        )}
      
        {/* Solana代币列表 */}
        <Card bg={cardBg} shadow={cardShadow} borderColor={borderColor}>
          <CardHeader>
            <Heading size="md" color={textColor}>Solana SPL 代币</Heading>
          </CardHeader>
          
          <CardBody>
                   {/* 代币列表 */}
                   {loading ? (
                     <Flex direction="column" justify="center" align="center" h="300px">
                       <Spinner size="xl" color={accentColor} thickness="4px" />
                       <Text mt={4} color={mutedColor} fontSize="lg">加载代币中...</Text>
                       <Text fontSize="sm" color={mutedColor} mt={2}>正在从区块链获取最新数据</Text>
                     </Flex>
                   ) : !walletConnected ? (
                     <Alert status="info" borderRadius="lg" bg={secondaryBgColor}>
                       <AlertIcon color={accentColor} />
                       <Box>
                         <AlertTitle color={textColor}>Solana钱包未连接</AlertTitle>
                         <AlertDescription color={mutedColor}>
                           请先连接您的Phantom钱包以查看Solana代币资产
                         </AlertDescription>
                       </Box>
                     </Alert>
                   ) : filteredTokens.length === 0 ? (
                     <Alert status="warning" borderRadius="lg" bg={secondaryBgColor}>
                       <AlertIcon />
                       <Box>
                         <AlertTitle color={textColor}>没有找到代币</AlertTitle>
                         <AlertDescription color={mutedColor}>
                           {searchQuery ? '尝试使用不同的搜索词或清除搜索条件' : '您的钱包中暂无代币资产'}
                         </AlertDescription>
                       </Box>
                     </Alert>
                   ) : (
                     <Box overflowX="auto">
                       <Table variant="simple" size={isMobile ? 'sm' : 'md'}>
                         <Thead>
                           <Tr>
                             <Th color={mutedColor} fontWeight="semibold">Solana代币</Th>
                             <Th isNumeric color={mutedColor} fontWeight="semibold">余额</Th>
                             <Th isNumeric color={mutedColor} fontWeight="semibold">价值 (USD)</Th>
                             {!isMobile && <Th color={mutedColor} fontWeight="semibold">SPL代币地址</Th>}
                             <Th></Th>
                           </Tr>
                         </Thead>
                         <Tbody>
                           {filteredTokens.map((token, index) => (
                             <Tr key={index} _hover={{ bg: secondaryBgColor }} transition="all 0.2s">
                               <Td>
                                 <Flex align="center">
                                   <Image
                                     src={token.TokenInfo?.logoURI || DEFAULT_TOKEN_ICON}
                                     alt={token.TokenInfo?.symbol || 'Token'}
                                     boxSize={isMobile ? '32px' : '40px'}
                                     borderRadius="full"
                                     mr={3}
                                     fallbackSrc={DEFAULT_TOKEN_ICON}
                                     border="2px solid"
                                     borderColor={borderColor}
                                   />
                                   <VStack align="start" spacing={0}>
                                     <Text fontWeight="semibold" color={textColor}>
                                       {token.TokenInfo?.symbol || formatAddress(token.Mint)}
                                     </Text>
                                     <Text fontSize="xs" color={mutedColor}>
                                       {token.TokenInfo?.name || '未知代币'}
                                     </Text>
                                   </VStack>
                                 </Flex>
                               </Td>
                               <Td isNumeric>
                                 <VStack align="end" spacing={0}>
                                   <Text fontWeight="medium" color={textColor}>
                                     {formatNumber(token.UIAmount)}
                                   </Text>
                                   <Text fontSize="xs" color={mutedColor}>
                                     {token.TokenInfo?.symbol || 'TOKEN'}
                                   </Text>
                                 </VStack>
                               </Td>
                               <Td isNumeric>
                                 <Text fontWeight="medium" color={token.USDValue ? successColor : mutedColor}>
                                   {token.USDValue ? `$${token.USDValue.toFixed(2)}` : '-'}
                                 </Text>
                               </Td>
                               {!isMobile && (
                                 <Td>
                                   <HStack>
                                     <Text fontSize="sm" color={mutedColor}>
                                       {formatAddress(token.Mint)}
                                     </Text>
                                     <IconButton
                                       icon={<ExternalLinkIcon />}
                                       size="xs"
                                       variant="ghost"
                                       aria-label="查看合约"
                                       onClick={() => window.open(`https://solscan.io/token/${token.Mint}`, '_blank')}
                                     />
                                   </HStack>
                                 </Td>
                               )}
                               <Td>
                                 <Button 
                                   size="sm" 
                                   variant="outline" 
                                   colorScheme="blue"
                                   isDisabled={!token.UIAmount || token.UIAmount === 0}
                                 >
                                   发送
                                 </Button>
                               </Td>
                             </Tr>
                           ))}
                         </Tbody>
                       </Table>
                     </Box>
                   )}
                 </CardBody>
          </Card>
        </VStack>
      </Container>
    );
 };
 
 export default TokensPage;
