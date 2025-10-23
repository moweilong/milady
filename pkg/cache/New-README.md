# onex Cache 模块

## 目录

- [概述](#概述)
- [架构设计](#架构设计)
- [核心接口](#核心接口)
- [缓存类型详解](#缓存类型详解)
- [存储层实现](#存储层实现)
- [键生成与处理机制](#键生成与处理机制)
- [并发与同步](#并发与同步)
- [快速开始](#快速开始)
- [高级用法示例](#高级用法示例)
- [配置选项](#配置选项)
- [错误处理](#错误处理)
- [性能优化](#性能优化)
- [典型用例](#典型用例)
- [注意事项](#注意事项)

## 概述

`pkg/cache` 是 onex 项目中的一个功能丰富、设计灵活的 Go 缓存库，提供了多种缓存管理方式和高级特性。该模块采用泛型设计，支持链式缓存、二级缓存（L2）和可加载缓存等多种缓存模式，并提供了多种后端存储实现，如高性能内存缓存（Ristretto）和分布式缓存（Redis）。

缓存模块的主要目标是提供统一的缓存抽象，同时支持多种缓存策略的组合使用，以满足不同场景下的性能和功能需求。

## 架构设计

### 分层架构

`pkg/cache` 模块采用清晰的分层架构设计，主要包含以下几层：

1. **接口层**：定义统一的 `Cache[T]` 和 `Store` 接口，作为所有缓存实现的标准
2. **策略层**：实现各种缓存策略（链式缓存、二级缓存、可加载缓存）
3. **存储层**：提供具体的存储实现（Redis、Ristretto内存缓存）
4. **工具层**：提供键生成、错误处理等通用工具

### 设计理念

- **接口分离原则**：清晰分离缓存策略和存储实现
- **组合优于继承**：通过组合不同的缓存策略和存储实现，灵活构建缓存解决方案
- **泛型支持**：利用Go泛型提供类型安全的缓存操作
- **上下文感知**：所有操作都支持context，便于超时控制和取消操作

## 特性

- **泛型支持**：使用 Go 泛型提供类型安全的缓存操作
- **多种缓存模式**：支持链式缓存、二级缓存、可加载缓存等
- **多后端存储**：支持内存缓存（Ristretto）、分布式缓存（Redis）等
- **灵活的键处理**：支持字符串键、自定义键生成和对象哈希键
- **TTL 支持**：完整支持缓存项的生存时间管理
- **异步操作**：支持异步写入和同步机制
- **上下文感知**：全面支持 context.Context 进行超时控制和取消操作

## 核心接口

### Cache[T] 接口

`Cache[T]` 是整个缓存模块的核心接口，定义了通用的缓存操作，采用泛型设计支持任意类型的值：

```go
// Cache represents the interface for all caches
type Cache[T any] interface {
	// Set 将值存入缓存
	Set(ctx context.Context, key any, obj T) error
	// Get 从缓存获取值
	Get(ctx context.Context, key any) (T, error)
	// SetWithTTL 将值存入缓存并设置过期时间
	SetWithTTL(ctx context.Context, key any, obj T, ttl time.Duration) error
	// GetWithTTL 从缓存获取值并返回剩余过期时间
	GetWithTTL(ctx context.Context, key any) (T, time.Duration, error)
	// Del 从缓存中删除值
	Del(ctx context.Context, key any) error
	// Clear 清空缓存
	Clear(ctx context.Context) error
	// Wait 等待异步操作完成
	Wait(ctx context.Context)
}

// ErrKeyNotFound 表示键不存在的错误
var ErrKeyNotFound = errors.New("key not found")

// Store 接口定义了底层存储实现的标准接口，被各种缓存策略使用
// Store 定义了缓存存储的接口
type Store interface {
	// Get 从存储中获取值
	Get(ctx context.Context, key string) ([]byte, error)
	// GetWithTTL 从存储中获取值并返回剩余过期时间
	GetWithTTL(ctx context.Context, key string) ([]byte, time.Duration, error)
	// Set 将值存入存储
	Set(ctx context.Context, key string, value []byte) error
	// SetWithTTL 将值存入存储并设置过期时间
	SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error
	// Del 从存储中删除值
	Del(ctx context.Context, key string) error
	// Clear 清空存储
	Clear(ctx context.Context) error
	// Wait 等待存储初始化完成
	Wait(ctx context.Context)
}
```

## 缓存类型

### 1. 链式缓存 (ChainCache)

链式缓存将多个缓存实例组合成一个缓存链，当从某层缓存获取数据成功后，会自动将数据同步到前面的缓存层，提高后续访问性能。

```go
// 创建链式缓存
chain := cache.NewChain(localCache, remoteCache)
```

### 2. 二级缓存 (L2Cache)

二级缓存结合了本地内存缓存（Ristretto）和远程缓存，提供高性能的缓存访问。本地缓存未命中时自动查询远程缓存。

```go
// 创建自定义配置的 L2 缓存
l2Cache := cache.NewL2(redisStore, 
	cache.L2WithNumCounters(1e6),
	cache.L2WithMaxCost(512<<20), // 512MB
	cache.L2WithMetrics(true),
)
```

### 3. 可加载缓存 (LoadableCache)

可加载缓存提供了数据自动加载的能力，当缓存未命中时，自动调用加载函数获取数据并写入缓存。

```go
// 定义加载函数
loadFunc := func(ctx context.Context, key any) (MyData, error) {
	// 从数据源加载数据
	return fetchDataFromDatabase(key), nil
}

// 创建可加载缓存
loadableCache := cache.NewLoadable(loadFunc, baseCache)
```

## 存储实现

### 1. Redis 存储

使用 Redis 作为分布式缓存后端，支持完整的缓存操作和 TTL 管理。

```go
import "github.com/onexstack/onex/pkg/cache/store/redis"

// 创建 Redis 客户端
redisClient := redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

// 创建 Redis 存储
redisStore := redis.NewRedis(redisClient)
```

### 2. Ristretto 内存存储

使用 Ristretto 库提供高性能的内存缓存实现，适用于需要快速访问的场景。

```go
import "github.com/onexstack/onex/pkg/cache/store/ristretto"

// 创建 Ristretto 客户端
ristrettoClient, _ := ristretto.NewCache(&ristretto.Config{
	NumCounters: 1e7,
	MaxCost:     1 << 30, // 1GB
	BufferItems: 64,
})

// 创建 Ristretto 存储
ristrettoStore := ristretto.NewRistretto(ristrettoClient)
```

## 键生成机制

缓存模块支持多种键类型：

1. **字符串键**：直接使用字符串作为键
2. **自定义键**：实现 `KeyGetter` 接口的对象
3. **对象哈希**：通过 MD5 哈希生成唯一键

## 快速开始

### 基本用法

```go
import (
	"context"
	"github.com/onexstack/onex/pkg/cache"
)

func main() {
	// 创建缓存实例（这里以 L2 缓存为例）
	myCache := cache.NewL2(redisStore)
	ctx := context.Background()

	// 设置缓存
	err := myCache.Set(ctx, "my-key", myValue)
	if err != nil {
		// 处理错误
	}

	// 获取缓存
	value, err := myCache.Get(ctx, "my-key")
	if err != nil {
		// 处理缓存未命中
	}

	// 设置带 TTL 的缓存
	err = myCache.SetWithTTL(ctx, "my-key-with-ttl", myValue, 30*time.Second)

	// 删除缓存
	err = myCache.Del(ctx, "my-key")
}
```

### 高级用法：组合多种缓存

```go
// 创建 Redis 存储
redisStore := redis.NewRedis(redisClient)

// 创建 L2 缓存（本地 + 远程）
l2Cache := cache.NewL2(redisStore)

// 创建可加载缓存
loadableCache := cache.NewLoadable(loadFunc, l2Cache)

// 此时 loadableCache 具备了自动加载、本地缓存和远程缓存的所有特性
```

## 配置选项

### L2 缓存配置

- `L2WithDisableCache`：启用或禁用本地缓存
- `L2WithNumCounters`：设置计数器数量（影响命中率统计精度）
- `L2WithMaxCost`：设置最大缓存容量
- `L2WithMetrics`：启用或禁用指标收集

## 错误处理

缓存模块定义了标准错误（如 `ErrKeyNotFound`），用于处理常见的缓存操作失败场景。链式缓存会收集所有缓存层的错误并返回综合信息。

## 性能优化

- 使用 L2 缓存减少对远程缓存的访问
- 为频繁访问的数据设置合理的 TTL
- 使用链式缓存组合多种缓存策略
- 利用可加载缓存实现数据预加载和自动刷新

## 注意事项

- 确保适当处理缓存未命中的情况
- 注意设置合理的 TTL 以平衡数据新鲜度和性能
- 在高并发场景下，考虑缓存穿透、缓存击穿和缓存雪崩问题
- 对于可加载缓存，确保加载函数是线程安全的