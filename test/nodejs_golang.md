性能
顺序编程模型更自然
部署简单
语法简单，学习成本低
类型系统，编译期发现问题，重构时更容易发现所修改代码的相关代码
适用场景更广，不限于http restful/api，tcp强交互都适用
独角兽公司、大型商业项目背书：docker，uber，七牛，猎豹移动，滴滴，B站，360推送系统单机100w在线，
今日头条，阿里中间件，百度运维，巨人网络，神仙道，国外直播系统单机50w在线，等等


https://www.zhihu.com/question/21409296
另外，云风博客中曾说过这样一句话：
我发现我花了四年时间锤炼自己用 C 语言构建系统的能力，试图找到一个规范，可以更好的编写软件。
结果发现只是对 Go 的模仿。缺乏语言层面的支持，只能是一个拙劣的模仿。



3、Go成功的项目
nsq：bitly开源的消息队列系统，性能非常高，目前他们每天处理数十亿条的消息
docker:基于lxc的一个虚拟打包工具，能够实现PAAS平台的组建。
packer:用来生成不同平台的镜像文件，例如VM、vbox、AWS等，作者是vagrant的作者
skynet：分布式调度框架
Doozer：分布式同步工具，类似ZooKeeper
Heka：mazila开源的日志处理系统
cbfs：couchbase开源的分布式文件系统
tsuru：开源的PAAS平台，和SAE实现的功能一模一样
groupcache：memcahe作者写的用于Google下载系统的缓存系统
god：类似redis的缓存系统，但是支持分布式和扩展性
gor：网络流量抓包和重放工具以下是一些公司，只是一小部分：




我们为什么会选择 Golang
https://ruby-china.org/topics/26838

从 Node.js 到 Golang 的迁徙之路
https://juejin.im/entry/584780928e450a006c1b801c



使用Go语言代替node.js实践
http://www.jiangmiao.org/blog/2932.html

Bowery为什么放弃Node.js，转向Go？
https://studygolang.com/articles/2326


如何看待 TJ 宣布退出 Node.js 开发，转向 Go？
https://www.zhihu.com/question/24373004
TJ何许人也？他medium自我介绍：TJ Holowaychuk，程序员兼艺术家，
Koa、Co、Express、jade、mocha、node-canvas、commander.js等知名开源项目的创建和贡献者。
社区影响：https://nodejsmodules.org 第一页出现次数最多的那个少年

express作者TJ Holowaychuk放弃Node.js投奔Go怀抱- NinJa911's Blog
http://blog.ninja911.com/blog-show-blog_id-81.html
