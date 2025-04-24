import React from 'react';
import {
  Box,
  Flex,
  Text,
  Stack,
  Image,
  Badge,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  useColorModeValue,
  Divider,
  Tooltip,
} from '@chakra-ui/react';

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

interface TradingPairsListProps {
  pairs: TradingPair[];
  selectedPair: TradingPair;
  onSelectPair: (pair: TradingPair) => void;
}

const TradingPairsList: React.FC<TradingPairsListProps> = ({ pairs, selectedPair, onSelectPair }) => {
  const bgSelected = useColorModeValue('blue.50', 'blue.900');
  const textColor = useColorModeValue('gray.800', 'white');
  const secondaryTextColor = useColorModeValue('gray.600', 'gray.400');
  const borderColor = useColorModeValue('gray.200', 'gray.700');
  
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
    <Box 
      overflowY="auto" 
      maxH="calc(100vh - 250px)" 
      css={{
        '&::-webkit-scrollbar': {
          width: '4px',
        },
        '&::-webkit-scrollbar-track': {
          width: '6px',
        },
        '&::-webkit-scrollbar-thumb': {
          background: borderColor,
          borderRadius: '24px',
        },
      }}
    >
      <Table variant="simple" size="sm">
        <Thead>
          <Tr>
            <Th color={secondaryTextColor}>交易对</Th>
            <Th color={secondaryTextColor} isNumeric>最新价格</Th>
            <Th color={secondaryTextColor} isNumeric>24h涨跌</Th>
          </Tr>
        </Thead>
        <Tbody>
          {pairs.map((pair) => (
            <Tr 
              key={pair.id}
              onClick={() => onSelectPair(pair)}
              bg={selectedPair.id === pair.id ? bgSelected : 'transparent'}
              cursor="pointer"
              _hover={{ bg: bgSelected }}
              transition="background-color 0.2s"
            >
              <Td>
                <Flex align="center">
                  <Image src={pair.logo} boxSize="24px" mr={2} borderRadius="full" />
                  <Text fontWeight="medium" color={textColor}>{pair.base}</Text>
                  <Text color={secondaryTextColor} ml={1}>/{pair.quote}</Text>
                </Flex>
              </Td>
              <Td isNumeric fontWeight="medium" color={textColor}>
                {formatNumber(pair.price, getPriceDecimals(pair.price))}
              </Td>
              <Td isNumeric>
                <Tooltip label={`24h成交量: ${formatNumber(pair.volume)} ${pair.quote}`} placement="top">
                  <Text color={pair.change >= 0 ? 'green.500' : 'red.500'} fontWeight="medium">
                    {pair.change >= 0 ? '+' : ''}{pair.change}%
                  </Text>
                </Tooltip>
              </Td>
            </Tr>
          ))}
        </Tbody>
      </Table>

      {pairs.length === 0 && (
        <Flex justify="center" align="center" h="200px" direction="column">
          <Text color={secondaryTextColor} fontSize="lg">没有找到匹配的交易对</Text>
        </Flex>
      )}
    </Box>
  );
};

export default TradingPairsList;