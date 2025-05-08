import { Box, Flex, Heading, VStack, HStack, Text, Badge, Divider, useColorModeValue, Button, Table, Thead, Tbody, Tr, Th, Td, Icon, Input } from '@chakra-ui/react';
import { useState } from 'react';
import { FaArrowUp, FaArrowDown, FaSearch, FaFilter, FaInfoCircle, FaChevronDown, FaChevronRight } from 'react-icons/fa';
import './HomePage.css';

// 模拟代币数据
const tokenData = [
  { id: 1, name: 'Bitcoin', symbol: 'BTC', price: 65432.10, change: 2.5, volume: '12.5B', status: 'up', apy: 4.2, weight: 25 },
  { id: 2, name: 'Ethereum', symbol: 'ETH', price: 3521.45, change: -1.2, volume: '8.3B', status: 'down', apy: 3.8, weight: 20 },
  { id: 3, name: 'Binance Coin', symbol: 'BNB', price: 612.78, change: 0.8, volume: '2.1B', status: 'up', apy: 5.1, weight: 15 },
  { id: 4, name: 'Solana', symbol: 'SOL', price: 142.35, change: 5.2, volume: '3.7B', status: 'up', apy: 6.5, weight: 12 },
  { id: 5, name: 'Cardano', symbol: 'ADA', price: 0.58, change: -0.5, volume: '1.2B', status: 'down', apy: 2.9, weight: 8 },
  { id: 6, name: 'XRP', symbol: 'XRP', price: 0.52, change: 1.3, volume: '1.8B', status: 'up', apy: 3.2, weight: 7 },
];

// 代币卡片组件
interface TokenCardProps {
  token: {
    id: number;
    name: string;
    symbol: string;
    price: number;
    change: number;
    volume: string;
    status: string;
    apy: number;
    weight: number;
  };
  isActive?: boolean;
}

const TokenCard = ({ token, isActive = false }: TokenCardProps) => {
  return (
    <Box 
      className={`token-card ${isActive ? 'active' : ''}`}
      p={3} 
      borderRadius="lg"
      bg="background.card"
      border="1px solid"
      borderColor="border.dark"
      transition="all 0.3s ease"
      _hover={{ borderColor: 'border.hover' }}
      cursor="pointer"
      w="100%"
      maxW="480px"
    >
      <Flex alignItems="center" gap={2} w="100%">
        <Box 
          w="24px" 
          h="24px" 
          borderRadius="full" 
          bg={`${token.symbol.toLowerCase()}.color`} 
          mr={2}
        />
        <Text fontSize="sm">
          <Text as="span" fontWeight="medium" mr={1.5}>{token.symbol}</Text>
          现已可在平台交易
        </Text>
        <Icon as={FaChevronRight} ml="auto" color="gray.500" />
      </Flex>
    </Box>
  );
};

