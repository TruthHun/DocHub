![LOGO](static/Home/default/img/logo.png)

- 文库简介
- 技术栈构成
- 功能特点
- 使用教程
- 演示站点
- 移动端文库模板
- 商业方向
- TODO
- 关于作者
- 赞助我

## 文库简介
这是一个蛮遥远的梦想...

还在读大学的时候，当时就想着要搭建一个类似百度文库的文库站点，实现文档在线浏览。

然而，时间一晃，好几年过去了，自己终于亲自动手开发实现了一套开源文库程序。

尽管离百度文库还有着巨大的差距，尽管可(一)能(定)会存在Bug，但是，整套程序从前端到后端到数据库再到丑陋的LOGO设计，都是自己独立完成的，满满的成就感...

### 程序命名
起初开发这套程序，是想自己做一个IT行业的文库站点，也就是现在的[IT文库(http://wenku.it)](http://wenku.it)，当初给文库起名就叫`IT文库`，但是决定开源出来之后，觉得这样不妥，于是起了个叫`DocHub`的名字。

毕竟，有个Git仓库管理的站点叫`GitHub`，那文档(Doc)仓库也就是文库，我干嘛不叫`DocHub`呢？

`DocHub`的中文名叫`多哈`，程序猿嘛，写代码就要开心，开心就要`哈哈哈哈哈哈哈哈`，多`哈`...(好幼稚)



## 主要技术栈

### 后端

Go语言框架[Beego](https://beego.me)

### 前端

基于Bootstrap的前端框架[Flat-UI](https://github.com/designmodo/Flat-UI)

### 数据库

MySQL，数据存储

### 环境依赖

#### Libreoffice(或Openoffice)，用于将office文档转PDF

使用命令:
```
soffice --headless --invisible --convert-to pdf path/to/officefile --outdir path/to/outdir
```

#### pdf2svg

> 注意，这个要用最新版的`pdf2svg`

pdf2svg，用于将PDF转成svg矢量图片，以供阅读。

使用命令：
```
pdf2svg <in file.pdf> <out file.svg> [<page no>]
```

#### calibre

用于将`mobi`、`chm`、`epub`等文档转成PDF，然后再将`pdf`转成`svg`

> 注意：目前`mobi`、`epub`等文档的在线阅读功能还没有实现

#### 阿里云OSS
存储office文档、PDF文档以及svg等文件

> 注意：目前只支持阿里云的OSS，暂时不支持其他云存储(后期我再抽时间开发其他云存储)，不支持本地化存储

## 功能特点
### 文档在线阅读
`DocHub`文库通过`svg`矢量图来实现文档阅读体验的，我知道的文库站点中，[新浪爱问](http://ishare.iask.sina.com.cn/)就是通过`png`等图片提供文档阅读体验的。

`SVG`相比`png`、`jpeg`等图片格式有很大的优势，至少放大不会失真，而且与JPEG 和 GIF 图像比起来，svg尺寸更小，可压缩性更强，`DocHub`通过gzip，将svg文件压缩，一般情况下，能减少70%的文件大小，比如200kb的svg，gzip压缩后，只有60kb左右的大小。

使用svg，大大提升了加载速度，优化了内容的阅读体验。

#### office文档在线阅读

这个需要经过两层转化:
```
office --> pdf --> svg
```
![office文档在线阅读](static/tutorial/preview.png)

> 之前有考虑过office文档不经过转化，然后直接在线浏览的，但是方案比较复杂，部署不容易，至少我没部署成功过...

#### PDF文档在线阅读

将PDF文档通过`pdf2svg`转化，提供在线阅读

> 没有使用mozila的`pdf.js`作为PDF文档阅读的实现方案，主要是我没有解决`pdf.js`分片分页加载的问题，每次都需要将整个PDF文档下载下来才能提供阅读。如果文档大的话，用户需要等待好长时间，而且也比较耗费服务器带宽资源。

#### mobi、epub、chm文档在线阅读【TODO】
使用`calibre`将文档转成PDF，然后pdf再转svg。

> 目前该功能还没实现，epub、mobi等文档，现在还是暂时不能在线阅读

### 全文搜索【TODO】
全文搜索功能，之前是使用`coreseek`开发实现了的，但是现在`coreseek`的官网都已经挂了(IT的江湖里，有着它的传说，却不见了它的身影)...打算用`elasticsearch`重新实现这个功能。

### 积分功能
用户签到、上传分享文档，获得积分奖励；用户下载文档，需要消耗积分

### 阅读文档水印功能
在提供阅读的svg文件上添加水印

### 简介的页面

没有哪一个时代不是看脸的...

- 首页

> ![首页](static/tutorial/index.png)

- 文档阅读页

>![文档阅读页](static/tutorial/preview.png)

- 用户中心

> ![用户中心](static/tutorial/ucenter.png)


- 管理后台

> ![管理后台](static/tutorial/admin.png)


- 搜索结果
> ![搜索结果](static/tutorial/search.png)

> ![搜索结果](static/tutorial/search1.png)








