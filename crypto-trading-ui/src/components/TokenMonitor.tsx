import { Box, Heading, VStack, HStack, Text, Badge, Divider, useColorModeValue } from '@chakra-ui/react';
import { FaArrowUp, FaArrowDown } from 'react-icons/fa';

// 模拟代币数据
const tokenData = [
  { id: 1, name: 'Bitcoin', symbol: 'BTC', price: 65432.10, change: 2.5, volume: '12.5B', status: 'up' },
  { id: 2, name: 'Ethereum', symbol: 'ETH', price: 3521.45, change: -1.2, volume: '8.3B', status: 'down' },
  { id: 3, name: 'Binance Coin', symbol: 'BNB', price: 612.78, change: 0.8, volume: '2.1B', status: 'up' },
  { id: 4, name: 'Solana', symbol: 'SOL', price: 142.35, change: 5.2, volume: '3.7B', status: 'up' },
  { id: 5, name: 'Cardano', symbol: 'ADA', price: 0.58, change: -0.5, volume: '1.2B', status: 'down' },
  { id: 6, name: 'XRP', symbol: 'XRP', price: 0.52, change: 1.3, volume: '1.8B', status: 'up' },
  { id: 7, name: 'Dogecoin', symbol: 'DOGE', price: 0.12, change: -2.1, volume: '0.9B', status: 'down' },
  { id: 8, name: 'Polkadot', symbol: 'DOT', price: 7.82, change: 0.3, volume: '0.7B', status: 'up' },
  { id: 9, name: 'Avalanche', symbol: 'AVAX', price: 35.67, change: 3.2, volume: '1.5B', status: 'up' },
  { id: 10, name: 'Chainlink', symbol: 'LINK', price: 15.23, change: -0.8, volume: '0.6B', status: 'down' },
  { id: 11, name: 'Polygon', symbol: 'MATIC', price: 0.78, change: 1.5, volume: '0.8B', status: 'up' },
  { id: 12, name: 'Uniswap', symbol: 'UNI', price: 8.45, change: -1.7, volume: '0.5B', status: 'down' },
];

const TokenMonitor = () => {
  return (
    <Box p={4} bg="background.dark" borderRadius="md">
      <Heading size="lg" mb={6} color="whiteAlpha.900">
        代币监控
      </Heading>
      
      <VStack spacing={4} align="stretch">
        {tokenData.map((token) => (
          <Box 
            key={token.id} 
            p={4} 
            bg="background.card" 
            borderRadius="md" 
            borderLeft="4px solid" 
            borderColor={token.status === 'up' ? 'trading.up' : 'trading.down'}
            _hover={{ bg: 'background.hover', transform: 'translateY(-2px)' }}
            transition="all 0.3s ease"
            className="token-card"
            boxShadow="0 2px 10px rgba(0, 0, 0, 0.2)"
            mb={2}
          >
            <HStack justifyContent="space-between" mb={2}>
              <HStack>
                <Text fontWeight="bold" fontSize="lg">{token.name}</Text>
                <Badge colorScheme={token.status === 'up' ? 'green' : 'red'} variant="subtle">
                  {token.symbol}
                </Badge>
              </HStack>
              <Text 
                fontWeight="bold" 
                fontSize="lg"
                color={token.status === 'up' ? 'trading.up' : 'trading.down'}
              >
                ${token.price.toLocaleString()}
              </Text>
            </HStack>

            <HStack justifyContent="space-between">
              <HStack>
                <Text fontSize="sm" color="whiteAlpha.700">24h变化:</Text>
                <HStack color={token.change > 0 ? 'trading.up' : 'trading.down'}>
                  {token.change > 0 ? <FaArrowUp size="12px" /> : <FaArrowDown size="12px" />}
                  <Text fontSize="sm" fontWeight="medium">{Math.abs(token.change)}%</Text>
                </HStack>
              </HStack>
              <Text fontSize="sm" color="whiteAlpha.700">成交量: {token.volume}</Text>
            </HStack>
          </Box>
        ))}
      </VStack>
    </Box>
  );
};

export default TokenMonitor;
