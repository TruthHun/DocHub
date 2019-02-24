/*
Package cos 腾讯云对象存储服务 COS(Cloud Object Storage) Go SDK。


COS API Version

封装了 V5 版本的 XML API 。


Usage

在项目的 _example 目录下有各个 API 的使用示例 。


Authentication

默认所有 API 都是匿名访问. 如果想添加认证信息的话,可以通过自定义一个 http.Client 来添加认证信息.

比如, 使用内置的 AuthorizationTransport 来为请求增加 Authorization Header 签名信息:

	client := cos.NewClient(b, &http.Client{
			Transport: &cos.AuthorizationTransport{
				SecretID:  "COS_SECRETID",
				SecretKey: "COS_SECRETKEY",
			},
		})

*/
package cos
