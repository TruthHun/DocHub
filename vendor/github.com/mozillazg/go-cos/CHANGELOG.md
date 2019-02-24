# Changelog


## [0.9.0] (2018-08-04)

### 新增

* 新增 `c.Object.PresignedURL` 用于获取预签名授权 URL。
  可用于无需知道 SecretID 和 SecretKey 就可以上传和下载文件。
* 上传和下载 Object 的功能支持指定预签名授权 URL。

详见 PR 以及使用示例：

* https://github.com/mozillazg/go-cos/pull/5
* 通过预签名授权 URL 下载文件，示例：[object/getWithPresignedURL.go](./_example/object/getWithPresignedURL.go)
* 通过预签名授权 URL 上传文件，示例：[object/putWithPresignedURL.go](./_example/object/putWithPresignedURL.go)


## [0.8.0] (2018-05-26)

### 新增

* 新增 `func NewBaseURL(bucketURL string) (u *BaseURL, err error)` (via [91f7759])

### 变更

* `NewBucketURL` 函数使用新的 URL 域名规则。(via [7dcd701])     
  影响：如果有使用 `NewBucketURL` 函数生成 bucketURL 的话，使用时需要使用新的 Region 名称，
  详见 https://cloud.tencent.com/document/product/436/6224 ，未使用 `NewBucketURL` 函数不受影响


## [0.7.0] (2017-12-23)

### 新增

* 支持新增的 Put Object Copy API
* 新增 `github.com/mozillazg/go-cos/debug`，目前只包含 `DebugRequestTransport`


## [0.6.0] (2017-07-09)

### 新增

* 增加说明在某些情况下 ObjectPutHeaderOptions.ContentLength 必须要指定
* 增加 ObjectUploadPartOptions.ContentLength


## [0.5.0] (2017-06-28)

### 修复

* 修复 ACL 相关 API 突然失效的问题.
  (因为 COS ACL 相关 API 的 request 和 response xml body 的结构发生了变化)

### 删除

* 删除调试用的 DebugRequestTransport(把它移动到 examples/ 中)


## [0.4.0] (2017-06-24)

### 新增

* 增加 AuthorizationTransport 辅助添加认证信息

### 修改

* 去掉 API 中的 authTime 参数，默认不再自动添加 Authorization header
  改为通过自定义 client 的方式来添加认证信息


## [0.3.0] (2017-06-23)

### 新增

* 完成剩下的所有 API


## [0.2.0] (2017-06-10)

### 变更

* 调用 bucket 相关 API 时不再需要 bucket 参数, 把参数移到 service 中
* 把参数 signStartTime, signEndTime, keyStartTime, keyEndTime 合并为 authTime


## 0.1.0 (2017-06-10)

### 新增

* 完成 Service API
* 完成大部分 Bucket API(还剩一个 Put Bucket Lifecycle)


[0.9.0]: https://github.com/mozillazg/go-cos/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/mozillazg/go-cos/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/mozillazg/go-cos/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/mozillazg/go-cos/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/mozillazg/go-cos/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/mozillazg/go-cos/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/mozillazg/go-cos/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/mozillazg/go-cos/compare/v0.1.0...v0.2.0

[91f7759]: https://github.com/mozillazg/go-cos/commit/91f7759958f9631e8997f47d30ae4044455fc971
[7dcd701]: https://github.com/mozillazg/go-cos/commit/7dcd701975f483d57525b292ab31d0f9a6c8866c
