import React from 'react'
import { Routes, Route } from 'react-router-dom'
import { Box, useColorModeValue } from '@chakra-ui/react'

// 导入组件
import Navbar from './components/Navbar'
import Footer from './components/Footer'

// 导入页面
import XenHome from './pages/XenHome'
import XenSlots from './pages/XenSlots'
import TokensPage from './pages/TokensPage'

function App() {
  return (
    <Box 
      minH="100vh" 
      display="flex" 
      flexDirection="column"
      backgroundAttachment="fixed"
      bg={useColorModeValue('background.light', 'gray.900')}
      color={useColorModeValue('gray.800', 'white')}
      transition="background 0.3s ease, color 0.3s ease"
    >
      <Navbar />
      <Box 
        flex="1" 
        width="100%" 
        maxW="1008px" 
        mx="auto" 
        px={0}
      >
        <Routes>
          <Route path="/" element={<XenHome />} />
          <Route path="/xen-slots" element={<XenSlots />} />
          <Route path="/tokens" element={<TokensPage />} />
        </Routes>
      </Box>
      <Footer />
    </Box>
  )

}

export default App