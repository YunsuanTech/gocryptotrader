import { Box, Flex, Text, Button, HStack, useColorModeValue, Icon } from '@chakra-ui/react';
import { Link as RouterLink } from 'react-router-dom';
import { FaBitcoin } from 'react-icons/fa';

const Navbar = () => {
  
  return (

    <Box
      as="nav"
      position="fixed"
      top={0}
      left={0}
      right={0}
      zIndex={10}
      bg="background.header"
      borderBottom="1px solid"
      borderColor="border.dark"
      h="60px"
    >
      
      <Flex
        h="100%"
        px={4}
        alignItems="center"
        justifyContent="space-between"
        maxW="1600px"
        mx="auto"
      >
        <Flex alignItems="center">
          <Icon as={FaBitcoin} w={8} h={8} color="brand.500" mr={2} />
          <Text fontSize="xl" fontWeight="bold" color="whiteAlpha.900">
            Crypto Trading
          </Text>
        </Flex>

        <HStack spacing={4}>
          <Button
            as={RouterLink}
            to="/"
            variant="ghost"
            colorScheme="brand"
            size="sm"
          >
            首页
          </Button>
          <Button
            as={RouterLink}
            to="/trading"
            variant="ghost"
            colorScheme="brand"
            size="sm"
          >
            交易
          </Button>
        </HStack>
      </Flex>
    </Box>
  );

};

export default Navbar;