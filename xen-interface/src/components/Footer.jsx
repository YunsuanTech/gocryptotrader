import React from 'react'
import {
  Box,
  Container,
  Stack,
  SimpleGrid,
  Text,
  useColorModeValue,
  Divider,
  Flex,
} from '@chakra-ui/react'

const ListHeader = ({ children }) => {
  return (
    <Text fontWeight={'500'} fontSize={'lg'} mb={2}>
      {children}
    </Text>
  )
}

const Footer = () => {
  const textColor = useColorModeValue('gray.700', 'gray.200');
  const borderColor = useColorModeValue('rgba(0, 0, 0, 0.04)', 'rgba(255, 255, 255, 0.08)');
  
  return (
    <Box
      bg="transparent"
      color={textColor}
      pt={4}
      pb={6}
    >
      <Container maxW={'1008px'} px={0}>
        <Divider borderColor={borderColor} mb={6} />
        <Flex
          direction={{ base: 'column', md: 'row' }}
          justify="space-between"
          align={{ base: 'center', md: 'flex-end' }}
        >
          <Text 
            fontSize={{ base: 'xs', md: 'sm' }}
            fontWeight="medium"
            letterSpacing="wide"
            textAlign={{ base: 'center', md: 'left' }}
            opacity={0.8}
          >
            © {new Date().getFullYear()} XEN Trade. All rights reserved
          </Text>
        </Flex>
      </Container>
    </Box>
  )
}

export default Footer