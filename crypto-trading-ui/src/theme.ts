import { extendTheme } from '@chakra-ui/react';

// 扩展Chakra UI默认主题
const theme = extendTheme({
  config: {
    initialColorMode: 'dark',
    useSystemColorMode: false,
  },
  colors: {
    brand: {
      50: '#e6f7ff',
      100: '#b3e0ff',
      200: '#80caff',
      300: '#4db3ff',
      400: '#1a9dff',
      500: '#0080ff', // 主色调
      600: '#0066cc',
      700: '#004d99',
      800: '#003366',
      900: '#001a33',
    },
    trading: {
      up: '#00c853', // 上涨颜色
      down: '#ff3d00', // 下跌颜色
      neutral: '#9e9e9e', // 中性颜色
    },
    // 暗色主题背景色
    background: {
      dark: '#121212', // 主背景色
      card: '#1E1E1E', // 卡片背景色
      hover: '#2A2A2A', // 悬停背景色
      header: '#1A1A1A', // 头部背景色
      gray: '#1A1A1A', // 灰色背景
    },
    // 边框颜色
    border: {
      dark: 'rgba(255, 255, 255, 0.1)', // 暗色主题边框
      hover: 'rgba(255, 255, 255, 0.2)', // 悬停边框
    },
    // 强调色
    accent: {
      primary: '#3A74FF', // 主要强调色
      primaryHover: '#2A64EF', // 主要强调色悬停
      secondary: '#2A2A2A', // 次要强调色
    },
    // 主色调
    primary: {
      500: '#3A74FF',
      hover: '#2A64EF',
    },
    // 代币颜色
    btc: { color: '#F7931A' },
    eth: { color: '#627EEA' },
    bnb: { color: '#F3BA2F' },
    sol: { color: '#00FFA3' },
    ada: { color: '#0033AD' },
    xrp: { color: '#23292F' },
    // 图表颜色
    chartreuse: '#00C853',
  },
  fonts: {
    body: "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif",
    heading: "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif",
    mono: "'SFMono-Regular', Menlo, Monaco, Consolas, 'Liberation Mono', monospace",
  },
  styles: {
    global: (props: { colorMode: string }) => ({
      body: {
        bg: props.colorMode === 'dark' ? 'background.dark' : 'gray.50',
        color: props.colorMode === 'dark' ? 'whiteAlpha.900' : 'gray.800',
      },
    }),
  },
  components: {
    Button: {
      baseStyle: {
        fontWeight: 'medium',
        borderRadius: 'md',
      },
      variants: {
        solid: (props: { colorMode: string }) => ({
          bg: props.colorMode === 'dark' ? 'brand.500' : 'brand.500',
          color: 'white',
          _hover: {
            bg: props.colorMode === 'dark' ? 'brand.600' : 'brand.400',
          },
        }),
      },
    },
    Table: {
      variants: {
        simple: (props: { colorMode: string }) => ({
          th: {
            borderColor: props.colorMode === 'dark' ? 'border.dark' : 'gray.200',
            fontWeight: 'medium',
            color: props.colorMode === 'dark' ? 'gray.400' : 'gray.600',
            fontSize: 'sm',
          },
          td: {
            borderColor: props.colorMode === 'dark' ? 'border.dark' : 'gray.100',
            fontSize: 'sm',
          },
          tbody: {
            tr: {
              _hover: {
                bg: props.colorMode === 'dark' ? 'background.hover' : 'gray.50',
              },
            },
          },
        }),
      },
    },
    Card: {
      baseStyle: (props: { colorMode: string }) => ({
        container: {
          bg: props.colorMode === 'dark' ? 'background.card' : 'white',
          borderRadius: 'md',
          boxShadow: props.colorMode === 'dark' ? 'none' : 'sm',
          borderWidth: '1px',
          borderColor: props.colorMode === 'dark' ? 'border.dark' : 'gray.200',
        },
      }),
    },
    Tabs: {
      variants: {
        line: (props: { colorMode: string }) => ({
          tablist: {
            borderColor: props.colorMode === 'dark' ? 'border.dark' : 'gray.200',
          },
          tab: {
            color: props.colorMode === 'dark' ? 'gray.400' : 'gray.600',
            _selected: {
              color: props.colorMode === 'dark' ? 'white' : 'brand.500',
              borderColor: props.colorMode === 'dark' ? 'brand.500' : 'brand.500',
            },
          },
        }),
      },
    },
  },
});

export default theme;