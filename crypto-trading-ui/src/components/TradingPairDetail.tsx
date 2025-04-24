import React from 'react';
import {
  Box,
  Flex,
  Text,
  Heading,
  Badge,
  HStack,
  VStack,
  useColorModeValue,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  StatArrow,
  Grid,
  GridItem,
  Divider,
  Button,
  Icon,
  Stack,
  Image,
} from '@chakra-ui/react';
import { InfoIcon, ArrowUpIcon, ArrowDownIcon } from '@chakra-ui/icons';

interface TradingPair {
  id: number;
  base: string;
  quote: string;
  price: number;
  change: number;
  volume: number;
  high: number;
  low: number;
  marketCap: string;
  logo: string;
}

interface TradingPairDetailProps {
  pair: TradingPair;
  timeframe: string;
}

const TradingPairDetail: React.FC<TradingPairDetailProps> = ({ pair, timeframe }) => {
  const bgColor = useColorModeValue('white', '#10141f');
  const textColor = useColorModeValue('gray.800', 'white');
  const secondaryTextColor = useColorModeValue('gray.600', 'gray.400');
  const borderColor = useColorModeValue('gray.200', 'gray.700');
  const cardBgColor = useColorModeValue('gray.50', '#1a1f2e');
  const buttonBgColor = useColorModeValue('blue.500', '#512da8');
  const highlightColor = useColorModeValue('green.500', 'green.300');
  
  // 格式化数字显示
  const formatNumber = (num: number, digits: number = 2) => {
    return num.toLocaleString(undefined, { minimumFractionDigits: digits, maximumFractionDigits: digits });
  };

  // 根据价格决定显示的小数位数
  const getPriceDecimals = (price: number) => {
    if (price >= 1000) return 2;
    if (price >= 100) return 2;
    if (price >= 10) return 3;
    if (price >= 1) return 4;
    if (price >= 0.1) return 5;
    if (price >= 0.01) return 6;
    return 8;
  };

  return (
    <Box bg={bgColor} borderRadius="xl" boxShadow="0 8px 20px rgba(0,0,0,.2)" overflow="hidden">
      {/* 交易对信息头部 */}
      <Flex p={4} bg={cardBgColor} alignItems="center" justifyContent="space-between">
        <Flex alignItems="center">
          {pair.logo && (
            <Box mr={3}>
              <Image src={pair.logo} boxSize="40px" borderRadius="full" alt={pair.base} />  
            </Box>
          )}
          <Box>
            <Heading size="md" color={textColor}>{pair.base}/{pair.quote}</Heading>
            <Flex alignItems="center" mt={1}>
              <Text fontSize="xl" fontWeight="bold" color={textColor} mr={2}>
                {formatNumber(pair.price, getPriceDecimals(pair.price))}
              </Text>
              <Badge colorScheme={pair.change >= 0 ? 'green' : 'red'} px={2} py={0.5} borderRadius="md">
                <Flex alignItems="center">
                  {pair.change >= 0 ? <ArrowUpIcon boxSize={3} mr={1} /> : <ArrowDownIcon boxSize={3} mr={1} />}
                  {Math.abs(pair.change).toFixed(2)}%
                </Flex>
              </Badge>
            </Flex>
          </Box>
        </Flex>
        <Button size="sm" bg={buttonBgColor} color="white" _hover={{ bg: 'purple.600' }}>
          交易规则
        </Button>
      </Flex>
      
      {/* 市场数据统计 */}
      <Grid templateColumns={{ base: 'repeat(2, 1fr)', md: 'repeat(4, 1fr)' }} gap={4} p={4} bg={bgColor}>
        <Stat>
          <StatLabel color={secondaryTextColor}>24小时最高</StatLabel>
          <StatNumber fontSize="lg" color={textColor}>
            {formatNumber(pair.high, getPriceDecimals(pair.high))}
          </StatNumber>
        </Stat>
        
        <Stat>
          <StatLabel color={secondaryTextColor}>24小时最低</StatLabel>
          <StatNumber fontSize="lg" color={textColor}>
            {formatNumber(pair.low, getPriceDecimals(pair.low))}
          </StatNumber>
        </Stat>
        
        <Stat>
          <StatLabel color={secondaryTextColor}>24小时成交量</StatLabel>
          <StatNumber fontSize="lg" color={textColor}>
            {formatNumber(pair.volume, 0)}
          </StatNumber>
        </Stat>
        
        <Stat>
          <StatLabel color={secondaryTextColor}>市值</StatLabel>
          <StatNumber fontSize="lg" color={textColor}>
            {pair.marketCap}
          </StatNumber>
        </Stat>
      </Grid>
      
      <Divider borderColor={borderColor} />
      
      {/* 交易规则区域 */}
      <Box p={5} bg={bgColor}>
        <Heading size="md" mb={4} color={textColor}>交易规则</Heading>
        
        <Grid templateColumns={{ base: '1fr', md: 'repeat(2, 1fr)' }} gap={6}>
          {/* 买入规则 */}
          <Box p={5} bg={cardBgColor} borderRadius="xl" boxShadow="md">
            <Flex justify="space-between" align="center" mb={4}>
              <Heading size="sm" color={textColor}>买入 {pair.base}</Heading>
              <Badge colorScheme="green" px={2} py={1} borderRadius="md">推荐</Badge>
            </Flex>
            
            <VStack spacing={4} align="stretch">
              <Box>
                <Text fontWeight="medium" color={textColor} mb={1}>交易限制</Text>
                <Flex justify="space-between" bg="rgba(0,0,0,0.1)" p={2} borderRadius="md">
                  <Text color={secondaryTextColor}>最小交易量</Text>
                  <Text color={textColor}>0.001 {pair.base}</Text>
                </Flex>
              </Box>
              
              <Box>
                <Text fontWeight="medium" color={textColor} mb={1}>手续费</Text>
                <Flex justify="space-between" bg="rgba(0,0,0,0.1)" p={2} borderRadius="md">
                  <Text color={secondaryTextColor}>标准费率</Text>
                  <Text color={textColor}>0.2%</Text>
                </Flex>
                <Flex justify="space-between" bg="rgba(0,0,0,0.1)" p={2} borderRadius="md" mt={1}>
                  <Text color={secondaryTextColor}>VIP费率</Text>
                  <Text color={highlightColor}>0.1%</Text>
                </Flex>
              </Box>
              
              <Box>
                <Text fontWeight="medium" color={textColor} mb={1}>价格信息</Text>
                <Flex justify="space-between" bg="rgba(0,0,0,0.1)" p={2} borderRadius="md">
                  <Text color={secondaryTextColor}>当前价格</Text>
                  <Text color={textColor}>{formatNumber(pair.price, getPriceDecimals(pair.price))} {pair.quote}</Text>
                </Flex>
                <Flex justify="space-between" bg="rgba(0,0,0,0.1)" p={2} borderRadius="md" mt={1}>
                  <Text color={secondaryTextColor}>价格波动</Text>
                  <Text color={pair.change >= 0 ? 'green.400' : 'red.400'}>{pair.change >= 0 ? '+' : ''}{pair.change}%</Text>
                </Flex>
              </Box>
              
              <Button bg={buttonBgColor} color="white" size="md" width="100%" _hover={{ bg: 'purple.600' }}>
                开始交易
              </Button>
            </VStack>
          </Box>
          
          {/* 交易提示 */}
          <Box p={5} bg={cardBgColor} borderRadius="xl" boxShadow="md">
            <Heading size="sm" mb={4} color={textColor}>交易提示</Heading>
            
            <VStack spacing={4} align="stretch">
              <Box>
                <Flex align="center" mb={2}>
                  <InfoIcon color={highlightColor} mr={2} />
                  <Text fontWeight="medium" color={textColor}>风险提示</Text>
                </Flex>
                <Text color={secondaryTextColor} fontSize="sm">
                  数字货币交易存在较高风险，价格波动较大，请谨慎投资，合理配置资产。
                </Text>
              </Box>
              
              <Box>
                <Flex align="center" mb={2}>
                  <InfoIcon color={highlightColor} mr={2} />
                  <Text fontWeight="medium" color={textColor}>交易时间</Text>
                </Flex>
                <Text color={secondaryTextColor} fontSize="sm">
                  交易市场全天24小时开放，无休市时间，可随时进行买卖操作。
                </Text>
              </Box>
              
              <Box>
                <Flex align="center" mb={2}>
                  <InfoIcon color={highlightColor} mr={2} />
                  <Text fontWeight="medium" color={textColor}>资金安全</Text>
                </Flex>
                <Text color={secondaryTextColor} fontSize="sm">
                  平台采用多重安全防护措施，冷热钱包分离存储，保障用户资产安全。
                </Text>
              </Box>
              
              <Box>
                <Flex align="center" mb={2}>
                  <InfoIcon color={highlightColor} mr={2} />
                  <Text fontWeight="medium" color={textColor}>交易建议</Text>
                </Flex>
                <Text color={secondaryTextColor} fontSize="sm">
                  建议使用限价单进行交易，可以更好地控制买入价格和风险。
                </Text>
              </Box>
            </VStack>
          </Box>
        </Grid>
      </Box>
    </Box>
  );
};

export default TradingPairDetail;