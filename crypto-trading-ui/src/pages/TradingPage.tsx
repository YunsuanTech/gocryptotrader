import { Box, Flex, useColorModeValue } from '@chakra-ui/react';
import TokenMonitor from '../components/TokenMonitor';
import TradeRules from '../components/TradeRules';
import './TradingPage.css';

const TradingPage = () => {
  return (
    <Flex 
      w="100%" 
      h="100%" 
      overflow="hidden"
      bg="background.dark"
      className="trading-container"
    >
      {/* 左侧代币监控组件 - 可滚动 */}
      <Box 
        flex="3" 
        pr={4} 
        overflowY="auto"
        h="100%"
        className="custom-scrollbar"
      >
        <TokenMonitor />
      </Box>

      {/* 右侧交易规则组件 - 固定 */}
      <Box 
        flex="1" 
        position="sticky" 
        top={0} 
        h="100%"
        borderLeft="1px solid" 
        borderColor="border.dark"
        pl={4}
        className="rule-card"
      >
        <TradeRules />
      </Box>
    </Flex>
  );
};

export default TradingPage;