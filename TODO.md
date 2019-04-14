## 待开发功能

- [ ] 增加权限管理功能
- [ ] 增加用户等级划分
- [ ] 增加第三方登录的支持：QQ登录、微信登录、微博登录、GitHub登录、Gitee登录
- [ ] 增强用户中心，全面参照百度文库和新浪爱问共享资源的用户中心设计和开发功能
- [ ] 邀请注册功能，增强站点运营
- [ ] 微信小程序 【预计4.0版本】
- [ ] 控制面板echarts统计图表 —— 统计每天、每周、每月各种数据增长曲线图，如文档量、用户注册量、每天签到量等
- [ ] 容器分布式应用部署，session的共享问题（redis实现）；日志问题
- [ ] 广告管理功能模块
- [ ] 充值功能(支付宝/微信充值)
- [ ] 增加爬虫采集功能
- [ ] 改造response函数，参数使用struct
- [ ] 积分商城
- [ ] 评论管理；用户可以在前台删除自己的评论；删除评论之后，文档评分发生变化
- [ ] 程序以微服务形式做成模块化，拆分成：web模块、爬虫模块、文档转换模块、全文搜索模块、用户模块、云存储模块
- [ ] 管理后台，管控是否开放注册控制
- [ ] url路由链接调整？
- [ ] 凌晨自动更新站点地图
- [ ] 文库频道，可在管理后台更新频道图片封面
- [ ] 文档分类数据统计修正(管理后台处理)
- [ ] 频道图片上传和更换功能
- [ ] 管理后台增加一个上传文档必须达到的最低积分要求，避免一些新用户上传垃圾文档刷积分
- [ ] 更换[KindEditor](https://www.oschina.net/news/104631/kindeditor-upload-vulnerability)
- [ ] 文档上传间隔频率控制


## DocHub v2.3
- [ ] 用户注册和登录成功之后的跳转优化
- [x] 图片裁剪质量优化
- [x] `文档管理`文档删除的错误
- [x] 移除对zoneinfo.zip的依赖
- [x] calibre 文档转换优化
- [x] 解决程序不支持utf8mb4数据库字符编码问题（ERROR 1071 (42000): Specified key was too long; max key length is 767 bytes），索引字段太长导致
- [ ] 完成部署文档
    - [ ] Windows 部署文档
    - [ ] Linux(Ubuntu) 部署文档
- [ ] 完成云存储配置部署文档
    - [ ] 阿里云OSS
    - [ ] 本地存储minio
    - [ ] 腾讯云存储cos
    - [ ] 七牛云存储qiniu
    - [ ] 百度云存储bos
    - [ ] 华为云存储obs
    - [ ] 又拍云upyun


```
CREATE DATABASE dochub CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

```
CREATE DATABASE dochub CHARACTER SET utf8 COLLATE utf8_general_ci;
```


## DocHub v2.2
- [x] 多样化存储实现
    - [x] 阿里云 - `OSS`
    - [x] 本地存储 - `minio`
    - [x] 腾讯云存储 - `cos`
    - [x] 七牛云存储 - `qiniu`
    - [x] 百度云存储 - `bos`
    - [x] 华为云存储 - `obs`
    - [x] 又拍云 - `upyun`
    - [ ] 金山云
    - [ ] UCloud
    - [ ] 青云
- [x] SEO优化，参考新浪爱问文库，从文档中提取文本，放到HTML页面隐藏显示的div中
- [x] 积分问题[#40](https://github.com/TruthHun/DocHub/issues/40)
- [x] 封面、横幅裁剪，不再依赖云存储做图片处理
- [x] `sudo`支持
- [x] 下载文档出错，下载id为100的文档，可能下载到id位80的文档（MySQL查询语句错误造成的）
- [x] 管理后台测试的时候，提示先保存
- [x] 把引入的外部js、css公共资源库替换成本地资源库，以方便内网部署
- [x] 增加虚拟根目录，路径为`./virtualroot`，并把`.well-known`加入到默认静态目录中，以便申请`let's encrypt`证书
- [x] 在管理后台增加注册邮箱验证开关，用户在注册的时候可以不验证邮箱
- [x] 移除管理后台系统设置的频道管理，直接根据频道排序，在首页展示，避免造成困扰
- [x] 优化文档上传流程
- [x] 解决文件路径问题（调用 cmd 的时候，用文件的绝对路径）
- [x] 文档评分人数统计问题(评分出现-1的情况)
- [x] 文档分享人可把文档设置为不可下载状态
- [x] 邮箱配置更改的时候进行验证



```
DROP TABLE `hc_ad`, `hc_admin`, `hc_ad_position`, `hc_banner`, `hc_category`, `hc_coin_log`, `hc_collect`, `hc_collect_folder`, `hc_document`, `hc_document_comment`, `hc_document_illegal`, `hc_document_info`, `hc_document_recycle`, `hc_document_remark`, `hc_document_store`, `hc_doc_text`, `hc_free_down`, `hc_friend`, `hc_pages`, `hc_relate`, `hc_report`, `hc_search_log`, `hc_seo`, `hc_sign`, `hc_suggest`, `hc_sys`, `hc_word`;
```

---------------

文档上传流程优化：
    1. 未登录用户，不允许上传文档
    2. 已登录用户，积分达不到管理后台规定的积分数，不允许上传文档
    3. 用户上传文档的时候，前端计算文档 MD5 的值，然后请求后台对比值是否存在，存在了，则表示文档已经存在，直接`秒传`
    4. 用户上传的文档在管理后台不存在，则把原文档上传到云存储，文档信息(分类、大小、md5等)、用户积分变化等存入数据库
    5. 数据入库成功之后，再进行文档转换，文档最终转换成功之后，更新文档状态
    
    (文档处理，分已存在和未存在的文档进行处理)




## DocHub v2.1

- [x] 修复搜索的时候，MySQL like 查询，文档在排序的时候查询不到数据的问题
- [x] 文档图标优化 
- [x] PC端个人中心页面大调整
- [x] 程序后端代码优化
- [x] 使用`govender`管理依赖包，方便需要做二次开发的朋友
- [x] epub、mobi等文档转PDF参数优化
- [x] 实现移动端支持。默认启用移动端，可后台`系统设置`进行控制。
    - [x] 首页
    - [x] 列表页
    - [x] 阅读页
    - [x] 搜索页
    - [x] 登录页
    - [x] 注册页
    - [x] 找回密码
    - [x] 个人中心
        - [x] 文档管理
        - [x] 收藏管理
        - [x] 文档编辑功能(移动端隐藏)
        - [x] 积分日志
    - [x] 文档上传(移动端隐藏)

## DocHub v2.0


- [x] 文档阅读页面再优化，修复v1.1版本优化造成的Bug——剩余未阅读页数不准确、无法正确翻页等问题。
- [x] 管理后台，暂时隐藏未开发的`广告管理`和`采集管理`功能
- [x] 文档上传，显示上传进度
- [x] 后台可配置最大上传的文件大小
- [x] 除了数据库之外的配置项，其余配置项在后台可管理和配置
- [x] 文档举报原因，后台可配置
- [x] 被禁用户还能正常登录的Bug
- [x] Sitemap站点地图定时自动更新和生成，也支持管理后台一键生成
- [x] 后端代码持续优化...
- [x] 增加svgo压缩 https://github.com/svg/svgo [这个库很强大，但是为了使用这个功能，需要安装个nodejs环境，而且动态语言实现的，肯定不如静态语言的性能好，届时找时间review一下他的js代码，然后用Go实现一个试下]
- [x] 程序安装功能
- [x] 增加邮箱配置测试，测试是否能正常发送邮件
- [x] 检测OSS配置，是否能连通
- [x] 显示日志文件列表、大小以及下载地址
- [x] 增加ElasticSearch数据统计，显示索引数据情况，以及索引更新等相关操作
- [x] 单页优化
- [x] 批量更新索引
- [x] 文档新建和更新时的TimeUpdate，即更新时间
- [x] 删除文档或者恢复文档的时候删除或者更新索引
- [x] 管理后台登录验证码管理（在修改密码的时候进行修改）
- [x] 文档管理优化
- [x] 制作docker镜像
- [x] 使用ElasticSearch搭建全文搜索引擎
    - [x] 全文搜索使用elasticsearch
    
    
暂时移除sitemap和索引定时更新功能，2.1版本再增加和优化

- [x] pdftotext
- [x] email
- [x] logs
- [x] pdf2svg
- [x] imagemagick
- [x] oss
- [x] soffice
- [x] calibre


ElasticSearch 部署示例：

```
mkdir -p /www/elasticsearch/dochub/data && sudo chmod 0777 -R /www/elasticsearch/dochub/data
sudo docker run -d -p 9300:9300 -p 9200:9200 --restart always -v /www/elasticsearch/dochub/data:/usr/share/elasticsearch/data --name dochub-search truthhun/elasticsearch:6.2.4.ik
```
其中 `data`目录是索引数据存放目录，必须有读写权限，如执行下面语句，赋予读写权限：
```
sudo chmod 0777 -R /www/elasticsearch/dochub/data
```

记得屏蔽对外的9200、9300端口

## 测试
- [x] 发送邮件测试


## DocHub v1.1
- [x] OSS存储代码代码优化（review了一下，之前的代码太乱了）
- [x] 重新设计登录页面。之前的登录页面确实太丑了(不过现在的页面好像也好不到哪去...)
- [x] 用户头像和文档封面等默认图片优化，在加载图片的时候直接在前端使用`onerror`，不再在后端查询oss中图片是否存在以及不存在时返回默认图片。
- [x] 所有相关配置项，为了配置的方便，都统一放到app.conf文件。配置文件中的每一项，都加上了详尽的配置说明。
- [x] mobi、epub、chm、txt等格式文档在线浏览的实现支持。
- [x] 解决邮件发送问题，统一使用SMTP发送邮件，并实现对TLS/SSL邮件的发送支持。使用了https://github.com/go-gomail/gomail库。
- [x] 文档阅读页面性能优化
- [x] 文档预览页数限制(可在`管理后台`->`系统设置`->`文档最大预览页数`做限制。这样的好处就是，如果一个300页的文档，只提供100页给用户阅读，可以减少服务器后端PDF转svg的资源开销，也可以促进用户使用积分下载文档...新浪爱问共享资料就是这么干的...)