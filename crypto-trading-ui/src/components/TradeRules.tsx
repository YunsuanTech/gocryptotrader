import { Box, Heading, VStack, Text, Divider, Badge, HStack, Icon } from '@chakra-ui/react';
import { FaCheckCircle, FaTimesCircle, FaInfoCircle } from 'react-icons/fa';

const TradeRules = () => {
  return (
    <Box p={4} bg="background.dark" h="100%">
      <Heading size="lg" mb={6} color="whiteAlpha.900">
        交易规则
      </Heading>
      
      {/* 买入规则 */}
      <Box mb={6}>
        <HStack mb={3}>
          <Badge colorScheme="green" fontSize="md" px={2} py={1}>
            买入规则
          </Badge>
        </HStack>
        
        <VStack align="start" spacing={3} pl={2}>
          <HStack>
            <Icon as={FaCheckCircle} color="trading.up" />
            <Text>价格低于20日均线</Text>
          </HStack>
          <HStack>
            <Icon as={FaCheckCircle} color="trading.up" />
            <Text>RSI指标低于30</Text>
          </HStack>
          <HStack>
            <Icon as={FaCheckCircle} color="trading.up" />
            <Text>成交量增加50%以上</Text>
          </HStack>
          <HStack>
            <Icon as={FaInfoCircle} color="blue.400" />
            <Text>市场情绪指数低于40时谨慎买入</Text>
          </HStack>
        </VStack>
      </Box>
      
      <Divider my={4} borderColor="border.dark" />
      
      {/* 卖出规则 */}
      <Box mb={6}>
        <HStack mb={3}>
          <Badge colorScheme="red" fontSize="md" px={2} py={1}>
            卖出规则
          </Badge>
        </HStack>
        
        <VStack align="start" spacing={3} pl={2}>
          <HStack>
            <Icon as={FaTimesCircle} color="trading.down" />
            <Text>价格高于50日均线</Text>
          </HStack>
          <HStack>
            <Icon as={FaTimesCircle} color="trading.down" />
            <Text>RSI指标高于70</Text>
          </HStack>
          <HStack>
            <Icon as={FaTimesCircle} color="trading.down" />
            <Text>价格涨幅超过30%</Text>
          </HStack>
          <HStack>
            <Icon as={FaInfoCircle} color="blue.400" />
            <Text>市场情绪指数高于80时考虑卖出</Text>
          </HStack>
        </VStack>
      </Box>

      <Divider my={4} borderColor="border.dark" />
      
      {/* 风险管理 */}
      <Box>
        <HStack mb={3}>
          <Badge colorScheme="purple" fontSize="md" px={2} py={1}>
            风险管理
          </Badge>
        </HStack>
        
        <VStack align="start" spacing={3} pl={2}>
          <HStack>
            <Icon as={FaInfoCircle} color="purple.400" />
            <Text>单笔交易不超过总资金的5%</Text>
          </HStack>
          <HStack>
            <Icon as={FaInfoCircle} color="purple.400" />
            <Text>设置15%止损位</Text>
          </HStack>
          <HStack>
            <Icon as={FaInfoCircle} color="purple.400" />
            <Text>高波动市场减少仓位</Text>
          </HStack>
        </VStack>
      </Box>
    </Box>
  );
};

export default TradeRules;