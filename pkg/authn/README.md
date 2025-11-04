# authn 模块开发文档

## 1. 模块概述

`authn` 模块是 Milady 项目中的认证核心模块，提供了完整的身份验证功能，包括令牌生成、验证、解析和销毁，以及密码加密与验证功能。该模块采用接口抽象与实现分离的设计模式，支持灵活的令牌存储机制。

## 2. 目录结构

```
pkg/authn/
├── authn.go           # 核心接口定义和密码处理工具
└── jwt/               # JWT认证实现
    ├── jwt.go         # JWTAuth主要实现
    ├── store.go       # 令牌存储接口定义
    ├── token.go       # 令牌信息结构体
    └── store/         # 存储实现
        └── redis/     # Redis存储实现
            └── redis.go
```

## 3. 核心接口

### 3.1 IToken 接口

`IToken` 接口定义了令牌的基本行为，位于 authn.go 中：

```go
// IToken defines methods to implement a generic token.
type IToken interface {
	// Get token string.
	GetToken() string
	// Get token type.
	GetTokenType() string
	// Get token expiration timestamp.
	GetExpiresAt() int64
	// JSON encoding
	EncodeToJSON() ([]byte, error)
}
```

**功能说明**：
- `GetToken()`: 获取令牌字符串
- `GetTokenType()`: 获取令牌类型（如 "Bearer"）
- `GetExpiresAt()`: 获取令牌过期时间戳
- `EncodeToJSON()`: 将令牌信息序列化为JSON

### 3.2 Authenticator 接口

`Authenticator` 接口定义了认证器的核心方法，位于 authn.go 中：

```go
// Authenticator defines methods used for token processing.
type Authenticator interface {
	// Sign is used to generate a token.
	Sign(ctx context.Context, userID string) (IToken, error)

	// Destroy is used to destroy a token.
	Destroy(ctx context.Context, accessToken string) error

	// ParseClaims parse the token and return the claims.
	ParseClaims(ctx context.Context, accessToken string) (*jwt.RegisteredClaims, error)

	// Release used to release the requested resources.
	Release() error
}
```

**功能说明**：
- `Sign()`: 为用户生成认证令牌
- `Destroy()`: 销毁令牌（防止重用）
- `ParseClaims()`: 解析令牌并返回声明信息
- `Release()`: 释放资源

### 3.3 Storer 接口

`Storer` 接口定义了令牌存储的基本操作，位于 store.go 中：

```go
// Storer token storage interface.
type Storer interface {
	// Store token data and specify expiration time.
	Set(ctx context.Context, accessToken string, expiration time.Duration) error

	// Delete token data from storage.
	Delete(ctx context.Context, accessToken string) (bool, error)

	// Check if token exists.
	Check(ctx context.Context, accessToken string) (bool, error)

	// Close the storage.
	Close() error
}
```

**功能说明**：
- `Set()`: 存储令牌数据并设置过期时间
- `Delete()`: 从存储中删除令牌数据
- `Check()`: 检查令牌是否存在
- `Close()`: 关闭存储连接

## 4. 核心实现

### 4.1 JWTAuth 实现

`JWTAuth` 结构体是 `Authenticator` 接口的JWT实现，位于 jwt.go 中：

**主要功能**：
- 基于JWT标准实现令牌生成与验证
- 支持防止重放攻击（通过存储已销毁的令牌）
- 可配置的令牌选项（过期时间、签名方法等）
- 支持国际化错误消息

**核心方法**：
- `Sign()`: 生成JWT令牌，包含标准Claims信息
- `parseToken()`: 解析并验证JWT令牌
- `Destroy()`: 销毁令牌，将其存入存储以防止重用
- `ParseClaims()`: 解析令牌Claims，并检查是否已被销毁
- `Release()`: 释放存储资源

### 4.2 tokenInfo 实现

`tokenInfo` 结构体是 `IToken` 接口的实现，位于 token.go 中：

**主要字段**：
- `Token`: 令牌字符串
- `Type`: 令牌类型（如 "Bearer"）
- `ExpiresAt`: 令牌过期时间戳

