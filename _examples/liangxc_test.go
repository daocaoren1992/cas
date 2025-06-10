package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"gopkg.in/cas.v2"
	"net/http"
	"net/url"
)

// 自定义 Gin 中间件：基于 CAS 的登录校验
func CasAuthMiddleware(casClient *cas.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用 go-cas 包装请求
		casClient.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cas.IsAuthenticated(r) {
				// 未认证，跳转 CAS 登录
				urlToRedirectTo, _ := casClient.LoginUrlForRequest(r)
				// 替换 "login/login" 为 "login"
				//urlToRedirectTo = strings.Replace(urlToRedirectTo, "login/login", "login", 1)
				urlToRedirectTo, _ = url.QueryUnescape(urlToRedirectTo)
				urlToRedirectTo = "http://authserver.hcc.edu.cn/authserver/login?service=http://im-front-test.oceghome.com:8002/secure"
				println("redirct----->" + urlToRedirectTo)
				// 设置响应头部 Content-Type 为 HTML
				c.Header("Content-Type", "text/html; charset=UTF-8")

				// 输出 JavaScript，重定向到指定 URL
				c.Writer.WriteString("<script language='javascript'>window.location.href='" + urlToRedirectTo + "'</script>")
				//cas.RedirectToLogin(w, r)
				//c.Abort()
				return
			}

			// 用户已登录，将用户信息注入到 Gin 的上下文
			username := cas.Username(r)
			c.Set("cas_user", username)

			// 将处理后的 request 回传回 Gin
			c.Request = r
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
	}
}

func main() {

	// 硬编码设置日志等级，模拟命令行参数
	flag.Set("v", "2")              // 设置日志等级为 2
	flag.Set("logtostderr", "true") // 设置日志输出到终端

	// 解析 flag
	flag.Parse()

	// 必须调用 Flush，确保日志会写到终端
	defer glog.Flush()
	// CAS Server 地址（请替换为你自己的 CAS 服务器）
	casURL, _ := url.Parse("http://authserver.hcc.edu.cn/authserver")

	// 创建 go-cas 客户端
	casClient := cas.NewClient(&cas.Options{
		URL: casURL,
	})

	// Gin 启动
	r := gin.Default()

	// 使用 CAS 登录中间件保护的路由
	protected := r.Group("/secure")
	protected.Use(CasAuthMiddleware(casClient))
	{
		protected.GET("/", func(c *gin.Context) {
			// 从上下文获取 CAS 用户名
			username := c.GetString("cas_user")
			c.JSON(200, gin.H{
				"message":  "Welcome!",
				"username": username,
			})
		})
	}

	// 登出（直接重定向到 CAS logout）
	r.GET("/logout", func(c *gin.Context) {
		cas.RedirectToLogout(c.Writer, c.Request)
	})

	// 启动服务
	r.Run(":8002")
}
