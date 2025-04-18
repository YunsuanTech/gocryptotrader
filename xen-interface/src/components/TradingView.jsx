import React from 'react'
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
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
} from '@chakra-ui/react'

const TradingView = ({ symbol = 'BTC/USDT' }) => {
  const bgColor = useColorModeValue('white', 'gray.800')
  const borderColor = useColorModeValue('gray.200', 'gray.700')
  const textColor = useColorModeValue('gray.600', 'gray.400')
  
  // 模拟市场数据
  const marketData = {
    price: '65,432.21',
    change: '+2.5%',
    high: '65,980.50',
    low: '63,245.75',
    volume: '1.2B USDT',
    status: 'up'
  }
  
  return (
    <Box 
      bg={bgColor} 
      borderRadius="lg" 
      boxShadow="sm"
      borderWidth="1px"
      borderColor={borderColor}
      overflow="hidden"
      mb={6}
    >
      {/* 交易对信息头部 */}
      <Flex 
        justify="space-between" 
        align="center" 
        p={4} 
        borderBottomWidth="1px"
        borderBottomColor={borderColor}
      >
        <Box>
          <Heading size="md">{symbol}</Heading>
          <HStack spacing={4} mt={1}>
            <Text fontSize="xl" fontWeight="bold">{marketData.price}</Text>
            <Badge colorScheme={marketData.status === 'up' ? 'green' : 'red'} fontSize="md" px={2} py={1}>
              {marketData.change}
            </Badge>
          </HStack>
        </Box>
        <HStack spacing={6} display={{ base: 'none', md: 'flex' }}>
          <Stat size="sm">
            <StatLabel>24h高</StatLabel>
            <StatNumber fontSize="md">{marketData.high}</StatNumber>
          </Stat>
          <Stat size="sm">
            <StatLabel>24h低</StatLabel>
            <StatNumber fontSize="md">{marketData.low}</StatNumber>
          </Stat>
          <Stat size="sm">
            <StatLabel>24h成交量</StatLabel>
            <StatNumber fontSize="md">{marketData.volume}</StatNumber>
          </Stat>
        </HStack>
      </Flex>
      
      {/* 图表区域 */}
      <Box p={4}>
        <Tabs variant="enclosed" colorScheme="blue" size="sm">
          <TabList>
            <Tab>K线图</Tab>
            <Tab>深度图</Tab>
          </TabList>
          <TabPanels>
            <TabPanel p={0} pt={4}>
              <Box 
                height="400px" 
                position="relative"
                border="1px dashed"
                borderColor={borderColor}
                borderRadius="md"
                display="flex"
                alignItems="center"
                justifyContent="center"
              >
                <Text color={textColor}>K线图区域 (TradingView 集成)</Text>
              </Box>
            </TabPanel>
            <TabPanel p={0} pt={4}>
              <Box 
                height="400px" 
                position="relative"
                border="1px dashed"
                borderColor={borderColor}
                borderRadius="md"
                display="flex"
                alignItems="center"
                justifyContent="center"
              >
                <Text color={textColor}>深度图区域</Text>
              </Box>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </Box>
    </Box>
  )
}

export default TradingView