// 资产表格组件
const AssetTable = () => {
  const [activeTab, setActiveTab] = useState('global');
  const [showLend, setShowLend] = useState(true);
  
  return (
    <Box w="full" mt={8}>
      <Flex justifyContent="space-between" alignItems="center" mb={6}>
        <Flex 
          as="div"
          role="group"
          bg="background.gray" 
          p={1.5} 
          borderRadius="lg"
          alignItems="center"
          justifyContent="center"
          gap={1}
        >
          <Button 
            size="sm"
            variant={activeTab === 'global' ? 'solid' : 'ghost'}
            bg={activeTab === 'global' ? 'accent.primary' : 'transparent'}
            color={activeTab === 'global' ? 'white' : 'gray.400'}
            h="8"
            px={3}
            onClick={() => setActiveTab('global')}
          >
            全局
          </Button>
          <Button 
            size="sm"
            variant={activeTab === 'isolated' ? 'solid' : 'ghost'}
            bg={activeTab === 'isolated' ? 'accent.primary' : 'transparent'}
            color={activeTab === 'isolated' ? 'white' : 'gray.400'}
            h="8"
            px={3}
            onClick={() => setActiveTab('isolated')}
          >
            隔离
          </Button>
          <Button 
            size="sm"
            variant={activeTab === 'staked' ? 'solid' : 'ghost'}
            bg={activeTab === 'staked' ? 'accent.primary' : 'transparent'}
            color={activeTab === 'staked' ? 'white' : 'gray.400'}
            h="8"
            px={3}
            onClick={() => setActiveTab('staked')}
          >
            原生质押
          </Button>
        </Flex>
        
        <Flex alignItems="center" gap={3} ml={10}>
          <Text 
            fontSize="sm" 
            fontWeight="normal" 
            color={showLend ? 'white' : 'gray.400'}
            cursor="pointer"
            onClick={() => setShowLend(true)}
          >
            借出
          </Text>
          <Box 
            as="button"
            role="switch"
            w="9"
            h="5"
            bg="chartreuse"
            borderRadius="full"
            position="relative"
            onClick={() => setShowLend(!showLend)}
          >
            <Box 
              w="4"
              h="4"
              bg="white"
              borderRadius="full"
              position="absolute"
              top="0.5"
              left={showLend ? "0.5" : "auto"}
              right={!showLend ? "0.5" : "auto"}
              transition="all 0.2s ease"
            />
          </Box>
          <Text 
            fontSize="sm" 
            fontWeight="normal" 
            color={!showLend ? 'white' : 'gray.400'}
            cursor="pointer"
            onClick={() => setShowLend(false)}
          >
            借入
          </Text>
        </Flex>
        
        <Box position="relative" color="gray.400" ml="auto" mr={4}>
          <Button
            variant="ghost"
            w="9"
            h="9"
            display="flex"
            alignItems="center"
            justifyContent="center"
            borderRadius="md"
            _hover={{ bg: 'background.hover' }}
          >
            <Icon as={FaSearch} boxSize={4} />
          </Button>
        </Box>
        
        <Button
          variant="ghost"
          display="flex"
          alignItems="center"
          gap={2}
          bg="background.card"
          px={3}
          py={2}
          h="9"
          borderRadius="md"
          _hover={{ bg: 'background.hover' }}
        >
          <Icon as={FaFilter} boxSize={4} />
          <Text>所有代币</Text>
          <Icon as={FaChevronDown} boxSize={3} ml="auto" opacity={0.5} />
        </Button>
      </Flex>
      
      <Box w="full" overflowX="auto">
        <Table variant="simple">
          <Thead>
            <Tr>
              <Th width="210px">
                <Text color="gray.400" fontSize="sm" fontWeight="light">资产</Text>
              </Th>
              <Th width="170px">
                <Flex color="gray.400" fontSize="sm" fontWeight="light" justifyContent="flex-end" alignItems="center" gap={1} cursor="pointer">
                  <Text>价格</Text>
                  <Icon as={FaInfoCircle} boxSize={3.5} />
                </Flex>
              </Th>
              <Th width="120px">
                <Flex color="gray.400" fontSize="sm" fontWeight="light" justifyContent="flex-end" alignItems="center" gap={1} cursor="pointer">
                  <Text>APY</Text>
                  <Icon as={FaInfoCircle} boxSize={3.5} />
                </Flex>
              </Th>
              <Th width="150px">
                <Flex color="gray.400" fontSize="sm" fontWeight="light" justifyContent="flex-end" alignItems="center" gap={1} cursor="pointer">
                  <Text>权重</Text>
                  <Icon as={FaInfoCircle} boxSize={3.5} />
                </Flex>
              </Th>
            </Tr>
          </Thead>
          <Tbody>
            {tokenData.map((token) => (
              <Tr key={token.id} _hover={{ bg: 'background.hover' }} cursor="pointer">
                <Td>
                  <Flex alignItems="center" gap={2}>
                    <Box 
                      w="32px" 
                      h="32px" 
                      borderRadius="full" 
                      bg={`${token.symbol.toLowerCase()}.color`} 
                    />
                    <Box>
                      <Text fontWeight="medium">{token.name}</Text>
                      <Text fontSize="xs" color="gray.400">{token.symbol}</Text>
                    </Box>
                  </Flex>
                </Td>
                <Td isNumeric>
                  <Text fontWeight="medium">${token.price.toLocaleString()}</Text>
                </Td>
                <Td isNumeric>
                  <Text 
                    fontWeight="medium"
                    color={token.apy > 4 ? 'trading.up' : 'white'}
                  >
                    {token.apy}%
                  </Text>
                </Td>
                <Td isNumeric>
                  <Text fontWeight="medium">{token.weight}%</Text>
                </Td>
              </Tr>
            ))}
          </Tbody>
        </Table>
      </Box>
    </Box>
  );
};

const HomePage = () => {
  
  return (
    <Box w="100%" h="100%" overflow="auto" className="home-container custom-scrollbar">
      <Flex 
        w="full" 
        flexDir="column" 
        justifyContent="center" 
        alignItems="center"
      >
                  {/* 资产列表表格 */}
                  <Box pt={4} pb={16} px={4} w="full" maxW="7xl">
            <AssetTable />
          </Box>

      </Flex>
    </Box>
  );
};

export default HomePage;