### 4.3 Redis 存储实现

`Store` 结构体是 `Storer` 接口的Redis实现，位于 redis.go 中：

**主要功能**：
- 在Redis中存储和管理令牌
- 支持自定义键前缀
- 提供令牌的设置、删除、检查和资源释放功能

**实现细节**：
- 使用 `github.com/redis/go-redis/v9` 作为Redis客户端
- 通过键前缀区分不同类型的令牌存储
- 令牌存储为Redis字符串，值为"1"，使用过期时间自动清理

## 5. 密码处理工具

`authn` 包提供了两个密码处理工具函数，位于 authn.go 中：

### 5.1 Encrypt 函数

```go
// Encrypt encrypts the plain text with bcrypt.
func Encrypt(source string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(source), bcrypt.DefaultCost)
	return string(hashedBytes), err
}
```

**功能**：使用bcrypt算法加密明文密码

### 5.2 Compare 函数

```go
// Compare compares the encrypted text with the plain text if it's the same.
func Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
```

**功能**：验证明文密码与加密密码是否匹配

## 6. 安全特性

### 6.1 防止重放攻击

该模块通过以下机制防止令牌重放攻击：
- 令牌销毁后，将其存储在Redis中
- 验证令牌时，检查令牌是否在已销毁列表中
- 利用Redis的过期时间机制，自动清理过期的已销毁令牌

### 6.2 令牌安全

- 使用标准JWT结构，包含完整的声明信息
- 支持可配置的签名方法（默认HS256）
- 令牌包含过期时间，自动失效

## 7. 国际化支持

模块集成了i18n支持，所有错误消息都可以根据上下文进行国际化：
- 预定义了标准错误消息和对应的国际化ID
- 错误消息使用`i18n.FromContext(ctx).LocalizeT()`方法获取本地化文本

## 8. 使用示例

### 8.1 创建JWT认证器

```go
import (
	"context"
	"github.com/moweilong/milady/pkg/authn/jwt"
	"github.com/moweilong/milady/pkg/authn/jwt/store/redis"
)

// 创建Redis存储
redisStore := redis.NewStore(&redis.Config{
	Addr:      "localhost:6379",
	KeyPrefix: "milady:token:",
})

// 创建JWT认证器
jwtAuth := jwt.New(redisStore, 
	jwt.WithSigningKey([]byte("your-secret-key")),
	jwt.WithExpired(2*time.Hour),
	jwt.WithIssuer("milady-service"),
)
```

### 8.2 生成令牌

```go
token, err := jwtAuth.Sign(ctx, "user123")
if err != nil {
	// 处理错误
}
```

### 8.3 验证令牌

```go
claims, err := jwtAuth.ParseClaims(ctx, tokenString)
if err != nil {
	// 处理错误（令牌无效或已过期）
}
userID := claims.Subject // 获取用户ID
```

### 8.4 销毁令牌

```go
err := jwtAuth.Destroy(ctx, tokenString)
if err != nil {
	// 处理错误
}
```

## 9. 设计特点

### 9.1 接口抽象

通过接口抽象，实现了：
- 认证逻辑与存储实现的解耦
- 支持多种认证方式和存储后端
- 便于单元测试和功能扩展

### 9.2 可配置性

使用选项模式（Option Pattern）提供灵活配置：
- 可自定义签名方法、密钥、过期时间等
- 可配置令牌头部信息
- 可自定义验证逻辑

### 9.3 安全性

- 使用bcrypt进行密码加密（防彩虹表攻击）
- 实现令牌销毁机制（防重放攻击）
- JWT标准实现，支持完整的令牌验证

## 10. 依赖关系

- `github.com/golang-jwt/jwt/v4`: JWT处理库
- `golang.org/x/crypto/bcrypt`: 密码加密库
- `github.com/redis/go-redis/v9`: Redis客户端
- `github.com/go-kratos/kratos/v2/errors`: 错误处理
- `github.com/nicksnyder/go-i18n/v2/i18n`: 国际化支持