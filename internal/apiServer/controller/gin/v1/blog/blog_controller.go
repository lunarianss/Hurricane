// Copyright 2024 Benjamin Lee <cyan0908@163.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package controller

import (
	"github.com/lunarianss/Hurricane/internal/apiServer/service"
)

type BlogController struct {
	blogService *service.BlogService
}

func NewBlogController(blogSrv *service.BlogService) *BlogController {
	return &BlogController{blogService: blogSrv}
}
