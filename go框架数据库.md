## 转发记录
-- 创建主表（包含通用字段和代币专用字段）
CREATE TABLE solana_transfers (
    signature TEXT PRIMARY KEY,  -- 签名作为主键，确保唯一性
    network TEXT NOT NULL,      -- 网络类型 
    send_time DATETIME,         -- 发送时间
    sender TEXT,                -- 发送者地址
    receiver TEXT,              -- 接收者地址
    -- 代币相关字段
    is_token_transfer BOOLEAN NOT NULL CHECK (is_token_transfer IN (0, 1)), -- 是否代币转账
    amount_display REAL NOT NULL,       ---金额
    token_mint_address TEXT CHECK (         -- 代币合约地址 (代币转账时必填)
        (is_token_transfer = 1 AND token_mint_address IS NOT NULL) OR
        (is_token_transfer = 0 AND token_mint_address IS NULL)
    )
);
-- 创建索引
CREATE INDEX idx_sender ON solana_transfers (sender);
CREATE INDEX idx_receiver ON solana_transfers (receiver);
CREATE INDEX IF NOT EXISTS idx_token_mint ON solana_transfers (token_mint_address);

## 交易记录
CREATE TABLE IF NOT EXISTS transaction_records (
    transaction_id INTEGER PRIMARY KEY AUTOINCREMENT,
    token_address TEXT NOT NULL,
    rule_id INTEGER,
    type TEXT NOT NULL CHECK (type IN ('buy', 'sell')),
    amount REAL NOT NULL CHECK (amount >= 0),
    price REAL NOT NULL CHECK (price >= 0),
    timestamp INTEGER NOT NULL,
    tx_hash TEXT UNIQUE,  -- 区块链交易哈希（可选字段）
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'failed')),  -- 交易状态（可选）
    FOREIGN KEY (token_address) REFERENCES token_monitor(token_address),
    FOREIGN KEY (rule_id) REFERENCES trading_rules(id)
);

transaction_id
自增主键，唯一标识每笔交易

token_address
外键关联 token_monitor 表

记录交易对应的代币地址

rule_id
外键关联 trading_rules 表（可为NULL）
记录触发交易的规则ID（允许手动交易）

type
交易类型：buy/sell
CHECK约束保证类型有效性

amount
交易数量（基于代币精度）
非负校验保证数据有效性

price
交易时单价（计价单位需与系统一致）
非负校验保证数据有效性

timestamp
交易时间（Unix时间戳）
与现有表保持时间格式统一

tx_hash
区块链交易哈希（如需对接链上数据）
UNIQUE约束防止重复记录

status
交易状态跟踪（待处理/已确认/失败）
帮助监控交易生命周期

### 监控代币表
CREATE TABLE IF NOT EXISTS token_monitor (
    token_address TEXT PRIMARY KEY NOT NULL,
    token_name TEXT NOT NULL,
    price REAL,
    token_decimals INTEGER NOT NULL CHECK (token_decimals >= 0),
    amount REAL, --现持有数量
    buy_amount REAL,
    buy_price REAL,
    buy_time INTEGER,  -- Unix 时间戳
    sell_percentage REAL,
    total_sell_price REAL,
    increase REAL, --涨幅
    last_sell_time INTEGER,  -- Unix 时间戳
    is_monitoring INTEGER NOT NULL DEFAULT 1  -- 0: false, 1: true
);

token_address (TEXT)  
代币合约地址，作为唯一标识符（Solana 地址通常是 44 个字符的 Base58 编码字符串）。
设置为 PRIMARY KEY NOT NULL，确保每条记录的唯一性且不能为空。

token_name (TEXT)  
代币名称，用于标识代币的名称。
设置为 NOT NULL，假设名称是必须提供的。

price (REAL)  
当前代币价格

token_decimals (INTEGER)  
代币精度，表示代币的小数位数。
设置为 NOT NULL，并添加约束 CHECK (token_decimals >= 0)，确保精度为非负整数。

buy_amount (REAL)  
买入数量，表示购买的代币数量。
使用 REAL 类型存储浮点数，允许为空（如果没有买入记录）。

buy_price (REAL)  
买入价格，表示购买时的单价。
使用 REAL 类型，允许为空（如果没有买入记录）。

buy_time (INTEGER)  
买入时间，存储为 Unix 时间戳（秒数）。
使用 INTEGER 类型，便于时间比较和计算，允许为空（如果没有买入记录）。

sell_percentage (REAL)  
卖出数量百分比，例如 25.5 表示卖出 25.5%。
使用 REAL 类型，允许为空（如果没有卖出记录）。

total_sell_price (REAL)  
总卖出价格，表示所有卖出交易的总金额。
使用 REAL 类型，允许为空（如果没有卖出记录）。

last_sell_time (INTEGER)  
最后一次卖出时间，存储为 Unix 时间戳。
使用 INTEGER 类型，允许为空（如果没有卖出记录）。

is_monitoring (INTEGER)  
是否监听，布尔值（0 表示 false，1 表示 true）。
设置为 NOT NULL DEFAULT 1，默认值为 1，表示正在监听。


### 买卖规则表
CREATE TABLE trading_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,  -- 唯一标识
    rule_name TEXT NOT NULL,              -- 规则名称
    rule_type TEXT NOT NULL,              -- 规则类型（'buy' 或 'sell'）
    buy_price REAL,                       -- 买入价格阈值（可为空）
    sell_price REAL,                      -- 卖出价格阈值（可为空）
    condition TEXT NOT NULL,              -- 规则条件
    priority REAL DEFAULT 0,              -- 规则优先级或权重
    description TEXT                      -- 规则描述
);

id (INTEGER)  
主键，自增，用于唯一标识每条规则。

rule_name (TEXT)  
规则名称，必须填写，例如“低价买入”或“高价卖出”。

rule_type (TEXT)  
规则类型，必须填写，取值为 'buy'（买入）或 'sell'（卖出），用于区分买卖规则。

buy_price (REAL)  
买入价格阈值，例如 100.0，表示当价格低于此值时触发买入。

可为空：如果规则不依赖买入价格，可以不填写。

sell_price (REAL)  
卖出价格阈值，例如 200.0，表示当价格高于此值时触发卖出。

可为空：如果规则不依赖卖出价格，可以不填写。

condition (TEXT)  
规则条件，必须填写，例如 "price < buy_price" 或 "price > sell_price"，定义触发规则的逻辑。

priority (REAL)  
规则优先级，默认值为 0，用于排序或决定规则执行顺序。

description (TEXT)  
规则描述，可选，用于记录规则的详细说明

