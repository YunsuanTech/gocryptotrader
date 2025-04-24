import React, { useState, useEffect } from 'react';
import {
  Box,
  Flex,
  Grid,
  GridItem,
  Heading,
  Text,
  Input,
  InputGroup,
  InputRightElement,
  Button,
  useColorModeValue,
  Stack,
  Icon,
  Divider,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  Image,
  Badge,
} from '@chakra-ui/react';
import { SearchIcon, ChevronDownIcon, InfoIcon } from '@chakra-ui/icons';
import TradingPairsList from '../components/TradingPairsList';
import TradingPairDetail from '../components/TradingPairDetail';

// 模拟交易对数据
const mockTradingPairs = [
  { id: 1, base: 'BTC', quote: 'USDT', price: 63245.78, change: 2.34, volume: 1245678.90, high: 63500.00, low: 62100.00, marketCap: '1.2T', logo: 'https://cryptologos.cc/logos/bitcoin-btc-logo.png' },
  { id: 2, base: 'ETH', quote: 'USDT', price: 3078.45, change: -1.23, volume: 987654.32, high: 3120.00, low: 3050.00, marketCap: '370B', logo: 'https://cryptologos.cc/logos/ethereum-eth-logo.png' },
  { id: 3, base: 'SOL', quote: 'USDT', price: 142.67, change: 5.67, volume: 456789.01, high: 145.00, low: 135.00, marketCap: '61B', logo: 'https://cryptologos.cc/logos/solana-sol-logo.png' },
  { id: 4, base: 'XRP', quote: 'USDT', price: 0.5234, change: 0.87, volume: 345678.90, high: 0.53, low: 0.51, marketCap: '28B', logo: 'https://cryptologos.cc/logos/xrp-xrp-logo.png' },
  { id: 5, base: 'ADA', quote: 'USDT', price: 0.4523, change: -2.45, volume: 234567.89, high: 0.47, low: 0.44, marketCap: '16B', logo: 'https://cryptologos.cc/logos/cardano-ada-logo.png' },
  { id: 6, base: 'DOGE', quote: 'USDT', price: 0.1345, change: 3.56, volume: 123456.78, high: 0.14, low: 0.13, marketCap: '19B', logo: 'https://cryptologos.cc/logos/dogecoin-doge-logo.png' },
  { id: 7, base: 'DOT', quote: 'USDT', price: 6.78, change: -0.45, volume: 98765.43, high: 6.85, low: 6.70, marketCap: '8.5B', logo: 'https://cryptologos.cc/logos/polkadot-new-dot-logo.png' },
  { id: 8, base: 'LINK', quote: 'USDT', price: 14.56, change: 1.23, volume: 87654.32, high: 14.70, low: 14.30, marketCap: '8.2B', logo: 'https://cryptologos.cc/logos/chainlink-link-logo.png' },
];

const TradingPage: React.FC = () => {
  const [selectedPair, setSelectedPair] = useState(mockTradingPairs[0]);
  const [searchQuery, setSearchQuery] = useState('');
  const [timeframe, setTimeframe] = useState('15分钟');
  const [filteredPairs, setFilteredPairs] = useState(mockTradingPairs);
  
  const bgColor = useColorModeValue('white', '#10141f');
  const textColor = useColorModeValue('gray.800', 'white');
  const borderColor = useColorModeValue('gray.200', 'gray.700');
  const buttonBgActive = useColorModeValue('blue.500', 'blue.400');
  const buttonBgInactive = useColorModeValue('gray.100', 'gray.700');
  
  useEffect(() => {
    // 根据搜索查询过滤交易对
    const filtered = mockTradingPairs.filter(pair => 
      pair.base.toLowerCase().includes(searchQuery.toLowerCase()) || 
      pair.quote.toLowerCase().includes(searchQuery.toLowerCase())
    );
    setFilteredPairs(filtered);
  }, [searchQuery]);

  const handlePairSelect = (pair: any) => {
    setSelectedPair(pair);
  };

  return (
    <Box bg={bgColor} minH="100vh" p={4}>
      <Grid
        templateColumns={{ base: '1fr', md: '350px 1fr' }}
        gap={6}
      >
        {/* 左侧交易对列表 */}
        <GridItem w="100%" bg={bgColor} borderRadius="lg" boxShadow="md" p={4}>
          <Stack spacing={4}>
            <Flex justify="space-between" align="center">
              <Heading size="md" color={textColor}>交易对</Heading>
              <Flex>
                <Menu>
                  <MenuButton as={Button} rightIcon={<ChevronDownIcon />} size="sm" variant="outline">
                    所有市场
                  </MenuButton>
                  <MenuList>
                    <MenuItem>USDT 市场</MenuItem>
                    <MenuItem>BTC 市场</MenuItem>
                    <MenuItem>ETH 市场</MenuItem>
                  </MenuList>
                </Menu>
              </Flex>
            </Flex>
            
            <InputGroup>
              <Input 
                placeholder="搜索交易对..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                borderColor={borderColor}
              />
              <InputRightElement>
                <SearchIcon color="gray.500" />
              </InputRightElement>
            </InputGroup>
            
            <Divider borderColor={borderColor} />
            
            <TradingPairsList 
              pairs={filteredPairs} 
              selectedPair={selectedPair} 
              onSelectPair={handlePairSelect} 
            />
          </Stack>
        </GridItem>
        
        {/* 右侧交易详情 */}
        <GridItem w="100%" bg={bgColor} borderRadius="lg" boxShadow="md">
          <Stack spacing={0}>
            <Flex p={4} justify="space-between" align="center" borderBottom="1px solid" borderColor={borderColor}>
              <Flex align="center">
                <Image src={selectedPair.logo} boxSize="32px" mr={2} borderRadius="full" />
                <Heading size="md" color={textColor}>{selectedPair.base}/{selectedPair.quote}</Heading>
                <Badge ml={2} colorScheme={selectedPair.change >= 0 ? 'green' : 'red'} borderRadius="full" px={2}>
                  {selectedPair.change >= 0 ? '+' : ''}{selectedPair.change}%
                </Badge>
              </Flex>
              
              <Flex>
                <Stack direction="row" spacing={1}>
                  <Button 
                    size="sm" 
                    bg={timeframe === '5分钟' ? buttonBgActive : buttonBgInactive}
                    color={timeframe === '5分钟' ? 'white' : textColor}
                    onClick={() => setTimeframe('5分钟')}
                  >
                    5分钟
                  </Button>
                  <Button 
                    size="sm" 
                    bg={timeframe === '15分钟' ? buttonBgActive : buttonBgInactive}
                    color={timeframe === '15分钟' ? 'white' : textColor}
                    onClick={() => setTimeframe('15分钟')}
                  >
                    15分钟
                  </Button>
                  <Button 
                    size="sm" 
                    bg={timeframe === '1小时' ? buttonBgActive : buttonBgInactive}
                    color={timeframe === '1小时' ? 'white' : textColor}
                    onClick={() => setTimeframe('1小时')}
                  >
                    1小时
                  </Button>
                  <Button 
                    size="sm" 
                    bg={timeframe === '1天' ? buttonBgActive : buttonBgInactive}
                    color={timeframe === '1天' ? 'white' : textColor}
                    onClick={() => setTimeframe('1天')}
                  >
                    1天
                  </Button>
                </Stack>
              </Flex>
            </Flex>
            
            <TradingPairDetail pair={selectedPair} timeframe={timeframe} />
          </Stack>
        </GridItem>
      </Grid>
    </Box>
  );
};

export default TradingPage;