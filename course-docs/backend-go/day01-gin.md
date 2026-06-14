# DAY01 - Gin 基础

## Gin 是什么

Gin 是 Go 生态中常用的 Web 框架，适合快速编写 REST API。

## 最小示例

```go
package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
	r.Run(":8080")
}
```

## 知识点

- `gin.Default()` 会创建带日志和恢复中间件的路由引擎。
- `GET` 用来注册 GET 请求路由。
- `c.JSON` 用来返回 JSON 响应。
