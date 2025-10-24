package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	v1 "github.com/moweilong/milady/pkg/api/apiserver/v1"
	"github.com/moweilong/milady/pkg/core"

	"k8s.io/klog/v2"
)

// Healthz 服务健康检查.
func (h *Handler) Healthz(c *gin.Context) {
	klog.FromContext(c.Request.Context()).Info("Healthz handler is called", "method", "Healthz", "status", "healthy")
	core.WriteResponse(c, v1.HealthzResponse{
		Status:    v1.ServiceStatus_Healthy,
		Timestamp: time.Now().Format(time.DateTime),
	}, nil)
}
