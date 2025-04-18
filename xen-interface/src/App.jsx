import React from 'react'
import { Routes, Route } from 'react-router-dom'
import { Box, useColorModeValue } from '@chakra-ui/react'

// 导入组件
import Navbar from './components/Navbar'
import Footer from './components/Footer'

// 导入页面
import XenHome from './pages/XenHome'
import XenSlots from './pages/XenSlots'

function App() {
  const bgColor = useColorModeValue('linear-gradient(to bottom, #f8f9fd 0%, #e6f7ff 100%)', 'linear-gradient(to bottom, #1a1a2e 0%, #0d0d1a 100%)')
  
  return (
    <Box 
      minH="100vh" 
      display="flex" 
      flexDirection="column"
      bg={bgColor}
      backgroundAttachment="fixed"
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
        </Routes>
      </Box>
      <Footer />
    </Box>
  )

}

export default App