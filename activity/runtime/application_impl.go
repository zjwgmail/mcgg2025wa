package runtime

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
)

type Application struct {
	engine      http.Handler
	handler     map[string][]func(r *gin.RouterGroup, hand ...*gin.HandlerFunc)
	routers     []Router
	mux         sync.RWMutex
	middlewares map[string]interface{}
}

type Router struct {
	HttpMethod, RelativePath, handler string
}

type Routers struct {
	List []Router
}

// 初始化应用上下文
func NewApplication() *Application {

	return &Application{
		middlewares: make(map[string]interface{}),
		handler:     make(map[string][]func(r *gin.RouterGroup, hand ...*gin.HandlerFunc)),
		routers:     make([]Router, 0),
	}
}

// SetEngine 设置路由引擎
func (e *Application) SetEngine(engine http.Handler) {

	e.engine = engine
}

// GetEngine 获取路由引擎
func (e Application) GetEngine() http.Handler {
	return e.engine
}

// SetRouter 设置路由表
func (e Application) SetRouter() []Router {
	switch e.engine.(type) {
	case *gin.Engine:
		routes := e.engine.(*gin.Engine).Routes()
		for _, router := range routes {
			e.routers = append(e.routers, Router{
				RelativePath: router.Path,
				handler:      router.Handler,
				HttpMethod:   router.Method,
			})
		}
	}
	return e.routers
}

// GetRouter 获取路由表
func (e Application) GetRouter() []Router {
	return e.routers
}

// SetMiddleware 设置中间件
func (e Application) SetMiddleware(key string, middleware interface{}) {
	e.mux.Lock()
	defer e.mux.Unlock()
	e.middlewares[key] = middleware
}

// GetMiddleware 获取所有中间件
func (e *Application) GetMiddleware() map[string]interface{} {
	return e.middlewares
}

// GetMiddlewareKey 获取对应key的中间件
func (e *Application) GetMiddlewareKey(key string) interface{} {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.middlewares[key]
}

func (e *Application) SetHandler(key string, routerGroup func(r *gin.RouterGroup, hand ...*gin.HandlerFunc)) {
	e.mux.Lock()
	defer e.mux.Unlock()
	e.handler[key] = append(e.handler[key], routerGroup)
}

func (e *Application) GetHandler() map[string][]func(r *gin.RouterGroup, hand ...*gin.HandlerFunc) {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.handler
}

func (e *Application) GetHandlerPrefix(key string) []func(r *gin.RouterGroup, hand ...*gin.HandlerFunc) {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.handler[key]
}
