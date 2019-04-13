package models

import (
	"fmt"
	"time"

	"github.com/TruthHun/DocHub/helper"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

func install() {
	installAdmin()
	installCategory()
	installFriendlinks()
	installPages()
	installSeo()
	installSys()
	installCfg()
	NewSys().UpdateGlobalConfig()
	NewConfig().UpdateGlobalConfig()
}

//安装管理员初始数据
func installAdmin() {
	var admin = Admin{
		Id:       1,
		Username: "admin",
		Password: helper.MD5Crypt("admin"),
		Email:    "TruthHun@QQ.COM",
		Code:     "芝麻开门",
	}
	beego.Info("初始化管理员数据")
	if _, _, err := orm.NewOrm().ReadOrCreate(&admin, "Id"); err != nil {
		helper.Logger.Error("初始化管理员数据失败：" + err.Error())
	}
}

//安装系统初始数据
func installSys() {
	var sys = Sys{
		Id: 1,
		TplEmailReg: `<style type="text/css">
				p{text-indent: 2em;}
			</style>
			<div><strong>尊敬的用户</strong></div>
			<p>您好，非常感谢您对DocHub文库(<a href="https://github.com/TruthHun/DocHub" target="_blank" title="DocHub文库">DocHub</a>)的关注和热爱</p>
			<p>您本次申请注册成为DocHub文库会员的邮箱验证码是: <strong style="font-size: 30px;color: red;">{code}</strong></p>
			<p>如果非您本人操作，请忽略该邮件。</p>`,
		TplEmailFindPwd: `<style type="text/css">
				p{text-indent: 2em;}
			</style>
			<div><strong>尊敬的用户</strong></div>
			<p>您好，非常感谢您对DocHub文库(<a href="https://github.com/TruthHun/DocHub" target="_blank" title="DocHub文库">DocHub</a>)的关注和热爱</p>
			<p>您本次申请找回密码的邮箱验证码是: <strong style="font-size: 30px;color: red;">{code}</strong></p>
			<p>如果非您本人操作，请忽略该邮件。</p>`,
		Trends:            "1,2,3,4,5",
		Site:              "DocHub(多哈)文库",
		Reward:            5,
		Sign:              5,
		Question:          "DocHub文库的中文名是什么？",
		Answer:            "多哈",
		ListRows:          10,
		TimeExpireHotspot: 604800,
		TimeExpireRelate:  604800,
		MaxFile:           52428800, //50M
		CoinReg:           10,       //注册奖励金币
		MobileOn:          true,
		ReportReasons: `1:垃圾广告
2:淫秽色情
3:虚假中奖
4:敏感信息
5:人身攻击
6:骚扰他人`, //举报原因
		Watermark:     "DocHub", //文档水印
		StoreType:     string(StoreOss),
		CheckRegEmail: true,
	}
	orm.NewOrm().ReadOrCreate(&sys, "Id")
}

//安装友链初始数据
func installFriendlinks() {
	var friend = new(Friend)
	if orm.NewOrm().QueryTable(friend).Filter("id__gt", 0).One(friend); friend.Id > 0 {
		return
	}

	now := int(time.Now().Unix())
	var friends = []Friend{
		Friend{
			Id:         1,
			Title:      "书栈网",
			Link:       "https://www.bookstack.cn",
			Status:     true,
			Sort:       1,
			TimeCreate: now,
		},
		Friend{
			Id:         2,
			Title:      "掘金量化",
			Link:       "https://www.myquant.cn",
			Status:     true,
			Sort:       2,
			TimeCreate: now,
		},
		Friend{
			Id:         4,
			Title:      "南宁引力互动科技",
			Link:       "http://www.gxyinli.com",
			Status:     true,
			Sort:       3,
			TimeCreate: now,
		},
		Friend{
			Id:         3,
			Title:      "HC-CMS",
			Link:       "http://www.hc-cms.com",
			Status:     true,
			Sort:       4,
			TimeCreate: now,
		},
	}
	if _, err := orm.NewOrm().InsertMulti(len(friends), friends); err != nil {
		helper.Logger.Error("初始化友链数据失败：" + err.Error())
	}
}

//安装单页初始数据
//存在唯一索引Alias，已存在的数据不会继续写入
func installPages() {
	//存在单页了，则表明已经初始化过数据
	var page = new(Pages)
	if orm.NewOrm().QueryTable(page).Filter("id__gt", 0).One(page); page.Id > 0 {
		return
	}

	now := int(time.Now().Unix())
	var pages = []Pages{
		Pages{
			Name:        "关于我们",
			Alias:       "about",
			Title:       "关于我们",
			Keywords:    "关于我们,about us,dochub",
			Description: "这是关于我们的单页",
			Content:     "这是关于我们的单页内容",
			TimeCreate:  now,
			Status:      true,
		},
		Pages{
			Name:        "文库协议",
			Alias:       "agreement",
			Title:       "关于我们",
			Keywords:    "文库协议,agreement,dochub",
			Description: "这是文库协议的单页",
			Content:     "这是文库协议的单页内容",
			TimeCreate:  now,
			Status:      true,
		},
		Pages{
			Name:        "意见反馈",
			Alias:       "feedback",
			Title:       "意见反馈",
			Keywords:    "意见反馈,feedback,dochub",
			Description: "这是意见反馈的单页",
			Content:     "这是意见反馈的单页内容",
			TimeCreate:  now,
			Status:      true,
		},
		Pages{
			Name:        "免责声明",
			Alias:       "response",
			Title:       "免责声明",
			Keywords:    "免责声明,response,dochub",
			Description: "这是免责声明的单页",
			Content:     "这是免责声明的单页内容",
			TimeCreate:  now,
			Status:      true,
		},
		Pages{
			Name:        "联系我们",
			Alias:       "contact",
			Title:       "意见反馈",
			Keywords:    "意见反馈,contact,dochub",
			Description: "这是联系我们的单页",
			Content:     "这是联系我们的单页内容",
			TimeCreate:  now,
			Status:      true,
		},
	}
	orm.NewOrm().InsertMulti(len(pages), &pages)
}

//安装SEO初始数据
//存在唯一索引Page字段，已存在数据，不会继续写入
func installSeo() {
	seo := new(Seo)
	if orm.NewOrm().QueryTable(seo).Filter("id__gt", 0).One(seo); seo.Id > 0 {
		return
	}
	var seos = []Seo{
		Seo{
			Name:        "首页",
			Page:        "PC-Index",
			Title:       "{title} - {sitename}",
			Keywords:    "{keywords}",
			Description: "{description}",
		},
		Seo{
			Name:        "列表页",
			Page:        "PC-List",
			Title:       "{title} - {sitename}",
			Keywords:    "{keywords}",
			Description: "{description}",
		},
		Seo{
			Name:        "文档上传页",
			Page:        "PC-Upload",
			Title:       "{title} - {sitename}",
			Keywords:    "{keywords}",
			Description: "{description}",
		},
		Seo{
			Name:        "文档预览页",
			Page:        "PC-View",
			Title:       "{title} - {sitename}",
			Keywords:    "{keywords}",
			Description: "{description}",
		},
		Seo{
			Name:        "用户中心文档列表页",
			Page:        "PC-Ucenter-Doc",
			Title:       "{title} - {sitename}",
			Keywords:    "{keywords}",
			Description: "{description}",
		},
		Seo{
			Name:        "用户中心积分记录页",
			Page:        "PC-Ucenter-Coin",
			Title:       "{title} - {sitename}",
			Keywords:    "{keywords}",
			Description: "{description}",
		},
		Seo{
			Name:        "用户中心收藏夹页",
			Page:        "PC-Ucenter-Folder",
			Title:       "{title} - {sitename}",
			Keywords:    "{keywords}",
			Description: "{description}",
		},
		Seo{
			Name:        "找回密码页",
			Page:        "PC-Findpwd",
			Title:       "{title} - {sitename}",
			Keywords:    "{keywords}",
			Description: "{description}",
		},
		Seo{
			Name:        "用户注册页",
			Page:        "PC-Reg",
			Title:       "{title} - {sitename}",
			Keywords:    "{keywords}",
			Description: "{description}",
		},
		Seo{
			Name:        "用户登录页",
			Page:        "PC-Login",
			Title:       "{title} - {sitename}",
			Keywords:    "{keywords}",
			Description: "{description}",
		},
		Seo{
			Name:        "单页",
			Page:        "PC-Pages",
			Title:       "{title} - {sitename}",
			Keywords:    "{keywords}",
			Description: "{description}",
		},
		Seo{
			Name:        "搜索结果页",
			Page:        "PC-Search",
			Title:       "{title} - {sitename}",
			Keywords:    "{keywords}",
			Description: "{description}",
		},
	}
	orm.NewOrm().InsertMulti(len(seos), &seos)
}

//安装分类初始数据
//带有主键id数据的初始化，如果已经存在数据，则不会继续写入
func installCategory() {
	//存在分类了，则表明已经初始化过数据
	var cate = new(Category)
	o := orm.NewOrm()
	defer func() {
		var cates []Category
		o.QueryTable(GetTableCategory()).Filter("Pid__in", 0).All(&cates)
		for _, item := range cates {
			if item.Cover == "" {
				item.Cover = fmt.Sprintf("/static/Home/default/img/cover-%v.png", item.Alias)
				o.Update(&item)
			}
		}
	}()

	if o.QueryTable(cate).Filter("id__gt", 0).One(cate); cate.Id > 0 {
		return
	}

	sql := `INSERT INTO hc_category (Id, Pid, Title, Cnt, Sort, Alias, Status) VALUES
		(1, 0, '教育频道', 0, 0, 'edu', 1),
		(2, 0, '专业资料', 0, 1, 'pro', 1),
		(3, 0, '实用文档', 0, 2, 'pra', 1),
		(4, 0, '资格考试', 0, 3, 'exam', 1),
		(5, 0, '生活休闲', 0, 4, 'life', 1),
		(7, 1, '幼儿教育', 0, 0, '', 1),
		(8, 1, '小学教育', 0, 0, '', 1),
		(9, 1, '初中教育', 0, 0, '', 1),
		(10, 1, '高中教育', 0, 0, '', 1),
		(11, 1, '职业教育', 0, 0, '', 1),
		(12, 1, '成人教育', 0, 0, '', 1),
		(13, 1, '文库题库', 0, 0, '', 1),
		(15, 7, '幼儿读物', 0, 0, '', 1),
		(16, 7, '少儿英语', 0, 1, '', 1),
		(17, 7, '唐诗宋词', 0, 0, '', 1),
		(18, 7, '育儿理论经验', 0, 0, '', 1),
		(19, 7, '育儿知识', 0, 0, '', 1),
		(20, 7, '家庭教育', 0, 0, '', 1),
		(21, 2, '人文社科', 0, 0, '', 1),
		(22, 2, '经营营销', 0, 0, '', 1),
		(23, 2, '工程科技', 0, 0, '', 1),
		(24, 2, 'IT/计算机', 0, 0, '', 1),
		(25, 2, '自然科学', 0, 0, '', 1),
		(26, 2, '医疗卫生', 0, 0, '', 1),
		(27, 2, '农林渔牧', 0, 0, '', 1),
		(28, 24, '互联网', 0, 0, '', 1),
		(29, 24, '电脑基础知识', 0, 0, '', 1),
		(30, 24, '计算机软件及应用', 0, 0, '', 1),
		(31, 24, '计算机硬件及网络', 0, 0, '', 1),
		(32, 8, '语文', 0, 0, '', 1),
		(33, 8, '数学', 0, 0, '', 1),
		(34, 8, '英语', 0, 0, '', 1),
		(35, 8, '作文', 0, 0, '', 1),
		(36, 8, '其它课程', 0, 0, '', 1),
		(37, 9, '作文库', 0, 0, '', 1),
		(38, 9, '语文', 0, 0, '', 1),
		(39, 9, '数学', 0, 0, '', 1),
		(40, 9, '英语', 0, 0, '', 1),
		(41, 9, '物理', 0, 0, '', 1),
		(42, 9, '化学', 0, 0, '', 1),
		(43, 9, '历史', 0, 0, '', 1),
		(44, 9, '生物', 0, 0, '', 1),
		(45, 9, '地理', 0, 0, '', 1),
		(46, 9, '政治', 0, 0, '', 1),
		(47, 9, '中考', 0, 0, '', 1),
		(48, 9, '科学', 0, 0, '', 1),
		(49, 9, '竞赛', 0, 0, '', 1),
		(50, 9, '其它课程', 0, 0, '', 1),
		(52, 1, '高等教育', 0, 0, '', 1),
		(53, 52, '理学', 0, 0, '', 1),
		(54, 52, '工学', 0, 0, '', 1),
		(55, 52, '经济学', 0, 0, '', 1),
		(56, 52, '医学', 0, 0, '', 1),
		(57, 52, '管理学', 0, 0, '', 1),
		(58, 52, '文学', 0, 0, '', 1),
		(59, 52, '哲学', 0, 0, '', 1),
		(60, 52, '历史学', 0, 0, '', 1),
		(61, 52, '法学', 0, 0, '', 1),
		(62, 52, '教育学', 0, 0, '', 1),
		(63, 52, '农学', 0, 0, '', 1),
		(64, 52, '军事', 0, 0, '', 1),
		(65, 52, '艺术', 0, 0, '', 1),
		(66, 52, '研究生入学考试', 0, 0, '', 1),
		(67, 52, '院校资料', 0, 0, '', 1),
		(68, 52, '其它', 0, 9, '', 1),
		(69, 13, '中考题库', 0, 0, '', 1),
		(70, 13, '高考题库', 0, 0, '', 1),
		(71, 13, '公务员题库', 0, 0, '', 1),
		(72, 13, '外语题库', 0, 0, '', 1),
		(73, 13, '考研题库', 0, 0, '', 1),
		(74, 12, '成考', 0, 0, '', 1),
		(75, 12, '自考', 0, 0, '', 1),
		(76, 12, '专升本', 0, 0, '', 1),
		(77, 12, '电大', 0, 0, '', 1),
		(78, 12, '远程、网络教育', 0, 0, '', 1),
		(79, 11, '中职中专', 0, 0, '', 1),
		(80, 11, '职高对口', 0, 0, '', 1),
		(81, 11, '职业技术培训', 0, 0, '', 1),
		(82, 11, '其它', 0, 0, '', 1),
		(83, 10, '语文', 0, 0, '', 1),
		(84, 10, '数学', 0, 0, '', 1),
		(85, 10, '英语', 0, 0, '', 1),
		(86, 10, '物理', 0, 0, '', 1),
		(87, 10, '化学', 0, 0, '', 1),
		(88, 10, '历史', 0, 0, '', 1),
		(89, 10, '生物', 0, 0, '', 1),
		(90, 10, '地理', 0, 0, '', 1),
		(91, 10, '思想政治', 0, 0, '', 1),
		(92, 10, '高中作文', 0, 0, '', 1),
		(93, 10, '学科竞赛', 0, 0, '', 1),
		(94, 10, '其它课程', 0, 0, '', 1),
		(95, 21, '法律资料', 0, 0, '', 1),
		(96, 21, '军事/政治', 0, 0, '', 1),
		(97, 21, '广告/传媒', 0, 0, '', 1),
		(98, 21, '设计/艺术', 0, 0, '', 1),
		(99, 21, '教育学/心理学', 0, 0, '', 1),
		(100, 21, '文化/宗教', 0, 0, '', 1),
		(101, 21, '哲学/历史', 0, 0, '', 1),
		(102, 21, '文学研究', 0, 0, '', 1),
		(103, 21, '社会学', 0, 0, '', 1),
		(104, 22, '经济/市场', 0, 0, '', 1),
		(105, 22, '金融/投资', 0, 0, '', 1),
		(106, 22, '人力资源管理', 0, 0, '', 1),
		(107, 22, '财务管理', 0, 0, '', 1),
		(108, 22, '生产/经营管理', 0, 0, '', 1),
		(109, 22, '企业管理', 0, 0, '', 1),
		(110, 22, '公共/行政管理', 0, 0, '', 1),
		(111, 22, '销售/营销', 0, 0, '', 1),
		(112, 23, '信息与通信', 0, 0, '', 1),
		(113, 23, '电子/电路', 0, 0, '', 1),
		(114, 23, '建筑/土木', 0, 0, '', 1),
		(115, 23, '城乡/园林规划', 0, 0, '', 1),
		(116, 23, '环境科学/食品科学', 0, 0, '', 1),
		(117, 23, '电力/水利', 0, 0, '', 1),
		(118, 23, '交通运输', 0, 0, '', 1),
		(119, 23, '能源/化工', 0, 0, '', 1),
		(120, 23, '机械/仪表', 0, 0, '', 1),
		(121, 23, '冶金/矿山/地质', 0, 0, '', 1),
		(122, 23, '纺织/轻工业', 0, 0, '', 1),
		(123, 23, '材料科学', 0, 0, '', 1),
		(124, 23, '兵器/核科学', 0, 0, '', 1),
		(125, 25, '数学', 0, 0, '', 1),
		(126, 25, '物理', 0, 0, '', 1),
		(127, 25, '化学', 0, 0, '', 1),
		(128, 25, '生物学', 0, 0, '', 1),
		(129, 25, '天文/地理', 0, 0, '', 1),
		(130, 26, '临床医学', 0, 0, '', 1),
		(131, 26, '基础医学', 0, 0, '', 1),
		(132, 26, '预防医学', 0, 0, '', 1),
		(133, 26, '中医中药', 0, 0, '', 1),
		(134, 26, '药学', 0, 0, '', 1),
		(135, 27, '农学', 0, 0, '', 1),
		(136, 27, '林学', 0, 0, '', 1),
		(137, 27, '畜牧兽医', 0, 0, '', 1),
		(138, 27, '水产渔业', 0, 0, '', 1),
		(139, 3, '求职/职场', 0, 0, '', 1),
		(140, 3, '计划/解决方案', 0, 0, '', 1),
		(141, 3, '总结/汇报', 0, 0, '', 1),
		(142, 3, '党团工作', 0, 0, '', 1),
		(143, 3, '工作范文', 0, 0, '', 1),
		(144, 3, '表格/模板', 0, 0, '', 1),
		(145, 139, '简历', 0, 0, '', 1),
		(146, 139, '面试', 0, 0, '', 1),
		(147, 139, '职业规划', 0, 0, '', 1),
		(148, 139, '自我管理与提升', 0, 0, '', 1),
		(149, 139, '笔试', 0, 0, '', 1),
		(150, 139, '社交礼仪', 0, 0, '', 1),
		(151, 140, '学习计划', 0, 0, '', 1),
		(152, 140, '工作计划', 0, 0, '', 1),
		(153, 140, '商业计划', 0, 0, '', 1),
		(154, 140, '营销/活动策划', 0, 0, '', 1),
		(155, 140, '解决方案', 0, 0, '', 1),
		(156, 140, '其它', 0, 0, '', 1),
		(157, 141, '学习总结', 0, 0, '', 1),
		(158, 141, '实习总结', 0, 0, '', 1),
		(159, 141, '工作总结/汇报', 0, 0, '', 1),
		(160, 141, '其它', 0, 0, '', 1),
		(161, 142, '入党/转正申请', 0, 0, '', 1),
		(162, 142, '思想汇报/心得体会', 0, 0, '', 1),
		(163, 142, '党团建设', 0, 0, '', 1),
		(164, 142, '其它', 0, 0, '', 1),
		(165, 143, '制度/规范', 0, 0, '', 1),
		(166, 143, '行政公文', 0, 0, '', 1),
		(167, 143, '演讲/主持', 0, 0, '', 1),
		(168, 143, '其它', 0, 0, '', 1),
		(169, 144, '合同协议', 0, 0, '', 1),
		(170, 144, '书信模板', 0, 0, '', 1),
		(171, 144, '表格类模板', 0, 0, '', 1),
		(172, 144, '调查/报告', 0, 0, '', 1),
		(173, 4, '财会类', 0, 0, '', 1),
		(174, 4, '公务员类', 0, 0, '', 1),
		(175, 4, '学历类', 0, 0, '', 1),
		(176, 4, '建筑类', 0, 0, '', 1),
		(177, 4, '外语类', 0, 0, '', 1),
		(178, 4, '资格类', 0, 0, '', 1),
		(179, 4, '外贸类', 0, 0, '', 1),
		(180, 4, '医药类', 0, 0, '', 1),
		(181, 4, '计算机类', 0, 0, '', 1),
		(182, 173, '注册会计师', 0, 0, '', 1),
		(183, 173, '价格鉴证师', 0, 0, '', 1),
		(184, 173, '证券从业资格', 0, 0, '', 1),
		(185, 173, '经济师', 0, 0, '', 1),
		(186, 173, '初级经济师', 0, 0, '', 1),
		(187, 173, '中级经济师', 0, 0, '', 1),
		(188, 173, '注册税务师', 0, 0, '', 1),
		(189, 173, '会计从业资格', 0, 0, '', 1),
		(190, 173, '银行从业资格', 0, 0, '', 1),
		(191, 173, '初级会计职称', 0, 0, '', 1),
		(192, 173, '中级会计职称', 0, 0, '', 1),
		(193, 173, '高级会计职称', 0, 0, '', 1),
		(194, 173, '统计师', 0, 0, '', 1),
		(195, 173, '资产评估师', 0, 0, '', 1),
		(196, 173, 'ACCA/CAT', 0, 0, '', 1),
		(197, 173, '精算师', 0, 0, '', 1),
		(198, 173, '基金从业', 0, 0, '', 1),
		(199, 173, '期货从业资格', 0, 0, '', 1),
		(200, 173, '内部审计师', 0, 0, '', 1),
		(201, 173, '中级审计师', 0, 0, '', 1),
		(202, 173, '助理理财规划师', 0, 0, '', 1),
		(203, 173, '理财规划师', 0, 0, '', 1),
		(204, 174, '国家公务员', 0, 0, '', 1),
		(205, 174, '地方公务员', 0, 0, '', 1),
		(206, 174, '政法干警', 0, 0, '', 1),
		(207, 174, '事业单位', 0, 0, '', 1),
		(208, 174, '公选', 0, 0, '', 1),
		(209, 174, '招警', 0, 0, '', 1),
		(210, 174, '信用社', 0, 0, '', 1),
		(211, 174, '三支一扶', 0, 0, '', 1),
		(212, 174, '军转干', 0, 0, '', 1),
		(213, 174, '村官', 0, 0, '', 1),
		(214, 175, '中考', 0, 0, '', 1),
		(215, 175, '小升初', 0, 0, '', 1),
		(216, 175, '考研', 0, 0, '', 1),
		(217, 175, '高考', 0, 0, '', 1),
		(218, 175, '会计硕士', 0, 0, '', 1),
		(219, 175, '法律硕士', 0, 0, '', 1),
		(220, 176, '一级建造师', 0, 0, '', 1),
		(221, 176, '二级建造师', 0, 0, '', 1),
		(222, 176, '造价工程师', 0, 0, '', 1),
		(223, 176, '公路造价工程师', 0, 0, '', 1),
		(224, 176, '监理工程师', 0, 0, '', 1),
		(225, 176, '质量工程师', 0, 0, '', 1),
		(226, 176, '房地产估价师', 0, 0, '', 1),
		(227, 176, '房地产经纪人', 0, 0, '', 1),
		(228, 176, '计量师 造价员', 0, 0, '', 1),
		(229, 176, '安全评价师', 0, 0, '', 1),
		(230, 176, '资产评估师', 0, 0, '', 1),
		(231, 176, '咨询工程师', 0, 0, '', 1),
		(232, 176, '房地产评估师', 0, 0, '', 1),
		(234, 176, '土地代理人', 0, 0, '', 1),
		(235, 176, '给排水工程师', 0, 0, '', 1),
		(236, 176, '一级建筑师', 0, 0, '', 1),
		(237, 176, '二级建筑师', 0, 0, '', 1),
		(238, 176, '化工工程师', 0, 0, '', 1),
		(239, 176, '暖通工程师', 0, 0, '', 1),
		(240, 176, '结构工程师', 0, 0, '', 1),
		(241, 176, '安全工程师', 0, 0, '', 1),
		(242, 176, '招标师', 0, 0, '', 1),
		(243, 176, '测绘工程师', 0, 0, '', 1),
		(244, 176, '城市规划师', 0, 0, '', 1),
		(245, 176, '岩土工程师', 0, 0, '', 1),
		(246, 176, '电气工程师', 0, 0, '', 1),
		(247, 176, '土地估价师', 0, 0, '', 1),
		(248, 176, '设备监理师', 0, 0, '', 1),
		(249, 176, '物业管理师', 0, 0, '', 1),
		(250, 176, '通信工程师', 0, 0, '', 1),
		(251, 176, '环境影响评价师', 0, 0, '', 1),
		(252, 177, '澳洲留学', 0, 0, '', 1),
		(253, 177, '英国留学', 0, 0, '', 1),
		(254, 177, '雅思', 0, 0, '', 1),
		(255, 177, '托福', 0, 0, '', 1),
		(256, 177, 'GRE', 0, 0, '', 1),
		(257, 177, '出国留学', 0, 0, '', 1),
		(258, 177, '英语四级', 0, 0, '', 1),
		(259, 177, '英语六级', 0, 0, '', 1),
		(260, 177, 'BEC', 0, 0, '', 1),
		(261, 177, 'GMAT', 0, 0, '', 1),
		(262, 177, '自考英语', 0, 0, '', 1),
		(263, 177, '职称英语', 0, 0, '', 1),
		(264, 177, '公共英语', 0, 0, '', 1),
		(265, 177, '职称日语', 0, 0, '', 1),
		(266, 177, '口译笔译', 0, 0, '', 1),
		(267, 177, '英语三级', 0, 0, '', 1),
		(268, 177, '专四专八', 0, 0, '', 1),
		(269, 177, 'ACT', 0, 0, '', 1),
		(270, 177, 'SAT', 0, 0, '', 1),
		(271, 178, '国家司法', 0, 0, '', 1),
		(272, 178, '幼儿教师资格证', 0, 0, '', 1),
		(273, 178, '小学教师资格证', 0, 0, '', 1),
		(274, 178, '中学教师资格证', 0, 0, '', 1),
		(275, 178, '人力资源管理师三级', 0, 0, '', 1),
		(276, 178, '企业法律顾问', 0, 0, '', 1),
		(277, 178, '管理咨询师', 0, 0, '', 1),
		(278, 178, '项目管理师', 0, 0, '', 1),
		(279, 178, '企业培训师', 0, 0, '', 1),
		(280, 178, '社会工作者', 0, 0, '', 1),
		(281, 178, '出版资格', 0, 0, '', 1),
		(282, 178, '广告师', 0, 0, '', 1),
		(283, 178, '公共营养师', 0, 0, '', 1),
		(284, 178, '心理咨询师', 0, 0, '', 1),
		(285, 178, '驾照考试', 0, 0, '', 1),
		(286, 179, '国际商务师', 0, 0, '', 1),
		(287, 179, '外销员', 0, 0, '', 1),
		(288, 179, '单证员', 0, 0, '', 1),
		(289, 179, '货运代理', 0, 0, '', 1),
		(290, 179, '物流师', 0, 0, '', 1),
		(291, 179, '报关员', 0, 0, '', 1),
		(292, 179, '跟单员', 0, 0, '', 1),
		(293, 180, '执业中药师', 0, 0, '', 1),
		(294, 180, '执业西药师', 0, 0, '', 1),
		(295, 180, '公卫执业医师', 0, 0, '', 1),
		(296, 180, '公卫执业助理', 0, 0, '', 1),
		(297, 180, '药学职称', 0, 0, '', 1),
		(298, 180, '中药学职称', 0, 0, '', 1),
		(299, 180, '临床执业医师', 0, 0, '', 1),
		(300, 180, '临床助理医师', 0, 0, '', 1),
		(301, 180, '中医执业医师', 0, 0, '', 1),
		(302, 180, '中医助理医师', 0, 0, '', 1),
		(303, 180, '中西医执业医师', 0, 0, '', 1),
		(304, 180, '中西医助理医师', 0, 0, '', 1),
		(305, 180, '口腔执业医师', 0, 0, '', 1),
		(306, 180, '口腔助理医师', 0, 0, '', 1),
		(307, 180, '护士资格', 0, 0, '', 1),
		(308, 180, '内科主治医师', 0, 0, '', 1),
		(309, 180, '外科主治医师', 0, 0, '', 1),
		(310, 180, '妇产科主治医师', 0, 0, '', 1),
		(311, 180, '医学检验', 0, 0, '', 1),
		(312, 181, '职称计算机', 0, 0, '', 1),
		(313, 5, '星座运势', 0, 0, '', 1),
		(314, 313, '手相面相', 0, 0, '', 1),
		(315, 313, '占卜算命', 0, 0, '', 1),
		(316, 313, '星座运势', 0, 0, '', 1),
		(317, 313, '风水学', 0, 0, '', 1),
		(318, 5, '兴趣爱好', 0, 0, '', 1),
		(319, 318, '体育/运动', 0, 0, '', 1),
		(320, 318, '音乐', 0, 0, '', 1),
		(321, 318, '旅游购物', 0, 0, '', 1),
		(322, 318, '美容化妆', 0, 0, '', 1),
		(323, 318, '影视/动漫', 0, 0, '', 1),
		(324, 318, '保健养生', 0, 0, '', 1),
		(325, 318, '随笔', 0, 0, '', 1),
		(326, 318, '摄影摄像', 0, 0, '', 1),
		(327, 318, '幽默滑稽', 0, 0, '', 1),
		(328, 5, '娱乐八卦', 0, 0, '', 1),
		(329, 328, '明星', 0, 0, '', 1),
		(330, 328, '花边', 0, 0, '', 1),
		(331, 328, '资讯', 0, 0, '', 1),
		(332, 5, '其它', 0, 0, '', 1),
		(333, 332, '其它', 0, 0, '', 1),
		(334, 10, '高考', 0, 0, '', 1);
`
	o.Raw(sql).Exec()

}

//初始化配置项
func installCfg() {
	var configs []Config

	//邮箱
	cateEmail := string(ConfigCateEmail)
	cfgEmail := []Config{
		Config{
			Title:       "主机",
			Description: "请填写邮箱HOST，当前仅支持SMTP。示例：smtpdm.aliyun.com",
			Key:         "host",
			Value:       "",
			Category:    cateEmail,
		},
		Config{
			Title:       "端口",
			Description: "邮箱服务端口",
			Key:         "port",
			Value:       "",
			Category:    cateEmail,
		},
		Config{
			Title:       "用户名",
			Description: "邮箱用户名",
			Key:         "username",
			Value:       "",
			Category:    cateEmail,
		},
		Config{
			Title:       "密码",
			Description: "邮箱密码",
			Key:         "password",
			Value:       "",
			Category:    cateEmail,
		},
		Config{
			Title:       "回件收件邮箱",
			Description: "用于接收邮件回件的邮箱。留空则表示使用发件邮箱作为收件邮箱",
			Key:         "replyto",
			Value:       "",
			Category:    cateEmail,
		},
		Config{
			Title:       "测试邮箱地址",
			Description: "在测试邮箱配置是否成功的时候，接收测试邮件的邮箱地址",
			Key:         "test",
			Value:       "",
			Category:    cateEmail,
		},
	}

	//日志
	cateLogs := string(ConfigCateLog)
	cfgLogs := []Config{
		Config{
			Title:       "保留时长(天)",
			Description: "日志保留时长，至少一天",
			Key:         "max_days",
			Value:       "7",
			InputType:   InputNumber,
			Category:    cateLogs,
		},
		Config{
			Title:       "日志文件最大行数",
			Description: "日志文件最大行数，默认为10000行，用于拆分较大日志文件",
			Key:         "max_lines",
			Value:       "10000",
			Category:    cateLogs,
		},
	}

	//依赖
	cateDepend := string(ConfigCateDepend)
	cfgDepend := []Config{
		Config{
			Title:       "PDF2SVG",
			Description: "PDF转SVG命令工具，默认为pdf2svg",
			Key:         "pdf2svg",
			Value:       "pdf2svg",
			Category:    cateDepend,
		},
		Config{
			Title:       "Soffice",
			Description: "libreoffice/openoffice将office文档转PDF文档的工具，默认为soffice",
			Key:         "soffice",
			Value:       "soffice",
			Category:    cateDepend,
		},
		Config{
			Title:       "Soffice转化超时时间(秒)",
			Description: "转换office文档的超时时间，避免转化失败还占用服务器资源，默认1800秒",
			Key:         "soffice-expire",
			Value:       "1800",
			InputType:   InputNumber,
			Category:    cateDepend,
		},
		Config{
			Title:       "Calibre",
			Description: "calibre文档转换命令，将mobi等转PDF，默认为ebook-convert",
			Key:         "calibre",
			Value:       "ebook-convert",
			Category:    cateDepend,
		},
		Config{
			Title:       "PDF2TEXT",
			Description: "从pdf中提取文本的工具，默认为pdftotext",
			Key:         "pdftotext",
			Value:       "pdftotext",
			Category:    cateDepend,
		},
		Config{
			Title:       "ImageMagick",
			Description: "图片转换工具命令，用于将svg转png，默认为convert",
			Key:         "imagemagick",
			Value:       "convert",
			Category:    cateDepend,
		},
		Config{
			Title:       "SVGO",
			Description: "node模块，svg压缩工具，清除svg多余字符",
			Key:         "svgo",
			Value:       "svgo",
			Category:    cateDepend,
		},
		//Config{
		//	Title:       "启用SVGO",
		//	Description: "是否启用svgo，默认为false",
		//	Key:         "svgo-on",
		//	Value:       "false",
		//	InputType:   InputBool,
		//	Category:    cateDepend,
		//},
	}

	//全文搜索
	cateES := string(ConfigCateElasticSearch)
	cfgES := []Config{
		Config{
			Title:       "是否开启",
			Description: "是否开启ElasticSearch作为全文搜索引擎",
			Key:         "on",
			Value:       "false",
			InputType:   InputBool,
			Category:    cateES,
		},
		Config{
			Title:       "索引名称",
			Description: "请输入索引名称，默认为dochub",
			Key:         "index",
			Value:       "dochub",
			Category:    cateES,
		},
		Config{
			Title:       "服务地址",
			Description: "ElasticSearch Host，如：http://localhost:9200，带http",
			Key:         "host",
			Value:       "",
			Category:    cateES,
		},
	}

	// 阿里云
	cateOss := string(StoreOss)
	cfgOss := []Config{
		Config{
			Title:       "AccessKey",
			Description: "阿里云 AccessKey",
			Key:         "access-key",
			Value:       "",
			InputType:   InputText,
			Category:    cateOss,
		},
		Config{
			Title:       "SecretKey",
			Description: "阿里云 SecretKey",
			Key:         "secret-key",
			Value:       "",
			InputType:   InputText,
			Category:    cateOss,
		},
		Config{
			Title:       "Endpoint",
			Description: "阿里云 OSS endpoint，如果与服务器同属于同一内网，建议填内网 endpoint",
			Key:         "endpoint",
			Value:       "",
			InputType:   InputText,
			Category:    cateOss,
		},
		Config{
			Title:       "公共读Bucket",
			Description: "阿里云 OSS 具有公共读权限的 Bucket",
			Key:         "public-bucket",
			Value:       "",
			InputType:   InputText,
			Category:    cateOss,
		},
		Config{
			Title:       "公共读Bucket域名",
			Description: "阿里云 OSS 具有公共读权限的 Bucket 所绑定的域名，带 http:// 或者 https://",
			Key:         "public-bucket-domain",
			Value:       "",
			InputType:   InputText,
			Category:    cateOss,
		},

		Config{
			Title:       "私有Bucket",
			Description: "阿里云 OSS 创建的私有 Bucket",
			Key:         "private-bucket",
			Value:       "",
			InputType:   InputText,
			Category:    cateOss,
		},
		Config{
			Title:       "私有Bucket域名",
			Description: "阿里云 OSS 创建的私有 Bucket 所绑定的域名，带 http:// 或者 https://",
			Key:         "private-bucket-domain",
			Value:       "",
			InputType:   InputText,
			Category:    cateOss,
		},
		Config{
			Title:       "过期时间",
			Description: "文档下载签名链接有效时长(秒)",
			Key:         "expire",
			Value:       "3600",
			InputType:   InputNumber,
			Category:    cateOss,
		},
	}

	// 百度云
	cateBos := string(StoreBos)
	cfgBos := []Config{
		Config{
			Title:       "AccessKey",
			Description: "百度云 AccessKey",
			Key:         "access-key",
			Value:       "",
			InputType:   InputText,
			Category:    cateBos,
		},
		Config{
			Title:       "SecretKey",
			Description: "百度云 SecretKey",
			Key:         "secret-key",
			Value:       "",
			InputType:   InputText,
			Category:    cateBos,
		},
		Config{
			Title:       "Endpoint",
			Description: "百度云 BOS endpoint",
			Key:         "endpoint",
			Value:       "",
			InputType:   InputText,
			Category:    cateBos,
		},
		Config{
			Title:       "公共读Bucket",
			Description: "百度云 BOS 具有公共读权限的 Bucket",
			Key:         "public-bucket",
			Value:       "",
			InputType:   InputText,
			Category:    cateBos,
		},
		Config{
			Title:       "公共读Bucket域名",
			Description: "百度云 BOS 具有公共读权限的 Bucket 所绑定的域名，带 http:// 或者 https://",
			Key:         "public-bucket-domain",
			Value:       "",
			InputType:   InputText,
			Category:    cateBos,
		},

		Config{
			Title:       "私有Bucket",
			Description: "百度云 BOS 创建的私有 Bucket",
			Key:         "private-bucket",
			Value:       "",
			InputType:   InputText,
			Category:    cateBos,
		},
		Config{
			Title:       "私有Bucket域名",
			Description: "百度云 BOS 创建的私有 Bucket 所绑定的域名，带 http:// 或者 https://",
			Key:         "private-bucket-domain",
			Value:       "",
			InputType:   InputText,
			Category:    cateBos,
		},
		Config{
			Title:       "过期时间",
			Description: "文档下载签名链接有效时长(秒)",
			Key:         "expire",
			Value:       "3600",
			InputType:   InputNumber,
			Category:    cateBos,
		},
	}

	// 腾讯云
	cateCos := string(StoreCos)
	cfgCos := []Config{
		Config{
			Title:       "AccessKey",
			Description: "AccessKey",
			Key:         "access-key",
			Value:       "",
			InputType:   InputText,
			Category:    cateCos,
		},
		Config{
			Title:       "SecretKey",
			Description: "SecretKey",
			Key:         "secret-key",
			Value:       "",
			InputType:   InputText,
			Category:    cateCos,
		},
		Config{
			Title:       "AppID",
			Description: "腾讯云 COS AppID",
			Key:         "app-id",
			Value:       "",
			InputType:   InputText,
			Category:    cateCos,
		},
		Config{
			Title:       "区域",
			Description: "COS 区域，即 Region",
			Key:         "region",
			Value:       "",
			InputType:   InputText,
			Category:    cateCos,
		},
		Config{
			Title:       "公共读Bucket",
			Description: "具有公共读权限的 Bucket",
			Key:         "public-bucket",
			Value:       "",
			InputType:   InputText,
			Category:    cateCos,
		},
		Config{
			Title:       "公共读Bucket域名",
			Description: "具有公共读权限的 Bucket 所绑定的域名，带 http:// 或者 https://",
			Key:         "public-bucket-domain",
			Value:       "",
			InputType:   InputText,
			Category:    cateCos,
		},

		Config{
			Title:       "私有Bucket",
			Description: "私有 Bucket",
			Key:         "private-bucket",
			Value:       "",
			InputType:   InputText,
			Category:    cateCos,
		},
		Config{
			Title:       "私有Bucket域名",
			Description: "私有 Bucket 所绑定的域名，带 http:// 或者 https://",
			Key:         "private-bucket-domain",
			Value:       "",
			InputType:   InputText,
			Category:    cateCos,
		},
		Config{
			Title:       "过期时间",
			Description: "文档下载签名链接有效时长(秒)",
			Key:         "expire",
			Value:       "3600",
			InputType:   InputNumber,
			Category:    cateCos,
		},
	}

	// 华为云
	cateObs := string(StoreObs)
	cfgObs := []Config{
		Config{
			Title:       "AccessKey",
			Description: "AccessKey",
			Key:         "access-key",
			Value:       "",
			InputType:   InputText,
			Category:    cateObs,
		},
		Config{
			Title:       "SecretKey",
			Description: "SecretKey",
			Key:         "secret-key",
			Value:       "",
			InputType:   InputText,
			Category:    cateObs,
		},
		Config{
			Title:       "Endpoint",
			Description: "endpoint",
			Key:         "endpoint",
			Value:       "",
			InputType:   InputText,
			Category:    cateObs,
		},
		Config{
			Title:       "公共读Bucket",
			Description: "具有公共读权限的 Bucket",
			Key:         "public-bucket",
			Value:       "",
			InputType:   InputText,
			Category:    cateObs,
		},
		Config{
			Title:       "公共读Bucket域名",
			Description: "具有公共读权限的 Bucket 所绑定的域名，带 http:// 或者 https://",
			Key:         "public-bucket-domain",
			Value:       "",
			InputType:   InputText,
			Category:    cateObs,
		},

		Config{
			Title:       "私有Bucket",
			Description: "私有 Bucket",
			Key:         "private-bucket",
			Value:       "",
			InputType:   InputText,
			Category:    cateObs,
		},
		Config{
			Title:       "私有Bucket域名",
			Description: "私有 Bucket 所绑定的域名，带 http:// 或者 https://",
			Key:         "private-bucket-domain",
			Value:       "",
			InputType:   InputText,
			Category:    cateObs,
		},
		Config{
			Title:       "过期时间",
			Description: "文档下载签名链接有效时长(秒)",
			Key:         "expire",
			Value:       "3600",
			InputType:   InputNumber,
			Category:    cateObs,
		},
	}

	// Minio
	cateMinio := string(StoreMinio)
	cfgMinio := []Config{
		Config{
			Title:       "AccessKey",
			Description: "Minio AccessKey",
			Key:         "access-key",
			Value:       "",
			InputType:   InputText,
			Category:    cateMinio,
		},
		Config{
			Title:       "SecretKey",
			Description: "Minio SecretKey",
			Key:         "secret-key",
			Value:       "",
			InputType:   InputText,
			Category:    cateMinio,
		},
		Config{
			Title:       "Endpoint",
			Description: "Minio endpoint",
			Key:         "endpoint",
			Value:       "",
			InputType:   InputText,
			Category:    cateMinio,
		},
		Config{
			Title:       "公共读Bucket",
			Description: "具有公共读权限的 Bucket",
			Key:         "public-bucket",
			Value:       "",
			InputType:   InputText,
			Category:    cateMinio,
		},
		Config{
			Title:       "公共读Bucket域名",
			Description: "具有公共读权限的 Bucket 所绑定的域名，带 http:// 或者 https://",
			Key:         "public-bucket-domain",
			Value:       "",
			InputType:   InputText,
			Category:    cateMinio,
		},

		Config{
			Title:       "私有Bucket",
			Description: "私有Bucket",
			Key:         "private-bucket",
			Value:       "",
			InputType:   InputText,
			Category:    cateMinio,
		},
		Config{
			Title:       "私有Bucket域名",
			Key:         "private-bucket-domain",
			Description: "私有 Bucket 所绑定的域名，带 http:// 或者 https://",
			Value:       "",
			InputType:   InputText,
			Category:    cateMinio,
		},
		Config{
			Title:       "过期时间",
			Description: "文档下载签名链接有效时长(秒)",
			Key:         "expire",
			Value:       "3600",
			InputType:   InputNumber,
			Category:    cateMinio,
		},
	}

	// Qiniu
	cateQiniu := string(StoreQiniu)
	cfgQiniu := []Config{
		Config{
			Title:       "AccessKey",
			Description: "AccessKey",
			Key:         "access-key",
			Value:       "",
			InputType:   InputText,
			Category:    cateQiniu,
		},
		Config{
			Title:       "SecretKey",
			Description: "SecretKey",
			Key:         "secret-key",
			Value:       "",
			InputType:   InputText,
			Category:    cateQiniu,
		},
		Config{
			Title:       "公共读Bucket",
			Description: "具有公共读权限的 Bucket",
			Key:         "public-bucket",
			Value:       "",
			InputType:   InputText,
			Category:    cateQiniu,
		},
		Config{
			Title:       "公共读Bucket域名",
			Description: "具有公共读权限的 Bucket 所绑定的域名，带 http:// 或者 https://",
			Key:         "public-bucket-domain",
			Value:       "",
			InputType:   InputText,
			Category:    cateQiniu,
		},

		Config{
			Title:       "私有Bucket",
			Description: "私有Bucket",
			Key:         "private-bucket",
			Value:       "",
			InputType:   InputText,
			Category:    cateQiniu,
		},
		Config{
			Title:       "私有Bucket域名",
			Description: "私有 Bucket 所绑定的域名，带 http:// 或者 https://",
			Key:         "private-bucket-domain",
			Value:       "",
			InputType:   InputText,
			Category:    cateQiniu,
		},
		Config{
			Title:       "过期时间",
			Description: "文档下载签名链接有效时长(秒)",
			Key:         "expire",
			Value:       "3600",
			InputType:   InputNumber,
			Category:    cateQiniu,
		},
	}

	// 又拍云
	cateUpyun := string(StoreUpyun)
	cfgUpyun := []Config{
		Config{
			Title:       "Operator",
			Description: "又拍云操作员",
			Key:         "operator",
			Value:       "",
			InputType:   InputText,
			Category:    cateUpyun,
		},
		Config{
			Title:       "Password",
			Description: "又拍云操作员密码",
			Key:         "password",
			Value:       "",
			InputType:   InputText,
			Category:    cateUpyun,
		},
		Config{
			Title:       "公共读Bucket",
			Description: "具有公共读权限的 Bucket",
			Key:         "public-bucket",
			Value:       "",
			InputType:   InputText,
			Category:    cateUpyun,
		},
		Config{
			Title:       "公共读Bucket域名",
			Description: "具有公共读权限的 Bucket 所绑定的域名，带 http:// 或者 https://",
			Key:         "public-bucket-domain",
			Value:       "",
			InputType:   InputText,
			Category:    cateUpyun,
		},

		Config{
			Title:       "私有Bucket",
			Description: "私有Bucket，需要URL签名才能访问",
			Key:         "private-bucket",
			Value:       "",
			InputType:   InputText,
			Category:    cateUpyun,
		},
		Config{
			Title:       "私有Bucket Secret",
			Description: "即 访问控制 的 Token 防盗链密钥",
			Key:         "secret",
			Value:       "",
			InputType:   InputText,
			Category:    cateUpyun,
		},
		Config{
			Title:       "私有Bucket域名",
			Description: "私有 Bucket 所绑定的域名，带 http:// 或者 https://",
			Key:         "private-bucket-domain",
			Value:       "",
			InputType:   InputText,
			Category:    cateUpyun,
		},
		Config{
			Title:       "过期时间",
			Description: "文档下载签名链接有效时长(秒)",
			Key:         "expire",
			Value:       "3600",
			InputType:   InputNumber,
			Category:    cateUpyun,
		},
	}

	configs = append(configs, cfgEmail...)
	configs = append(configs, cfgLogs...)
	configs = append(configs, cfgDepend...)
	configs = append(configs, cfgES...)
	configs = append(configs, cfgOss...)
	configs = append(configs, cfgBos...)
	configs = append(configs, cfgCos...)
	configs = append(configs, cfgObs...)
	configs = append(configs, cfgMinio...)
	configs = append(configs, cfgQiniu...)
	configs = append(configs, cfgUpyun...)

	o := orm.NewOrm()
	for _, cfg := range configs {
		// 逐条写入数据库
		if helper.Debug {
			beego.Info("==如果因数据已存在而导致数据写入失败，则请忽略==")
		}
		o.Insert(&cfg)

	}
	o.QueryTable(NewConfig()).Filter("Category", "depend").Filter("Key__in", "svgo-on").Delete()
}
