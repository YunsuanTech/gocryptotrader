import React from 'react';
import { ChakraProvider, Box, useColorMode } from '@chakra-ui/react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';

import theme from './theme';
import Navbar from './components/Navbar';
import HomePage from './pages/HomePage';
import TradingPage from './pages/TradingPage';

// 应用容器组件，强制使用暗色模式
const AppContainer = ({ children }: { children: React.ReactNode }) => {
  const { setColorMode } = useColorMode();
  
  // 确保应用始终使用暗色模式
  React.useEffect(() => {
    setColorMode('dark');
  }, [setColorMode]);

  return (
    <Box minH="100vh" w="100%" bg="background.dark">
      {children}
    </Box>
  );
};

function App() {
  return (
    <ChakraProvider theme={theme}>
      <Router>
        <AppContainer>
          <Navbar />
          <Box as="main" pt="60px" w="100%" h="calc(100vh - 60px)" overflow="hidden">
            <Routes>
              <Route path="/" element={<HomePage />} />
              <Route path="/home" element={<HomePage />} />
              <Route path="/trading" element={<TradingPage />} />
              {/* 其他路由可以在这里添加 */}
            </Routes>
          </Box>
        </AppContainer>
      </Router>
    </ChakraProvider>
  )
}

export default App
 