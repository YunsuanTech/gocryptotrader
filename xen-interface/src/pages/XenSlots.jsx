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
  CardFooter,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  Divider,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Badge,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  IconButton,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  Progress,
  useColorModeValue,
  useBreakpointValue,
} from '@chakra-ui/react';
import { ChevronDownIcon, CheckIcon, RepeatIcon, TimeIcon, InfoIcon } from '@chakra-ui/icons';

const XenSlots = () => {
  // 使用Chakra UI的颜色模式
  const bgColor = useColorModeValue('white', 'gray.800');
  const secondaryBgColor = useColorModeValue('gray.50', 'gray.700');
  const borderColor = useColorModeValue('gray.200', 'gray.700');
  const textColor = useColorModeValue('gray.600', 'gray.400');
  const accentColor = 'hsl(142, 76%, 36%)';
  const accentColorLight = 'rgba(24, 201, 100, 0.1)';
  
  // 状态管理
  const [activeTab, setActiveTab] = useState(0);
  const [activeSlots, setActiveSlots] = useState([]);
  const [completedSlots, setCompletedSlots] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  
  // 加载Xen记录
  useEffect(() => {
    // 监听WebSocket消息
    const subscription = websocketResponseHandlerService.shared.subscribe(message => {
      if (message.event === 'getxensbychainname' || message.event === 'getxensbystatusandchain') {
        // 处理获取Xen记录的响应
        if (message.data && Array.isArray(message.data)) {
          // 分类处理记录
          const active = message.data.filter(slot => ['pending', 'active', 'processing'].includes(slot.status));
          const completed = message.data.filter(slot => ['completed', 'claimed'].includes(slot.status));
          
          // 格式化数据
          const formattedActive = active.map(slot => ({
            id: slot.slot,
            createdAt: new Date(slot.executionTime).toLocaleDateString('zh-CN'),
            lockDays: slot.days,
            endDate: new Date(new Date(slot.executionTime).getTime() + slot.days * 24 * 60 * 60 * 1000).toLocaleDateString('zh-CN'),
            status: slot.status,
            progress: calculateProgress(slot.executionTime, slot.days),
            estimatedReward: `${slot.expectedReward || 0} XEN`
          }));
          
          const formattedCompleted = completed.map(slot => ({
            id: slot.slot,
            createdAt: new Date(slot.executionTime).toLocaleDateString('zh-CN'),
            lockDays: slot.days,
            endDate: new Date(slot.claimTime || new Date()).toLocaleDateString('zh-CN'),
            status: slot.status,
            reward: `${slot.expectedReward || 0} XEN`
          }));
          
          setActiveSlots(formattedActive);
          setCompletedSlots(formattedCompleted);
        }
        setIsLoading(false);
      } else if (message.event === 'updatexen') {
        // 处理更新Xen记录的响应
        if (!message.error) {
          // 刷新记录列表
          xenService.getXensByChainName('base');
        }
      }
    });
    
    // 初始加载Xen记录
    if (websocketResponseHandlerService.isConnected) {
      xenService.getXensByChainName('base');
    } else {
      // 尝试使用REST API获取数据
      fetchXenRecords();
    }
    
    return () => subscription.unsubscribe();
  }, []);
  
  // 使用REST API获取Xen记录
  const fetchXenRecords = async () => {
    try {
      const records = await xenService.fetchXensByChainName('base');
      // 分类处理记录
      const active = records.filter(slot => ['pending', 'active', 'processing'].includes(slot.status));
      const completed = records.filter(slot => ['completed', 'claimed'].includes(slot.status));
      
      // 格式化数据
      const formattedActive = active.map(slot => ({
        id: slot.slot,
        createdAt: new Date(slot.executionTime).toLocaleDateString('zh-CN'),
        lockDays: slot.days,
        endDate: new Date(new Date(slot.executionTime).getTime() + slot.days * 24 * 60 * 60 * 1000).toLocaleDateString('zh-CN'),
        status: slot.status,
        progress: calculateProgress(slot.executionTime, slot.days),
        estimatedReward: `${slot.expectedReward || 0} XEN`
      }));
      
      const formattedCompleted = completed.map(slot => ({
        id: slot.slot,
        createdAt: new Date(slot.executionTime).toLocaleDateString('zh-CN'),
        lockDays: slot.days,
        endDate: new Date(slot.claimTime || new Date()).toLocaleDateString('zh-CN'),
        status: slot.status,
        reward: `${slot.expectedReward || 0} XEN`
      }));
      
      setActiveSlots(formattedActive);
      setCompletedSlots(formattedCompleted);
    } catch (error) {
      console.error('获取Xen记录失败:', error);
      setError('获取Xen记录失败，请稍后重试');
    } finally {
      setIsLoading(false);
    }
  };
  
  // 计算进度百分比
  const calculateProgress = (startTimeStr, days) => {
    if (!startTimeStr) return 0;
    
    const startTime = new Date(startTimeStr).getTime();
    const endTime = startTime + (days * 24 * 60 * 60 * 1000);
    const currentTime = Date.now();
    
    if (currentTime >= endTime) return 100;
    if (currentTime <= startTime) return 0;
    
    return Math.floor(((currentTime - startTime) / (endTime - startTime)) * 100);
  };
  
  // 处理领取奖励
  const handleClaimReward = (slotId) => {
    try {
      // 更新Xen记录状态为已领取
      xenService.updateXen({
        slot: slotId,
        chainName: 'base',
        status: 'claimed'
      });
    } catch (error) {
      console.error('领取奖励失败:', error);
    }
  };
  
  // 处理复投
  const handleReinvest = (slotId, days) => {
    try {
      // 创建新的Xen记录
      xenService.addXen({
        slot: Date.now(), // 使用时间戳作为临时槽位ID
        chainName: 'base',
        count: 1,
        days: days,
        ranking: 0,
        amp: 0,
        eaa: 0,
        status: 'pending',
        executionTime: new Date()
      });
      
      // 更新原记录状态
      xenService.updateXen({
        slot: slotId,
        chainName: 'base',
        status: 'reinvested'
      });
    } catch (error) {
      console.error('复投失败:', error);
    }
  };
  
  // 响应式调整
  const isMobile = useBreakpointValue({ base: true, md: false });
  
  // 渲染活跃槽位表格
  const renderActiveSlots = () => {
    return (
      <Box overflowX="auto">
        <Table variant="simple">
          <Thead>
            <Tr>
              <Th>槽位ID</Th>
              <Th>创建日期</Th>
              <Th>锁定天数</Th>
              <Th>到期日期</Th>
              <Th>进度</Th>
              <Th>预计奖励</Th>
              <Th>操作</Th>
            </Tr>
          </Thead>
          <Tbody>
            {activeSlots.map((slot) => (
              <Tr key={slot.id}>
                <Td fontWeight="medium">#{slot.id}</Td>
                <Td>{slot.createdAt}</Td>
                <Td>{slot.lockDays}天</Td>
                <Td>{slot.endDate}</Td>
                <Td>
                  <Box width="150px">
                    <Progress 
                      value={slot.progress} 
                      size="sm" 
                      colorScheme="green" 
                      borderRadius="full"
                    />
                    <Text fontSize="xs" mt={1}>{slot.progress}%</Text>
                  </Box>
                </Td>
                <Td>{slot.estimatedReward}</Td>
                <Td>
                  <Menu>
                    <MenuButton
                      as={Button}
                      rightIcon={<ChevronDownIcon />}
                      size="sm"
                      colorScheme="green"
                      variant="outline"
                    >
                      操作
                    </MenuButton>
                    <MenuList>
                      <MenuItem icon={<InfoIcon />}>查看详情</MenuItem>
                      <MenuItem icon={<RepeatIcon />} isDisabled>提前结束</MenuItem>
                    </MenuList>
                  </Menu>
                </Td>
              </Tr>
            ))}
          </Tbody>
        </Table>
      </Box>
    );
  };
  
  // 渲染已完成槽位表格
  const renderCompletedSlots = () => {
    return (
      <Box overflowX="auto">
        <Table variant="simple">
          <Thead>
            <Tr>
              <Th>槽位ID</Th>
              <Th>创建日期</Th>
              <Th>锁定天数</Th>
              <Th>完成日期</Th>
              <Th>奖励</Th>
              <Th>状态</Th>
              <Th>操作</Th>
            </Tr>
          </Thead>
          <Tbody>
            {completedSlots.map((slot) => (
              <Tr key={slot.id}>
                <Td fontWeight="medium">#{slot.id}</Td>
                <Td>{slot.createdAt}</Td>
                <Td>{slot.lockDays}天</Td>
                <Td>{slot.endDate}</Td>
                <Td fontWeight="bold" color={accentColor}>{slot.reward}</Td>
                <Td>
                  <Badge colorScheme="green" variant="subtle" px={2} py={1} borderRadius="full">
                    可领取
                  </Badge>
                </Td>
                <Td>
                  <HStack spacing={2}>
                    <Button
                      leftIcon={<CheckIcon />}
                      colorScheme="green"
                      size="sm"
                      bg={accentColor}
                      _hover={{ bg: `${accentColor}/90` }}
                    >
                      领取
                    </Button>
                    <Button
                      leftIcon={<RepeatIcon />}
                      colorScheme="green"
                      size="sm"
                      variant="outline"
                    >
                      复投
                    </Button>
                  </HStack>
                </Td>
              </Tr>
            ))}
          </Tbody>
        </Table>
      </Box>
    );
  };
  
  // 渲染移动端卡片视图
  const renderMobileCards = (slots, isActive) => {
    return (
      <VStack spacing={4} align="stretch">
        {slots.map((slot) => (
          <Card key={slot.id} bg={bgColor} boxShadow="sm" borderWidth="1px" borderColor={borderColor}>
            <CardHeader pb={0}>
              <Flex justify="space-between" align="center">
                <Heading size="md">槽位 #{slot.id}</Heading>
                {isActive ? (
                  <Badge colorScheme="blue" variant="subtle" px={2} py={1} borderRadius="full">
                    进行中
                  </Badge>
                ) : (
                  <Badge colorScheme="green" variant="subtle" px={2} py={1} borderRadius="full">
                    可领取
                  </Badge>
                )}
              </Flex>
            </CardHeader>
            <CardBody>
              <VStack spacing={3} align="stretch">
                <Flex justify="space-between">
                  <Text color={textColor}>创建日期:</Text>
                  <Text fontWeight="medium">{slot.createdAt}</Text>
                </Flex>
                <Flex justify="space-between">
                  <Text color={textColor}>锁定天数:</Text>
                  <Text fontWeight="medium">{slot.lockDays}天</Text>
                </Flex>
                <Flex justify="space-between">
                  <Text color={textColor}>{isActive ? '到期日期:' : '完成日期:'}</Text>
                  <Text fontWeight="medium">{slot.endDate}</Text>
                </Flex>
                
                {isActive ? (
                  <>
                    <Flex justify="space-between">
                      <Text color={textColor}>进度:</Text>
                      <Text fontWeight="medium">{slot.progress}%</Text>
                    </Flex>
                    <Progress 
                      value={slot.progress} 
                      size="sm" 
                      colorScheme="green" 
                      borderRadius="full"
                    />
                    <Flex justify="space-between">
                      <Text color={textColor}>预计奖励:</Text>
                      <Text fontWeight="bold" color={accentColor}>{slot.estimatedReward}</Text>
                    </Flex>
                  </>
                ) : (
                  <Flex justify="space-between">
                    <Text color={textColor}>奖励:</Text>
                    <Text fontWeight="bold" color={accentColor}>{slot.reward}</Text>
                  </Flex>
                )}
              </VStack>
            </CardBody>
            <CardFooter pt={0}>
              {isActive ? (
                <Button
                  leftIcon={<InfoIcon />}
                  colorScheme="green"
                  variant="outline"
                  size="sm"
                  width="full"
                >
                  查看详情
                </Button>
              ) : (
                <SimpleGrid columns={2} spacing={2} width="full">
                  <Button
                    leftIcon={<CheckIcon />}
                    colorScheme="green"
                    size="sm"
                    bg={accentColor}
                    _hover={{ bg: `${accentColor}/90` }}
                  >
                    领取
                  </Button>
                  <Button
                    leftIcon={<RepeatIcon />}
                    colorScheme="green"
                    size="sm"
                    variant="outline"
                  >
                    复投
                  </Button>
                </SimpleGrid>
              )}
            </CardFooter>
          </Card>
        ))}
      </VStack>
    );
  };
  
  return (
    <Container maxW="1008px" py={8}>
      <VStack spacing={8} align="stretch">
        {/* 页面标题和摘要 */}
        <Box
          bg={bgColor}
          p={6}
          borderRadius="xl"
          boxShadow="sm"
          borderWidth="1px"
          borderColor={borderColor}
        >
          <VStack spacing={4} align="start">
            <Heading as="h1" size="xl">我的XEN槽位</Heading>
            <Text fontSize="lg" color={textColor}>
              管理您的XEN槽位，查看进度，领取奖励或复投以获得更多收益。
            </Text>
            <Button
              as={RouterLink}
              to="/"
              colorScheme="green"
              bg={accentColor}
              _hover={{ bg: `${accentColor}/90`, transform: 'translateY(-2px)' }}
              boxShadow="md"
            >
              创建新槽位
            </Button>
          </VStack>
        </Box>
        
        {/* 统计数据 */}
        <SimpleGrid columns={{ base: 1, md: 3 }} spacing={4}>
          <Card bg={bgColor} boxShadow="sm" borderWidth="1px" borderColor={borderColor}>
            <CardBody>
              <Stat>
                <StatLabel>活跃槽位</StatLabel>
                <StatNumber>{activeSlots.length}</StatNumber>
                <StatHelpText>总计锁定180天</StatHelpText>
              </Stat>
            </CardBody>
          </Card>
          
          <Card bg={bgColor} boxShadow="sm" borderWidth="1px" borderColor={borderColor}>
            <CardBody>
              <Stat>
                <StatLabel>待领取奖励</StatLabel>
                <StatNumber>665 XEN</StatNumber>
                <StatHelpText>来自2个已完成槽位</StatHelpText>
              </Stat>
            </CardBody>
          </Card>
          
          <Card bg={bgColor} boxShadow="sm" borderWidth="1px" borderColor={borderColor}>
            <CardBody>
              <Stat>
                <StatLabel>预计未来奖励</StatLabel>
                <StatNumber>1,780 XEN</StatNumber>
                <StatHelpText>来自3个活跃槽位</StatHelpText>
              </Stat>
            </CardBody>
          </Card>
        </SimpleGrid>
        
        {/* 槽位列表 */}
        <Box
          bg={bgColor}
          p={6}
          borderRadius="xl"
          boxShadow="sm"
          borderWidth="1px"
          borderColor={borderColor}
        >
          <Tabs colorScheme="green" index={activeTab} onChange={(index) => setActiveTab(index)}>
            <TabList>
              <Tab _selected={{ color: accentColor, borderColor: accentColor }}>活跃槽位 ({activeSlots.length})</Tab>
              <Tab _selected={{ color: accentColor, borderColor: accentColor }}>已完成槽位 ({completedSlots.length})</Tab>
            </TabList>
            
            <TabPanels>
              <TabPanel px={0}>
                {isMobile ? renderMobileCards(activeSlots, true) : renderActiveSlots()}
              </TabPanel>
              <TabPanel px={0}>
                {isMobile ? renderMobileCards(completedSlots, false) : renderCompletedSlots()}
              </TabPanel>
            </TabPanels>
          </Tabs>
        </Box>
      </VStack>
    </Container>
  );
};

export default XenSlots;