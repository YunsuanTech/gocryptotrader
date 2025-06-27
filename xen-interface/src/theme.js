import { extendTheme } from '@chakra-ui/react'

// 自定义颜色 - 使用pump.fun的浅绿色配色
const colors = {
  brand: {
    50: '#e6f9ef',
    100: '#d0f4e0',
    200: '#a2e9c1',
    300: '#74dea2',
    400: '#46d383',
    500: '#18c964', // 主色调 - pump.fun绿色
    600: '#13a954',
    700: '#0e8944',
    800: '#096834',
    900: '#054824',
  },
  // 背景色
  background: {
    light: '#f8f9fd', // 亮色模式背景色
    dark: '#1a1a2e', // 暗色模式背景色
    card: {
      light: 'transparent', // 透明容器
      dark: 'transparent'
    },
    hover: {
      light: 'rgba(24, 201, 100, 0.08)', // 半透明浅绿色悬停效果
      dark: 'rgba(24, 201, 100, 0.15)'
    }
  },
  // 渐变色定义
  gradients: {
    '135deg': {
      blue: 'linear-gradient(135deg, #0073ff 0%, #00c8ff 100%)',
      green: 'linear-gradient(135deg, #00c076 0%, #00e676 100%)',
      red: 'linear-gradient(135deg, #ff5c5c 0%, #ff8a8a 100%)'
    },
    '90deg': {
      blue: 'linear-gradient(90deg, #0073ff 0%, #00c8ff 100%)',
      green: 'linear-gradient(90deg, #00c076 0%, #00e676 100%)',
      red: 'linear-gradient(90deg, #ff5c5c 0%, #ff8a8a 100%)'
    }
  },
}

// 自定义字体
const fonts = {
  body: "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif",
  heading: "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif",
}

// 自定义组件样式
const components = {
  Button: {
    baseStyle: {
      fontWeight: 500,
      borderRadius: 'md',
    },
    variants: {
      solid: (props) => ({
        bg: props.colorScheme === 'brand' ? 'brand.500' : undefined,
        _hover: {
          bg: props.colorScheme === 'brand' ? 'brand.600' : undefined,
        },
      }),
      outline: (props) => ({
        borderColor: props.colorScheme === 'brand' ? 'brand.500' : undefined,
        color: props.colorScheme === 'brand' ? 'brand.500' : undefined,
        _hover: {
          bg: 'transparent',
          borderColor: props.colorScheme === 'brand' ? 'brand.600' : undefined,
          color: props.colorScheme === 'brand' ? 'brand.600' : undefined,
        },
      }),
      ghost: (props) => ({
        _hover: {
          bg: props.colorScheme === 'brand' ? 'rgba(24, 201, 100, 0.08)' : undefined,
        },
      }),
    },
  },
  Table: {
    variants: {
      simple: {
        th: {
          borderColor: 'rgba(0, 0, 0, 0.04)',
          _dark: {
            borderColor: 'rgba(255, 255, 255, 0.08)',
          },
          fontWeight: 600,
          textTransform: 'none',
          letterSpacing: 'normal',
        },
        td: {
          borderColor: 'rgba(0, 0, 0, 0.04)',
          _dark: {
            borderColor: 'rgba(255, 255, 255, 0.08)',
          },
        },
      },
    },
  },
  Card: {
    baseStyle: {
      container: {
        borderRadius: 'xl',
        boxShadow: 'none',
        overflow: 'hidden',
        bg: 'transparent',
        border: '1px solid',
        borderColor: 'rgba(0, 0, 0, 0.04)',
        _dark: {
          borderColor: 'rgba(255, 255, 255, 0.08)',
        },
      },
    },
  },
  Tabs: {
    variants: {
      enclosed: {
        tab: {
          fontWeight: 500,
          _selected: {
            color: 'brand.500',
            borderColor: 'rgba(0, 0, 0, 0.04)',
            borderBottomColor: 'transparent',
            _dark: {
              color: 'brand.300',
              borderColor: 'rgba(255, 255, 255, 0.08)',
              borderBottomColor: 'transparent',
            },
          },
        },
        tablist: {
          borderColor: 'rgba(0, 0, 0, 0.04)',
          _dark: {
            borderColor: 'rgba(255, 255, 255, 0.08)',
          },
        },
      },
      'soft-rounded': {
        tab: {
          fontWeight: 500,
          _selected: {
            color: 'white',
            bg: 'brand.500',
            _dark: {
              bg: 'brand.500',
            },
          },
        },
      },
    },
  },
}

// 全局样式
const styles = {
  global: (props) => ({
    body: {
      bg: props.colorMode === 'dark' ? 'background.dark' : 'background.light',
      color: props.colorMode === 'dark' ? 'white' : 'gray.800',
    },
    // 添加全局容器样式
    '.container-card': {
      bg: 'transparent',
      backdropFilter: 'blur(10px)',
      borderRadius: 'xl',
      boxShadow: 'none',
      border: '1px solid',
      borderColor: props.colorMode === 'dark' ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.04)',
    },
  }),
}

// 其他主题配置
const config = {
  initialColorMode: 'light',
  useSystemColorMode: false,
  cssVarPrefix: 'xen',
  disableTransitionOnChange: false, // 启用颜色模式切换时的过渡效果
}

// 创建自定义主题
const theme = extendTheme({
  colors,
  fonts,
  components,
  styles,
  config,
})

export default theme